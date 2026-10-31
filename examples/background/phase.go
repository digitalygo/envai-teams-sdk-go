// Package background demonstrates a phase job using an already-configured native
// client and signed-tool runner. Tests compile it without executing provider calls.
package background

import (
	"context"
	"encoding/json"
	"fmt"
	envai "github.com/digitalygo/envai-teams-sdk-go"
)

// RunPhase requires client.Config.NativeTasks and a runner constructed from that
// same client. EventID and RequestID must be unique and persisted by the caller.
func RunPhase(ctx context.Context, client *envai.Client, runner *envai.ToolRunner, originChat, workspace, originTurn, targetChat, eventID, requestID string) (envai.NativeTask, envai.PhaseClassificationResult, error) {
	notify := false
	task, err := client.RunDelegatedJob(ctx, originChat, envai.NativeDelegationRequest{
		WorkspaceID: workspace, EventID: eventID, OriginTurnID: originTurn, TargetChatID: targetChat,
		Prompt: "Review the requested change and return the complete result", NotifyOrigin: &notify, StrictModel: true,
	}, envai.BackgroundOptions{Tools: runner})
	if err != nil {
		return task, envai.PhaseClassificationResult{}, err
	}
	if task.Status != envai.TaskCompleted {
		return task, envai.PhaseClassificationResult{}, fmt.Errorf("phase job ended with %s", task.Status)
	}
	decision, err := client.ClassifyPhase(ctx, envai.ClassifyPhaseParams{ChatID: originChat, Body: envai.PhaseClassificationRequest{
		RequestID: requestID, TaskContext: json.RawMessage(`{"phase":"review"}`), PhaseID: "review", Attempt: 1, FinalOutput: task.Output,
	}})
	// Save successful task/decision responses; an uncertain classifier error must
	// pause the workflow rather than automatically reusing the request ID.
	return task, decision, err
}
