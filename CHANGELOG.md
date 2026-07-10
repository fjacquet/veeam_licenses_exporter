# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
