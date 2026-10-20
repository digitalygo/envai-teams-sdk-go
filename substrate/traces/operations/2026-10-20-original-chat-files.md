---
status: completed
created_at: 2026-10-20
files_edited:
  - attachments.go
  - attachments_test.go
  - attachments_native_test.go
  - client_test.go
  - contract.go
  - contract/openapi.json
  - contract/openapi.yaml
  - contract/provenance.json
  - contract/wiki.json
  - contract_provenance_test.go
  - docs/chat-attachments.md
  - docs/wiki.md
  - features.go
  - models_generated.go
  - native_capabilities_test.go
  - scripts/check_contracts.py
  - scripts/generate.py
  - testdata/capabilities/canonical-v3.json
  - testdata/capabilities/provenance.json
  - testdata/chat-attachments/original-ready.json
  - wiki.go
  - wiki_original_test.go
rationale:
  - Match the merged server and PHP original-file behavior.
  - Preserve public feature lists and existing source upload calls.
  - Reject unsupported original uploads before mutation.
supporting_docs:
  - https://github.com/digitalygo/envai-teams/commit/7192fc05ebec817bddc705186c8f195cc0dc5ec5
  - https://github.com/digitalygo/envai-teams-sdk-php/commit/93eafc02271035065ac975f91cd4f77d148c9d2c
---

# Original chat files SDK alignment

## Summary of changes

Added native-original capability guards, `ready` plus `not_required` polling, and optional original-only wiki staging. Current upstream schemas, generated models and provenance were refreshed; frozen v2 contracts remain unchanged.

## Technical reasoning

Native attachment uploads require both `chat_attachments` and `chat_attachment_originals`. Other attachment operations and messages retain only their previous `chat_attachments` requirement. Strict boolean feature maps normalize to enabled names while preserving `Capabilities.Features` as `[]string`. Duplicate map keys are rejected as ambiguous; homogeneous feature lists retain their previous behavior, including repeated names.

`UploadSourceFile` accepts an optional trailing processing argument. Omitted or explicit `convert` retains the file-only helper request; `original` adds a separate multipart field and requires `wiki_original_sources`. Low-level multipart validation rejects invalid, duplicate or misplaced processing fields before any request. Existing HTTP, schema, multipart and polling structures were reused without a new dependency.

## Impact assessment

Baseline: `6cceaebd8411095db40c78f2bf72c89c8a4aefe3`. All five current source snapshots match immutable server `7192fc05ebec817bddc705186c8f195cc0dc5ec5`. PHP `93eafc02271035065ac975f91cd4f77d148c9d2c` is the behavior reference. Sibling Python and TypeScript SDKs were aligned in the same task; server and PHP files were not changed.

Existing call sites remain valid. Original bytes, filenames and idempotency keys are preserved, with no automatic mutation retry or OCR fallback. There are no dependency, package version, workflow, migration, tag, release or deployment changes. Work remains local and uncommitted.

## Validation steps

The leader independently ran the repository schema checkers, generation and formatting checks, static analysis and all race-enabled package tests with a dedicated disposable PostgreSQL 18 database:

```bash
python scripts/check_contracts.py
python scripts/test_check_contracts.py
python scripts/generate.py
gofmt -w models_generated.go operations_generated.go events_generated.go open_models_generated_test.go
go mod download
go vet ./...
go test -race -count=1 -coverprofile=/tmp/envai-sdk-go-leader-coverage.out ./...
python scripts/check_coverage.py /tmp/envai-sdk-go-leader-coverage.out
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
npm test --prefix .github/release-tools
npm audit --audit-level=high --prefix .github/release-tools
git diff --check
```

Verification used the sibling Python SDK virtual environment for Python scripts, `ENVAI_TEST_POSTGRES_DSN` selecting only the disposable SDK database, and `GOTOOLCHAIN=go1.27.2`. Generated-file hashes before and after regeneration were identical; this substitutes for CI's clean-checkout `git diff --exit-code` while intentional changes are unstaged.

All packages passed, including real PostgreSQL integration. Aggregate statement coverage was 94.07% (3395/3609); the root package reported 94.7%. Every changed handwritten runtime function met the 80% delta threshold: capability normalization 86.8%, source-mode detection 85.7%, source upload 95.2%, source multipart validation 95.7%, and the remaining changed functions 100%. The generated capabilities unmarshaller reached 90.9%. Both contract-checker tests and all 91 release-tool tests passed. Audits found no release-tool dependency vulnerabilities; govulncheck found no reachable vulnerabilities. The workstation's older Go 1.27.1 standard-library findings were avoided with a temporary patched toolchain selector, without changing project requirements. Two required-module findings were unreachable according to govulncheck.

Quality judgment: PASS. Focused cross-SDK security judgment: PASS. Reviewed 22-file package SHA-256: `fe4954baaaee21f2dacb5386220cbbc0a4e30d5e7c378cefe0b419a317505dc9`, excluding this operation record. Other Go matrix versions remain CI checks. No production request, real user file or paid provider was used. The disposable database container was stopped and removed after combined verification.
