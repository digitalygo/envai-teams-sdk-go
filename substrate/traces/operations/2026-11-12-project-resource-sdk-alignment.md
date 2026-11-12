---
status: completed
created_at: 2026-11-12
files_edited:
  - README.md
  - client.go
  - client_test.go
  - contract/openapi.json
  - contract/openapi.yaml
  - contract/operations.json
  - contract/provenance.json
  - contract_provenance_test.go
  - docs/api-reference.md
  - docs/contract3-migration.md
  - docs/getting-started.md
  - docs/independent-jobs.md
  - operations_generated.go
  - pagination.go
  - project_log_query_test.go
  - project_operations.go
  - project_operations_test.go
  - project_tenancy_test.go
  - scripts/check_contracts.py
  - scripts/generate.py
  - scripts/sdk-operations.json
  - scripts/test_check_contracts.py
  - strict_validation.go
  - strict_validation_test.go
  - testdata/api-cases.json
  - testdata/operations.json
  - wiki.go
rationale:
  - Expose the backend's existing project-key resource and reporting operations through the Go SDK.
  - Preserve generated API compatibility, precise accounting values and safe pagination.
supporting_docs:
  - ../../../AGENTS.md
  - ../../../docs/api-reference.md
  - ../../../docs/contract3-migration.md
  - ../../../contract/provenance.json
  - https://github.com/digitalygo/1-1-5-envai-aaas/tree/ba021aa61e30e379880f809d45f90e5d3564d43e
---

# Project-resource SDK alignment

## Summary of changes

Aligned this SDK with backend `feature/user-company-projects`, pinned at
`ba021aa61e30e379880f809d45f90e5d3564d43e`. The feature branch has the same name
and was created from clean, synchronized `alpha` at
`3cc7cfc3d54ca8b2b20ad270d0f8677cea8c534c`.

Added 14 existing project-key operations: project skills, manual cron management,
project usage/logs, chat title updates and workspace/chat deletion. The reviewed
inventory is 92 operations, with 89 callable operations and three existing
locally rejected API-key management stubs, against 146 backend operations.

## Technical reasoning

Extended the existing allowlist and generator rather than introducing a second
transport or exposing human-only APIs. Generated artifacts remain reproducible,
and existing public skill request-map types remain unchanged.

- The low-level `GetProjectLogs` retains raw JSON for its response union; typed
  event, analytics and filter-catalog helpers validate the selected view before
  hydration. Usage and event pages preserve metadata and exact decimal strings.
- Added local numeric report iterators using the existing completion pagination
  pattern, with context cancellation, start-page preservation, metadata guards,
  error termination and consumer stop. No iterator follows response URLs.
- Generated repeated-parameter metadata drives repeated bracketed query keys.
  Log validation rejects malformed lists, scalar/plural conflicts and fields
  prohibited for the selected view.
- Only exact empty log-entry `agent` and `data` arrays are normalized to objects
  before strict validation and typed decoding. Canonical schema bytes and
  unrelated payload validation are not weakened.
- Workspace/chat deletion returns HTTP 200 cleanup state, not a purge receipt.
  Cron/skill deletion returns HTTP 204. The SDK never automatically cancels cron
  work or retries a mutation after 409, 402 or 503.
- Existing `APIError` preserves status, code and details. The canonical image
  configuration error is represented without changing the error hierarchy.
- Existing title, cron and billing events remain typed. Billing notices can have
  null turns before admission and are not terminal events. Historical ATP v2
  stays frozen; native task extensions remain opt-in.
- Refreshed OpenAPI and derived artifacts and pinned six source snapshots.
  Independent verification corrected the provenance repository identity to
  `digitalygo/1-1-5-envai-aaas`. Older capability and response captures retain
  their historical provenance.

## Impact assessment

The additions are unreleased source. No dependency, Go version declaration,
migration, CI workflow, release policy, tag or publication changed. This operation
does not commit or push the work. Backend and PHP sources were read only.

Project-key scope and the human-administration/financial-management exclusions
remain intact. No public cron-create method was invented. Documentation
explains raw versus typed log results, optional/null request-map values, and why
cancelling a turn does not necessarily remove an active-cron deletion guard.

## Validation steps

The orchestrator independently verified:

- `scripts/check_contracts.py` and its tests: six source snapshots, fixture
  hashes, canonical repository, 92-operation subset and frozen v2 passed.
- Regeneration plus gofmt: before/after hashes of all generated outputs matched.
- Formatting, `go vet ./...` and full uncached race tests passed.
- Aggregate statement coverage: 91.25%, above the unchanged 80% gate. Independent
  changed-statement coverage: 334/346, 96.53%. Remaining uncovered statements
  are defensive failure branches.
- The separate tooling module passed vet and uncached race tests against an
  isolated MySQL 8.4 database. Main-module tests used isolated PostgreSQL 18.
- Root and tooling `govulncheck@v1.9.0` passed with no affected vulnerabilities.
  Two required-module advisories were reported as not called by this code.
- Markdown structural checks, relative links and `git diff --check` passed.
- Independent quality judgment: PASS. Independent security review: PASS.

Checks used job-scoped `GOTOOLCHAIN=go1.27.4`, within the declared CI 1.27.x
matrix. The original host Go 1.27.1 reported standard-library advisories; the
patched job toolchain passed without changing the host default or repository
version requirements. Go 1.24 and 1.26 were not run locally. No live backend or
paid-provider test is claimed.

Evidence is retained in `/tmp/envai-sdk-go-leader.log`,
`/tmp/envai-sdk-go-tooling-leader.log`,
`/tmp/envai-other-sdks-delta-coverage.json`, and
`/tmp/envai-sdk-go-review-snapshot.json`. The reviewed non-trace SHA-256 is
`9865d484fad94c42b775661f7898692bf1fd9f67064d32a07c5f3fd04a0a127a`.

The backend advanced independently during verification without changing the
pinned contracts, request/resource definitions or routes. Temporary databases
were isolated per SDK, and their containers and private connection files are
removed at task closure.
