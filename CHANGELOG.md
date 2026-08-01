# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-08-01

### Breaking
- The published container image runs as uid **10001** (named user `licenses`),
  not `65532`. `Dockerfile.goreleaser` moves from `gcr.io/distroless/static:nonroot`
  to `alpine:latest`, matching the local `./Dockerfile` and the rest of the
  exporter family. Anyone pinning the container uid — a `securityContext.runAsUser`,
  ownership on a mounted secret or log volume — must update it. See ADR-0002.

### Security
- Bump `licenses-exporter-core` to v1.1.1, carrying the upstream `grpc`
  v1.82.0 → v1.83.0 fix for **GO-2026-6061**, a reachable vulnerability.
  Documentation-only release vs v1.1.0; no code changes in this repo.

### Added
- **`/livez` and `/readyz`**, both always 200 and reading no state, via
  `licenses-exporter-core` v1.1.1. Point Kubernetes probes and container
  healthchecks at these, never at `/metrics`.
- **`HEALTHCHECK`** against `http://127.0.0.1:9107/livez` in both `Dockerfile` and
  `Dockerfile.goreleaser`, and a matching `healthcheck:` in `docker-compose.yml`
  and `docker-compose.ghcr.yml`. The ghcr healthcheck needs an image from this
  release or later — earlier published images are distroless and carry no `wget`.

### Changed
- **`/health` now always returns 200**, with `starting`/`ok` as the body rather
  than the status code. It previously returned 503 until the first collection
  cycle completed, which restarted healthy pods under a `livenessProbe` and
  reported containers unhealthy for the whole start-up window. Anything asserting
  a 503 from `/health` must be updated.
- **`/metrics` now accepts `name[]` query parameters** for metric filtering, via
  the `prometheus/client_golang` v1.23.2 → v1.24.1 bump pulled in by
  `licenses-exporter-core` v1.1.1. Requests with no `name[]` parameter are
  unaffected. Note for operators: `client_golang` 1.24 also always enforces
  UTF-8 metric-name validation rather than the legacy scheme; neither this repo
  nor `licenses-exporter-core` sets `model.NameValidationScheme`, so there is no
  observable behavior change here.

## [0.1.2] - 2026-07-10

### Security
- Bump Go toolchain to `go 1.26.5` to patch **GO-2026-5856** (`crypto/tls`), reported by
  `govulncheck` / `make ci`.

### Added
- First multi-arch (`linux/amd64`, `linux/arm64`) GHCR container image publishing via
  GoReleaser `dockers_v2` — `ghcr.io/fjacquet/veeam_licenses_exporter` (tags `{version}`,
  `{major}.{minor}`, and `latest` on non-prerelease), with per-image CycloneDX SBOM.

## [0.1.0] - 2026-07-03

### Added
- Initial release: a Veeam license exporter reading the Veeam Backup Enterprise Manager
  REST API (`:9398`, session auth → `GET /api/licensing`) via a hand-rolled resty client,
  built on `github.com/fjacquet/licenses-exporter-core`. Emits the shared `license_` schema
  (`vendor="veeam"`, `unit="instances"`). Default metrics port `9107`. Requires Enterprise
  Manager. See ADR-0001. Released as **v0.1.0**: the EM `/api/licensing` field mapping is
  isolated (`internal/veeam/model.go`) and pending verification against a live Enterprise
  Manager; the parser is tolerant (absent-not-zero) until then.
