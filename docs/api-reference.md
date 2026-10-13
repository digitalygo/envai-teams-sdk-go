# Go API reference

[README](../README.md) · [Getting started](getting-started.md) · [Streaming](streaming.md) · [Client tools](client-tools.md)

## Calling conventions

Every HTTP operation has a method on `*envai.Client`. Pass a `context.Context` and the operation's named `Params` struct. Request/response types are in [models_generated.go](../models_generated.go); exact required fields, nullability, enums, and validation constraints are in the bundled [OpenAPI contract](../contract/openapi.yaml). IDs are opaque strings. Never derive one ID from another or use a local file path as an artifact ID.

Paths below are relative to your `/api/v1` base URL, except `GetJWKS`, which uses the configured origin-level JWKS URL. Bearer authentication is omitted for public operations and signed artifact downloads. These are primitive operations, not a recommendation to use provider webhooks or image-test endpoints in an ordinary application.

| Method | Params | Returns | HTTP path | Bearer |
| --- | --- | --- | --- | --- |
| `SendChatMessage` | `SendChatMessageParams` | `(ChatMessageAccepted, error)` | POST `/chats/{chat_id}/messages` | Yes |
| `StreamChatEvents` | `StreamChatEventsParams` | `(*EventStream, error)` | GET `/chats/{chat_id}/events` | Yes |
| `CancelChatTurn` | `CancelChatTurnParams` | `(TurnAccepted, error)` | POST `/chats/{chat_id}/cancel` | Yes |
| `SteerChatTurn` | `SteerChatTurnParams` | `(TurnAccepted, error)` | POST `/chats/{chat_id}/steer` | Yes |
| `ResolveChatTool` | `ResolveChatToolParams` | `error` | POST `/chats/{chat_id}/tool-output` | Yes |
| `ResolveChatApproval` | `ResolveChatApprovalParams` | `error` | POST `/chats/{chat_id}/approval` | Yes |
| `ForkChat` | `ForkChatParams` | `(ChatForked, error)` | POST `/chats/{chat_id}/fork` | Yes |
| `RestoreChat` | `RestoreChatParams` | `(RestoreChatAccepted, error)` | POST `/chats/{chat_id}/restore` | Yes |
| `GetCapabilities` | `GetCapabilitiesParams` | `(Capabilities, error)` | GET `/capabilities` | Yes |
| `HealthCheck` | `HealthCheckParams` | `(HealthCheckResult, error)` | GET `/up` | No |
| `GetJWKS` | `GetJWKSParams` | `(JWKSDocument, error)` | GET `/.well-known/jwks.json` | No |
| `CreateWorkspace` | `CreateWorkspaceParams` | `(Workspace, error)` | POST `/workspaces` | Yes |
| `ListWorkspaces` | `ListWorkspacesParams` | `([]Workspace, error)` | GET `/workspaces` | Yes |
| `GetWorkspace` | `GetWorkspaceParams` | `(Workspace, error)` | GET `/workspaces/{workspace_id}` | Yes |
| `UpdateWorkspace` | `UpdateWorkspaceParams` | `(Workspace, error)` | PUT `/workspaces/{workspace_id}` | Yes |
| `CreateUserTool` | `CreateUserToolParams` | `(UserTool, error)` | POST `/tools` | Yes |
| `ListUserTools` | `ListUserToolsParams` | `([]UserTool, error)` | GET `/tools` | Yes |
| `GetUserTool` | `GetUserToolParams` | `(UserTool, error)` | GET `/tools/{name}` | Yes |
| `UpdateUserTool` | `UpdateUserToolParams` | `(UserTool, error)` | PUT `/tools/{name}` | Yes |
| `DeleteUserTool` | `DeleteUserToolParams` | `error` | DELETE `/tools/{name}` | Yes |
| `CreateChat` | `CreateChatParams` | `(Chat, error)` | POST `/workspaces/{workspace_id}/chats` | Yes |
| `ListChats` | `ListChatsParams` | `([]Chat, error)` | GET `/workspaces/{workspace_id}/chats` | Yes |
| `GetChat` | `GetChatParams` | `(Chat, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}` | Yes |
| `ListCompletions` | `ListCompletionsParams` | `(CompletionCollection, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}/completions` | Yes |
| `GetCompletion` | `GetCompletionParams` | `(Completion, error)` | GET `/workspaces/{workspace_id}/chats/{chat_id}/completions/{completion_id}` | Yes |
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
| `CreateAPIKey` | `CreateAPIKeyParams` | `(json.RawMessage, error)` | POST `/api-keys` | Yes |
| `ListAPIKeys` | `ListAPIKeysParams` | `([]APIKey, error)` | GET `/api-keys` | Yes |
| `DeleteAPIKey` | `DeleteAPIKeyParams` | `error` | DELETE `/api-keys/{key_id}` | Yes |
| `TestGenerateImage` | `TestGenerateImageParams` | `(TestGenerateImageResult, error)` | POST `/tests/images/generate` | No |
| `CreateImageGeneration` | `CreateImageGenerationParams` | `(ImageGenerationAcceptedRecord, error)` | POST `/images/generations` | Yes |
| `GetImageGeneration` | `GetImageGenerationParams` | `(ImageGenerationStatusRecord, error)` | GET `/images/generations/{record}` | Yes |
| `UploadImage` | `UploadImageParams` | `(ImageUploadRecord, error)` | POST `/images/uploads` | Yes |
| `ListImageModels` | `ListImageModelsParams` | `([]ModelCatalogEntry, error)` | GET `/images/models` | Yes |
| `DownloadArtifact` | `DownloadArtifactParams` | `(*Download, error)` | GET `/artifacts/download` | No |
| `HandleProviderWebhook` | `HandleProviderWebhookParams` | `(WebhookOk, error)` | POST `/webhooks/providers/{provider}` | No |

## Models, optional fields, and collections

- `envai.Present(value)` includes an optional field. The zero-value `Optional[T]` omits it. `envai.Null[T]()` explicitly sends JSON null; use only when the field's schema allows null.
- Resource `data` wrappers are unwrapped into typed resources. Completion pages retain `Data`, `Links`, and `Meta`. This contract has no declared completion `page` request parameter; there is no automatic page iterator.
- API key creation distinguishes omitted expiry from explicit `"never"`. Treat the plaintext key in the creation response as a secret.
- Flexible metadata, schema, and union values use `json.RawMessage` or typed JSON maps. Unknown REST response properties are ignored; known fields keep their declared Go types.
- HTTP 204 methods return only an error. `Download` contains `Body`, `Header`, and `Status`; always call `Close`, including after partial reads. Reading a download is your responsibility and must remain within the request's context deadline.
- Multipart methods take `[]envai.MultipartPart`. You supply `Data`, `Name`, `Filename`, and `ContentType`. The SDK does not open local files or execute file contents.
- `DownloadArtifact` uses server-provided signed fields. Do not construct a signature, replace opaque IDs, or add bearer credentials to the signed URL.

## Configuration and errors

See [configuration](getting-started.md#configuration) for defaults and [error handling](streaming.md#error-handling) for typed errors. Request methods never automatically retry mutations. Generated methods share the same transport and credential handling.

## Event payloads

`DecodeEvent` validates ATP v3 and immutable numeric v2 envelopes before hydration. `Event.TypedPayload()` returns the corresponding concrete payload; `Event.Raw` retains original JSON. All 29 declared types are listed in [events_generated.go](../events_generated.go) and described by [events.json](../contract/events.json). Valid future event types remain available as raw data. Capabilities discovery does not restrict which valid event types can be decoded.

## Wiki and negotiated contracts

See [Wiki lifecycle and durable webhook intake](wiki.md) for all source, operation, catalog and JWKS methods, and [contract 3 migration](contract3-migration.md) for default selection, strict legacy reads, nullable costs and compatibility errors.
