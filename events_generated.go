package envai

import "encoding/json"

type ATPAgent struct {
	InstanceID string                     `json:"instance_id"`
	Name       string                     `json:"name"`
	Role       string                     `json:"role"`
	ParentID   *string                    `json:"parent_id"`
	Model      Optional[string]           `json:"model,omitzero"`
	Extra      map[string]json.RawMessage `json:"-"`
}

func (v *ATPAgent) UnmarshalJSON(b []byte) error {
	type plain ATPAgent
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"instance_id", "name", "role", "parent_id", "model"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPAgent(out)
	return nil
}
func (v ATPAgent) MarshalJSON() ([]byte, error) {
	type plain ATPAgent
	return marshalUnknown(plain(v), v.Extra)
}

type ATPUsage struct {
	InputTokens   int64                      `json:"input_tokens"`
	OutputTokens  int64                      `json:"output_tokens"`
	CostUSD       *float64                   `json:"cost_usd"`
	DurationMs    int64                      `json:"duration_ms"`
	SpriteMs      Optional[int64]            `json:"sprite_ms,omitzero"`
	SpriteRamGbMs Optional[int64]            `json:"sprite_ram_gb_ms,omitzero"`
	ToolCalls     Optional[int64]            `json:"tool_calls,omitzero"`
	Billing       Optional[ATPBillingState]  `json:"billing,omitzero"`
	UsageCost     Optional[ATPUsageCost]     `json:"usage_cost,omitzero"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ATPUsage) UnmarshalJSON(b []byte) error {
	type plain ATPUsage
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"input_tokens", "output_tokens", "cost_usd", "duration_ms", "sprite_ms", "sprite_ram_gb_ms", "tool_calls", "billing", "usage_cost"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPUsage(out)
	return nil
}
func (v ATPUsage) MarshalJSON() ([]byte, error) {
	type plain ATPUsage
	return marshalUnknown(plain(v), v.Extra)
}

type ATPToolKind string

type ATPPing map[string]json.RawMessage

type ATPChatTitleUpdated struct {
	Title string `json:"title"`
}

type ATPTurnStarted struct {
	TurnID      string                                     `json:"turn_id"`
	Input       string                                     `json:"input"`
	Attachments Optional[[]ATPChatAttachmentManifestEvent] `json:"attachments,omitzero"`
}

type ATPTurnCompleted struct {
	TurnID       string                    `json:"turn_id"`
	Usage        ATPUsage                  `json:"usage"`
	Cost         *float64                  `json:"cost"`
	ChatUsage    ATPUsage                  `json:"chat_usage"`
	ChatCost     *float64                  `json:"chat_cost"`
	FinishReason string                    `json:"finish_reason"`
	Billing      Optional[ATPBillingState] `json:"billing,omitzero"`
	UsageCost    Optional[ATPUsageCost]    `json:"usage_cost,omitzero"`
}

type ATPTurnCancelled struct {
	TurnID    string                    `json:"turn_id"`
	Reason    string                    `json:"reason"`
	Usage     ATPUsage                  `json:"usage"`
	Cost      *float64                  `json:"cost"`
	ChatUsage ATPUsage                  `json:"chat_usage"`
	ChatCost  *float64                  `json:"chat_cost"`
	Billing   Optional[ATPBillingState] `json:"billing,omitzero"`
	UsageCost Optional[ATPUsageCost]    `json:"usage_cost,omitzero"`
}

type ATPTurnFailed struct {
	TurnID    string                    `json:"turn_id"`
	Code      string                    `json:"code"`
	Message   string                    `json:"message"`
	Usage     ATPUsage                  `json:"usage"`
	Cost      *float64                  `json:"cost"`
	ChatUsage ATPUsage                  `json:"chat_usage"`
	ChatCost  *float64                  `json:"chat_cost"`
	Billing   Optional[ATPBillingState] `json:"billing,omitzero"`
	UsageCost Optional[ATPUsageCost]    `json:"usage_cost,omitzero"`
}

type ATPTurnSteered struct {
	TurnID        string           `json:"turn_id"`
	Text          string           `json:"text"`
	TargetAgentID Optional[string] `json:"target_agent_id,omitzero"`
}

type ATPAgentStarted struct {
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	Model    string  `json:"model"`
	ParentID *string `json:"parent_id"`
}

type ATPAgentPlan struct {
	Plan string `json:"plan"`
}

type ATPAgentTextDelta struct {
	Delta string `json:"delta"`
}

type ATPAgentTextDone struct {
	Text string `json:"text"`
}

type ATPAgentReasoningDelta struct {
	Delta string `json:"delta"`
}

type ATPAgentReasoningDone struct {
	Text string `json:"text"`
}

type ATPAgentDelegated struct {
	ChildInstanceID string `json:"child_instance_id"`
	ChildName       string `json:"child_name"`
	Mode            string `json:"mode"`
	Title           string `json:"title"`
	Task            string `json:"task"`
}

type ATPAgentToolStart struct {
	CallID   string                     `json:"call_id"`
	ToolName string                     `json:"tool_name"`
	Kind     ATPToolKind                `json:"kind"`
	Args     map[string]json.RawMessage `json:"args"`
}

type ATPAgentToolDelta struct {
	CallID string `json:"call_id"`
	Delta  string `json:"delta"`
}

type ATPAgentToolComplete struct {
	CallID  string          `json:"call_id"`
	Result  json.RawMessage `json:"result"`
	IsError bool            `json:"is_error"`
	Usage   ATPUsage        `json:"usage"`
}

type ATPAgentToolRequest struct {
	CallID    string                       `json:"call_id"`
	ToolName  string                       `json:"tool_name"`
	Kind      string                       `json:"kind"`
	Args      map[string]json.RawMessage   `json:"args"`
	Signature ATPAgentToolRequestSignature `json:"signature"`
}

type ATPAgentToolResult struct {
	CallID  string `json:"call_id"`
	Output  string `json:"output"`
	IsError bool   `json:"is_error"`
}

type ATPAgentEscalate struct {
	EscalationID string                     `json:"escalation_id"`
	AgentID      string                     `json:"agent_id"`
	Reason       string                     `json:"reason"`
	ToolName     string                     `json:"tool_name"`
	Args         map[string]json.RawMessage `json:"args"`
}

type ATPAgentEscalationResult struct {
	EscalationID string `json:"escalation_id"`
	Approve      bool   `json:"approve"`
	Reason       string `json:"reason"`
}

type ATPAgentArtifact struct {
	ArtifactID string `json:"artifact_id"`
	Name       string `json:"name"`
	MediaType  string `json:"media_type"`
	Size       int64  `json:"size"`
}

type ATPAgentContextCompacting struct {
	TurnID          Optional[string]  `json:"turn_id,omitzero"`
	Reason          Optional[string]  `json:"reason,omitzero"`
	ContextTokens   Optional[int64]   `json:"context_tokens,omitzero"`
	ThresholdTokens Optional[int64]   `json:"threshold_tokens,omitzero"`
	WindowTokens    Optional[int64]   `json:"window_tokens,omitzero"`
	Threshold       Optional[float64] `json:"threshold,omitzero"`
}

type ATPAgentContextCompacted struct {
	NewContext    string           `json:"new_context"`
	TurnID        Optional[string] `json:"turn_id,omitzero"`
	ContextTokens Optional[int64]  `json:"context_tokens,omitzero"`
}

type ATPAgentDone struct {
	Status         string           `json:"status"`
	Summary        string           `json:"summary"`
	Usage          ATPUsage         `json:"usage"`
	DurationMs     int64            `json:"duration_ms"`
	RequestedModel Optional[string] `json:"requested_model,omitzero"`
	Model          Optional[string] `json:"model,omitzero"`
}

type ATPAgentFailed struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ATPApprovalRequest struct {
	ApprovalID    string                     `json:"approval_id"`
	OriginAgentID string                     `json:"origin_agent_id"`
	ToolName      string                     `json:"tool_name"`
	Args          map[string]json.RawMessage `json:"args"`
	Reason        string                     `json:"reason"`
	Deadline      string                     `json:"deadline"`
}

type ATPApprovalResponse struct {
	ApprovalID string `json:"approval_id"`
	Approve    bool   `json:"approve"`
	Reason     string `json:"reason"`
}

type ATPApprovalExpire struct {
	ApprovalID string `json:"approval_id"`
}

type ATPBillingState struct {
	Status string                     `json:"status"`
	Extra  map[string]json.RawMessage `json:"-"`
}

func (v *ATPBillingState) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "BillingState", b); err != nil {
		return err
	}
	type plain ATPBillingState
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPBillingState(out)
	return nil
}
func (v ATPBillingState) MarshalJSON() ([]byte, error) {
	type plain ATPBillingState
	return marshalUnknown(plain(v), v.Extra)
}

type ATPCurrencyAmount struct {
	Currency string                     `json:"currency"`
	Amount   string                     `json:"amount"`
	Extra    map[string]json.RawMessage `json:"-"`
}

func (v *ATPCurrencyAmount) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "CurrencyAmount", b); err != nil {
		return err
	}
	type plain ATPCurrencyAmount
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"currency", "amount"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPCurrencyAmount(out)
	return nil
}
func (v ATPCurrencyAmount) MarshalJSON() ([]byte, error) {
	type plain ATPCurrencyAmount
	return marshalUnknown(plain(v), v.Extra)
}

type ATPUsageCost struct {
	Status                 string                     `json:"status"`
	Basis                  string                     `json:"basis"`
	ScopeStatus            string                     `json:"scope_status"`
	Revision               int64                      `json:"revision"`
	AsOf                   string                     `json:"as_of"`
	InvocationCount        int64                      `json:"invocation_count"`
	PendingInvocationCount int64                      `json:"pending_invocation_count"`
	Totals                 []ATPCurrencyAmount        `json:"totals"`
	PricedSubtotals        []ATPCurrencyAmount        `json:"priced_subtotals"`
	Extra                  map[string]json.RawMessage `json:"-"`
}

func (v *ATPUsageCost) UnmarshalJSON(b []byte) error {
	if err := validateDefinition("wiki.json", "UsageCost", b); err != nil {
		return err
	}
	type plain ATPUsageCost
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"status", "basis", "scope_status", "revision", "as_of", "invocation_count", "pending_invocation_count", "totals", "priced_subtotals"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPUsageCost(out)
	return nil
}
func (v ATPUsageCost) MarshalJSON() ([]byte, error) {
	type plain ATPUsageCost
	return marshalUnknown(plain(v), v.Extra)
}

type ATPChatAttachmentManifestEvent struct {
	AttachmentID string                     `json:"attachment_id"`
	OriginalName string                     `json:"original_name"`
	Revision     int64                      `json:"revision"`
	Sha256       string                     `json:"sha256"`
	Bytes        int64                      `json:"bytes"`
	Conversion   map[string]json.RawMessage `json:"conversion"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *ATPChatAttachmentManifestEvent) UnmarshalJSON(b []byte) error {
	type plain ATPChatAttachmentManifestEvent
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"attachment_id", "original_name", "revision", "sha256", "bytes", "conversion"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPChatAttachmentManifestEvent(out)
	return nil
}
func (v ATPChatAttachmentManifestEvent) MarshalJSON() ([]byte, error) {
	type plain ATPChatAttachmentManifestEvent
	return marshalUnknown(plain(v), v.Extra)
}

type ATPTaskStarted struct {
	TaskID       string                     `json:"task_id"`
	OriginTurnID string                     `json:"origin_turn_id"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *ATPTaskStarted) UnmarshalJSON(b []byte) error {
	type plain ATPTaskStarted
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"task_id", "origin_turn_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPTaskStarted(out)
	return nil
}
func (v ATPTaskStarted) MarshalJSON() ([]byte, error) {
	type plain ATPTaskStarted
	return marshalUnknown(plain(v), v.Extra)
}

type ATPTaskEvent struct {
	TaskID       string                     `json:"task_id"`
	OriginTurnID string                     `json:"origin_turn_id"`
	EventType    string                     `json:"event_type"`
	EventPayload map[string]json.RawMessage `json:"event_payload"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *ATPTaskEvent) UnmarshalJSON(b []byte) error {
	type plain ATPTaskEvent
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"task_id", "origin_turn_id", "event_type", "event_payload"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPTaskEvent(out)
	return nil
}
func (v ATPTaskEvent) MarshalJSON() ([]byte, error) {
	type plain ATPTaskEvent
	return marshalUnknown(plain(v), v.Extra)
}

type ATPTaskCompleted struct {
	TaskID        string                     `json:"task_id"`
	OriginTurnID  string                     `json:"origin_turn_id"`
	Status        string                     `json:"status"`
	Output        string                     `json:"output"`
	ErrorEnvelope string                     `json:"error"`
	ResultTurnID  Optional[string]           `json:"result_turn_id,omitzero"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ATPTaskCompleted) UnmarshalJSON(b []byte) error {
	type plain ATPTaskCompleted
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"task_id", "origin_turn_id", "status", "output", "error", "result_turn_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPTaskCompleted(out)
	return nil
}
func (v ATPTaskCompleted) MarshalJSON() ([]byte, error) {
	type plain ATPTaskCompleted
	return marshalUnknown(plain(v), v.Extra)
}

type ATPTaskFailed struct {
	TaskID        string                     `json:"task_id"`
	OriginTurnID  string                     `json:"origin_turn_id"`
	Status        string                     `json:"status"`
	Output        string                     `json:"output"`
	ErrorEnvelope string                     `json:"error"`
	ResultTurnID  Optional[string]           `json:"result_turn_id,omitzero"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ATPTaskFailed) UnmarshalJSON(b []byte) error {
	type plain ATPTaskFailed
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"task_id", "origin_turn_id", "status", "output", "error", "result_turn_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPTaskFailed(out)
	return nil
}
func (v ATPTaskFailed) MarshalJSON() ([]byte, error) {
	type plain ATPTaskFailed
	return marshalUnknown(plain(v), v.Extra)
}

type ATPTaskCancelled struct {
	TaskID        string                     `json:"task_id"`
	OriginTurnID  string                     `json:"origin_turn_id"`
	Status        string                     `json:"status"`
	Output        string                     `json:"output"`
	ErrorEnvelope string                     `json:"error"`
	ResultTurnID  Optional[string]           `json:"result_turn_id,omitzero"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func (v *ATPTaskCancelled) UnmarshalJSON(b []byte) error {
	type plain ATPTaskCancelled
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"task_id", "origin_turn_id", "status", "output", "error", "result_turn_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPTaskCancelled(out)
	return nil
}
func (v ATPTaskCancelled) MarshalJSON() ([]byte, error) {
	type plain ATPTaskCancelled
	return marshalUnknown(plain(v), v.Extra)
}

type ATPChatDeliveryQueued struct {
	TaskID       string                     `json:"task_id"`
	OriginTurnID string                     `json:"origin_turn_id"`
	Status       string                     `json:"status"`
	EventID      string                     `json:"event_id"`
	Extra        map[string]json.RawMessage `json:"-"`
}

func (v *ATPChatDeliveryQueued) UnmarshalJSON(b []byte) error {
	type plain ATPChatDeliveryQueued
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"task_id", "origin_turn_id", "status", "event_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPChatDeliveryQueued(out)
	return nil
}
func (v ATPChatDeliveryQueued) MarshalJSON() ([]byte, error) {
	type plain ATPChatDeliveryQueued
	return marshalUnknown(plain(v), v.Extra)
}

type ATPChatDeliveryDelivered struct {
	TaskID         string                     `json:"task_id"`
	OriginTurnID   string                     `json:"origin_turn_id"`
	Status         string                     `json:"status"`
	EventID        string                     `json:"event_id"`
	DeliveryTurnID string                     `json:"delivery_turn_id"`
	Extra          map[string]json.RawMessage `json:"-"`
}

func (v *ATPChatDeliveryDelivered) UnmarshalJSON(b []byte) error {
	type plain ATPChatDeliveryDelivered
	var out plain
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	extra, err := unknownFields(b, []string{"task_id", "origin_turn_id", "status", "event_id", "delivery_turn_id"})
	if err != nil {
		return err
	}
	out.Extra = extra
	*v = ATPChatDeliveryDelivered(out)
	return nil
}
func (v ATPChatDeliveryDelivered) MarshalJSON() ([]byte, error) {
	type plain ATPChatDeliveryDelivered
	return marshalUnknown(plain(v), v.Extra)
}

type ATPAgentToolRequestSignature struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Sig string `json:"sig"`
}

func (e Event) TypedPayload() (any, error) {
	switch e.Type {
	case "ping":
		var value ATPPing
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "chat.title.updated":
		var value ATPChatTitleUpdated
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "turn.started":
		var value ATPTurnStarted
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "turn.completed":
		var value ATPTurnCompleted
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "turn.cancelled":
		var value ATPTurnCancelled
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "turn.failed":
		var value ATPTurnFailed
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "turn.steered":
		var value ATPTurnSteered
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.started":
		var value ATPAgentStarted
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.plan":
		var value ATPAgentPlan
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.text.delta":
		var value ATPAgentTextDelta
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.text.done":
		var value ATPAgentTextDone
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.reasoning.delta":
		var value ATPAgentReasoningDelta
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.reasoning.done":
		var value ATPAgentReasoningDone
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.delegated":
		var value ATPAgentDelegated
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.tool.start":
		var value ATPAgentToolStart
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.tool.delta":
		var value ATPAgentToolDelta
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.tool.complete":
		var value ATPAgentToolComplete
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.tool.request":
		var value ATPAgentToolRequest
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.tool.result":
		var value ATPAgentToolResult
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.escalate":
		var value ATPAgentEscalate
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.escalation.result":
		var value ATPAgentEscalationResult
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.artifact":
		var value ATPAgentArtifact
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.context.compacting":
		var value ATPAgentContextCompacting
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.context.compacted":
		var value ATPAgentContextCompacted
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.done":
		var value ATPAgentDone
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "agent.failed":
		var value ATPAgentFailed
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "approval.request":
		var value ATPApprovalRequest
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "approval.response":
		var value ATPApprovalResponse
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "approval.expire":
		var value ATPApprovalExpire
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "task.started":
		var value ATPTaskStarted
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "task.event":
		var value ATPTaskEvent
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "task.completed":
		var value ATPTaskCompleted
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "task.failed":
		var value ATPTaskFailed
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "task.cancelled":
		var value ATPTaskCancelled
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "chat.delivery.queued":
		var value ATPChatDeliveryQueued
		err := json.Unmarshal(e.Data, &value)
		return value, err
	case "chat.delivery.delivered":
		var value ATPChatDeliveryDelivered
		err := json.Unmarshal(e.Data, &value)
		return value, err
	}
	return append(json.RawMessage(nil), e.Data...), nil
}
