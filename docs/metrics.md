# Metrics reference

`veeam_licenses_exporter` exposes one generic `license_` metric family, shared across the
`licenses_exporter` family via `github.com/fjacquet/licenses-exporter-core`. Vendors are
distinguished by **labels**, not by metric name — this exporter emits `vendor="veeam"`.
Every value is a raw fact straight from the Veeam Backup Enterprise Manager `/api/licensing`
resource: there is no exporter-computed compliance verdict or "days remaining" gauge. Derive
those in PromQL or alert rules from the raw facts below.

!!! warning "v0.1.0 — field mapping pending verification"
    The EM `/api/licensing` JSON field mapping (`internal/veeam/model.go`) is based on the
    documented Veeam licensing terms but has **not yet been verified against a live
    Enterprise Manager**. The parser is tolerant (absent-not-zero), so an unexpected field
    name yields an absent sample rather than a wrong one — but treat the field names below as
    provisional until confirmed against a real EM response. First release is tagged
    **v0.1.0** for this reason.

This table is the diff target for `--once --debug`, which dumps every collected sample
(sorted, exposition style) for live payload validation against a real Enterprise Manager.

## Example series

```text
license_seats_total{vendor="veeam",product="Enterprise Plus",unit="instances",instance="em-a"} 100
license_seats_used{vendor="veeam",product="Enterprise Plus",unit="instances",instance="em-a"} 42
license_expiration_timestamp_seconds{vendor="veeam",product="Enterprise Plus",instance="em-a"} 1.8015264e+09
```

## License facts

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `license_seats_total` | Gauge | `vendor, product, unit, instance` | Total licensed instance capacity (`LicensedInstancesNumber`). **Omitted** when the value is unlimited (`LicensedInstancesNumber <= 0`) — never a `0` or `9999` sentinel. |
| `license_seats_used` | Gauge | `vendor, product, unit, instance` | Currently consumed capacity (`UsedInstancesNumber`). Always emitted when known — an unlimited license still reports its used count. |
| `license_expiration_timestamp_seconds` | Gauge | `vendor, product, instance` | License expiration as a Unix timestamp, parsed from `ExpirationDate` (RFC3339). **Omitted entirely** when the license is perpetual (no `ExpirationDate`, or unparseable). |

## Health / state

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `license_up` | Gauge | `vendor, instance` | `1` if the Enterprise Manager's last collection cycle (session login → `GET /api/licensing` → logout) succeeded, `0` if it failed. Absent entirely until that EM's first cycle resolves. |
| `license_collector_last_success_timestamp_seconds` | Gauge | `vendor, instance` | Unix timestamp of the last successful collection for this EM. `time() - this` is the data-age/freshness signal. |
| `license_scrape_duration_seconds` | Gauge | `vendor, instance` | Wall-clock time spent collecting this EM during the last cycle. |
| `license_build_info` | Gauge | `version, goversion` | Constant `1`; carries the exporter's build metadata. The only series present before the first collection cycle completes. |

## Label semantics

| Label | Meaning / source |
|---|---|
| `vendor` | `"veeam"`. |
| `product` | The license's `Edition` (e.g. `Enterprise Plus`). Raw vendor identifier — no friendly-name mapping. Falls back to `"veeam"` if `Edition` is empty. |
| `unit` | Always `"instances"` — Veeam licenses count backed-up/protected instances, not CPUs or sockets. |
| `instance` | The configured Enterprise Manager's `instance` name from `config.yaml` (e.g. `em-a`). One process can poll many Enterprise Managers. |

## Design rules (raw facts, absent-not-zero)

- **Unlimited licenses omit `seats_total`.** A license with `LicensedInstancesNumber <= 0`
  emits only `license_seats_used` — never a `0`, and never a large sentinel standing in for
  "unlimited".
- **Perpetual licenses omit expiration.** A license with no (or unparseable) `ExpirationDate`
  never emits `license_expiration_timestamp_seconds` — there is no `9999`-year row to filter
  out in dashboards or alerts.
- **No `days_to_expiration` gauge, no perpetual sentinel.** `license_expiration_timestamp_seconds`
  carries the absolute Unix timestamp; compute days remaining in PromQL:
  `(license_expiration_timestamp_seconds - time()) / 86400`.
- **No exporter-computed `compliance_status`.** Over-allocation is
  `license_seats_used > license_seats_total`; policy belongs in PromQL/alert rules, not the
  exporter.
- **Absent, never zero.** An unparseable or missing capacity/used value yields an *absent*
  sample, never a fake `0` — a false `0` on a capacity metric would silently corrupt
  dashboards and over-allocation alerts.
- **Cold start.** Immediately after startup, before any Enterprise Manager's first collection
  cycle resolves, `/metrics` exposes **only** `license_build_info` — no `license_up` or
  per-target series exist yet, so a scrape during that window can never see a transient `0`
  or a flapping target.
- **Label-key consistency.** Every series of a given metric name carries the same label-key
  set, built from the shared constructors in `licenses-exporter-core` (see
  [ADR-0001](adr/0001-consume-core-resty-em.md)).

## Live validation

```bash
./bin/veeam_licenses_exporter --config config.yaml --once --debug
```

Runs a single collection cycle and prints every collected sample in sorted, Prometheus
exposition-style output — diff it against the tables above to catch a silently-absent
metric that `license_up` alone would not reveal. This is also the primary way to verify the
v0.1.0 field mapping above against a live Enterprise Manager.
