package envai

import (
	"context"
	"encoding/json"
)

type SendChatMessageParams struct {
	ChatID string
	Body   ChatMessageRequest
}

func (c *Client) SendChatMessage(ctx context.Context, req SendChatMessageParams) (ChatMessageAccepted, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out ChatMessageAccepted
	err := c.call(ctx, "sendChatMessage", p, req.Body, &out)
	return out, err
}

type StreamChatEventsParams struct {
	ChatID string
	After  Optional[int64]
}

func (c *Client) StreamChatEvents(ctx context.Context, req StreamChatEventsParams) (*EventStream, error) {
	return c.Events(ctx, req.ChatID, StreamOptions{After: req.After.Value})
}

type CancelChatTurnParams struct {
	ChatID string
}

func (c *Client) CancelChatTurn(ctx context.Context, req CancelChatTurnParams) (TurnAccepted, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out TurnAccepted
	err := c.call(ctx, "cancelChatTurn", p, nil, &out)
	return out, err
}

type SteerChatTurnParams struct {
	ChatID string
	Body   SteerRequest
}

func (c *Client) SteerChatTurn(ctx context.Context, req SteerChatTurnParams) (TurnAccepted, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out TurnAccepted
	err := c.call(ctx, "steerChatTurn", p, req.Body, &out)
	return out, err
}

type ResolveChatToolParams struct {
	ChatID string
	Body   ToolOutputRequest
}

func (c *Client) ResolveChatTool(ctx context.Context, req ResolveChatToolParams) error {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	return c.call(ctx, "resolveChatTool", p, req.Body, nil)
}

type ResolveChatApprovalParams struct {
	ChatID string
	Body   ApprovalRequest
}

func (c *Client) ResolveChatApproval(ctx context.Context, req ResolveChatApprovalParams) error {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	return c.call(ctx, "resolveChatApproval", p, req.Body, nil)
}

type ForkChatParams struct {
	ChatID string
	Body   ForkChatRequest
}

func (c *Client) ForkChat(ctx context.Context, req ForkChatParams) (ChatForked, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out ChatForked
	err := c.call(ctx, "forkChat", p, req.Body, &out)
	return out, err
}

type RestoreChatParams struct {
	ChatID string
	Body   RestoreChatRequest
}

func (c *Client) RestoreChat(ctx context.Context, req RestoreChatParams) (RestoreChatAccepted, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out RestoreChatAccepted
	err := c.call(ctx, "restoreChat", p, req.Body, &out)
	return out, err
}

type GetCapabilitiesParams struct {
}

func (c *Client) GetCapabilities(ctx context.Context, req GetCapabilitiesParams) (Capabilities, error) {
	p := map[string]any{}

	var out Capabilities
	err := c.call(ctx, "getCapabilities", p, nil, &out)
	return out, err
}

type HealthCheckParams struct {
}

func (c *Client) HealthCheck(ctx context.Context, req HealthCheckParams) (HealthCheckResult, error) {
	p := map[string]any{}

	var out HealthCheckResult
	err := c.call(ctx, "healthCheck", p, nil, &out)
	return out, err
}

type GetJWKSParams struct {
}

func (c *Client) GetJWKS(ctx context.Context, req GetJWKSParams) (JWKSDocument, error) {
	p := map[string]any{}

	var out JWKSDocument
	err := c.call(ctx, "getJwks", p, nil, &out)
	return out, err
}

type CreateWorkspaceParams struct {
	Body WorkspaceCreateRequest
}

func (c *Client) CreateWorkspace(ctx context.Context, req CreateWorkspaceParams) (Workspace, error) {
	p := map[string]any{}

	var out Workspace
	err := c.call(ctx, "createWorkspace", p, req.Body, &out)
	return out, err
}

type ListWorkspacesParams struct {
}

func (c *Client) ListWorkspaces(ctx context.Context, req ListWorkspacesParams) ([]Workspace, error) {
	p := map[string]any{}

	var out []Workspace
	err := c.call(ctx, "listWorkspaces", p, nil, &out)
	return out, err
}

type GetWorkspaceParams struct {
	WorkspaceID string
}

func (c *Client) GetWorkspace(ctx context.Context, req GetWorkspaceParams) (Workspace, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out Workspace
	err := c.call(ctx, "getWorkspace", p, nil, &out)
	return out, err
}

type UpdateWorkspaceParams struct {
	WorkspaceID string
	Body        WorkspaceUpdateRequest
}

func (c *Client) UpdateWorkspace(ctx context.Context, req UpdateWorkspaceParams) (Workspace, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out Workspace
	err := c.call(ctx, "updateWorkspace", p, req.Body, &out)
	return out, err
}

type CreateUserToolParams struct {
	Body UserToolCreateRequest
}

func (c *Client) CreateUserTool(ctx context.Context, req CreateUserToolParams) (UserTool, error) {
	p := map[string]any{}

	var out UserTool
	err := c.call(ctx, "createUserTool", p, req.Body, &out)
	return out, err
}

type ListUserToolsParams struct {
}

func (c *Client) ListUserTools(ctx context.Context, req ListUserToolsParams) ([]UserTool, error) {
	p := map[string]any{}

	var out []UserTool
	err := c.call(ctx, "listUserTools", p, nil, &out)
	return out, err
}

type GetUserToolParams struct {
	Name string
}

func (c *Client) GetUserTool(ctx context.Context, req GetUserToolParams) (UserTool, error) {
	p := map[string]any{}
	p["name"] = req.Name
	var out UserTool
	err := c.call(ctx, "getUserTool", p, nil, &out)
	return out, err
}

type UpdateUserToolParams struct {
	Name string
	Body UserToolUpdateRequest
}

func (c *Client) UpdateUserTool(ctx context.Context, req UpdateUserToolParams) (UserTool, error) {
	p := map[string]any{}
	p["name"] = req.Name
	var out UserTool
	err := c.call(ctx, "updateUserTool", p, req.Body, &out)
	return out, err
}

type DeleteUserToolParams struct {
	Name string
}

func (c *Client) DeleteUserTool(ctx context.Context, req DeleteUserToolParams) error {
	p := map[string]any{}
	p["name"] = req.Name
	return c.call(ctx, "deleteUserTool", p, nil, nil)
}

type CreateChatParams struct {
	WorkspaceID string
	Body        ChatCreateRequest
}

func (c *Client) CreateChat(ctx context.Context, req CreateChatParams) (Chat, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out Chat
	err := c.call(ctx, "createChat", p, req.Body, &out)
	return out, err
}

type ListChatsParams struct {
	WorkspaceID string
}

func (c *Client) ListChats(ctx context.Context, req ListChatsParams) ([]Chat, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out []Chat
	err := c.call(ctx, "listChats", p, nil, &out)
	return out, err
}

type GetChatParams struct {
	WorkspaceID string
	ChatID      string
}

func (c *Client) GetChat(ctx context.Context, req GetChatParams) (Chat, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	var out Chat
	err := c.call(ctx, "getChat", p, nil, &out)
	return out, err
}

type ListCompletionsParams struct {
	WorkspaceID       string
	ChatID            string
	Status            Optional[string]
	ModelOrchestrator Optional[string]
	StartedAtFrom     Optional[string]
	StartedAtTo       Optional[string]
	PerPage           Optional[int64]
}

func (c *Client) ListCompletions(ctx context.Context, req ListCompletionsParams) (CompletionCollection, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	if req.Status.Set {
		p["status"] = req.Status.Value
	}
	if req.ModelOrchestrator.Set {
		p["model_orchestrator"] = req.ModelOrchestrator.Value
	}
	if req.StartedAtFrom.Set {
		p["started_at[from]"] = req.StartedAtFrom.Value
	}
	if req.StartedAtTo.Set {
		p["started_at[to]"] = req.StartedAtTo.Value
	}
	if req.PerPage.Set {
		p["per_page"] = req.PerPage.Value
	}
	var out CompletionCollection
	err := c.call(ctx, "listCompletions", p, nil, &out)
	return out, err
}

type GetCompletionParams struct {
	WorkspaceID  string
	ChatID       string
	CompletionID string
}

func (c *Client) GetCompletion(ctx context.Context, req GetCompletionParams) (Completion, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	p["completion_id"] = req.CompletionID
	var out Completion
	err := c.call(ctx, "getCompletion", p, nil, &out)
	return out, err
}

type IndexWorkspaceFilesParams struct {
	WorkspaceID string
	Path        string
}

func (c *Client) IndexWorkspaceFiles(ctx context.Context, req IndexWorkspaceFilesParams) (WorkspaceFile, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["path"] = req.Path
	var out WorkspaceFile
	err := c.call(ctx, "indexWorkspaceFiles", p, nil, &out)
	return out, err
}

type CreateWorkspaceFileParams struct {
	WorkspaceID string
	Body        WorkspaceFileRequest
}

func (c *Client) CreateWorkspaceFile(ctx context.Context, req CreateWorkspaceFileParams) (WorkspaceFile, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out WorkspaceFile
	err := c.call(ctx, "createWorkspaceFile", p, req.Body, &out)
	return out, err
}

type UpdateWorkspaceFileParams struct {
	WorkspaceID string
	Body        WorkspaceFileRequest
}

func (c *Client) UpdateWorkspaceFile(ctx context.Context, req UpdateWorkspaceFileParams) (WorkspaceFile, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out WorkspaceFile
	err := c.call(ctx, "updateWorkspaceFile", p, req.Body, &out)
	return out, err
}

type DeleteWorkspaceFileParams struct {
	WorkspaceID string
	Path        string
	Recursive   Optional[bool]
}

func (c *Client) DeleteWorkspaceFile(ctx context.Context, req DeleteWorkspaceFileParams) error {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["path"] = req.Path
	if req.Recursive.Set {
		p["recursive"] = req.Recursive.Value
	}
	return c.call(ctx, "deleteWorkspaceFile", p, nil, nil)
}

type DownloadWorkspaceFileParams struct {
	WorkspaceID string
	Path        string
}

func (c *Client) DownloadWorkspaceFile(ctx context.Context, req DownloadWorkspaceFileParams) (*Download, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["path"] = req.Path
	return c.download(ctx, "downloadWorkspaceFile", p, nil)
}

type ConvertWorkspaceFileToPDFParams struct {
	WorkspaceID string
	Body        ConvertToPDFRequest
}

func (c *Client) ConvertWorkspaceFileToPDF(ctx context.Context, req ConvertWorkspaceFileToPDFParams) (*Download, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	return c.download(ctx, "convertWorkspaceFileToPdf", p, req.Body)
}

type CreateAgentParams struct {
	Body AgentCreateRequest
}

func (c *Client) CreateAgent(ctx context.Context, req CreateAgentParams) (Agent, error) {
	p := map[string]any{}

	var out Agent
	err := c.call(ctx, "createAgent", p, req.Body, &out)
	return out, err
}

type ListAgentsParams struct {
}

func (c *Client) ListAgents(ctx context.Context, req ListAgentsParams) ([]Agent, error) {
	p := map[string]any{}

	var out []Agent
	err := c.call(ctx, "listAgents", p, nil, &out)
	return out, err
}

type GetAgentParams struct {
	AgentID string
}

func (c *Client) GetAgent(ctx context.Context, req GetAgentParams) (Agent, error) {
	p := map[string]any{}
	p["agent_id"] = req.AgentID
	var out Agent
	err := c.call(ctx, "getAgent", p, nil, &out)
	return out, err
}

type UpdateAgentParams struct {
	AgentID string
	Body    AgentUpdateRequest
}

func (c *Client) UpdateAgent(ctx context.Context, req UpdateAgentParams) (Agent, error) {
	p := map[string]any{}
	p["agent_id"] = req.AgentID
	var out Agent
	err := c.call(ctx, "updateAgent", p, req.Body, &out)
	return out, err
}

type DeleteAgentParams struct {
	AgentID string
}

func (c *Client) DeleteAgent(ctx context.Context, req DeleteAgentParams) error {
	p := map[string]any{}
	p["agent_id"] = req.AgentID
	return c.call(ctx, "deleteAgent", p, nil, nil)
}

type CreateTeamParams struct {
	Body TeamCreateRequest
}

func (c *Client) CreateTeam(ctx context.Context, req CreateTeamParams) (Team, error) {
	p := map[string]any{}

	var out Team
	err := c.call(ctx, "createTeam", p, req.Body, &out)
	return out, err
}

type ListTeamsParams struct {
}

func (c *Client) ListTeams(ctx context.Context, req ListTeamsParams) ([]Team, error) {
	p := map[string]any{}

	var out []Team
	err := c.call(ctx, "listTeams", p, nil, &out)
	return out, err
}

type GetTeamParams struct {
	TeamID string
}

func (c *Client) GetTeam(ctx context.Context, req GetTeamParams) (Team, error) {
	p := map[string]any{}
	p["team_id"] = req.TeamID
	var out Team
	err := c.call(ctx, "getTeam", p, nil, &out)
	return out, err
}

type UpdateTeamParams struct {
	TeamID string
	Body   TeamUpdateRequest
}

func (c *Client) UpdateTeam(ctx context.Context, req UpdateTeamParams) (Team, error) {
	p := map[string]any{}
	p["team_id"] = req.TeamID
	var out Team
	err := c.call(ctx, "updateTeam", p, req.Body, &out)
	return out, err
}

type DeleteTeamParams struct {
	TeamID string
}

func (c *Client) DeleteTeam(ctx context.Context, req DeleteTeamParams) error {
	p := map[string]any{}
	p["team_id"] = req.TeamID
	return c.call(ctx, "deleteTeam", p, nil, nil)
}

type ListModelsParams struct {
}

func (c *Client) ListModels(ctx context.Context, req ListModelsParams) (ModelList, error) {
	p := map[string]any{}

	var out ModelList
	err := c.call(ctx, "listModels", p, nil, &out)
	return out, err
}

type CreateAPIKeyParams struct {
	Body APIKeyCreate
}

func (c *Client) CreateAPIKey(ctx context.Context, req CreateAPIKeyParams) (json.RawMessage, error) {
	p := map[string]any{}

	var out json.RawMessage
	err := c.call(ctx, "createApiKey", p, req.Body, &out)
	return out, err
}

type ListAPIKeysParams struct {
}

func (c *Client) ListAPIKeys(ctx context.Context, req ListAPIKeysParams) ([]APIKey, error) {
	p := map[string]any{}

	var out []APIKey
	err := c.call(ctx, "listApiKeys", p, nil, &out)
	return out, err
}

type DeleteAPIKeyParams struct {
	KeyID string
}

func (c *Client) DeleteAPIKey(ctx context.Context, req DeleteAPIKeyParams) error {
	p := map[string]any{}
	p["key_id"] = req.KeyID
	return c.call(ctx, "deleteApiKey", p, nil, nil)
}

type TestGenerateImageParams struct {
	Body []MultipartPart
}

func (c *Client) TestGenerateImage(ctx context.Context, req TestGenerateImageParams) (TestGenerateImageResult, error) {
	p := map[string]any{}

	var out TestGenerateImageResult
	err := c.call(ctx, "testGenerateImage", p, req.Body, &out)
	return out, err
}

type CreateImageGenerationParams struct {
	Body ImageGenerationRequest
}

func (c *Client) CreateImageGeneration(ctx context.Context, req CreateImageGenerationParams) (ImageGenerationAcceptedRecord, error) {
	p := map[string]any{}

	var out ImageGenerationAcceptedRecord
	err := c.call(ctx, "createImageGeneration", p, req.Body, &out)
	return out, err
}

type GetImageGenerationParams struct {
	Record string
}

func (c *Client) GetImageGeneration(ctx context.Context, req GetImageGenerationParams) (ImageGenerationStatusRecord, error) {
	p := map[string]any{}
	p["record"] = req.Record
	var out ImageGenerationStatusRecord
	err := c.call(ctx, "getImageGeneration", p, nil, &out)
	return out, err
}

type UploadImageParams struct {
	Body []MultipartPart
}

func (c *Client) UploadImage(ctx context.Context, req UploadImageParams) (ImageUploadRecord, error) {
	p := map[string]any{}

	var out ImageUploadRecord
	err := c.call(ctx, "uploadImage", p, req.Body, &out)
	return out, err
}

type ListImageModelsParams struct {
}

func (c *Client) ListImageModels(ctx context.Context, req ListImageModelsParams) ([]ModelCatalogEntry, error) {
	p := map[string]any{}

	var out []ModelCatalogEntry
	err := c.call(ctx, "listImageModels", p, nil, &out)
	return out, err
}

type DownloadArtifactParams struct {
	Generation string
	Artifact   string
	Expires    Optional[int64]
	Signature  string
}

func (c *Client) DownloadArtifact(ctx context.Context, req DownloadArtifactParams) (*Download, error) {
	p := map[string]any{}
	p["generation"] = req.Generation
	p["artifact"] = req.Artifact
	if req.Expires.Set {
		p["expires"] = req.Expires.Value
	}
	p["signature"] = req.Signature
	return c.download(ctx, "downloadArtifact", p, nil)
}

type HandleProviderWebhookParams struct {
	Provider             string
	CallbackToken        Optional[string]
	WebhookID            Optional[string]
	WebhookTimestamp     Optional[string]
	WebhookSignature     Optional[string]
	XFalWebhookRequestID Optional[string]
	XFalWebhookUserID    Optional[string]
	XFalWebhookTimestamp Optional[string]
	XFalWebhookSignature Optional[string]
	XBFLSignature        Optional[string]
	Body                 ProviderWebhookPayload
}

func (c *Client) HandleProviderWebhook(ctx context.Context, req HandleProviderWebhookParams) (WebhookOk, error) {
	p := map[string]any{}
	p["provider"] = req.Provider
	if req.CallbackToken.Set {
		p["callback_token"] = req.CallbackToken.Value
	}
	if req.WebhookID.Set {
		p["webhook-id"] = req.WebhookID.Value
	}
	if req.WebhookTimestamp.Set {
		p["webhook-timestamp"] = req.WebhookTimestamp.Value
	}
	if req.WebhookSignature.Set {
		p["webhook-signature"] = req.WebhookSignature.Value
	}
	if req.XFalWebhookRequestID.Set {
		p["X-Fal-Webhook-Request-Id"] = req.XFalWebhookRequestID.Value
	}
	if req.XFalWebhookUserID.Set {
		p["X-Fal-Webhook-User-Id"] = req.XFalWebhookUserID.Value
	}
	if req.XFalWebhookTimestamp.Set {
		p["X-Fal-Webhook-Timestamp"] = req.XFalWebhookTimestamp.Value
	}
	if req.XFalWebhookSignature.Set {
		p["X-Fal-Webhook-Signature"] = req.XFalWebhookSignature.Value
	}
	if req.XBFLSignature.Set {
		p["X-BFL-Signature"] = req.XBFLSignature.Value
	}
	var out WebhookOk
	err := c.call(ctx, "handleProviderWebhook", p, req.Body, &out)
	return out, err
}

type UploadWikiSourceParams struct {
	WorkspaceID    string
	IdempotencyKey string
	Body           []MultipartPart
}

func (c *Client) UploadWikiSource(ctx context.Context, req UploadWikiSourceParams) (WikiSource, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["Idempotency-Key"] = req.IdempotencyKey
	var out WikiSource
	err := c.call(ctx, "uploadWikiSource", p, req.Body, &out)
	return out, err
}

type GetWikiSourceParams struct {
	WorkspaceID string
	FileID      string
}

func (c *Client) GetWikiSource(ctx context.Context, req GetWikiSourceParams) (WikiSource, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["file_id"] = req.FileID
	var out WikiSource
	err := c.call(ctx, "getWikiSource", p, nil, &out)
	return out, err
}

type DownloadWikiSourceParams struct {
	WorkspaceID string
	FileID      string
}

func (c *Client) DownloadWikiSource(ctx context.Context, req DownloadWikiSourceParams) (*Download, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["file_id"] = req.FileID
	return c.download(ctx, "downloadWikiSource", p, nil)
}

type WithdrawWikiSourceParams struct {
	WorkspaceID string
	Body        WithdrawWikiSourceBody
}

func (c *Client) WithdrawWikiSource(ctx context.Context, req WithdrawWikiSourceParams) (WikiSource, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out WikiSource
	err := c.call(ctx, "withdrawWikiSource", p, req.Body, &out)
	return out, err
}

type CreateWikiInboxRunParams struct {
	WorkspaceID    string
	IdempotencyKey string
	Body           CreateWikiInboxRunBody
}

func (c *Client) CreateWikiInboxRun(ctx context.Context, req CreateWikiInboxRunParams) (WikiOperation, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["Idempotency-Key"] = req.IdempotencyKey
	var out WikiOperation
	err := c.call(ctx, "createWikiInboxRun", p, req.Body, &out)
	return out, err
}

type CreateWikiOutboxRunParams struct {
	WorkspaceID    string
	IdempotencyKey string
	Body           CreateWikiOutboxRunBody
}

func (c *Client) CreateWikiOutboxRun(ctx context.Context, req CreateWikiOutboxRunParams) (WikiOperation, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["Idempotency-Key"] = req.IdempotencyKey
	var out WikiOperation
	err := c.call(ctx, "createWikiOutboxRun", p, req.Body, &out)
	return out, err
}

type CreateWikiSyncRunParams struct {
	WorkspaceID    string
	IdempotencyKey string
	Body           CreateWikiSyncRunBody
}

func (c *Client) CreateWikiSyncRun(ctx context.Context, req CreateWikiSyncRunParams) (WikiOperation, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["Idempotency-Key"] = req.IdempotencyKey
	var out WikiOperation
	err := c.call(ctx, "createWikiSyncRun", p, req.Body, &out)
	return out, err
}

type GetWikiOperationParams struct {
	WorkspaceID string
	OperationID string
}

func (c *Client) GetWikiOperation(ctx context.Context, req GetWikiOperationParams) (WikiOperation, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["operation_id"] = req.OperationID
	var out WikiOperation
	err := c.call(ctx, "getWikiOperation", p, nil, &out)
	return out, err
}

type ListWikiCategoriesParams struct {
	WorkspaceID string
}

func (c *Client) ListWikiCategories(ctx context.Context, req ListWikiCategoriesParams) (Categories, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out Categories
	err := c.call(ctx, "listWikiCategories", p, nil, &out)
	return out, err
}

type GetWikiCategoryIndexParams struct {
	WorkspaceID     string
	CategoryID      string
	From            Optional[int64]
	To              Optional[int64]
	All             Optional[bool]
	Type            Optional[string]
	CatalogRevision Optional[int64]
}

func (c *Client) GetWikiCategoryIndex(ctx context.Context, req GetWikiCategoryIndexParams) (CategoryIndex, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["category_id"] = req.CategoryID
	if req.From.Set {
		p["from"] = req.From.Value
	}
	if req.To.Set {
		p["to"] = req.To.Value
	}
	if req.All.Set {
		p["all"] = req.All.Value
	}
	if req.Type.Set {
		p["type"] = req.Type.Value
	}
	if req.CatalogRevision.Set {
		p["catalog_revision"] = req.CatalogRevision.Value
	}
	var out CategoryIndex
	err := c.call(ctx, "getWikiCategoryIndex", p, nil, &out)
	return out, err
}

type GetWikiWebhookJWKSParams struct {
}

func (c *Client) GetWikiWebhookJWKS(ctx context.Context, req GetWikiWebhookJWKSParams) (JWKSDocument, error) {
	p := map[string]any{}

	var out JWKSDocument
	err := c.call(ctx, "getWikiWebhookJwks", p, nil, &out)
	return out, err
}

type StreamChatEventsLegacyParams struct {
	ChatID string
	After  Optional[int64]
}

func (c *Client) StreamChatEventsLegacy(ctx context.Context, req StreamChatEventsLegacyParams) (*EventStream, error) {
	return c.LegacyEvents(ctx, req.ChatID, StreamOptions{After: req.After.Value})
}

type GetCapabilitiesLegacyParams struct {
}

func (c *Client) GetCapabilitiesLegacy(ctx context.Context, req GetCapabilitiesLegacyParams) (CapabilitiesLegacy, error) {
	p := map[string]any{}

	var out CapabilitiesLegacy
	err := c.call(ctx, "getCapabilitiesLegacy", p, nil, &out)
	return out, err
}
