# Wiki V1 with negotiated contract 3

The Go SDK exposes all 65 shared public operations, including the 11 Wiki/source/JWKS operations. It selects contract 3 automatically. The base URL still ends in `/api/v1`; ATP uses `/atp/v3/...` beneath it. No endpoint is invented for internal writer capabilities.

## Upload first, then index explicitly

```go
client, err := envai.New(envai.Config{BaseURL: baseURL, APIKey: apiKey})
if err != nil { return err }
source, err := client.UploadSourceFile(ctx, workspaceID, "report.pdf", uploadKey, "application/pdf")
if err != nil { return err }
current, err := client.GetWikiSource(ctx, envai.GetWikiSourceParams{
    WorkspaceID: workspaceID, FileID: source.FileID,
})
// Poll current source metadata until conversion completes. Conversion alone does
// not start a Wiki operation, index the source, or emit a file-processed event.
operation, err := client.CreateWikiInboxRun(ctx, envai.CreateWikiInboxRunParams{
    WorkspaceID: workspaceID, IdempotencyKey: operationKey,
    Body: envai.CreateWikiInboxRunBody{WebhookURL: "https://your-app.example/wiki"},
})
```

`UploadSourceFile` streams a regular file instead of copying it into memory. The original limit is inclusive: 1,000,000,000 bytes. Keep the file unchanged until the request completes, and retain its original bytes and stable key for explicit replay after reconciling an ambiguous result. `UploadWikiSource` also accepts `MultipartPart{Reader: reader, Size: size}` or small `Data` byte slices. HTTP 202 means the original was acquired durably, not that conversion or indexing succeeded. The source retains `conversion.error_code`, `recoverable`, and `next_retry_at`; only a durably scheduled, eligible automatic recovery is recoverable.

Supported originals are MD/TXT, CSV, XLSX, PDF/DOCX, MP3 and MP4 audio. MP4 does not imply frame analysis. Backend/provider limitations remain explicit errors; the SDK does not promise successful conversion or substitute placeholder text. The application controls polling cadence and its deadline. The compilable `examples/wiki` program shows upload/status/inbox sequencing with offline tests.

`CreateWikiOutboxRun` and `CreateWikiSyncRun` take their corresponding typed bodies and an explicit idempotency key. `GetWikiOperation` returns phase counts, file outcomes, captured batch sizes, diagnostics and current usage. Identical replay returns the existing operation; different payload with the same key conflicts. Queue full is HTTP 429 `wiki_queue_full` with a retry hint; it is not a slot reservation. Mutations and uploads are never retried automatically, including ambiguous transport failures, 429, 5xx or missing acknowledgment.

## Immediate withdrawal and exhaustive catalog reads

```go
withdrawn, err := client.WithdrawWikiSource(ctx, envai.WithdrawWikiSourceParams{
    WorkspaceID: workspaceID, Body: envai.WithdrawWikiSourceBody{FileID: fileID},
})
categories, err := client.ListWikiCategories(ctx, envai.ListWikiCategoriesParams{WorkspaceID: workspaceID})
page, err := client.GetWikiCategoryIndex(ctx, envai.GetWikiCategoryIndexParams{
    WorkspaceID: workspaceID, CategoryID: categoryID,
    From: envai.Present(int64(1)), To: envai.Present(int64(25)),
    Type: envai.Present("raw"), CatalogRevision: envai.Present(categories.CatalogRevision),
})
```

Withdrawal immediately removes RAW visibility and references. It does not start cleanup, alter artifacts or emit an event. Explicit outbox processing performs cleanup/deletion. Delayed conversion or inbox deliveries must not reverse tombstones. `DownloadWikiSource` returns an authenticated streaming body that must be closed; current visibility is rechecked by the server, rather than exposing an S3 URL.

Catalog GET is exhaustive and deterministic, ordered by insertion time and entry ID. Choose `All: Present(true)` or inclusive, one-based `From`/`To`; omit `Type` to include both `raw` and `artifact`. Preserve the catalog revision between pages, and restart on `catalog_revision_changed`. Counts, actual/requested ranges and completeness are typed. No revision grants access to withdrawn RAWs.

Artifact writing/editing, compare-and-swap revisions, reference requests, Decisions QUERY, and inbox/outbox confirmation are actor-bound internal runtime capabilities in V1. They are not public SDK writer endpoints. Existing public artifact download and other inventory operations remain available; generic filesystem APIs do not bypass managed wiki protections.

## Provider accounting is not customer billing

`BillingState.Status` is `inactive`. `UsageCost` preserves decimal-string amounts and original currencies, `status`, `basis`, open/closed scope, revision/as-of and invocation counts. Pending totals remain a nil slice that marshals as JSON null; genuinely empty complete scopes retain a non-nil empty slice that marshals as `[]`. Priced subtotals are only the known portion, not a final total. Do not sum different currencies, invent an FX conversion, or treat an unknown number as zero.

Source conversion consumption belongs to upload. Inbox does not bill it again; sync unites phase invocations. Failed work may have positive cost and successful work may have pending cost. Polling, redelivery and idempotent HTTP replay are not additional provider invocations. Late reconciliation changes current GET cost metadata, not functional outcome or immutable terminal events.

`Completion.CostUSD`, `ATPUsage.CostUSD`, and terminal `Cost`/`ChatCost` are now `*float64`. Nil is unknown; a non-nil pointer to zero is known zero. Use `UsageCost` decimal strings as the authoritative precision-preserving accounting surface. Optional fields retain missing/null/value distinctions through `Optional[T]`. Wiki models preserve unknown fields in `Extra` without floating-point conversion.

Completion GET/list negotiate the same header. `TerminalSnapshot` is optional immutable metadata, bound to its containing completion, and includes event identity/time/type plus typed metadata and optional accepted-version provenance. It does not recover assistant text or deltas. Current cost revisions can advance without modifying this snapshot.

## Verify exact outgoing Wiki bytes

```go
verifier := &envai.WikiWebhookVerifier{Client: client, WorkspaceID: workspaceID}
event, err := verifier.Verify(ctx, "POST", request.Header, exactRawBody)
```

Read at most the configured limit before calling Verify; the default verifier limit is 1 MiB. Never parse/reserialize before verification. Four headers must occur exactly once: `Envai-Webhook-Id`, `Envai-Webhook-Timestamp`, `Envai-Webhook-Key-Id`, `Envai-Webhook-Signature`. The signing input is `envai-webhook-v1`, kid, logical event ID, positive decimal delivery timestamp, and lowercase SHA-256 of the exact raw body, joined with LF and no trailing LF. Ed25519 keys must explicitly use `OKP`, `Ed25519`, `EdDSA`, `sig`; key/signature base64url must be unpadded and canonical.

Only the configured Envai origin's `/.well-known/wiki-webhook-jwks.json` is used. Bearer credentials are not sent to this public route. Keys are cached for 300 seconds, with one bounded refresh for an unknown kid. The default replay tolerance is ±300 seconds around the delivery attempt, not the logical occurrence time. Retries retain the body/event ID and receive fresh signed timestamps.

`VerifiedWikiEvent` distinguishes `FileProcessed`, `FileDeleted` and terminal `Operation`. File events have source revisions and no invented operation cost. Persist source revision/deletion tombstones, and reread current source metadata before downstream work. Out-of-order events must not resurrect withdrawn files. ElevenLabs/provider HMAC callbacks remain a separate protocol and are never accepted as outgoing Wiki signatures. Existing client-tool signature preimages are unchanged.

## Durable receipt and application outbox

The SDK provides a PostgreSQL `database/sql` integration. Register the PostgreSQL driver your application uses, explicitly apply `envai.WikiInboxSchema` under your deployment process, then:

```go
inbox := &envai.WikiWebhookInbox{DB: db, Verifier: verifier}
outcome, err := inbox.ReceiveToOutbox(ctx, "production-envai", request.Method, request.Header, exactRawBody)
if err != nil { http.Error(w, "verification or storage failed", 500); return }
if outcome == envai.WikiReceiptCollision { http.Error(w, "event collision", 409); return }
w.WriteHeader(http.StatusNoContent)
```

Receipts are scoped by connection + workspace + event ID and compare exact payload hashes. Same ID/different bytes is a collision. One transaction inserts the receipt and durable local application job. A crash/failure before commit rolls both back; after commit redelivery is a duplicate. The verified raw body is stored as BYTEA, without reserialization. Constructors never run DDL.

`Receive` alternatively accepts a callback with the same restricted `WikiTransaction` executor to persist your own outbox shape. The callback must only write local transactional work; do not commit/rollback (including SQL transaction-control statements), switch connections, execute nontransactional statements, or perform network/external effects. It must return an error on any failure. Process the outbox separately with your own idempotency policy; atomic receipt does not guarantee exactly-once remote side effects. Only accepted/duplicate outcomes may receive 2xx. Missing tables, verification errors and commit failures must not be acknowledged.

This integration is tested against real PostgreSQL, including concurrent independent processes. It makes no claim of MySQL or SQLite support. Generic in-memory call stores do not become durable Wiki receivers.
