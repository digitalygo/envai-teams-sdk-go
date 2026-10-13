# Wiki / contract 3 validation

Validated against Go SDK alpha `3ae53b16682b2b51d088d4f33995218ff62471d2` on 2026-10-13. All 76 source blobs and file modes were materialized and independently matched to Git identities. Alpha was refreshed before implementation and again after final verification; it remained unchanged. Local history is a synthetic baseline and must never be pushed. Publication uses guarded real-tree overlays.

The five shared contracts are byte-pinned to verified backend commit `c27963d0848540679bccf30d2aeab354097b1a38`. Each local Git blob independently matches the published backend tree. The immutable original numeric v2 schema has blob `02b13b23cd06eacd81059d1ed99e0027769d75e4`. Shared synthetic HTTP-kernel responses now cover pending completion accounting and immutable snapshot provenance, a scheduled conversion diagnostic, and unchanged historical numeric completion data; actual response headers are retained separately. No deployment is implied by this pin.

## Results

- Go 1.26.9 on Linux; PostgreSQL 18.3; Node 24.19.0 / npm 11.9.0.
- Complete Go suite with real PostgreSQL, `-race`, no test cache and coverage: **497 test/subtest passes**, no failures. The subprocess-only helper is skipped in the parent and executed by six child processes.
- Main SDK statement coverage: **94.6%**. Aggregate including all examples: **3054/3253, 93.88%**.
- Changed production Go statements: **1971/2046, 96.33%**. Every changed Go production file is between **86.36% and 100%**. Detailed file results are in `wiki-delta-coverage.json`. The report intersects Go coverage blocks with changed lines and includes every block of new production files.
- Standard-library trace / bytecode-line coverage of changed Python generator lines: **44/44, 100%**; provenance checker **24/25, 96%**. Its deliberately inactive pre-publication candidate branch is the uncovered line.
- Generator rerun is byte-identical; five source schema hashes, all attributed fixtures and the complete **65-operation** inventory pass. Provenance mutation regression tests pass, including every shared fixture.
- `go vet ./...`, workflow actionlint, and Git whitespace checks pass.
- Exact CI `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...` passes with **zero reachable or imported-package vulnerabilities**. It reports two module-only advisories in the pre-existing indirect `golang.org/x/text v0.14.0` dependency (GO-2026-6629 and GO-2026-5970); neither vulnerable package/path is used by this SDK. This is not a claim that the entire transitive module set has no advisories.
- Release-tooling install, semantic-release adapter patch, all **91 release tests**, and npm high-severity audit pass; npm reports no vulnerabilities. No release was run or manually published.

## Verified boundaries

Default automatic contract 3, capability preflight, namespaced mutation routing, missing/wrong/repeated acknowledgments, unsupported/old servers, no mutation retry or downgrade, read-only restrictions, explicit legacy APIs bound to 2, finite legacy replay EOF, strict numeric v2 / nullable v3 before hydration, mixed replay, reconnect selection and original event bytes.

Streaming multipart upload, inclusive 1,000,000,000-byte validation, exact path/query/header encoding, idempotency validation, durable source/status/download, conversion diagnostics, immediate withdrawal, inbox/outbox/sync, all/ranged catalog reads and revision parameters, typed unknown-field retention and decimal/currency provenance, atomic completion page rejection and snapshot turn binding. Internal artifact writer/CAS/QUERY capabilities are documented without fabricated public endpoints.

Exact shared RFC8032 Ed25519 vector, signed raw bytes, header duplication/casing/malformed values, timestamp bounds, old logical events with fresh delivery timestamps, wrong workspace, tampering, key metadata and canonical encoding, bounded unknown-kid refresh, expiry and concurrent verifier calls. Provider HMAC callbacks and existing tool signature preimages remain separate.

Real PostgreSQL intake checks verify atomic receipt + application outbox commit, rollback/retry after callback failure, missing tables/store failure, verification before database work, same-event/different-body collision, independent connection scopes, and **six independent competing processes yielding one accepted receipt/job and five duplicates**. The restricted callback executor does not expose Commit/Rollback. Application callback misuse and external-effect idempotency remain application responsibilities.

## Reproducible checks

```sh
python scripts/check_contracts.py
python scripts/test_check_contracts.py
python scripts/generate.py
gofmt -w models_generated.go operations_generated.go events_generated.go open_models_generated_test.go
git diff --exit-code
go mod download
go vet ./...
ENVAI_TEST_POSTGRES_DSN='postgres://.../disposable_database?sslmode=disable' \
  go test -race -count=1 -coverprofile=coverage.out ./...
python scripts/check_coverage.py coverage.out
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
actionlint .github/workflows/*.yml
npm ci --ignore-scripts --prefix .github/release-tools
node .github/release-tools/patch-semantic-release.mjs
npm test --prefix .github/release-tools
npm audit --audit-level=high --prefix .github/release-tools
git diff --check
```

The disposable PostgreSQL server and tests were launched in one shell lifetime on an isolated test port. CI now provisions PostgreSQL 18 for every Go matrix job and supplies the test-only DSN. Go 1.24 and 1.27 were not executed locally; those remain fresh remote matrix gates. PostgreSQL is the supported SQL inbox adapter; no real MySQL/SQLite claim is made.

Synthetic fixtures, local HTTP fakes and public signature vectors only. No Envai/provider paid calls, customer data, new credentials, production migration, deployment, backend/FFS merge, tag rewriting, manual release, push or SDK PR was performed by this work.

## Independent review corrections

The external boundary review reproduced acceptance of negative/overflowing completion costs, nullable nested legacy usage and `data:null` completion pages, plus loss of allowed unknown inline usage fields. These are fixed with retained independent regressions. All cost-bearing descendants below usage/chat-usage (including lists) are checked before hydration; unrelated root extensions remain opaque. Every generated named open object, including inline schemas, preserves exact unknown JSON values. The generated open-model matrix covers all such shapes.

Finite v2 replay is now buffered and validated atomically under a configurable 64 MiB default snapshot limit before exposing any event. Late malformed, null-bearing, mixed-v3, truncated or gapped tails fail without yielding a prefix; cancellation and EOF are covered. V3 continues incrementally. The final complete PostgreSQL/race and coverage results above include these changes and the real published backend fixture pin.

Read-only actual GitHub metadata plus the existing locked Semantic Release dry-run predict release type `major` and version `v1.0.0-alpha.2` on alpha. The real channel note declares `alpha`; there is no stable v1 tag, so this continues the unreleased 1.0.0 line. The existing Go module-major guard passes unchanged. Local simulated refs were unchanged and no publish callback ran. No module-path migration or tag modification was needed.
