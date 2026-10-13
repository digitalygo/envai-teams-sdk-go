# Signed client tools

[README](../README.md) · [Getting started](getting-started.md) · [API reference](api-reference.md) · [Streaming](streaming.md)

## Three explicit steps

1. Define an application handler with `envai.Tool` and register it locally using `NewRegistry`.
2. Publish the intended definitions to the account with `CreateUserTool`. `registry.Definitions()` produces `UserToolCreateRequest` values. Review these remote mutations deliberately; rerunning a create is not an upsert.
3. Enable the selected tool on the intended agent through its `external_tools_enabled` configuration. Registration alone does not grant an agent access.

Create a `ToolRunner` with that registry, a call store, and a stable account/deployment namespace, then provide it as `TurnOptions.Tools`. `RunTurn` scopes requests to the turn it starts and forwards matching result/terminal observations to the runner. The [echo example](../examples/echo/main.go) demonstrates a harmless local handler and file store, without automatically changing remote tool or agent configuration.

Run the echo example from a checkout with `go run ./examples/echo` after setting `ENVAI_BASE_URL`, `ENVAI_API_KEY`, `ENVAI_CHAT_ID`, `ENVAI_STATE_DIR`, and `ENVAI_ACCOUNT_SCOPE`. The chat's configured agent must already have `echo` enabled. This sends a real, potentially billable turn.

## Verification and schemas

Before running a handler, the SDK validates the ATP envelope and expected chat, verifies an Ed25519 signature, validates arguments against the local schema, and obtains an atomic call claim. The signed message is exactly `call_id + "\n" + tool_name + "\n" + RFC8785(args)`. The key set is fetched from the configured trusted JWKS endpoint without bearer authentication. Keys are cached for five minutes by default, with one refresh for an unknown key ID.

The signature protects the call ID, tool name, and canonical arguments. It does not bind the chat, turn, account, timestamp, or freshness. Only pass events from the trusted authenticated stream; do not execute arbitrary saved events as new work. For manual `Handle` usage, enforce the intended turn yourself and feed matching result/terminal events to `ObserveContext`, checking its errors.

Handlers receive owned canonical JSON bytes, including binary64 number semantics. Use strings for exact integers beyond ±(2^53−1). Local schemas require an object root. An omitted `$schema` defaults to Draft 2020-12; explicit Draft 4, 6, 7, 2019-09, and 2020-12 are supported. Nested validation and format checks apply; remote references are disabled. Use a declared modern dialect for cross-language tools: TypeScript supports Draft 6 and newer but not Draft 4.

## Authorization and sandboxing

A valid signature does not establish permission for every side effect. Use `RunnerOptions.Authorize` and enforce account/resource authorization in the handler. Handlers are ordinary application Go code; the SDK does not sandbox processes or grant filesystem capabilities.

For custom file or command tools, restrict allowed directories and operations, account for symlinks and path traversal, validate argument lists, avoid shell interpolation, bound output, and apply the operating system's isolation and least-privilege controls. Never provide a general shell or unrestricted filesystem handler merely because its request is signed. Protect secrets from handler output and logs.

## Durable claims and recovery

`FileCallStore` uses an application-owned private directory with atomic claims and durable writes. Keep the same directory and namespace across restarts and API-key rotation. Separate accounts/deployments need separate namespaces. Local filesystems must support atomic mkdir/rename and fsync; this implementation does not support network filesystems. For multiple hosts, implement `CallStore` with a shared transactional database. `MemoryCallStore` is non-durable and only suitable for tests or intentionally ephemeral work.

States are `started`, `submitted`, `resolved`, and `ambiguous`. A claim is recorded before handler effects. `submitted` is persisted before delivery. Existing records are never automatically reclaimed. Matching results create resolved records, including when the request was not seen; terminal observations create admission barriers. Malformed records and missing claims during transitions fail closed without overwriting original evidence.

An empty claim directory is ambiguous, not safe to reuse. Never delete or expire records to make a stalled call run again without reconciling its effects. A crash between an external effect and saving state cannot be solved by local deduplication alone. Use `ToolInvocation.CallID` as a downstream idempotency key, a transactional operation, or an explicit reconciliation procedure. There is no exactly-once guarantee.

## Execution and output

Concurrency defaults to four handlers. Each tool timeout defaults to 120 seconds and must be 1–600 seconds. Verification and authorization time count against the local deadline, and the default one-second submission margin reduces the execution window. Callers can configure the margin. The server's deadline began before local receipt; this is a local bound, not proof of remaining server time.

Handlers must check their context before side effects and propagate it to cancellable dependencies. Result/terminal observations cancel matching local handlers. Go cannot forcibly terminate code that ignores cancellation, and such a handler retains its concurrency slot until it exits.

Return a `ToolResult` string or use `JSONResult(value)` for one JSON serialization. Output must be valid nonempty UTF-8 and at most 128 KiB. Invalid/unknown tools, rejected authorization, schema failures, handler failures, busy slots, and invalid output produce sanitized controlled error outputs. Handler arguments and output are not automatically logged.

Output is submitted once. Network ambiguity is recorded and surfaced. `409 tool_already_resolved` marks the Go call as resolved but still returns the HTTP error; it does not prove which client supplied the accepted result. Other 404/409 responses cannot establish earlier success. Check the corresponding result event and authoritative downstream state before any manual recovery.
