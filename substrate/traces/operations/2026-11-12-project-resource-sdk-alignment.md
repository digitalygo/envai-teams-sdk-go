---
status: completed
created_at: 2026-11-12
updated_at: 2026-11-14
files_edited:
  - contract/events.json
  - docs/model-selection.md
  - model_turn_test.go
  - models_generated.go
  - stream.go
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
  - Preserve exact per-turn model selection through typed requests and RunTurn.
  - Pin the reviewed model-only server commit and reject boundary Unicode White_Space in model values, with no adoption of unrelated server API changes.
supporting_docs:
  - https://github.com/digitalygo/envai-teams/tree/7f4085dbfdd3f208b82716119cff1860fbd502bb
  - ../../../docs/model-selection.md
  - ../../../AGENTS.md
  - ../../../docs/api-reference.md
  - ../../../docs/contract3-migration.md
  - ../../../contract/provenance.json
  - https://github.com/digitalygo/1-1-5-envai-aaas/tree/ba021aa61e30e379880f809d45f90e5d3564d43e
  - https://github.com/digitalygo/envai-teams/commit/f970714c0a358ff94b14f15eabde197c55861e74
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

## Update: 2026-11-13 per-turn model selection

### Summary and reasoning

Starting from SDK `791db34` on `feature/user-company-projects`, refreshed the
current contracts to backend `7f4085dbfdd3f208b82716119cff1860fbd502bb`, which
already included the requested `alpha` merge after pull. Backend PR 58 was
current and push reported no further changes. GitHub metadata now resolves
the backend to `digitalygo/envai-teams`; prior source names above are historical.

Added `Optional[string]` model fields to `ChatMessageRequest` and `TurnOptions`.
`RunTurn` copies the option into its fresh request. Invalid explicit null,
non-string, empty and padded model values fail before message submission,
without normalization or paid retry. The override does not become a default,
change the team, require native tasks, prefetch a catalog or authorize routing.
Generated `LLMModelEntry.Selectable` remains guidance, not a filter or gate.

### Impact and verification

No dependency, workflow, version, endpoint or backend source change was added.
All six source snapshots match the backend pin; frozen v2 and historical
captures are unchanged. The new guide covers optional values, idempotency,
model errors, event identities and resident-only replay.

Independent contract checks and checker tests, repeated generation, formatting,
vet and full uncached race tests passed with Go 1.27.4. Root tests used an
isolated PostgreSQL 18 database; tooling tests passed vet and race checks with
isolated MySQL 8.4. Final overall statement coverage was 91.28%, with 17/17
statements in changed executable blocks covered, including generated decoding.
Both `govulncheck@v1.9.0` scans passed with zero reachable vulnerabilities;
two root-module and three tooling-module advisories affected required modules
but no imported vulnerable package or called vulnerable symbol.

Quality judgment and security review returned PASS for non-trace snapshot
`30ecabf7a0131b5932d946328191777dc94abe3ffea6bb07d7b4166d758b86e0`.
Go 1.24/1.26 and actionlint were not run locally. No live provider, deployment
or release verification is implied. Evidence remains under
`/tmp/envai-model-sdk-verify/leader-go*`; task containers, their volumes and
private connection files were removed after verification. Unrelated resources
remain intact.

## Update: 2026-11-14 model input validation follow-up

### Follow-up summary of changes

A minor finding on the model input was corrected. The four SDK OpenAPI copies and the six source provenance and current-docs assertions were synced to the reviewed server commit [f970714c0a35](https://github.com/digitalygo/envai-teams/commit/f970714c0a358ff94b14f15eabde197c55861e74), which holds only the canonical `not.pattern` that forbids the 25 exact Go Unicode `White_Space` code points at the start or end of `model`. The Go contract checker pin and provenance were updated; the SDK runtime is unchanged.

### Follow-up technical reasoning

The server runtime was already correct and is unchanged; the published schema is now as strict. The rule uses a YAML-decoded Unicode positive class, so it avoids Python final line feed and ECMAScript NEL escape differences. The string type, `minLength`, optional omission and non-null wire behavior are unchanged, and nothing is coerced, trimmed, normalized or case-folded. Valid Unicode IDs, interior whitespace and the `FEFF` and `001C` to `001F` code points that Go does not treat as space stay accepted. Null and omission semantics of other DTO types are not tightened globally.

The server working `alpha` branch has also merged unrelated Wiki PR 60 and image PR 61. The SDKs intentionally pin the reviewed model-only commit instead of adopting those unrelated API features. The pin commit exists locally and its push is pending; it will be a public `alpha` ancestor before any SDK push, and publication is not claimed yet. The user explicitly approved pushing server `alpha` and the four SDK `feature/user-company-projects` branches, with repo-admin team membership and bypass verified by the leader and no policy change.

### Follow-up impact assessment

No package version, dependency, CI, endpoint, count, authentication, billing, provider, hold or restart change was made, and no deployment, paid or live provider test, or new release is claimed. No runtime change.

### Follow-up validation steps

- Go 1.27.4: root and tooling `go vet` and uncached race tests with SQL passed. Root coverage was 91.28% (6319/6923). The contract checkers and unit tests passed. `govulncheck` v1.9.0 passed in both modules with zero reachable vulnerabilities; the unchanged required-module advisories were not live called and are not part of this delta.
- The changed checker pin lines 9 and 59 are covered at 100% through Python coverage. No production runtime source changed.
- Documentation-only record update on 2026-11-14. Checked that the five operation records are the only files written by this step and that the reviewed source files remain identical to the manifest at `/tmp/envai-sdk-model-minor-20261113/review/manifest.json` (50 changed paths across the five repositories; trace records are excluded from the gate).
- Quality judgment `PASS` and security review `PASS` over combined source and test snapshot `cb56581c5f2e5c9f947901e63a56521d126a3bdee4a025afc48339952e902e60` (lexically sorted repository name and path, NUL, bytes, NUL). All 50 paths were read and their diffs matched. Security review found no authentication, payer, provider or retry change and verified the six pinned sources.
- This is local completion of the code task. Commit and push are pending, so green remote CI and remote publication are not claimed. Python 3.11 and 3.12, Go 1.24 and 1.26, other PHP editions, Laravel 11 and 12, and Node 22 were not run locally; the remote CI matrices run after commit and push. `git diff --check` passed in every repository.
- Test resources were limited to task-named isolated Podman PostgreSQL and MySQL containers; `envai-netbird-demo-*` resources were not touched. Cleanup stays pending until the publication tests complete, and the leader owns removal at closure.
