# 1. Consume licenses-exporter-core; hand-rolled resty client to Enterprise Manager

Date: 2026-07-02

## Status
Accepted

## Context
This exporter is the Veeam sibling in the licenses_exporter family. The
vendor-neutral engine lives in `github.com/fjacquet/licenses-exporter-core`.
Veeam license data is exposed only via the Veeam Backup Enterprise Manager REST
API (`:9398/api`), not the VBR REST API (`:9419`, which has no license endpoint).
There is no official Veeam Go SDK; the unofficial VeeamHub SDK targets VBR and does
not cover licensing.

## Decision
Depend on `licenses-exporter-core` and build every sample through its constructors.
`main.go` delegates the whole lifecycle to `core.Main`. Read license data from
Enterprise Manager via a hand-rolled `resty/v2` client (session auth → GET
/api/licensing → logout), matching the family's hand-rolled backup exporter
(`nbu_exporter`). No SDK dependency.

## Consequences
- Requires Enterprise Manager to be installed and reachable — documented in the README.
- Schema identity is guaranteed by construction — no local `license_` metric code.
- The EM /api/licensing JSON field mapping is isolated in `internal/veeam/model.go`
  and unverified against a live EM at first release → first tag is v0.1.0 until verified.
- Startup is fatal on an unbuildable-but-valid config (core behaviour); see the core CHANGELOG.
