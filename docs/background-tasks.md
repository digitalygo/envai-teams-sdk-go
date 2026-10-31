# Native background tasks (unreleased)

This source supports the reviewed native-task protocol. It does not establish compatibility of previously published packages. Enable it explicitly on the execution client:

```go
client, err := envai.New(envai.Config{
    BaseURL: baseURL, APIKey: apiKey, NativeTasks: true,
})
```

`NativeTasks` sends `X-Envai-Extensions: native-tasks-v1` on negotiated requests, including capabilities, SSE reconnects, admissions and tool output, alongside contract version 3. Native ATP responses must echo the extension; a missing acknowledgment raises a compatibility error and can mean an uncertain mutation outcome, so never blindly retry. The default client retains ordinary foreground behavior. `RunTurn` still stops at its accepted turn terminal. Setting `TurnOptions.StrictModel` or `ChatMessageRequest.StrictModel` to true requires this same native-enabled client and the advertised `strict_model` feature.

Use the generated typed APIs `DeliverWorkflowResult`, `DelegateToTeam`, `GetDetachedTask`, `CancelDetachedTask` and `ClassifyPhase` for direct requests. High-level helpers include `SubmitDelivery`, `DelegateTeam`, `GetTask`, `CancelTask`, `ClassifyNativePhase` and `WaitTask`. Mutations require their advertised capabilities. `notify_origin` requires `delegation_notify_origin`; workflows require `delegated_job_lifecycle`. Requests are not automatically retried. Preserve the same native mode when retrying an idempotency key; mode changes cannot replay an earlier admission.

## Cross-team work with signed tools

Build the tool runner with the same native-enabled client and a stable account namespace and durable call store. Then:

```go
notify := false // phase jobs collect output without waking the origin model
result, err := client.RunDelegatedJob(ctx, originChatID, envai.NativeDelegationRequest{
    WorkspaceID: workspaceID,
    EventID: uniqueEventID,
    OriginTurnID: originTurnID,
    TargetChatID: targetChatID,
    Prompt: "Review the complete change",
    NotifyOrigin: &notify,
    StrictModel: true,
}, envai.BackgroundOptions{Tools: runner})
if err != nil { return err }
switch result.Status {
case envai.TaskCompleted:
    // Persist result.Output and result.ResultTurnID before the next phase.
case envai.TaskFailed, envai.TaskCancelled, envai.TaskBlocked:
    // Preserve and handle this explicit outcome; it is not successful output.
}
```

`RunDelegatedJob` opens the target chat SSE and waits for its live boundary **before** submitting the delegation. It keeps that stream active through the final origin task outcome. Target coordinator and detached-child tools stay on the target stream; outputs go to that same chat. The initial receipt `turn_id` identifies only the first acknowledgment turn. The final task snapshot supplies full `output` and `result_turn_id` after all child cohorts and synthesis finish. This helper polls origin task status while servicing target SSE; it surfaces stream/tool failures and never treats an initial target terminal as the final job result.

For same-team detached work, open `OpenBackgroundSession` before admitting the message, keep it open through every required task terminal, then close it. For manual cross-team requests, likewise open the target session before `DelegateTeam` and close it only after `WaitTask` returns. Opening a session authorizes live tool dispatch for the whole selected execution chat, including later turns: use an isolated chat or application authorization policy where appropriate.

**Initial attachment is for new live work, not recovery of already-pending tool requests.** The initial replay is observed for resolved-call and terminal barriers but never executes historical tool requests. Reconcile an existing job before resuming it; a new session cannot promise to recover pending calls. `RunDelegatedJob` rejects nonzero initial cursors. If a duplicate event ID returns a coordinator turn present in initial replay, it returns an already-terminal task or an explicit reconciliation error for an active job; it does not silently wait with unserviced historical tools. Within an established session, bounded SSE reconnect uses the usual cursor/replay rules.

`task.event` is unwrapped from `event_payload`, with the outer agent/chat identity retained. Existing Ed25519 verification authenticates the original call ID, tool name and arguments before a handler or store claim. Call IDs deduplicate across foreground and task wrappers using the same account/chat store. Task barriers are distinct from origin-turn barriers, so an origin terminal does not suppress detached calls; the task's own terminal cancels active handlers and prevents late execution. `ToolInvocation.TaskID` and `OriginTurnID` identify wrapped work; its `TurnID` is empty because the wire wrapper has no turn. Signatures still do not authenticate the envelope or prove freshness: only use the trusted authenticated stream.

## Cancellation, classification and uncertain outcomes

`CancelTask(ctx, chatID, taskID, subtree)` sends explicit task cancellation. Cancelling a local context or closing a stream only stops local work. Cancelling an origin turn does not cancel detached tasks. Cancellation is best effort and cannot undo external effects. A cancelled task or failed job remains an explicit result state.

For phase classification, submit a unique `request_id`, non-null JSON `task_context`, positive `attempt`, `phase_id`, and full final output to `ClassifyPhase` (or `ClassifyNativePhase`). Save successful responses. A 502/503 or lost response is uncertain admission: pause and reconcile rather than retrying a potentially paid classifier call.

Task state, delivery receipts and deduplication last while the chat remains resident. `task_unknown` after restart/eviction means unknown outcome, never proof of nonexecution. `delivery_error` does not invalidate the task's output. No automatic paid provider calls, provisioning, or remote resource changes occur during the offline SDK tests.
