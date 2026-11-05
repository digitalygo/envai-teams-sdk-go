> The new Wiki/contract 3 APIs require a compatible server and a release containing this migration. The older published alpha tag below predates these APIs; see [migration](contract3-migration.md).

# Getting started

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

## Application lifecycle

Reuse a client across operations. Pass contexts with bounded deadlines; close event streams and downloads. Local cancellation stops this client's I/O and cooperative handlers. To cancel remote work, explicitly call `CancelChatTurn` and observe `turn.cancelled`.

Start with turn helpers for new messages. Use low-level streams when you need a custom lifecycle, after reading the [checkpoint and recovery rules](streaming.md). Add local tools only after you have an explicit authorization policy and durable call store.

## Offline checks

Tests use mocks and public signature vectors; they require no Envai credentials. Durable receipt/outbox integration tests additionally use a disposable PostgreSQL database selected by `ENVAI_TEST_POSTGRES_DSN`; CI provisions it automatically. `go test ./...` compiles examples and executes the quickstart against a local mock server. The README includes the full generator, provenance, race, vet, and coverage commands. Passing these checks demonstrates client behavior against the bundled contracts, not availability or configuration of your API deployment.

