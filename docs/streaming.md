# Streaming, recovery, and errors

[README](../README.md) · [Getting started](getting-started.md) · [API reference](api-reference.md) · [Client tools](client-tools.md)

## New turns

`RunTurn(ctx, chatID, message, options)` opens SSE, waits for a live `ping`, then submits the message once. The default live-readiness timeout is 30 seconds. Use `TurnOptions.LiveTimeout` and the parent context to bound readiness and total time. An HTTP 200 stream response alone does not establish live readiness.

The helper correlates the accepted turn ID. `OnEvent` receives only events of that turn. The result contains `ChatID`, `TurnID`, `LastSeq`, `OrchestratorText`, and the terminal `Event`. Only `turn.completed`, `turn.cancelled`, and `turn.failed` end a turn; `agent.done` does not. Failed/cancelled terminals are returned as events rather than ordinary Go errors, so inspect `result.Terminal.Type` even when `err == nil`.

Text accumulation uses the top-level orchestrator only. A final `agent.text.done` replaces accumulated deltas. Unknown valid event types remain observable. Handler callbacks and event payloads can contain private content; log only what your application intends to disclose.

## Existing streams and checkpoints

`Events(ctx, chatID, StreamOptions{...})` returns an `EventStream`. Call `Next` from one goroutine and always `Close` it. `Cursor` may be read concurrently. The cursor advances as a non-ping event is delivered, before your application commits its own effects. Save the event's sequence only after processing succeeds; the SDK cursor is not a transaction checkpoint for your application.

`ping` has sequence zero and does not advance the cursor. Duplicate sequences are skipped. An observable upward gap attempts replay within the reconnect budget; repeated or unrecoverable gaps return `*ReplayGapError`. Invalid envelopes or malformed known payloads return `*ProtocolError` immediately.

Reconnect is opt-in: `MaxReconnects` defaults to zero and its budget covers the stream lifetime. The default reconnect delay is 250 ms multiplied by the attempt number; the connection timeout defaults to 10 seconds. A context deadline is needed to bound an otherwise idle stream. Reconnects do not resend messages or tool outputs.

An `After` sequence is only meaningful while the same chat session remains resident. The API has no durable epoch or cursor-expired signal. Restart/eviction can reset sequences, and a lost prefix or reset is not always observable as a gap. After a known reset, reconcile durable application state rather than silently continuing from an old cursor. A new stream does not guarantee historical completeness.

Replay and live subscription are not atomic. Waiting for a ping before a new turn reduces the attach race, but cannot recover already lost events. Use one executor or a shared atomic call store when multiple subscribers could process the same client-tool request.

## Cancellation

Cancelling a Go context or closing the stream stops local work. It does not cancel the remote turn. `CancelChatTurn` is an explicit API mutation; only a `turn.cancelled` event confirms remote cancellation. A timeout or disconnect after submitting a mutation is ambiguous and must be reconciled rather than automatically submitted again.

## Error handling

Use `errors.As` with `*envai.APIError` to inspect `Status`, `Type`, `Code`, `RequestID`, and `RetryAfter`. Branch on stable codes and status, not human-readable message text. `Message` and `Details` may contain sensitive server data; `Error()` intentionally omits them. Transport, protocol, replay, cancellation, and application-handler failures also return Go errors.

No mutation is automatically retried. If enabled with `RetryAttempts`, eligible JSON GETs retry supported transient statuses with jitter and `Retry-After`. Stream reconnects use their own budget, and binary/partially consumed responses are never retried as buffered JSON.

Practical recovery rules:

- Authentication/access error: verify account access and credentials; do not loop on the same rejected key.
- Validation error: inspect safe error metadata and correct the request.
- Replay gap or EOF before a terminal: preserve the last processed checkpoint and reconcile the turn.
- Local tool store error: stop execution, preserve original records, and reconcile side effects before repair.
- Ambiguous mutation/output delivery: check authoritative application/server results before considering another action.

## API compatibility limits

Approval resolution currently responds with HTTP 409. A steer acknowledgement does not guarantee changed execution. Restore acknowledgements do not guarantee durable conversation history; restore does not itself start a new turn. Fork requires a separate empty destination chat in the same owner/workspace and both chats idle. Treat usage/cost returned by the API as reported data, not a replacement for billing records. These limits cannot be repaired by a client retry policy.
