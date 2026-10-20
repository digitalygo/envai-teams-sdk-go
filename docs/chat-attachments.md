# Chat attachments

This unreleased feature requires a matching backend advertising `chat_attachments` in contract-3 capabilities. Native uploads additionally require `chat_attachment_originals`. Other attachment operations and messages containing attachment IDs require only `chat_attachments`. The SDK checks capabilities once before the mutation and never falls back to sending the text alone.

Capabilities may advertise enabled names as a canonical homogeneous string list or as a strict string-to-boolean map. Required names must be present and enabled. A missing, false, unknown or malformed entry fails before the multipart request. `Capabilities.Features` remains the public `[]string` of enabled names, so a boolean map is normalized to that list. Unrelated advertised names never substitute for a required name.

Upload one immutable original with an upload idempotency key, wait for its ready state, then send ordered attachment IDs with a separate stable message idempotency key. A native original is ready when `lifecycle` is `ready` and `conversion.status` is `not_required`; an uploaded native record keeps polling. Legacy converted records still finish at `completed` or a terminal `failed` state. A message acknowledgment must contain those same IDs. Pending or failed selected files block the whole turn; remove that file before sending. Poll timeouts cancel only the local wait. Mutations are never automatically retried, even with a key: reconcile an uncertain response before explicitly replaying the same payload/key. Native retries answer HTTP 409 `attachment_retry_not_permitted`.

Native chat input accepts MD, TXT, CSV, PDF, MP3 and MP4. No conversion, OCR or transcription is scheduled for a native upload. Legacy converted records also accept XLSX and DOCX through their existing converters. MP4 visual frames are not analyzed. Originals reach compatible models as private content parts. There is no automatic Wiki indexing, guaranteed Office editing/export, or automatic mounting of original files in tools.

Legacy conversions retain the complete textual projection and must fit the server and model context budgets; oversized input fails explicitly and the converted attachment stays available. No silent truncation occurs. Originals and derivatives remain private and scoped to the owning user, workspace and chat. Forked/restored text does not grant access to originals. Status/list/download/retry/delete are separate operations. Download selects `kind=original` or `kind=derived`.

Draft retention defaults to 24 hours; accepted attachments and input snapshots default to 14 days. Idempotent replay is guaranteed while the retained input snapshot exists. Withdrawal blocks new reads immediately; physical cleanup may wait for live conversion leases. Deletion does not erase already accepted conversation content.

Conversion cost is exposed in the attachment's `usage_cost`; inference is attributed to the turn. Missing tariffs and uncertain outcomes stay pending with null totals. An explicit conversion retry creates a linked accounting cycle, preserving previous cost evidence. Wiki file outcomes may now include `deferred` for infrastructure failures; applications should display that outcome without describing the original document as defective.

```go
uploaded, err := client.UploadAttachmentFile(ctx, workspaceID, chatID, "invoice.pdf", "upload-invoice-1", "application/pdf")
if err != nil { return err }
ready, err := client.WaitForAttachment(ctx, workspaceID, chatID, uploaded.AttachmentID, time.Second, 10*time.Minute)
if err != nil { return err }
if ready.Lifecycle != "ready" || (ready.Conversion.Status != "completed" && ready.Conversion.Status != "not_required") { return errors.New("resolve or remove the failed attachment before sending") }
result, err := client.RunTurn(ctx, chatID, "Summarize the invoice.", envai.TurnOptions{AttachmentIDs: []string{ready.AttachmentID}, IdempotencyKey: "message-invoice-1"})
```

Low-level named methods include `UploadChatAttachment`, `ListChatAttachments`, `GetChatAttachment`, `RetryChatAttachment`, `DeleteChatAttachment`, and `DownloadChatAttachment`. `SendChatMessageParams.IdempotencyKey` and `ChatMessageRequest.AttachmentIDs` use `envai.Present(...)`. Context cancellation and polling deadlines are propagated. Files stream from disk without buffering the full upload.

See the matching backend [attachment guide](https://github.com/digitalygo/envai-teams/blob/7192fc05ebec817bddc705186c8f195cc0dc5ec5/docs/guides/chat-attachments.md) for limits, operator settings, and retention. Publishing this SDK does not deploy the backend.

With a stable message key, turn helpers retain the initial replay until the acknowledgment identifies the accepted turn. A previously completed turn can therefore finish from retained events without another inference call. The pre-send replay buffer is limited to 16 MiB; exceeding it fails before submitting the message. Resume a known turn with its event cursor when the retained stream is larger or a replay gap is reported.
