# Getting started

> The new Wiki/contract 3 APIs require a compatible server and a release containing this migration. The older published alpha tag below predates these APIs; see [migration](contract3-migration.md). The project resources described under [Project resources](#project-resources-unreleased) are also absent from that tag.

[README](../README.md) · [API reference](api-reference.md) · [Streaming](streaming.md) · [Client tools](client-tools.md)

## Install

Requires Go 1.24 or newer. Use a supported, security-patched toolchain.

```sh
mkdir envai-example
cd envai-example
go mod init example.com/envai-example
go get github.com/digitalygo/envai-teams-sdk-go@v1.0.0-alpha.1
```

[`v1.0.0-alpha.1`](https://github.com/digitalygo/envai-teams-sdk-go/releases/tag/v1.0.0-alpha.1) is a published prerelease available through the public Go module proxy. The explicit version opts into the alpha release; its APIs may change. Commit `go.mod` and `go.sum` and review upgrades deliberately. The module path is `github.com/digitalygo/envai-teams-sdk-go` and its package name is `envai`.

For unreleased changes, use a reviewed commit SHA or `@alpha`; a branch/commit download resolves to a pseudo-version. Prefer the tagged version for repeatable application installs.

To inspect and build all examples instead:

```sh
git clone --branch v1.0.0-alpha.1 https://github.com/digitalygo/envai-teams-sdk-go.git
cd envai-teams-sdk-go
go test ./...
go build ./examples/...
```

## First turn

Use the complete [quickstart program](../examples/quickstart/main.go). It reads these application-owned environment variables:

- `ENVAI_BASE_URL`: your trusted API URL, for example `https://api.example.com/api/v1`
- `ENVAI_API_KEY`: an API key for the account
- `ENVAI_TEAM_ID`: an existing team configured for that account

From a checkout, run `go run ./examples/quickstart`. The example creates a workspace and chat and submits a real turn, which can incur account charges. It does not delete the created resources. It limits the local operation to two minutes and prints the terminal type separately from assistant text. A failed or cancelled terminal is not a successful answer.

The SDK itself never reads environment variables. Supply credentials explicitly through `Config`; never commit them or include them in URLs. Keep IDs returned by the API rather than synthesizing or UUID-coercing them.

## Configuration

`envai.New(envai.Config{...})` validates configuration without making a request.

| Option | Default / behavior |
| --- | --- |
| `StrictValidation` | False for compatibility. Validate canonical HTTP parameters/request bodies before preflight and JSON success envelopes before unwrapping. |
| `NativeTasks` | False. Opt into `native-tasks-v1` for background APIs, discovery, streams and tool outputs. |
| `LegacyReadOnly` | False. Select historical v2 replay; cannot combine with native tasks. |
| `BaseURL` | Required. HTTPS origin or URL ending in `/api/v1`. An origin gets `/api/v1` appended. No credentials, query, or fragment. |
| `APIKey` | Explicit bearer key for authenticated operations. Public operations omit it. |
| `JWKSURL` | Origin-level `/.well-known/jwks.json`; set only to a trusted signing-key endpoint. Never receives the API key. |
| `RequestTimeout` | 60 seconds for ordinary requests. Set a context deadline for a whole workflow. |
| `MaxResponseBytes` | 8 MiB for buffered JSON responses. Downloads remain streamed. |
| `RetryAttempts` | 0, so ordinary request retries are disabled. Accepts 0–10 retries for eligible GETs. |
| `RetryDelay` | 250 ms base delay when retries are enabled. |
| `HTTPClient` | Optional caller-owned transport configuration. The SDK copies the client, disables redirects, and manages deadlines per request. |
| `AllowHTTP` | False. Enable only for an explicitly trusted local development server. |
| `Observer` | Optional callback for operation, status, attempt, and duration. No bodies, URLs, or credentials. |

The default transport honors standard proxy environment variables. Review proxy configuration before using secrets. If you inject an HTTP transport, keep TLS verification enabled and honor request cancellation. Redirects are refused.

## Project resources (unreleased)

These methods exist on the `feature/user-company-projects` branch only and are not part of the published `v1.0.0-alpha.1`. To try them, depend on a reviewed commit SHA (`go get github.com/digitalygo/envai-teams-sdk-go@<sha>`). This guide makes no claim about a future tag or version. They require a server that implements the pinned canonical contract (`digitalygo/envai-teams` at `7f4085dbfdd3f208b82716119cff1860fbd502bb`).

Before you start:

- Use a project API key. It selects the project; only set `XEnvaiProjectID` where you also hold a human session, and it must match.
- Treat deletion as access cut plus pending cleanup. `DeleteWorkspace` and `DeleteChat` return a `CleanupResult` whose intents are usually still `pending`; this is not proof of physical purge. A 409 conflict arises from active turns or active cron protection; active cron protection can persist even after turn cancellation, and cancelling a turn does not guarantee removal of the deletion guard. Inspect the error code and existing cron, occurrence, and turn state, taking explicit operator action where appropriate rather than treating turn cancellation as a generic recipe. The SDK just surfaces 409.
- Skill bodies are `map[string]json.RawMessage`, so use `json.RawMessage` values and omit keys you do not change.
- Chat crons can be listed, read, updated and cancelled, but not created through the API.
- Logs have three views. Use `GetProjectLogsEvents`, `GetProjectLogsAnalytics` and `GetProjectLogsFilterCatalog`; `GetProjectLogs` returns raw JSON. Use `ProjectUsagePages` and `ProjectLogPages` to iterate numeric pages on your configured origin.

```go
skills, err := client.ListProjectSkills(ctx, envai.ListProjectSkillsParams{ProjectID: projectID})
if err != nil { return err }

filters, err := client.GetProjectLogsFilterCatalog(ctx, envai.GetProjectLogsParams{
    ProjectID: projectID,
    From:      envai.Present("2026-11-01"),
    To:        envai.Present("2026-11-07"),
})
if err != nil { return err }
_, _ = skills, filters.Types
```

See the [API reference](api-reference.md#project-resources-deletion-and-observability) for every field, constraint and error.

## Application lifecycle

Reuse a client across operations. Pass contexts with bounded deadlines; close event streams and downloads. Local cancellation stops this client's I/O and cooperative handlers. To cancel remote work, explicitly call `CancelChatTurn` and observe `turn.cancelled`.

Start with turn helpers for new messages. Use low-level streams when you need a custom lifecycle, after reading the [checkpoint and recovery rules](streaming.md). Add local tools only after you have an explicit authorization policy and durable call store.

## Offline checks

Tests use mocks and public signature vectors; they require no Envai credentials. Durable receipt/outbox integration tests additionally use a disposable PostgreSQL database selected by `ENVAI_TEST_POSTGRES_DSN`; CI provisions it automatically. `go test ./...` compiles examples and executes the quickstart against a local mock server. The README includes the full generator, provenance, race, vet, and coverage commands. CI also runs `govulncheck`. A past local run on Go 1.27.1 reported standard-library advisories; use a patched Go release (CI resolves the latest 1.27 patch) and do not run production workloads on an unpatched runtime. Passing these checks demonstrates client behavior against the bundled contracts, not availability or configuration of your API deployment.
