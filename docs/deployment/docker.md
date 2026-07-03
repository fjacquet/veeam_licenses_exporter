# Docker deployment

The image (`Dockerfile`) is a non-root, multi-stage Alpine build: it runs as the
unprivileged `licenses` user (uid `10001`), listens on `9107`, and reads `config.yaml` from
`/etc/veeam_licenses_exporter/config.yaml`.

!!! warning "Requires Veeam Backup Enterprise Manager"
    This exporter reads license data exclusively from the **Veeam Backup Enterprise
    Manager** REST API (`:9398/api`). A standalone Veeam Backup & Replication server has no
    license REST endpoint — Enterprise Manager must be installed and reachable.

## Standalone container

```bash
docker run -d --name veeam_licenses_exporter -p 9107:9107 \
  -e VEEAM_EM_HOST=https://em.example.com:9398 \
  -e VEEAM_USERNAME=svc-ro \
  -e VEEAM_PASSWORD=... \
  -v /path/to/config.yaml:/etc/veeam_licenses_exporter/config.yaml:ro \
  ghcr.io/fjacquet/veeam_licenses_exporter:latest
```

`config.yaml` is the source of truth for which `${ENV}` references are actually consumed
(`${VEEAM_EM_HOST}`, `${VEEAM_USERNAME}`, `${VEEAM_PASSWORD}`, etc.) — every variable the
mounted config references must exist in the container's environment, or the exporter fails
fast at load with `config references unset environment variable "..."`. Secrets can
alternatively be supplied as a file via `passwordFile` in `config.yaml`, mounted as a
read-only volume instead of passed as an env var.

`/metrics` and `/health` are both served on `9107`; `/health` returns HTTP 200 with
`starting` until the first collection cycle completes for every enabled Enterprise Manager,
then `ok`.

## One-command demo stack (Compose)

```bash
docker compose up
```

`docker-compose.yml` builds the exporter from the local `Dockerfile` and brings up:

- **`veeam_licenses_exporter`** (`:9107`) — built locally, config mounted from `./config.yaml`.
- **`prometheus`** (`:9090`) — scrapes the exporter per `prometheus.yml` and loads the
  alert rules in `deploy/prometheus/license.rules.yml`.
- **`grafana`** (`:3000`, `admin`/`admin` by default) — auto-provisioned with the Prometheus
  datasource and the **Enterprise Licenses — Overview** dashboard
  (`grafana/dashboards/licenses-overview.json`); see [Dashboards](../dashboards.md).

The bundled `config.yaml` ships with placeholder `${VEEAM_*}` env references;
`docker-compose.yml` supplies default placeholder values for those so the stack starts
without any `.env` file, purely to demonstrate the wiring end-to-end. Override them (shell
env or a `.env` file next to `docker-compose.yml`) with real Enterprise Manager credentials
to point the demo at a live environment.

To run the **published** image instead of building locally:

```bash
docker compose -f docker-compose.ghcr.yml up -d
```

Pin a version with `VEEAM_LICENSES_EXPORTER_TAG` (defaults to `:latest`):

```bash
VEEAM_LICENSES_EXPORTER_TAG=0.2.1 docker compose -f docker-compose.ghcr.yml up -d
```

## Required permissions before first run

### Enterprise Manager — read-only account

The collector authenticates against Enterprise Manager's session API
(`POST /api/sessionMngr`), calls `GET /api/licensing`, then logs out
(`DELETE /api/logonSessions/current`) — once per collection cycle, with no persisted
session. Enterprise Manager's built-in roles are **Portal Administrator**, **Portal User**,
and **Restore Operator**; grant the service account the least-privileged role that can
authenticate and read the licensing resource in your Enterprise Manager version (verify
against your own EM — Veeam's per-endpoint RBAC granularity is not otherwise documented,
hence the v0.1.0 caveat below). No restore, backup-management, or configuration privileges
are required. Without sufficient read access, `Collect` fails and that Enterprise Manager's
cycle degrades to `license_up{vendor="veeam",...}=0` rather than blocking the whole
exporter.

!!! note "v0.1.0 — field mapping pending verification"
    The `/api/licensing` field mapping (`internal/veeam/model.go`) has not yet been
    verified against a live Enterprise Manager. See [docs/metrics.md](../metrics.md) and
    [ADR-0001](../adr/0001-consume-core-resty-em.md).

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `--config` | `config.yaml` | Path to the config file. |
| `--web.listen-address` | `:9107` | Address the HTTP server (metrics + health) binds to. |
| `--once` | `false` | Run a single collection cycle and exit instead of serving. |
| `--debug` | `false` | Debug-level logging; combined with `--once` it dumps every collected sample (sorted, exposition style) — see `docs/metrics.md`. |
| `--trace` | `false` | Logs repo-owned API responses for live payload validation. resty is a thin repo-owned client, so this only ever logs repo-owned request/response bodies — never SDK-level debug output that could leak the session token. |

Config reload is live: `SIGHUP`, or any write/create to the config file, triggers a
validated hot reload without a restart or any interruption to `/metrics`.
