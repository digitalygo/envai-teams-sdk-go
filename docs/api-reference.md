# Go API reference

[README](../README.md) · [Getting started](getting-started.md) · [Streaming](streaming.md) · [Client tools](client-tools.md)

## Calling conventions

Every reviewed SDK operation has a method on `*envai.Client`. Pass a `context.Context` and the operation's named `Params` struct. Request/response types are in [models_generated.go](../models_generated.go); exact required fields, nullability, enums, and validation constraints are in the bundled [OpenAPI contract](../contract/openapi.yaml). IDs are opaque strings. Never derive one ID from another or use a local file path as an artifact ID.

Paths below are relative to your `/api/v1` base URL, except `GetJWKS`, which uses the configured origin-level JWKS URL. Bearer authentication is omitted for public operations and signed artifact downloads. These are primitive operations, not a recommendation to use provider webhooks or image-test endpoints in an ordinary application.

| Method | Params | Returns | HTTP path | Bearer |
| --- | --- | --- | --- | --- |
| `DeliverWorkflowResult` | `DeliverWorkflowResultParams` | `(NativeDeliveryReceipt, error)` | POST `/atp/v3/chats/{chat_id}/deliveries` | Yes |
| `DelegateToTeam` | `DelegateToTeamParams` | `(TeamDelegationReceipt, error)` | POST `/atp/v3/chats/{chat_id}/delegations` | Yes |
| `GetDetachedTask` | `GetDetachedTaskParams` | `(DetachedTaskSnapshot, error)` | GET `/atp/v3/chats/{chat_id}/tasks/{task_id}` | Yes |
| `CancelDetachedTask` | `CancelDetachedTaskParams` | `(CancelDetachedTaskResult, error)` | POST `/atp/v3/chats/{chat_id}/tasks/{task_id}/cancel` | Yes |
| `ClassifyPhase` | `ClassifyPhaseParams` | `(PhaseClassificationResult, error)` | POST `/atp/v3/chats/{chat_id}/classify-phase` | Yes |
| `ClassifyChoice` | `ClassifyChoiceParams` | `(ChoiceClassificationResult, error)` | POST `/atp/v3/chats/{chat_id}/classify-choice` | Yes |
| `StartIndependentJob` | `StartIndependentJobParams` | `(IndependentJobReceipt, error)` | POST `/atp/v3/chats/{chat_id}/jobs` | Yes |
| `SendChatMessage` | `SendChatMessageParams` | `(ChatMessageAccepted, error)` | POST `/atp/v3/chats/{chat_id}/messages` | Yes |
| `StreamChatEvents` | `StreamChatEventsParams` | `(*EventStream, error)` | GET `/atp/v3/chats/{chat_id}/events` | Yes |
| `CancelChatTurn` | `CancelChatTurnParams` | `(TurnAccepted, error)` | POST `/atp/v3/chats/{chat_id}/cancel` | Yes |
| `SteerChatTurn` | `SteerChatTurnParams` | `(TurnAccepted, error)` | POST `/atp/v3/chats/{chat_id}/steer` | Yes |
| `ResolveChatTool` | `ResolveChatToolParams` | `error` | POST `/atp/v3/chats/{chat_id}/tool-output` | Yes |
| `ResolveChatApproval` | `ResolveChatApprovalParams` | `error` | POST `/atp/v3/chats/{chat_id}/approval` | Yes |
| `ForkChat` | `ForkChatParams` | `(ChatForked, error)` | POST `/atp/v3/chats/{chat_id}/fork` | Yes |
| `RestoreChat` | `RestoreChatParams` | `(RestoreChatAccepted, error)` | POST `/atp/v3/chats/{chat_id}/restore` | Yes |
| `GetCapabilities` | `GetCapabilitiesParams` | `(Capabilities, error)` | GET `/atp/v3/capabilities` | Yes |
| `HealthCheck` | `HealthCheckParams` | `(HealthCheckResult, error)` | GET `/up` | No |
| `GetJWKS` | `GetJWKSParams` | `(JWKSDocument, error)` | GET `/.well-known/jwks.json` | No |
| `CreateWorkspace` | `CreateWorkspaceParams` | `(Workspace, error)` | POST `/workspaces` | Yes |
| `ListWorkspaces` | `ListWorkspacesParams` | `([]Workspace, error)` | GET `/workspaces` | Yes |
| `GetWorkspace` | `GetWorkspaceParams` | `(Workspace, error)` | GET `/workspaces/{workspace_id}` | Yes |
| `UpdateWorkspace` | `UpdateWorkspaceParams` | `(Workspace, error)` | PUT `/workspaces/{workspace_id}` | Yes |
| `DeleteWorkspace` | `DeleteWorkspaceParams` | `(CleanupResult, error)` | DELETE `/workspaces/{workspace_id}` | Yes |
| `CreateProjectTool` | `CreateProjectToolParams` | `(ProjectTool, error)` | POST `/tools` | Yes |
| `ListProjectTools` | `ListProjectToolsParams` | `([]ProjectTool, error)` | GET `/tools` | Yes |
| `GetProjectTool` | `GetProjectToolParams` | `(ProjectTool, error)` | GET `/tools/{name}` | Yes |
| `UpdateProjectTool` | `UpdateProjectToolParams` | `(ProjectTool, error)` | PUT `/tools/{name}` | Yes |
| `DeleteProjectTool` | `DeleteProjectToolParams` | `error` | DELETE `/tools/{name}` | Yes |
| `ListProjectSkills` | `ListProjectSkillsParams` | `([]ProjectSkill, error)` | GET `/projects/{project_id}/skills` | Yes |
| `CreateProjectSkill` | `CreateProjectSkillParams` | `(ProjectSkill, error)` | POST `/projects/{project_id}/skills` | Yes |
| `GetProjectSkill` | `GetProjectSkillParams` | `(ProjectSkill, error)` | GET `/projects/{project_id}/skills/{skill_id}` | Yes |
| `UpdateProjectSkill` | `UpdateProjectSkillParams` | `(ProjectSkill, error)` | PUT `/projects/{project_id}/skills/{skill_id}` | Yes |
| `DeleteProjectSkill` | `DeleteProjectSkillParams` | `error` | DELETE `/projects/{project_id}/skills/{skill_id}` | Yes |
| `CreateChat` | `CreateChatParams` | `(Chat, error)` | POST `/workspaces/{workspace_id}/chats` | Yes |
| `ListChats` | `ListChatsParams` | `([]Chat, error)` | GET `/workspaces/{workspace_id}/chats` | Yes |
| `GetChat` | `GetChatParams` | `(Chat, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}` | Yes |
| `UpdateChat` | `UpdateChatParams` | `(Chat, error)` | PUT `/workspaces/{workspace_id}/chats/{chat_id}` | Yes |
| `DeleteChat` | `DeleteChatParams` | `(CleanupResult, error)` | DELETE `/workspaces/{workspace_id}/chats/{chat_id}` | Yes |
| `ListCompletions` | `ListCompletionsParams` | `(CompletionCollection, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}/completions` | Yes |
| `GetCompletion` | `GetCompletionParams` | `(Completion, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}/completions/{completion_id}` | Yes |
| `ListChatCrons` | `ListChatCronsParams` | `([]ChatCron, error)` | GET `/workspaces/{workspace_id}/crons` | Yes |
| `GetChatCron` | `GetChatCronParams` | `(ChatCron, error)` | GET `/workspaces/{workspace_id}/crons/{cron_id}` | Yes |
| `UpdateChatCron` | `UpdateChatCronParams` | `(ChatCron, error)` | PATCH `/workspaces/{workspace_id}/crons/{cron_id}` | Yes |
| `DeleteChatCron` | `DeleteChatCronParams` | `error` | DELETE `/workspaces/{workspace_id}/crons/{cron_id}` | Yes |
| `IndexWorkspaceFiles` | `IndexWorkspaceFilesParams` | `(WorkspaceFile, error)` | GET `/workspaces/{workspace_id}/files` | Yes |
| `CreateWorkspaceFile` | `CreateWorkspaceFileParams` | `(WorkspaceFile, error)` | POST `/workspaces/{workspace_id}/files` | Yes |
| `UpdateWorkspaceFile` | `UpdateWorkspaceFileParams` | `(WorkspaceFile, error)` | PUT `/workspaces/{workspace_id}/files` | Yes |
| `DeleteWorkspaceFile` | `DeleteWorkspaceFileParams` | `error` | DELETE `/workspaces/{workspace_id}/files` | Yes |
| `DownloadWorkspaceFile` | `DownloadWorkspaceFileParams` | `(*Download, error)` | GET `/workspaces/{workspace_id}/files/download` | Yes |
| `ConvertWorkspaceFileToPDF` | `ConvertWorkspaceFileToPDFParams` | `(*Download, error)` | POST `/workspaces/{workspace_id}/files/convert-to-pdf` | Yes |
| `CreateAgent` | `CreateAgentParams` | `(Agent, error)` | POST `/agents` | Yes |
| `ListAgents` | `ListAgentsParams` | `([]Agent, error)` | GET `/agents` | Yes |
| `GetAgent` | `GetAgentParams` | `(Agent, error)` | GET `/agents/{agent_id}` | Yes |
| `UpdateAgent` | `UpdateAgentParams` | `(Agent, error)` | PUT `/agents/{agent_id}` | Yes |
| `DeleteAgent` | `DeleteAgentParams` | `error` | DELETE `/agents/{agent_id}` | Yes |
| `CreateTeam` | `CreateTeamParams` | `(Team, error)` | POST `/teams` | Yes |
| `ListTeams` | `ListTeamsParams` | `([]Team, error)` | GET `/teams` | Yes |
| `GetTeam` | `GetTeamParams` | `(Team, error)` | GET `/teams/{team_id}` | Yes |
| `UpdateTeam` | `UpdateTeamParams` | `(Team, error)` | PUT `/teams/{team_id}` | Yes |
| `DeleteTeam` | `DeleteTeamParams` | `error` | DELETE `/teams/{team_id}` | Yes |
| `ListModels` | `ListModelsParams` | `(ModelList, error)` | GET `/models` | Yes |
| `CreateAPIKey` | `CreateAPIKeyParams` | `(json.RawMessage, error)` | POST `/api-keys` | Unsupported: human session + CSRF |
| `ListAPIKeys` | `ListAPIKeysParams` | `([]APIKey, error)` | GET `/api-keys` | Unsupported: human session + CSRF |
| `DeleteAPIKey` | `DeleteAPIKeyParams` | `error` | DELETE `/api-keys/{key_id}` | Unsupported: human session + CSRF |
| `TestGenerateImage` | `TestGenerateImageParams` | `(TestGenerateImageResult, error)` | POST `/tests/images/generate` | No |
| `CreateImageGeneration` | `CreateImageGenerationParams` | `(ImageGenerationAcceptedRecord, error)` | POST `/images/generations` | Yes |
| `GetImageGeneration` | `GetImageGenerationParams` | `(ImageGenerationStatusRecord, error)` | GET `/images/generations/{record}` | Yes |
| `UploadImage` | `UploadImageParams` | `(ImageUploadRecord, error)` | POST `/images/uploads` | Yes |
| `ListImageModels` | `ListImageModelsParams` | `([]ModelCatalogEntry, error)` | GET `/images/models` | Yes |
| `DownloadArtifact` | `DownloadArtifactParams` | `(*Download, error)` | GET `/artifacts/download` | No |
| `HandleProviderWebhook` | `HandleProviderWebhookParams` | `(WebhookOk, error)` | POST `/webhooks/providers/{provider}` | No |
| `UploadChatAttachment` | `UploadChatAttachmentParams` | `(ChatAttachment, error)` | POST `/workspaces/{workspace_id}/chats/{chat_id}/attachments` | Yes |
| `ListChatAttachments` | `ListChatAttachmentsParams` | `(ListChatAttachmentsResult, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}/attachments` | Yes |
| `GetChatAttachment` | `GetChatAttachmentParams` | `(ChatAttachment, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}/attachments/{attachment_id}` | Yes |
| `DeleteChatAttachment` | `DeleteChatAttachmentParams` | `(ChatAttachment, error)` | DELETE `/workspaces/{workspace_id}/chats/{chat_id}/attachments/{attachment_id}` | Yes |
| `RetryChatAttachment` | `RetryChatAttachmentParams` | `(ChatAttachment, error)` | POST `/workspaces/{workspace_id}/chats/{chat_id}/attachments/{attachment_id}/retry` | Yes |
| `DownloadChatAttachment` | `DownloadChatAttachmentParams` | `(*Download, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}/attachments/{attachment_id}/download` | Yes |
| `UploadWikiSource` | `UploadWikiSourceParams` | `(WikiSource, error)` | POST `/workspaces/{workspace_id}/files/sources` | Yes |
| `GetWikiSource` | `GetWikiSourceParams` | `(WikiSource, error)` | GET `/workspaces/{workspace_id}/files/sources/{file_id}` | Yes |
| `DownloadWikiSource` | `DownloadWikiSourceParams` | `(*Download, error)` | GET `/workspaces/{workspace_id}/files/sources/{file_id}/download` | Yes |
| `WithdrawWikiSource` | `WithdrawWikiSourceParams` | `(WikiSource, error)` | POST `/workspaces/{workspace_id}/wiki/outbox/files` | Yes |
| `CreateWikiInboxRun` | `CreateWikiInboxRunParams` | `(WikiOperation, error)` | POST `/workspaces/{workspace_id}/wiki/inbox/runs` | Yes |
| `CreateWikiOutboxRun` | `CreateWikiOutboxRunParams` | `(WikiOperation, error)` | POST `/workspaces/{workspace_id}/wiki/outbox/runs` | Yes |
| `CreateWikiSyncRun` | `CreateWikiSyncRunParams` | `(WikiOperation, error)` | POST `/workspaces/{workspace_id}/wiki/syncs` | Yes |
| `GetWikiOperation` | `GetWikiOperationParams` | `(WikiOperation, error)` | GET `/workspaces/{workspace_id}/wiki/operations/{operation_id}` | Yes |
| `ListWikiCategories` | `ListWikiCategoriesParams` | `(Categories, error)` | GET `/workspaces/{workspace_id}/wiki/categories` | Yes |
| `GetWikiCategoryIndex` | `GetWikiCategoryIndexParams` | `(CategoryIndex, error)` | GET `/workspaces/{workspace_id}/wiki/categories/{category_id}/index` | Yes |
| `GetWikiWebhookJWKS` | `GetWikiWebhookJWKSParams` | `(JWKSDocument, error)` | GET `/.well-known/wiki-webhook-jwks.json` | No |
| `StreamChatEventsLegacy` | `StreamChatEventsLegacyParams` | `(*EventStream, error)` | GET `/chats/{chat_id}/events` | Yes |
| `GetCapabilitiesLegacy` | `GetCapabilitiesLegacyParams` | `(CapabilitiesLegacy, error)` | GET `/capabilities` | Yes |
| `GetProjectUsage` | `GetProjectUsageParams` | `(UsageReportCollection, error)` | GET `/projects/{project_id}/usage` | Yes |
| `GetProjectLogs` | `GetProjectLogsParams` | `(json.RawMessage, error)` | GET `/projects/{project_id}/logs` | Yes |

The 14 rows for `DeleteWorkspace`, `UpdateChat`, `DeleteChat`, the four chat cron methods, the five project skill methods, `GetProjectUsage` and `GetProjectLogs` are unreleased on `feature/user-company-projects`; see [Project resources, deletion and observability](#project-resources-deletion-and-observability). Native background task methods (`DeliverWorkflowResult`, `DelegateToTeam`, `GetDetachedTask`, `CancelDetachedTask`, `ClassifyPhase`, `ClassifyChoice`, `StartIndependentJob`) and chat attachment methods are also unreleased surfaces, documented in [native background tasks](background-tasks.md), [independent jobs](independent-jobs.md), and [chat attachments](chat-attachments.md). Typed helpers that wrap `GetProjectLogs` and numeric page iterators are listed under project resources, outside this operation table.

## Models, optional fields, and collections

- `envai.Present(value)` includes an optional field. The zero-value `Optional[T]` omits it. `envai.Null[T]()` explicitly sends JSON null; use only when the field's schema allows null.
- Resource `data` wrappers are unwrapped into typed resources. Completion, project usage and project log pages retain `Data`, `Links`, and `Meta`. Use `CompletionPages`, `ProjectUsagePages` or `ProjectLogPages` for numeric traversal on the configured origin.
- API key creation distinguishes omitted expiry from explicit `"never"`. Treat the plaintext key in the creation response as a secret.
- Flexible metadata, schema, and union values use `json.RawMessage` or typed JSON maps. Unknown REST response properties are ignored; known fields keep their declared Go types.
- HTTP 204 methods return only an error. `Download` contains `Body`, `Header`, and `Status`; always call `Close`, including after partial reads. Reading a download is your responsibility and must remain within the request's context deadline.
- Multipart methods take `[]envai.MultipartPart`. You supply `Data`, `Name`, `Filename`, and `ContentType`. The SDK does not open local files or execute file contents.
- `DownloadArtifact` uses server-provided signed fields. Do not construct a signature, replace opaque IDs, or add bearer credentials to the signed URL.

## Project resources, deletion and observability

Unreleased on `feature/user-company-projects`; the published `v1.0.0-alpha.1` does not include these methods, and this page makes no claim about a future version. The canonical source is `digitalygo/envai-teams` at `7f4085dbfdd3f208b82716119cff1860fbd502bb` (146 server operations, 92 in the reviewed SDK subset, 89 callable, 3 API-key stubs). Project and workspace IDs are opaque; pass the values the server returned.

`X-Envai-Project-Id` is exposed as `XEnvaiProjectID Optional[string]` on cron, usage and log params. A project API key derives its project from the credential; a header that differs from it is rejected by the server with 403.

### Chats, workspaces and cleanup

| Method | Body or filters | Result and errors |
| --- | --- | --- |
| `UpdateChat` | `ChatUpdateRequest{Title Optional[string]}`; `envai.Null[string]()` clears the title. | `Chat`. Only `title` is accepted; `team_id`, `workspace_id`, `project_id` and ownership fields give 422. |
| `DeleteChat` | None. | `CleanupResult`. 409 when deletion is refused, such as active turns (`turn_in_progress`) or active cron schedules/occurrences. |
| `DeleteWorkspace` | None. | `CleanupResult`. 409 under matching conflict rules across its chats. |

`CleanupResult` carries `ResourceType` (`workspace`, `chat`, `project` or `company`), `ResourceID` and `Cleanup []CleanupIntentState`. Each intent has `ID`, `Kind`, `Status` (`pending`, `processing`, `completed` or `failed`), `Attempts`, nullable `DeadlineAt`, `StartedAt`, `FinishedAt` and `LastErrorCode`.

- A 200 means access was cut and an intent was created. `pending` is **not** confirmation of physical purge: content removal continues server-side, and the intent is closed only once removal is confirmed. No reviewed operation reads an intent back, so do not treat the 200 body as a purge receipt.
- Completion metadata, usage and cost accounting are retained after chat deletion; conversation content is not.
- The SDK never automatically cancels turns or modifies crons to make deletion succeed; it just surfaces HTTP 409 as `*APIError`. A 409 conflict can arise from an active turn (`turn_in_progress`) or active cron schedules and occurrences. Do not treat destructive turn cancellation as a generic recipe: true current backend active cron protection can persist even after turn cancellation, and cancelling a turn does not promise or ensure removal of the deletion guard. Inspect `apiErr.Code` and current cron, occurrence, and turn state. Where appropriate, resolve active dependencies through explicit operator actions (such as removing active chat crons via `DeleteChatCron`, waiting for running work, or cancelling a turn explicitly) before deciding whether to retry deletion.

```go
result, err := client.DeleteChat(ctx, envai.DeleteChatParams{
    WorkspaceID: workspaceID,
    ChatID:      chatID,
})
var apiErr *envai.APIError
switch {
case errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict:
    return fmt.Errorf("chat conflict (%s): %w", apiErr.Code, err)
case err != nil:
    return err
}
for _, intent := range result.Cleanup {
    log.Printf("cleanup %s %s attempts=%d", intent.ID, intent.Status, intent.Attempts)
}
```

### Chat crons

`ListChatCrons` filters by `OriginChatID` and `Status` (`active` or `completed`) and returns non-cancelled schedules. `GetChatCron` and `DeleteChatCron` take `CronID`. There is no `CreateChatCron`: creation is limited to the native cron tool during an active project-key turn, and a human session cannot start paid work.

`UpdateChatCron` takes `ChatCronUpdateRequest`, a partial update of an active schedule: `TargetMode` (`same_chat` or `new_chat`), `Prompt` (at most 16000 characters, not blank), `IntervalMinutes`, `DailyAt` (`HH:MM`), `Timezone`, `StartAt`, `RepeatCount` and `Until`. All fields are `Optional`; the schedule fields accept `envai.Null[T]()`. Only occurrences not yet materialized follow the new settings. Origin chat, team, project, key and turn attribution, status, counters and timestamps are immutable and give 422 if sent. `RepeatCount` cannot drop below `ScheduledCount`, and updating a schedule that is not active returns 409.

`DeleteChatCron` returns only an error (HTTP 204). It cancels the schedule and fences occurrences not yet admitted; a turn that was already admitted is not cancelled.

### Project skills

`ProjectSkill` has `ID`, `ProjectID`, nullable `CreatedByUserID`, `Name`, nullable `Description`, `Content` and timestamps. `ListProjectSkills` orders by name and ID; a skill from another project answers 404.

`ProjectSkillCreateRequest` and `ProjectSkillUpdateRequest` are `map[string]json.RawMessage`, **not** structs. Each value is raw JSON, so build samples with `json.RawMessage`. An omitted key is not sent; `json.RawMessage("null")` is meaningful only for `description`.

```go
created, err := client.CreateProjectSkill(ctx, envai.CreateProjectSkillParams{
    ProjectID: projectID,
    Body: envai.ProjectSkillCreateRequest{
        "name":        json.RawMessage(`"incident_response"`),
        "description": json.RawMessage(`"Safe incident handling"`),
        "content":     json.RawMessage(`"# Incident response\n\nEscalate after triage.\n"`),
    },
})
if err != nil { return err }

// Clear only the description; name and content are untouched.
_, err = client.UpdateProjectSkill(ctx, envai.UpdateProjectSkillParams{
    ProjectID: projectID,
    SkillID:   created.ID,
    Body: envai.ProjectSkillUpdateRequest{
        "description": json.RawMessage("null"),
    },
})
```

Constraints from the contract: `name` and `content` are required on create. `name` is 1 to 64 characters matching `^[a-z][a-z0-9_-]{0,63}$` and cannot be one of the reserved built-in names. `content` is 1 to 65535 characters of Markdown without NUL bytes, and whitespace is preserved exactly. Ownership and audit fields (`id`, `project_id`, `created_by_user_id`, `company_id`, `user_id`, `workspace_id`, `created_at`, `updated_at`) are prohibited. With `Config.StrictValidation` the SDK also rejects any key outside `name`, `description` and `content` locally, before HTTP. `DeleteProjectSkill` hard deletes the skill and returns only an error (HTTP 204).

### Usage and logs

`GetProjectUsage` accepts `From` and `To` (UTC dates `YYYY-MM-DD`, window at most 90 days, default 30), `APIKeyID`, `Page` and `PerPage` (1 to 100). It returns a `UsageReportCollection` with one `UsageReportDay` per UTC day, plus `Links` and `Meta`. Amounts are exact decimal strings per currency; unknown-cost invocations are counted separately. Never sum currencies or read unknown cost as zero.

`GetProjectLogs` is the low-level method and returns the response body as `json.RawMessage`, because the shape depends on `View`. Prefer the typed helpers, each of which sets `View` itself and fails locally if you pass a different one:

| Helper | View | Returns |
| --- | --- | --- |
| `GetProjectLogsEvents` | `events` | `ProjectLogCollection` (`Data`, `Links`, `Meta` preserved) |
| `GetProjectLogsAnalytics` | `analytics` | `ProjectActivityAnalytics`, unwrapped from `data` |
| `GetProjectLogsFilterCatalog` | `filters` | `ProjectEventTypeFilters`, unwrapped from `data` |

Filters: `ChatID`, `APIKeyID`, `Type`, and the list filters `Types`, `WorkspaceIDs` and `APIKeyIDs`, each `Optional[[]string]` sent as repeated `types[]=`, `workspace_ids[]=` and `api_key_ids[]=` entries (1 to 50 distinct values; an empty or nil list fails locally). Scalar `Type` cannot be combined with `Types`, nor `APIKeyID` with `APIKeyIDs`. `analytics` rejects `Type`, every list filter, `Page` and `PerPage`; `filters` additionally rejects `APIKeyID` and `ChatID`. Analytics and filters are limited to 90 UTC days. Events are metadata only, retained about 30 days by default and best-effort, so an empty day is not proof of no activity. Empty `agent` and `data` metadata returned as `[]` are normalized to `{}`.

```go
analytics, err := client.GetProjectLogsAnalytics(ctx, envai.GetProjectLogsParams{
    ProjectID: projectID,
    From:      envai.Present("2026-11-01"),
    To:        envai.Present("2026-11-07"),
    ChatID:    envai.Present(chatID),
})

for page, err := range client.ProjectLogPages(ctx, envai.GetProjectLogsParams{
    ProjectID: projectID,
    Types:     envai.Present([]string{"turn.started", "turn.completed"}),
    PerPage:   envai.Present(int64(50)),
}) {
    if err != nil { return err }
    for _, entry := range page.Data {
        _ = entry.Type
    }
    // break stops fetching further pages.
}
```

### Numeric page iterators

`ProjectUsagePages(ctx, GetProjectUsageParams)` and `ProjectLogPages(ctx, GetProjectLogsParams)` return `iter.Seq2` values like `CompletionPages`. They start at `Page` (default 1), request each page by number through the same client and origin, and stop after `Meta.LastPage`. `ProjectLogPages` always uses the `events` view and the typed helper, so it yields `ProjectLogCollection`. Local origin only: the iterators never follow a server-supplied `next` URL. They check the context before each request, stop on `break`, yield a `ProtocolError` once for a non-positive start page or inconsistent `current_page`/`last_page` metadata, and yield transport errors once.

### Errors and retries

Every refusal is a generic `*APIError` with `Status`, `Type`, `Code`, `Message`, `Details`, `RequestID` and `RetryAfter`. The SDK does not branch on or rewrite codes. A 402 billing refusal or a 503 is not retried for any mutation. The contract now declares a 503 `provider_configuration_error` (type and code) for `CreateImageGeneration` when a provider in the routing chain is not configured. Treat it as an operator configuration failure, not a transient fault.

### Unchanged boundaries

- No new ATP event types: the 39 typed events are unchanged. `chat.cron.wakeup`, `chat.title.updated`, `billing.limit_reached` and `billing.credit_exhausted` were already typed. Billing events may have a null `turn_id` before admission or a real `turn_id` midturn; they are never terminal. A null cost on a terminal remains unknown, not zero.
- The v2 historical contract stays frozen, and native task behavior remains opt-in through `Config.NativeTasks`.
- The SDK has no human administration or financial operations. `CreateAPIKey`, `ListAPIKeys` and `DeleteAPIKey` remain local stubs that return `ErrAPIKeyManagementUnsupported`.

## Configuration and errors

See [configuration](getting-started.md#configuration) for defaults and [error handling](streaming.md#error-handling) for typed errors. Request methods never automatically retry mutations. Generated methods share the same transport and credential handling.

## Event payloads

`DecodeEvent` validates ATP v3 and immutable numeric v2 envelopes before hydration. `Event.TypedPayload()` returns the corresponding concrete payload; `Event.Raw` retains original JSON. All 39 declared types are listed in [events_generated.go](../events_generated.go) and described by [events.json](../contract/events.json). Valid future event types remain available as raw data. Capabilities discovery does not restrict which valid event types can be decoded.

## Wiki and negotiated contracts

See [Wiki lifecycle and durable webhook intake](wiki.md) for all source, operation, catalog and JWKS methods, and [contract 3 migration](contract3-migration.md) for default selection, strict legacy reads, nullable costs and compatibility errors.

## Project tenancy and authentication

Workspaces, chats, agents and teams replace UserID ownership with ProjectID and nullable CreatedByUserID audit provenance. Completion adds ProjectID and APIKeyID, and UserID is now a nullable *string. Use project ownership checks, never the key creator. Workspace metadata is a json.RawMessage alias supporting objects and lists, including the server default [].

Tool registration is project-wide, not user-wide. Use ProjectTool methods and models. Deprecated UserTool aliases forward to project-scoped operations; they do not retain user-wide visibility. A project key selects its project, and any supplied project header must match it. Definition normalization ignores server-owned identity and audit metadata.

The byte-exact server snapshot contains 146 operations. The explicit scripts/sdk-operations.json allowlist retains a reviewed 92-operation subset, not entire-server parity: 89 callable operations plus 3 discoverable API-key administration stubs. API-key list/create/delete fail locally before HTTP or capability discovery with ErrAPIKeyManagementUnsupported (errors.Is supported). The server requires a human session and CSRF for /api/v1/api-keys; use its authenticated application flow. This client never bypasses that boundary. Human financial and administration operations are not generated.

Explicit historical legacy completion reads retain their frozen numeric/owner contract; current strict response validation still requires project ownership fields.
