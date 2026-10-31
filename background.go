package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"
)

// TaskStatus is a resident task outcome. A missing task is an unknown outcome.
type DeliveryStatus string

const (
	DeliveryCompleted DeliveryStatus = "completed"
	DeliveryBlocked   DeliveryStatus = "blocked"
	DeliveryFailed    DeliveryStatus = "failed"
	DeliveryCancelled DeliveryStatus = "cancelled"
)

type DeliveryReceiptStatus string

const (
	DeliveryQueued    DeliveryReceiptStatus = "queued"
	DeliveryDelivered DeliveryReceiptStatus = "delivered"
	ReceiptCancelled  DeliveryReceiptStatus = "cancelled"
	ReceiptFailed     DeliveryReceiptStatus = "failed"
)

type PhaseDecision string

const (
	PhaseBlocked    PhaseDecision = "BLOCKED"
	PhaseNotBlocked PhaseDecision = "NOT BLOCKED"
)

type TaskStatus string

const (
	TaskRunning   TaskStatus = "running"
	TaskBlocked   TaskStatus = "blocked"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
	TaskCancelled TaskStatus = "cancelled"
)

type NativeTask struct {
	TaskID        string     `json:"task_id"`
	OriginTurnID  string     `json:"origin_turn_id"`
	Status        TaskStatus `json:"status"`
	ParentTaskID  string     `json:"parent_task_id,omitempty"`
	TargetChatID  string     `json:"target_chat_id,omitempty"`
	Output        string     `json:"output,omitempty"`
	Error         string     `json:"error,omitempty"`
	DeliveryError string     `json:"delivery_error,omitempty"`
	ResultTurnID  string     `json:"result_turn_id,omitempty"`
}

func (t NativeTask) Terminal() bool {
	return t.Status == TaskBlocked || t.Status == TaskCompleted || t.Status == TaskFailed || t.Status == TaskCancelled
}

type NativeDelegationRequest struct {
	WorkspaceID  string `json:"workspace_id"`
	EventID      string `json:"event_id"`
	OriginTurnID string `json:"origin_turn_id"`
	TargetChatID string `json:"target_chat_id"`
	Prompt       string `json:"prompt"`
	NotifyOrigin *bool  `json:"notify_origin,omitempty"`
	StrictModel  bool   `json:"strict_model,omitempty"`
}
type NativeDelegationReceipt struct {
	TaskID       string `json:"task_id"`
	TargetChatID string `json:"target_chat_id"`
	// TurnID identifies the initial acknowledgment, never the final job outcome.
	TurnID string `json:"turn_id"`
}
type NativeCancellation struct {
	TaskID  string     `json:"task_id"`
	Status  TaskStatus `json:"status"`
	Subtree bool       `json:"subtree"`
}
type NativePhaseRequest struct {
	RequestID   string          `json:"request_id"`
	TaskContext json.RawMessage `json:"task_context"`
	PhaseID     string          `json:"phase_id"`
	Attempt     int             `json:"attempt"`
	FinalOutput string          `json:"final_output"`
}
type NativePhaseResult struct {
	Decision     PhaseDecision `json:"decision"`
	Model        string        `json:"model"`
	RequestID    string        `json:"request_id"`
	InvocationID string        `json:"invocation_id"`
}

func (c *Client) requireFeatures(ctx context.Context, names ...string) error {
	if c.config.LegacyReadOnly || !c.config.NativeTasks {
		return &CompatibilityError{Reason: "background APIs require contract 3 and Config.NativeTasks"}
	}
	caps, err := c.GetCapabilities(ctx, GetCapabilitiesParams{})
	if err != nil {
		return err
	}
	for _, name := range names {
		if !featureAdvertised(caps.Features, name) {
			return &CompatibilityError{Reason: "server does not advertise " + name}
		}
	}
	return nil
}
func (c *Client) SubmitDelivery(ctx context.Context, chatID string, request NativeDeliveryRequest) (NativeDeliveryReceipt, error) {
	var out NativeDeliveryReceipt
	if err := c.requireFeatures(ctx, "native_deliveries"); err != nil {
		return out, err
	}
	err := c.call(ctx, "deliverWorkflowResult", map[string]any{"chat_id": chatID}, request, &out)
	return out, err
}
func (c *Client) DelegateTeam(ctx context.Context, chatID string, request NativeDelegationRequest) (NativeDelegationReceipt, error) {
	var out NativeDelegationReceipt
	names := []string{"cross_team_delegation", "delegated_job_lifecycle"}
	if request.NotifyOrigin != nil {
		names = append(names, "delegation_notify_origin")
	}
	if request.StrictModel {
		names = append(names, "strict_model")
	}
	if err := c.requireFeatures(ctx, names...); err != nil {
		return out, err
	}
	err := c.call(ctx, "delegateToTeam", map[string]any{"chat_id": chatID}, request, &out)
	return out, err
}
func (c *Client) GetTask(ctx context.Context, chatID, taskID string) (NativeTask, error) {
	var out NativeTask
	if err := c.requireFeatures(ctx, "detached_background"); err != nil {
		return out, err
	}
	err := c.call(ctx, "getDetachedTask", map[string]any{"chat_id": chatID, "task_id": taskID}, nil, &out)
	if err == nil && out.TaskID != taskID {
		err = &ProtocolError{"task response identity mismatch"}
	}
	return out, err
}
func (c *Client) CancelTask(ctx context.Context, chatID, taskID string, subtree bool) (NativeCancellation, error) {
	var out NativeCancellation
	if err := c.requireFeatures(ctx, "detached_background"); err != nil {
		return out, err
	}
	err := c.call(ctx, "cancelDetachedTask", map[string]any{"chat_id": chatID, "task_id": taskID}, map[string]bool{"subtree": subtree}, &out)
	if err == nil && (out.TaskID != taskID || out.Status != TaskCancelled || out.Subtree != subtree) {
		err = &ProtocolError{"invalid cancellation response"}
	}
	return out, err
}

// ClassifyPhase never retries admission, including an uncertain 502/503 result.
func (c *Client) ClassifyNativePhase(ctx context.Context, chatID string, request NativePhaseRequest) (NativePhaseResult, error) {
	var out NativePhaseResult
	if request.Attempt < 1 || len(request.TaskContext) == 0 || string(bytes.TrimSpace(request.TaskContext)) == "null" || !json.Valid(request.TaskContext) {
		return out, errors.New("envai: classification requires positive attempt and non-null JSON task context")
	}
	if err := c.requireFeatures(ctx, "phase_classification"); err != nil {
		return out, err
	}
	err := c.call(ctx, "classifyPhase", map[string]any{"chat_id": chatID}, request, &out)
	if err == nil && out.RequestID != request.RequestID {
		err = &ProtocolError{"classification request identity mismatch"}
	}
	return out, err
}

// WaitTask observes the final resident task snapshot, including failed/cancelled
// states. It never substitutes the initial coordinator turn for a job outcome.
// Keep a BackgroundSession on the execution chat alive while waiting for tools.
func (c *Client) WaitTask(ctx context.Context, chatID, taskID string, interval time.Duration) (NativeTask, error) {
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}
	if err := c.requireFeatures(ctx, "delegated_job_lifecycle"); err != nil {
		return NativeTask{}, err
	}
	for {
		task, err := c.GetTask(ctx, chatID, taskID)
		if err != nil {
			return task, err
		}
		if task.TaskID != taskID {
			return task, &ProtocolError{"task response identity mismatch"}
		}
		if task.Terminal() {
			return task, nil
		}
		if task.Status != TaskRunning {
			return task, &ProtocolError{"unknown task status"}
		}
		if err = wait(ctx, interval); err != nil {
			return task, err
		}
	}
}

// validateNativeIdentity also protects generated low-level operation methods.
func validateNativeIdentity(id string, params map[string]any, body any, raw []byte) error {
	var result map[string]json.RawMessage
	if id != "getDetachedTask" && id != "cancelDetachedTask" && id != "delegateToTeam" && id != "classifyPhase" && id != "deliverWorkflowResult" {
		return nil
	}
	if json.Unmarshal(raw, &result) != nil {
		return &ProtocolError{"invalid native response"}
	}
	requestBytes, _ := json.Marshal(body)
	var request map[string]json.RawMessage
	_ = json.Unmarshal(requestBytes, &request)
	matches := func(key string, expected any) bool {
		var actual string
		return json.Unmarshal(result[key], &actual) == nil && actual != "" && actual == expected
	}
	switch id {
	case "getDetachedTask", "cancelDetachedTask":
		if !matches("task_id", params["task_id"]) {
			return &ProtocolError{"task response identity mismatch"}
		}
		if id == "cancelDetachedTask" {
			var status string
			var subtree bool
			if len(result) != 3 || (string(result["subtree"]) != "true" && string(result["subtree"]) != "false") || json.Unmarshal(result["status"], &status) != nil || status != "cancelled" || json.Unmarshal(result["subtree"], &subtree) != nil {
				return &ProtocolError{"invalid cancellation response"}
			}
			var expected bool
			if json.Unmarshal(request["subtree"], &expected) != nil || subtree != expected {
				return &ProtocolError{"cancellation subtree mismatch"}
			}
		}
	default:
		key := map[string]string{"delegateToTeam": "target_chat_id", "classifyPhase": "request_id", "deliverWorkflowResult": "event_id"}[id]
		var expected string
		_ = json.Unmarshal(request[key], &expected)
		if !matches(key, expected) {
			return &ProtocolError{"native response correlation mismatch"}
		}
	}
	return nil
}
