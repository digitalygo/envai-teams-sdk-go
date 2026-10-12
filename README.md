# Envai Teams Go SDK

Typed Go client for the Envai HTTP API and ATP v2 event stream. Language compatibility: Go 1.24+. For production use a currently supported Go release with all security patches (validated with Go 1.26.9).

This is a **pre-release source implementation** on the `alpha` branch. Installing from source does not imply a published release, and this SDK does not deploy the server. Pin a reviewed commit for reproducible use.

The implementation is pinned to `digitalygo/envai-teams@55275860c9a950eaad8e39e9d2b728149084779c`. `contract/openapi.yaml` and `contract/events.json` are the byte-exact source snapshots. All 52 operation IDs have named typed methods. Examples use explicit application-owned handlers; there are no built-in filesystem or shell tools.

## Use

In an authorized checkout, run `go test -race ./...`. Consumers can pin a reviewed commit with `go get github.com/digitalygo/envai-teams-sdk-go@<commit>` using their normal GitHub private-module credentials. Do not put credentials in URLs or source control. No release tag exists yet.

```go
client, err := envai.New(envai.Config{
    BaseURL: "https://your-envai.example/api/v1",
    APIKey: os.Getenv("ENVAI_API_KEY"),
})
if err != nil { return err }

workspace, err := client.CreateWorkspace(ctx, envai.CreateWorkspaceParams{
    Body: envai.WorkspaceCreateRequest{Name: "Research"},
})
if err != nil { return err }
chat, err := client.CreateChat(ctx, envai.CreateChatParams{
    WorkspaceID: workspace.ID,
    Body: envai.ChatCreateRequest{TeamID: "team_YOUR_EXISTING_TEAM_ID"},
})
```

An origin-only base URL is normalized to `/api/v1`. HTTPS is required unless `AllowHTTP` is explicitly enabled for development. Public JWKS is derived from the API origin; `JWKSURL` can explicitly select a different trusted HTTPS control-plane origin and never receives the bearer token. Redirects are not followed.

The API methods match operation IDs (`sendChatMessage` → `SendChatMessage`, `getJwks` → `GetJWKS`). Each accepts `context.Context` and a named `...Params` struct. Resource `data` wrappers are unwrapped, while paginated completion results preserve `Data`, `Links`, and `Meta`. Binary operations return a `Download`; always close it. Multipart uploads use explicit `[]MultipartPart`; the SDK never reads a local path automatically.

Models preserve absence separately from explicit null through `Optional[T]`:

```go
request := envai.WorkspaceUpdateRequest{
    Name: envai.Present("New name"),
}
```

Use `envai.Null[T]()` only for fields whose server schema permits null. Unset optionals are omitted. Flexible schema/metadata/union fields use `json.RawMessage` or a typed JSON map rather than forcing a misleading shape. Unknown REST fields are ignored by typed decoding; events retain their complete original `Raw` bytes.

## Streaming and turns

```go
result, err := client.RunTurn(ctx, chat.ID, "Summarize the project", envai.TurnOptions{
    Stream: envai.StreamOptions{MaxReconnects: 2},
    OnEvent: func(event envai.Event) error {
        payload, err := event.TypedPayload()
        if err != nil { return err }
        _ = payload
        return nil
    },
})
```

`RunTurn` opens SSE, waits for a live `ping`, sends the message, correlates the returned turn ID and requires `turn.completed`, `turn.cancelled`, or `turn.failed`. It returns the terminal event, orchestrator text and last sequence. A failed or cancelled terminal is returned as that terminal; inspect `result.Terminal.Type`. Transport termination without a terminal is an error. `OnEvent` receives events of the accepted turn only. Typed payloads cover all 28 declared types, with raw preservation for future types.

Low-level methods remain available: `Events`, `Next`, `Cursor`, `Close`, `SendChatMessage`, and `CancelChatTurn`. Closing a local stream cancels its local network I/O; it does not cancel the remote turn. Remote cancellation requires `CancelChatTurn`, followed by observing `turn.cancelled`.

A stream rejects simultaneous readers. Its cursor is safe to inspect concurrently. `ping` has sequence zero and does not move the cursor. Replay duplicates are skipped; an upward sequence gap triggers bounded recovery and an unrecoverable gap raises `ReplayGapError`. Reconnect is opt-in through `MaxReconnects`; the budget applies to the stream's lifetime. `After` checkpoints are only best-effort within the same resident server runtime.

## Client tools

1. Implement handlers in your application and register them explicitly with `NewRegistry`.
2. Register their definitions using `CreateUserTool` (or existing deployment tooling).
3. Enable the tool name in the intended remote agent's `external_tools_enabled` list.
4. Provide a call store and a stable account/deployment namespace, then pass the runner to `RunTurn`.

`examples/echo` demonstrates a complete harmless handler. It does not automatically register or modify remote resources. Run it only after configuring a chat/team with the `echo` tool. Required environment variables: `ENVAI_BASE_URL`, `ENVAI_API_KEY`, `ENVAI_CHAT_ID`, `ENVAI_STATE_DIR`, `ENVAI_ACCOUNT_SCOPE`.

```go
store, err := envai.NewFileCallStore("./private-call-state")
if err != nil { return err }
runner, err := envai.NewToolRunner(client, registry, store, envai.RunnerOptions{
    Namespace: "stable-account-id:production",
    MaxConcurrent: 4,
})
```

The runner validates the ATP envelope, expected chat, client-tool kind, Ed25519 signature and RFC 8785 canonical arguments before accessing a handler. JWKS is cached; an unknown key ID refreshes once for that verification. Handlers receive the same canonicalized owned JSON bytes that were verified, including RFC 8785 binary64 number semantics. Integers beyond the safe range may be rounded; use strings when exact large integer values are required. Local JSON Schema validation includes nested constraints. External schema references are disabled. Optional `Authorize` adds application policy before execution; handler arguments and outputs are not logged.

A valid signature authenticates `call_id`, tool name, and arguments. It does **not** sign the chat/turn/account envelope or prove freshness. Only feed a runner events from your trusted, authenticated stream. `RunTurn` restricts tool handling to its newly accepted turn. Applications manually replaying historical requests must reconcile their status before execution.

`FileCallStore` atomically claims a call before side effects using a private directory and durable fsync/rename. Existing records, including corrupt/unfinished records, are never automatically reclaimed. This fails closed after a crash. Keep the directory and its namespace stable across process restarts and API-key rotation. Separate accounts/deployments must use separate namespaces. Do not delete state to recover a stalled call without first reconciling its side effect. The local filesystem must support atomic mkdir/rename and fsync; network filesystems are not supported by this implementation. Multi-host consumers should implement `CallStore` against a shared transactional database. `MemoryCallStore` is explicitly non-durable and intended for tests.

Bounded handlers receive a cancellable context. The local deadline includes signature verification and authorization time and reserves a submission margin. Terminal/tool-result events cancel matching active handlers and persist duplicate-prevention records even when observed before the request. Low-level callers must check errors from `Observe` / `ObserveContext`; a failed persistent write requires reconciliation. Go cannot forcibly stop a handler that ignores its context; such a handler retains its concurrency slot until it returns. For consequential effects, application-level idempotency keyed by `CallID`, transactions, or reconciliation is still required. No exactly-once guarantee is made.

The server output limit is 128 KiB UTF-8. Strings are sent unchanged; `JSONResult` serializes structured values once. Unknown tools, denied authorization, schema violations, handler failures, busy slots, invalid UTF-8, empty or oversized output produce sanitized controlled errors. Although the schema lacks `minLength`, the current gateway rejects empty output; the runner emits `empty_output` rather than an invalid submission. Ambiguous network delivery is recorded and never blindly retried. A 409 `tool_already_resolved` records resolved state but is still surfaced as the HTTP error; it does not prove which client supplied the accepted output.

## Known server limitations

- Replay is in memory; sequences reset after restart/eviction and no epoch/cursor-expired signal exists. Durable/lossless resume cannot be promised. Reset/reconciliation policy belongs to the application.
- Replay and subscription are not atomic. Waiting for `ping` before a new message reduces that race, but does not fix every existing-turn attach case.
- There is no device-claim/session-routing protocol. All subscribers receive a chat's events. Use one executor or a shared atomic call store.
- Approval is a 409 stub; the low-level method preserves that error. Steer records guidance but does not currently change the LLM loop.
- Remote context persistence is best-effort. A restore acknowledgement does not establish durability.
- Signed artifact downloads use `artifact`, optional `expires`, `generation`, then `signature`; the old PHP `path` contract is not used.

No automatic mutation retries occur. Optional GET retries apply only to JSON operations and retryable statuses, with exponential jitter and `Retry-After`. Stream/binary partial reads are never retried as ordinary GETs. Errors expose status, type, code, details, request ID and retry delay; their string representation and client/config formatting redact sensitive information. `Observer` receives only operation/status/attempt/duration.

## Development and verification

```sh
python3 -m pip install -r scripts/requirements.txt
python3 scripts/generate.py
gofmt -w models_generated.go operations_generated.go events_generated.go
go mod download
go test -race -coverprofile=coverage.out ./...
go vet ./...
python3 scripts/check_coverage.py coverage.out
```

The generator creates models, named operations and event payloads; transport, replay and tool execution are handwritten. CI verifies generated outputs are reproducible and the pinned contract hashes are unchanged, then runs unit/integration mocks, the race detector, static checks and a coverage floor. Shared test vectors are identical to the TypeScript/Python/PHP compatibility suite.

## License

MIT, Copyright (c) 2026 DigItalyGo. The original repository license is preserved. Dependencies retain their own licenses.
