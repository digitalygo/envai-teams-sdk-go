# Wiki / contract 3 validation

Validated against Go SDK alpha `3ae53b16682b2b51d088d4f33995218ff62471d2` on 2026-10-13. All 76 source blobs and file modes were materialized and independently matched to Git identities. Alpha was refreshed before implementation and again after final verification; it remained unchanged. Local history is a synthetic baseline and must never be pushed. Publication uses guarded real-tree overlays.

The five shared contracts are byte-pinned in `contract/provenance.json`. The backend publication commit is deliberately pending: this is an explicitly marked working-tree snapshot, not a claim that the new contract was deployed or published. The release coordinator must replace that candidate provenance with the verified backend commit and rerun the contract checks before publication.

## Results

- Go 1.26.9 on Linux; PostgreSQL 18.3; Node 24.19.0 / npm 11.9.0.
- Complete Go suite with real PostgreSQL, `-race`, no test cache and coverage: **364 test/subtest passes**, no failures. The subprocess-only helper is skipped in the parent and executed by six child processes.
- Main SDK statement coverage: **91.8%**. Aggregate including all examples: **1865/2054, 90.80%**.
- Changed production Go statements: **780/844, 92.42%**. Every changed Go production file is between **86.36% and 100%**. Detailed file results are in `wiki-delta-coverage.json`. The report intersects Go coverage blocks with changed lines and includes every block of new production files.
- Standard-library trace / bytecode-line coverage of changed Python generator lines: **38/38, 100%**; provenance checker **24/25, 96%**. Its not-yet-active final-publication pin branch is the uncovered line.
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
gofmt -w models_generated.go operations_generated.go events_generated.go
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
