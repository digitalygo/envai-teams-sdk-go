package envai

import (
	"context"
	"encoding/json"
)

type DeliverWorkflowResultParams struct {
	ChatID string
	Body   NativeDeliveryRequest
}

func (c *Client) DeliverWorkflowResult(ctx context.Context, req DeliverWorkflowResultParams) (NativeDeliveryReceipt, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out NativeDeliveryReceipt
	err := c.call(ctx, "deliverWorkflowResult", p, req.Body, &out)
	return out, err
}

type DelegateToTeamParams struct {
	ChatID string
	Body   TeamDelegationRequest
}

func (c *Client) DelegateToTeam(ctx context.Context, req DelegateToTeamParams) (TeamDelegationReceipt, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out TeamDelegationReceipt
	err := c.call(ctx, "delegateToTeam", p, req.Body, &out)
	return out, err
}

type GetDetachedTaskParams struct {
	ChatID string
	TaskID string
}

func (c *Client) GetDetachedTask(ctx context.Context, req GetDetachedTaskParams) (DetachedTaskSnapshot, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	p["task_id"] = req.TaskID
	var out DetachedTaskSnapshot
	err := c.call(ctx, "getDetachedTask", p, nil, &out)
	return out, err
}

type CancelDetachedTaskParams struct {
	ChatID string
	TaskID string
	Body   CancelDetachedTaskBody
}

func (c *Client) CancelDetachedTask(ctx context.Context, req CancelDetachedTaskParams) (CancelDetachedTaskResult, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	p["task_id"] = req.TaskID
	var out CancelDetachedTaskResult
	err := c.call(ctx, "cancelDetachedTask", p, req.Body, &out)
	return out, err
}

type ClassifyPhaseParams struct {
	ChatID string
	Body   PhaseClassificationRequest
}

func (c *Client) ClassifyPhase(ctx context.Context, req ClassifyPhaseParams) (PhaseClassificationResult, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out PhaseClassificationResult
	err := c.call(ctx, "classifyPhase", p, req.Body, &out)
	return out, err
}

type ClassifyChoiceParams struct {
	ChatID string
	Body   ChoiceClassificationRequest
}

func (c *Client) ClassifyChoice(ctx context.Context, req ClassifyChoiceParams) (ChoiceClassificationResult, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out ChoiceClassificationResult
	err := c.call(ctx, "classifyChoice", p, req.Body, &out)
	return out, err
}

type StartIndependentJobParams struct {
	ChatID string
	Body   IndependentJobRequest
}

func (c *Client) StartIndependentJob(ctx context.Context, req StartIndependentJobParams) (IndependentJobReceipt, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	var out IndependentJobReceipt
	err := c.call(ctx, "startIndependentJob", p, req.Body, &out)
	return out, err
}

type SendChatMessageParams struct {
	ChatID         string
	IdempotencyKey Optional[string]
	Body           ChatMessageRequest
}

func (c *Client) SendChatMessage(ctx context.Context, req SendChatMessageParams) (ChatMessageAccepted, error) {
	p := map[string]any{}
	p["chat_id"] = req.ChatID
	if req.IdempotencyKey.Set {
		p["Idempotency-Key"] = req.IdempotencyKey.Value
	}
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
	XEnvaiProjectID Optional[string]
	Body            WorkspaceCreateRequest
}

func (c *Client) CreateWorkspace(ctx context.Context, req CreateWorkspaceParams) (Workspace, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out Workspace
	err := c.call(ctx, "createWorkspace", p, req.Body, &out)
	return out, err
}

type ListWorkspacesParams struct {
	XEnvaiProjectID Optional[string]
}

func (c *Client) ListWorkspaces(ctx context.Context, req ListWorkspacesParams) ([]Workspace, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
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

type DeleteWorkspaceParams struct {
	WorkspaceID string
}

func (c *Client) DeleteWorkspace(ctx context.Context, req DeleteWorkspaceParams) (CleanupResult, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	var out CleanupResult
	err := c.call(ctx, "deleteWorkspace", p, nil, &out)
	return out, err
}

type CreateProjectToolParams struct {
	XEnvaiProjectID Optional[string]
	Body            ProjectToolCreateRequest
}

func (c *Client) CreateProjectTool(ctx context.Context, req CreateProjectToolParams) (ProjectTool, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out ProjectTool
	err := c.call(ctx, "createProjectTool", p, req.Body, &out)
	return out, err
}

// Deprecated: use CreateProjectTool. Tools are scoped to the key project, not the user.
type CreateUserToolParams = CreateProjectToolParams

// Deprecated: use CreateProjectTool. This alias uses project scope.
func (c *Client) CreateUserTool(ctx context.Context, req CreateUserToolParams) (ProjectTool, error) {
	return c.CreateProjectTool(ctx, req)
}

type ListProjectToolsParams struct {
	XEnvaiProjectID Optional[string]
}

func (c *Client) ListProjectTools(ctx context.Context, req ListProjectToolsParams) ([]ProjectTool, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out []ProjectTool
	err := c.call(ctx, "listProjectTools", p, nil, &out)
	return out, err
}

// Deprecated: use ListProjectTools. Tools are scoped to the key project, not the user.
type ListUserToolsParams = ListProjectToolsParams

// Deprecated: use ListProjectTools. This alias uses project scope.
func (c *Client) ListUserTools(ctx context.Context, req ListUserToolsParams) ([]ProjectTool, error) {
	return c.ListProjectTools(ctx, req)
}

type GetProjectToolParams struct {
	Name            string
	XEnvaiProjectID Optional[string]
}

func (c *Client) GetProjectTool(ctx context.Context, req GetProjectToolParams) (ProjectTool, error) {
	p := map[string]any{}
	p["name"] = req.Name
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out ProjectTool
	err := c.call(ctx, "getProjectTool", p, nil, &out)
	return out, err
}

// Deprecated: use GetProjectTool. Tools are scoped to the key project, not the user.
type GetUserToolParams = GetProjectToolParams

// Deprecated: use GetProjectTool. This alias uses project scope.
func (c *Client) GetUserTool(ctx context.Context, req GetUserToolParams) (ProjectTool, error) {
	return c.GetProjectTool(ctx, req)
}

type UpdateProjectToolParams struct {
	Name            string
	XEnvaiProjectID Optional[string]
	Body            ProjectToolUpdateRequest
}

func (c *Client) UpdateProjectTool(ctx context.Context, req UpdateProjectToolParams) (ProjectTool, error) {
	p := map[string]any{}
	p["name"] = req.Name
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out ProjectTool
	err := c.call(ctx, "updateProjectTool", p, req.Body, &out)
	return out, err
}

// Deprecated: use UpdateProjectTool. Tools are scoped to the key project, not the user.
type UpdateUserToolParams = UpdateProjectToolParams

// Deprecated: use UpdateProjectTool. This alias uses project scope.
func (c *Client) UpdateUserTool(ctx context.Context, req UpdateUserToolParams) (ProjectTool, error) {
	return c.UpdateProjectTool(ctx, req)
}

type DeleteProjectToolParams struct {
	Name            string
	XEnvaiProjectID Optional[string]
}

func (c *Client) DeleteProjectTool(ctx context.Context, req DeleteProjectToolParams) error {
	p := map[string]any{}
	p["name"] = req.Name
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	return c.call(ctx, "deleteProjectTool", p, nil, nil)
}

// Deprecated: use DeleteProjectTool. Tools are scoped to the key project, not the user.
type DeleteUserToolParams = DeleteProjectToolParams

// Deprecated: use DeleteProjectTool. This alias uses project scope.
func (c *Client) DeleteUserTool(ctx context.Context, req DeleteUserToolParams) error {
	return c.DeleteProjectTool(ctx, req)
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

type UpdateChatParams struct {
	WorkspaceID string
	ChatID      string
	Body        ChatUpdateRequest
}

func (c *Client) UpdateChat(ctx context.Context, req UpdateChatParams) (Chat, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	var out Chat
	err := c.call(ctx, "updateChat", p, req.Body, &out)
	return out, err
}

type DeleteChatParams struct {
	WorkspaceID string
	ChatID      string
}

func (c *Client) DeleteChat(ctx context.Context, req DeleteChatParams) (CleanupResult, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	var out CleanupResult
	err := c.call(ctx, "deleteChat", p, nil, &out)
	return out, err
}

type ListCompletionsParams struct {
	WorkspaceID       string
	ChatID            string
	Status            Optional[string]
	ModelOrchestrator Optional[string]
	StartedAtFrom     Optional[string]
	StartedAtTo       Optional[string]
	Page              Optional[int64]
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
	if req.Page.Set {
		p["page"] = req.Page.Value
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

type ListChatCronsParams struct {
	WorkspaceID     string
	XEnvaiProjectID Optional[string]
	OriginChatID    Optional[string]
	Status          Optional[string]
}

func (c *Client) ListChatCrons(ctx context.Context, req ListChatCronsParams) ([]ChatCron, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	if req.OriginChatID.Set {
		p["origin_chat_id"] = req.OriginChatID.Value
	}
	if req.Status.Set {
		p["status"] = req.Status.Value
	}
	var out []ChatCron
	err := c.call(ctx, "listChatCrons", p, nil, &out)
	return out, err
}

type GetChatCronParams struct {
	WorkspaceID     string
	XEnvaiProjectID Optional[string]
	CronID          string
}

func (c *Client) GetChatCron(ctx context.Context, req GetChatCronParams) (ChatCron, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	p["cron_id"] = req.CronID
	var out ChatCron
	err := c.call(ctx, "getChatCron", p, nil, &out)
	return out, err
}

type UpdateChatCronParams struct {
	WorkspaceID     string
	XEnvaiProjectID Optional[string]
	CronID          string
	Body            ChatCronUpdateRequest
}

func (c *Client) UpdateChatCron(ctx context.Context, req UpdateChatCronParams) (ChatCron, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	p["cron_id"] = req.CronID
	var out ChatCron
	err := c.call(ctx, "updateChatCron", p, req.Body, &out)
	return out, err
}

type DeleteChatCronParams struct {
	WorkspaceID     string
	XEnvaiProjectID Optional[string]
	CronID          string
}

func (c *Client) DeleteChatCron(ctx context.Context, req DeleteChatCronParams) error {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	p["cron_id"] = req.CronID
	return c.call(ctx, "deleteChatCron", p, nil, nil)
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
	XEnvaiProjectID Optional[string]
	Body            AgentCreateRequest
}

func (c *Client) CreateAgent(ctx context.Context, req CreateAgentParams) (Agent, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out Agent
	err := c.call(ctx, "createAgent", p, req.Body, &out)
	return out, err
}

type ListAgentsParams struct {
	XEnvaiProjectID Optional[string]
}

func (c *Client) ListAgents(ctx context.Context, req ListAgentsParams) ([]Agent, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
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
	XEnvaiProjectID Optional[string]
	Body            TeamCreateRequest
}

func (c *Client) CreateTeam(ctx context.Context, req CreateTeamParams) (Team, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out Team
	err := c.call(ctx, "createTeam", p, req.Body, &out)
	return out, err
}

type ListTeamsParams struct {
	XEnvaiProjectID Optional[string]
}

func (c *Client) ListTeams(ctx context.Context, req ListTeamsParams) ([]Team, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
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
	XEnvaiProjectID Optional[string]
}

func (c *Client) ListModels(ctx context.Context, req ListModelsParams) (ModelList, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out ModelList
	err := c.call(ctx, "listModels", p, nil, &out)
	return out, err
}

type CreateAPIKeyParams struct {
	XEnvaiProjectID Optional[string]
	Body            APIKeyCreate
}

func (c *Client) CreateAPIKey(ctx context.Context, req CreateAPIKeyParams) (json.RawMessage, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out json.RawMessage
	err := c.call(ctx, "createApiKey", p, req.Body, &out)
	return out, err
}

type ListAPIKeysParams struct {
	XEnvaiProjectID Optional[string]
}

func (c *Client) ListAPIKeys(ctx context.Context, req ListAPIKeysParams) ([]APIKey, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out []APIKey
	err := c.call(ctx, "listApiKeys", p, nil, &out)
	return out, err
}

type DeleteAPIKeyParams struct {
	KeyID           string
	XEnvaiProjectID Optional[string]
}

func (c *Client) DeleteAPIKey(ctx context.Context, req DeleteAPIKeyParams) error {
	p := map[string]any{}
	p["key_id"] = req.KeyID
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
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
	IdempotencyKey Optional[string]
	Body           ImageGenerationRequest
}

func (c *Client) CreateImageGeneration(ctx context.Context, req CreateImageGenerationParams) (ImageGenerationAcceptedRecord, error) {
	p := map[string]any{}
	if req.IdempotencyKey.Set {
		p["Idempotency-Key"] = req.IdempotencyKey.Value
	}
	var out ImageGenerationAcceptedRecord
	err := c.call(ctx, "createImageGeneration", p, req.Body, &out)
	return out, err
}

type GetImageGenerationParams struct {
	Record          string
	XEnvaiProjectID Optional[string]
}

func (c *Client) GetImageGeneration(ctx context.Context, req GetImageGenerationParams) (ImageGenerationStatusRecord, error) {
	p := map[string]any{}
	p["record"] = req.Record
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out ImageGenerationStatusRecord
	err := c.call(ctx, "getImageGeneration", p, nil, &out)
	return out, err
}

type UploadImageParams struct {
	XEnvaiProjectID Optional[string]
	Body            []MultipartPart
}

func (c *Client) UploadImage(ctx context.Context, req UploadImageParams) (ImageUploadRecord, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	var out ImageUploadRecord
	err := c.call(ctx, "uploadImage", p, req.Body, &out)
	return out, err
}

type ListImageModelsParams struct {
	XEnvaiProjectID Optional[string]
}

func (c *Client) ListImageModels(ctx context.Context, req ListImageModelsParams) ([]ModelCatalogEntry, error) {
	p := map[string]any{}
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
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

type UploadChatAttachmentParams struct {
	WorkspaceID    string
	ChatID         string
	IdempotencyKey string
	Body           []MultipartPart
}

func (c *Client) UploadChatAttachment(ctx context.Context, req UploadChatAttachmentParams) (ChatAttachment, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	p["Idempotency-Key"] = req.IdempotencyKey
	var out ChatAttachment
	err := c.call(ctx, "uploadChatAttachment", p, req.Body, &out)
	return out, err
}

type ListChatAttachmentsParams struct {
	WorkspaceID string
	ChatID      string
}

func (c *Client) ListChatAttachments(ctx context.Context, req ListChatAttachmentsParams) (ListChatAttachmentsResult, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	var out ListChatAttachmentsResult
	err := c.call(ctx, "listChatAttachments", p, nil, &out)
	return out, err
}

type GetChatAttachmentParams struct {
	WorkspaceID  string
	ChatID       string
	AttachmentID string
}

func (c *Client) GetChatAttachment(ctx context.Context, req GetChatAttachmentParams) (ChatAttachment, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	p["attachment_id"] = req.AttachmentID
	var out ChatAttachment
	err := c.call(ctx, "getChatAttachment", p, nil, &out)
	return out, err
}

type DeleteChatAttachmentParams struct {
	WorkspaceID  string
	ChatID       string
	AttachmentID string
}

func (c *Client) DeleteChatAttachment(ctx context.Context, req DeleteChatAttachmentParams) (ChatAttachment, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	p["attachment_id"] = req.AttachmentID
	var out ChatAttachment
	err := c.call(ctx, "deleteChatAttachment", p, nil, &out)
	return out, err
}

type RetryChatAttachmentParams struct {
	WorkspaceID    string
	ChatID         string
	AttachmentID   string
	IdempotencyKey string
}

func (c *Client) RetryChatAttachment(ctx context.Context, req RetryChatAttachmentParams) (ChatAttachment, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	p["attachment_id"] = req.AttachmentID
	p["Idempotency-Key"] = req.IdempotencyKey
	var out ChatAttachment
	err := c.call(ctx, "retryChatAttachment", p, nil, &out)
	return out, err
}

type DownloadChatAttachmentParams struct {
	WorkspaceID  string
	ChatID       string
	AttachmentID string
	Kind         string
}

func (c *Client) DownloadChatAttachment(ctx context.Context, req DownloadChatAttachmentParams) (*Download, error) {
	p := map[string]any{}
	p["workspace_id"] = req.WorkspaceID
	p["chat_id"] = req.ChatID
	p["attachment_id"] = req.AttachmentID
	p["kind"] = req.Kind
	return c.download(ctx, "downloadChatAttachment", p, nil)
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

type ListProjectSkillsParams struct {
	ProjectID string
}

func (c *Client) ListProjectSkills(ctx context.Context, req ListProjectSkillsParams) ([]ProjectSkill, error) {
	p := map[string]any{}
	p["project_id"] = req.ProjectID
	var out []ProjectSkill
	err := c.call(ctx, "listProjectSkills", p, nil, &out)
	return out, err
}

type CreateProjectSkillParams struct {
	ProjectID string
	Body      ProjectSkillCreateRequest
}

func (c *Client) CreateProjectSkill(ctx context.Context, req CreateProjectSkillParams) (ProjectSkill, error) {
	p := map[string]any{}
	p["project_id"] = req.ProjectID
	var out ProjectSkill
	err := c.call(ctx, "createProjectSkill", p, req.Body, &out)
	return out, err
}

type GetProjectSkillParams struct {
	ProjectID string
	SkillID   string
}

func (c *Client) GetProjectSkill(ctx context.Context, req GetProjectSkillParams) (ProjectSkill, error) {
	p := map[string]any{}
	p["project_id"] = req.ProjectID
	p["skill_id"] = req.SkillID
	var out ProjectSkill
	err := c.call(ctx, "getProjectSkill", p, nil, &out)
	return out, err
}

type UpdateProjectSkillParams struct {
	ProjectID string
	SkillID   string
	Body      ProjectSkillUpdateRequest
}

func (c *Client) UpdateProjectSkill(ctx context.Context, req UpdateProjectSkillParams) (ProjectSkill, error) {
	p := map[string]any{}
	p["project_id"] = req.ProjectID
	p["skill_id"] = req.SkillID
	var out ProjectSkill
	err := c.call(ctx, "updateProjectSkill", p, req.Body, &out)
	return out, err
}

type DeleteProjectSkillParams struct {
	ProjectID string
	SkillID   string
}

func (c *Client) DeleteProjectSkill(ctx context.Context, req DeleteProjectSkillParams) error {
	p := map[string]any{}
	p["project_id"] = req.ProjectID
	p["skill_id"] = req.SkillID
	return c.call(ctx, "deleteProjectSkill", p, nil, nil)
}

type GetProjectUsageParams struct {
	ProjectID       string
	XEnvaiProjectID Optional[string]
	From            Optional[string]
	To              Optional[string]
	APIKeyID        Optional[string]
	Page            Optional[int64]
	PerPage         Optional[int64]
}

func (c *Client) GetProjectUsage(ctx context.Context, req GetProjectUsageParams) (UsageReportCollection, error) {
	p := map[string]any{}
	p["project_id"] = req.ProjectID
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	if req.From.Set {
		p["from"] = req.From.Value
	}
	if req.To.Set {
		p["to"] = req.To.Value
	}
	if req.APIKeyID.Set {
		p["api_key_id"] = req.APIKeyID.Value
	}
	if req.Page.Set {
		p["page"] = req.Page.Value
	}
	if req.PerPage.Set {
		p["per_page"] = req.PerPage.Value
	}
	var out UsageReportCollection
	err := c.call(ctx, "getProjectUsage", p, nil, &out)
	return out, err
}

type GetProjectLogsParams struct {
	ProjectID       string
	XEnvaiProjectID Optional[string]
	View            Optional[string]
	From            Optional[string]
	To              Optional[string]
	APIKeyID        Optional[string]
	ChatID          Optional[string]
	Type            Optional[string]
	Types           Optional[[]string]
	WorkspaceIDs    Optional[[]string]
	APIKeyIDs       Optional[[]string]
	Page            Optional[int64]
	PerPage         Optional[int64]
}

func (c *Client) GetProjectLogs(ctx context.Context, req GetProjectLogsParams) (json.RawMessage, error) {
	p := map[string]any{}
	p["project_id"] = req.ProjectID
	if req.XEnvaiProjectID.Set {
		p["X-Envai-Project-Id"] = req.XEnvaiProjectID.Value
	}
	if req.View.Set {
		p["view"] = req.View.Value
	}
	if req.From.Set {
		p["from"] = req.From.Value
	}
	if req.To.Set {
		p["to"] = req.To.Value
	}
	if req.APIKeyID.Set {
		p["api_key_id"] = req.APIKeyID.Value
	}
	if req.ChatID.Set {
		p["chat_id"] = req.ChatID.Value
	}
	if req.Type.Set {
		p["type"] = req.Type.Value
	}
	if req.Types.Set {
		p["types[]"] = req.Types.Value
	}
	if req.WorkspaceIDs.Set {
		p["workspace_ids[]"] = req.WorkspaceIDs.Value
	}
	if req.APIKeyIDs.Set {
		p["api_key_ids[]"] = req.APIKeyIDs.Value
	}
	if req.Page.Set {
		p["page"] = req.Page.Value
	}
	if req.PerPage.Set {
		p["per_page"] = req.PerPage.Value
	}
	var out json.RawMessage
	err := c.call(ctx, "getProjectLogs", p, nil, &out)
	return out, err
}
