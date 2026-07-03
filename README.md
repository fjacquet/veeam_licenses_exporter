# veeam_licenses_exporter

[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

A **Veeam Backup Enterprise Manager** license exporter for the Prometheus/Grafana stack,
built on
[`github.com/fjacquet/licenses-exporter-core`](https://github.com/fjacquet/licenses-exporter-core).
It periodically polls the Enterprise Manager `/api/licensing` resource for one or more
Enterprise Managers and normalizes capacity/usage/expiration data into the shared
`license_` Prometheus schema, exposed via **both** a Prometheus `/metrics` endpoint and an
OTLP metric push, fed from a single shared snapshot.

> [!IMPORTANT]
> **Requires Veeam Backup Enterprise Manager (:9398).** This exporter reads license data
> exclusively from the Enterprise Manager REST API (`:9398/api`). A standalone Veeam
> Backup & Replication server (`:9419`) has **no** license REST endpoint — Enterprise
> Manager must be installed and reachable for this exporter to work.

> **v0.1.0** — license field mapping pending verification against a live Enterprise
> Manager. The EM `/api/licensing` JSON field mapping is isolated in
> `internal/veeam/model.go`; the parser is tolerant (absent-not-zero), so an unexpected
> field yields an absent sample rather than a wrong one — but treat the mapping as
> provisional until confirmed against a real EM response (see
> [docs/metrics.md](docs/metrics.md)).

Part of the `licenses_exporter` family; shares the `license_` schema via
`licenses-exporter-core` — see [ADR-0001](docs/adr/0001-consume-core-resty-em.md).

## Metrics

One `license_` prefix shared across the family; vendors are distinguished by labels, not by
metric name:

| Metric | Labels | Notes |
|---|---|---|
| `license_seats_total` | `vendor,product,unit,instance` | Omitted for unlimited licenses (`LicensedInstancesNumber <= 0`) — never a `0`/`9999` sentinel. |
| `license_seats_used` | `vendor,product,unit,instance` | Raw fact, always emitted when known. |
| `license_expiration_timestamp_seconds` | `vendor,product,instance` | Omitted entirely for perpetual licenses. |
| `license_up` | `vendor,instance` | `1`/`0` per source's last collection cycle. |
| `license_collector_last_success_timestamp_seconds` | `vendor,instance` | Unix timestamp of the last successful collection. |
| `license_scrape_duration_seconds` | `vendor,instance` | Time spent collecting that source. |
| `license_build_info` | `version,goversion` | Constant `1`; exporter build metadata. |

No exporter-computed `days_to_expiration` or compliance verdict — derive those in PromQL /
alert rules from the raw facts above. An unparseable value yields an absent sample, never a
fake `0`. At cold start only `license_build_info` is emitted; per-target series appear once
each Enterprise Manager's first collection cycle resolves. See
[docs/metrics.md](docs/metrics.md) for the full reference.

## Quick start

```bash
make cli
./bin/veeam_licenses_exporter --config config.yaml
# metrics: http://localhost:9107/metrics   health: http://localhost:9107/health
```

Useful flags: `--once --debug` runs a single collection cycle and dumps every collected
sample (sorted, exposition style) instead of serving; `--trace` logs repo-owned API response
bodies for live payload validation.

## Configuration

The Veeam collector is toggled in `config.yaml` (`veeam.enabled`), not via environment
variables. Secrets are referenced as `${ENV}` placeholders inside `config.yaml` (or via
`passwordFile` for file-based secrets); a `.env` file is a convenience for local `${ENV}`
expansion, never the source of truth. See `config.yaml` for a full example (one or more
Enterprise Managers):

```yaml
veeam:
  enabled: true
  servers:
    - instance: primary
      host: ${VEEAM_EM_HOST}       # e.g. https://em.example.com:9398
      username: ${VEEAM_USERNAME}
      password: ${VEEAM_PASSWORD}
      insecureSkipVerify: false    # opt-in only, for self-signed EM certs
```

### Enterprise Manager read-only account

The collector authenticates against Enterprise Manager's session API, calls
`GET /api/licensing`, then logs out — stateless, once per collection cycle. Enterprise
Manager's built-in roles are **Portal Administrator**, **Portal User**, and
**Restore Operator**; grant the service account the least-privileged role that can
authenticate and read the licensing resource in your Enterprise Manager version. No
restore, backup-management, or configuration privileges are required.

## Demo stack

```bash
docker compose up
```

Brings up the exporter (`:9107`), Prometheus, and Grafana, auto-provisioned. See
[docs/deployment/docker.md](docs/deployment/docker.md) and
[docs/dashboards.md](docs/dashboards.md).

## Development

```bash
make tools   # install golangci-lint, cyclonedx-gomod, govulncheck (pinned)
make ci      # gofmt check + vet + lint + race tests + govulncheck + build (the CI gate)
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
