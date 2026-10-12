package envai

import "encoding/json"

type ATPAgent struct {
	InstanceID string           `json:"instance_id"`
	Name       string           `json:"name"`
	Role       string           `json:"role"`
	ParentID   *string          `json:"parent_id"`
	Model      Optional[string] `json:"model,omitzero"`
}

type ATPUsage struct {
	InputTokens   int64           `json:"input_tokens"`
	OutputTokens  int64           `json:"output_tokens"`
	CostUSD       float64         `json:"cost_usd"`
	DurationMs    int64           `json:"duration_ms"`
	SpriteMs      Optional[int64] `json:"sprite_ms,omitzero"`
	SpriteRamGbMs Optional[int64] `json:"sprite_ram_gb_ms,omitzero"`
	ToolCalls     Optional[int64] `json:"tool_calls,omitzero"`
}

type ATPToolKind string

type ATPPing map[string]json.RawMessage

type ATPTurnStarted struct {
	TurnID string `json:"turn_id"`
	Input  string `json:"input"`
}

type ATPTurnCompleted struct {
	TurnID       string   `json:"turn_id"`
	Usage        ATPUsage `json:"usage"`
	Cost         float64  `json:"cost"`
	ChatUsage    ATPUsage `json:"chat_usage"`
	ChatCost     float64  `json:"chat_cost"`
	FinishReason string   `json:"finish_reason"`
}

type ATPTurnCancelled struct {
	TurnID    string   `json:"turn_id"`
	Reason    string   `json:"reason"`
	Usage     ATPUsage `json:"usage"`
	Cost      float64  `json:"cost"`
	ChatUsage ATPUsage `json:"chat_usage"`
	ChatCost  float64  `json:"chat_cost"`
}

type ATPTurnFailed struct {
	TurnID    string   `json:"turn_id"`
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Usage     ATPUsage `json:"usage"`
	Cost      float64  `json:"cost"`
	ChatUsage ATPUsage `json:"chat_usage"`
	ChatCost  float64  `json:"chat_cost"`
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
	Status     string   `json:"status"`
	Summary    string   `json:"summary"`
	Usage      ATPUsage `json:"usage"`
	DurationMs int64    `json:"duration_ms"`
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
	}
	return append(json.RawMessage(nil), e.Data...), nil
}
