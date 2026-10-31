package envai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var backgroundCapabilities = strings.Replace(testCapabilities, `"features":[]`, `"features":["native_deliveries","cross_team_delegation","detached_background","phase_classification","delegated_job_lifecycle","delegation_notify_origin","strict_model"]`, 1)

func backgroundHarness(t *testing.T, handler func(context.Context, ToolInvocation) (ToolResult, error), schema string, status int) (*ToolRunner, *atomic.Int32, *ToolOutputRequest) {
	t.Helper()
	f := fixture[signatureFixtures](t, "signatures.json")
	posts := &atomic.Int32{}
	out := &ToolOutputRequest{}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, backgroundCapabilities)
			return
		}
		if r.URL.Path == "/.well-known/jwks.json" {
			json.NewEncoder(w).Encode(f.JWKS)
			return
		}
		if r.Header.Get("X-Envai-Extensions") != "native-tasks-v1" {
			t.Error("background output omitted extension")
		}
		posts.Add(1)
		json.NewDecoder(r.Body).Decode(out)
		w.WriteHeader(status)
	})
	c.config.NativeTasks = true
	registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(schema), Handler: handler})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewToolRunner(c, registry, NewMemoryCallStore(), RunnerOptions{Namespace: "background-harness"})
	if err != nil {
		t.Fatal(err)
	}
	return runner, posts, out
}
func wrappedTool(t *testing.T) Event {
	e := signedEvent(t, `{}`)
	e.V = 3
	e.Type = "task.event"
	e.TurnID = nil
	e.Data, _ = json.Marshal(NativeTaskEvent{TaskID: "task_1", OriginTurnID: "cmpl_01ARZ3NDEKTSV4RRFFQ69G5FAV", EventType: "agent.tool.request", EventPayload: e.Data})
	return e
}
func taskTerminal(t *testing.T, typ string) Event {
	e := wrappedTool(t)
	e.Type = typ
	e.Agent = nil
	e.Data, _ = json.Marshal(map[string]string{"task_id": "task_1", "origin_turn_id": "cmpl_01ARZ3NDEKTSV4RRFFQ69G5FAV", "status": strings.TrimPrefix(typ, "task."), "output": "final", "error": "", "result_turn_id": "turn_final"})
	return e
}
func TestBackgroundWrappedToolsAfterOriginTerminal(t *testing.T) {
	var executed atomic.Int32
	runner, posts, _ := backgroundHarness(t, func(ctx context.Context, i ToolInvocation) (ToolResult, error) {
		executed.Add(1)
		if i.TaskID != "task_1" || i.OriginTurnID != "cmpl_01ARZ3NDEKTSV4RRFFQ69G5FAV" || i.TurnID != "" {
			t.Errorf("wrong invocation identity %+v", i)
		}
		return ToolResult{Output: "ok"}, nil
	}, `{"type":"object"}`, 204)
	e := wrappedTool(t)
	terminal := fixture[[]Event](t, "events-valid.json")
	for _, v := range terminal {
		if v.Terminal() {
			turn := "cmpl_01ARZ3NDEKTSV4RRFFQ69G5FAV"
			v.TurnID = &turn
			if err := runner.ObserveBackground(context.Background(), v); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	for range 2 {
		if err := runner.HandleBackground(context.Background(), e, e.ChatID); err != nil {
			t.Fatal(err)
		}
	}
	if executed.Load() != 1 || posts.Load() != 1 {
		t.Fatal(executed.Load(), posts.Load())
	}
}
func TestBackgroundTaskTerminalAndBadSignature(t *testing.T) {
	for _, terminal := range []string{"task.completed", "task.failed", "task.cancelled"} {
		t.Run(terminal, func(t *testing.T) {
			runner, posts, _ := backgroundHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
				t.Fatal("executed after terminal")
				return ToolResult{}, nil
			}, `{"type":"object"}`, 204)
			if err := runner.ObserveBackground(context.Background(), taskTerminal(t, terminal)); err != nil {
				t.Fatal(err)
			}
			e := wrappedTool(t)
			if err := runner.HandleBackground(context.Background(), e, e.ChatID); err != nil {
				t.Fatal(err)
			}
			if posts.Load() != 0 {
				t.Fatal("posted")
			}
		})
	}
	runner, posts, _ := backgroundHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
		t.Fatal("executed bad signature")
		return ToolResult{}, nil
	}, `{"type":"object"}`, 204)
	e := wrappedTool(t)
	e.Data = []byte(strings.Replace(string(e.Data), `"args":{}`, `"args":{"tampered":true}`, 1))
	if err := runner.HandleBackground(context.Background(), e, e.ChatID); err == nil {
		t.Fatal("accepted tampered signed payload")
	}
	if posts.Load() != 0 {
		t.Fatal("posted")
	}
}
func TestWaitTaskFinalStatesAndUnknown(t *testing.T) {
	for _, status := range []string{"completed", "failed", "cancelled", "unknown", "missing", "wrong_identity"} {
		t.Run(status, func(t *testing.T) {
			var polls atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					fmt.Fprint(w, backgroundCapabilities)
					return
				}
				if status == "missing" {
					w.WriteHeader(404)
					fmt.Fprint(w, `{"error":{"type":"not_found","code":"task_unknown"}}`)
					return
				}
				n := polls.Add(1)
				state := status
				if n == 1 {
					state = "running"
				}
				id := "job"
				if status == "wrong_identity" {
					id = "other"
				}
				fmt.Fprintf(w, `{"task_id":%q,"origin_turn_id":"origin","status":%q,"output":"final result","result_turn_id":"synthesis"}`, id, state)
			})
			c.config.NativeTasks = true
			result, err := c.WaitTask(context.Background(), "chat", "job", time.Millisecond)
			if status == "unknown" || status == "missing" || status == "wrong_identity" {
				if err == nil {
					t.Fatal("accepted invalid/unknown outcome")
				}
				return
			}
			if err != nil || string(result.Status) != status || result.Output != "final result" || result.ResultTurnID != "synthesis" || polls.Load() != 2 {
				t.Fatal(result, err, polls.Load())
			}
		})
	}
}

func TestRunDelegatedJobKeepsTargetToolsAliveAfterAck(t *testing.T) {
	wrapped := wrappedTool(t)
	target := wrapped.ChatID
	origin := "cht_01ARZ3NDEKTSV4RRFFQ69G5FAW"
	var wrapper NativeTaskEvent
	if err := json.Unmarshal(wrapped.Data, &wrapper); err != nil {
		t.Fatal(err)
	}
	initialTurn := wrapper.OriginTurnID
	finalTurn := "cmpl_01ARZ3NDEKTSV4RRFFQ69G5FAX"
	var signedRequest ToolRequest
	if err := json.Unmarshal(wrapper.EventPayload, &signedRequest); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	var completed atomic.Bool
	var calls atomic.Int32
	f := fixture[signatureFixtures](t, "signatures.json")
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/capabilities"):
			if r.Header.Get("X-Envai-Extensions") != "native-tasks-v1" {
				t.Error("missing extension preflight")
			}
			fmt.Fprint(w, backgroundCapabilities)
		case r.URL.Path == "/.well-known/jwks.json":
			json.NewEncoder(w).Encode(f.JWKS)
		case strings.HasSuffix(r.URL.Path, "/events"):
			if r.Header.Get("X-Envai-Extensions") != "native-tasks-v1" {
				t.Error("missing extension SSE")
			}
			if !strings.Contains(r.URL.Path, target) {
				t.Error("wrong SSE chat")
			}
			w.Header().Set("Content-Type", "text/event-stream")
			emit := func(e Event) { b, _ := json.Marshal(e); fmt.Fprintf(w, "data: %s\n\n", b); w.(http.Flusher).Flush() }
			ping := fixture[[]Event](t, "events-valid.json")[0]
			ping.V = 3
			emit(ping)
			select {
			case <-started:
			case <-r.Context().Done():
				return
			}
			for _, e := range fixture[[]Event](t, "events-valid.json") {
				if e.Terminal() {
					e.V = 3
					e.Seq = 1
					e.TurnID = &initialTurn
					emit(e)
					break
				}
			}
			wrapped.Seq = 2
			emit(wrapped)
			<-r.Context().Done()
		case strings.HasSuffix(r.URL.Path, "/delegations"):
			w.WriteHeader(202)
			fmt.Fprintf(w, `{"task_id":"job","target_chat_id":%q,"turn_id":%q}`, target, initialTurn)
			close(started)
		case strings.HasSuffix(r.URL.Path, "/tool-output"):
			if r.Header.Get("X-Envai-Extensions") != "native-tasks-v1" {
				t.Error("missing extension tool output")
			}
			if !strings.Contains(r.URL.Path, target) {
				t.Error("wrong output chat")
			}
			var output ToolOutputRequest
			if err := json.NewDecoder(r.Body).Decode(&output); err != nil {
				t.Error(err)
			}
			if output.CallID != signedRequest.CallID || output.Output != "answer" {
				t.Errorf("wrong tool output: call=%q output=%q", output.CallID, output.Output)
			}
			calls.Add(1)
			completed.Store(true)
			w.WriteHeader(204)
		case strings.Contains(r.URL.Path, "/tasks/"):
			if r.URL.Path != "/api/v1/atp/v3/chats/"+origin+"/tasks/job" {
				t.Error("task outcome was not read from origin", r.URL.Path)
			}
			state := "running"
			if completed.Load() {
				state = "completed"
			}
			fmt.Fprintf(w, `{"task_id":"job","origin_turn_id":"origin","status":%q,"output":"final synthesis","result_turn_id":%q}`, state, finalTurn)
		default:
			t.Errorf("unexpected route %s", r.URL.Path)
		}
	})
	c.config.NativeTasks = true
	registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(context.Context, ToolInvocation) (ToolResult, error) { return ToolResult{Output: "answer"}, nil }})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewToolRunner(c, registry, NewMemoryCallStore(), RunnerOptions{Namespace: "background-test"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	notify := false
	result, err := c.RunDelegatedJob(ctx, origin, NativeDelegationRequest{WorkspaceID: "workspace", EventID: "event", OriginTurnID: "origin", TargetChatID: target, Prompt: "work", NotifyOrigin: &notify}, BackgroundOptions{Tools: runner})
	if err != nil || result.ResultTurnID != finalTurn || result.ResultTurnID == initialTurn || result.Output != "final synthesis" || calls.Load() != 1 {
		t.Fatal(result, err, calls.Load())
	}
}
func TestNativeNegotiationAndNoClassificationRetry(t *testing.T) {
	var mutations atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			writeCapabilities(w)
			return
		}
		mutations.Add(1)
	})
	c.config.NativeTasks = true
	_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{Message: "native"}})
	if err == nil || mutations.Load() != 0 {
		t.Fatal("native admission did not fail closed")
	}
	_, err = c.DelegateTeam(context.Background(), "chat", NativeDelegationRequest{StrictModel: true})
	if err == nil || mutations.Load() != 0 {
		t.Fatal("missing capability admitted")
	}
	_, err = c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{Message: "hello", StrictModel: Present(true)}})
	if err == nil || mutations.Load() != 0 {
		t.Fatal("strict model admitted")
	}
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, backgroundCapabilities)
			return
		}
		mutations.Add(1)
		w.WriteHeader(503)
		fmt.Fprint(w, `{"error":{"type":"unavailable"}}`)
	})
	c.config.RetryAttempts = 3
	c.config.NativeTasks = true
	_, err = c.ClassifyNativePhase(context.Background(), "chat", NativePhaseRequest{RequestID: "once", TaskContext: json.RawMessage(`{}`), PhaseID: "review", Attempt: 1, FinalOutput: "full"})
	if err == nil || mutations.Load() != 1 {
		t.Fatal("classification retried", mutations.Load(), err)
	}
}
func TestCompletionDiagnosticsAndNullEscape(t *testing.T) {
	event := signedEvent(t, `{}`)
	event.V = 3
	event.Type = "agent.done"
	event.Data = json.RawMessage(`{"status":"completed","summary":"done","duration_ms":1,"usage":{"input_tokens":1,"output_tokens":1,"cost_usd":0,"duration_ms":1},"requested_model":"configured","model":"provider"}`)
	wire, _ := json.Marshal(event)
	decoded, err := DecodeEvent(wire)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := decoded.TypedPayload()
	if err != nil || payload.(ATPAgentDone).RequestedModel.Value != "configured" {
		t.Fatal(payload, err)
	}
	event.V = 2
	wire, _ = json.Marshal(event)
	if _, err = DecodeEvent(wire); err == nil {
		t.Fatal("legacy schema relaxed")
	}
	if err = validateDefinition("wiki.json", "ReportedProblem", []byte(`{"file_id":"file","message":"ok"}`)); err != nil {
		t.Fatal(err)
	}
	if err = validateDefinition("wiki.json", "ReportedProblem", []byte(`{"file_id":"file","message":"bad\u0000text"}`)); err == nil {
		t.Fatal("null character accepted")
	}
}
func TestNativeExtensionDefaultAbsentAndStrictOptIn(t *testing.T) {
	var strict atomic.Bool
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		want := ""
		if strict.Load() {
			want = "native-tasks-v1"
		}
		if r.Header.Get("X-Envai-Extensions") != want {
			t.Errorf("extension=%q want=%q", r.Header.Get("X-Envai-Extensions"), want)
		}
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, backgroundCapabilities)
			return
		}
		fmt.Fprint(w, `{"turn_id":"accepted","status":"accepted"}`)
	})
	if _, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{Message: "foreground"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{Message: "strict", StrictModel: Present(true)}}); err == nil {
		t.Fatal("strict accepted without explicit mode")
	}
	strict.Store(true)
	c.config.NativeTasks = true
	if _, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{Message: "strict", StrictModel: Present(true)}}); err != nil {
		t.Fatal(err)
	}

}
func TestTaskTerminalWinsAuthorizationRace(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	runner, posts, _ := backgroundHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
		t.Error("handler ran after task terminal")
		return ToolResult{Output: "bad"}, nil
	}, `{"type":"object"}`, 204)
	runner.options.Authorize = func(context.Context, ToolInvocation, string) error { close(entered); <-release; return nil }
	e := wrappedTool(t)
	done := make(chan error, 1)
	go func() { done <- runner.HandleBackground(context.Background(), e, e.ChatID) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("authorization did not begin")
	}
	if err := runner.ObserveBackground(context.Background(), taskTerminal(t, "task.cancelled")); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if posts.Load() != 0 {
		t.Fatal("late output posted")
	}
}
func TestNativeCancellationAndDeliveryAPI(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Envai-Extensions") != "native-tasks-v1" {
			t.Error("missing native header")
		}
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, backgroundCapabilities)
			return
		}
		var request map[string]any
		json.NewDecoder(r.Body).Decode(&request)
		if strings.HasSuffix(r.URL.Path, "/cancel") {
			if request["subtree"] != true {
				t.Error(request)
			}
			fmt.Fprint(w, `{"task_id":"task","status":"cancelled","subtree":true}`)
			return
		}
		if request["status"] != "blocked" {
			t.Error(request)
		}
		fmt.Fprint(w, `{"event_id":"event","status":"queued","duplicate":false}`)
	})
	c.config.NativeTasks = true
	result, err := c.CancelTask(context.Background(), "chat", "task", true)
	if err != nil || result.Status != TaskCancelled || !result.Subtree {
		t.Fatal(result, err)
	}
	receipt, err := c.SubmitDelivery(context.Background(), "chat", NativeDeliveryRequest{WorkspaceID: "workspace", EventID: "event", OriginTurnID: "origin", TaskID: "task", Status: DeliveryBlocked, Output: "needs review"})
	if err != nil || receipt.Status != DeliveryQueued {
		t.Fatal(receipt, err)
	}
}
func TestNativeStreamReconnectRetainsExtension(t *testing.T) {
	var connects atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Envai-Extensions") != "native-tasks-v1" {
			t.Error("reconnect lost native mode")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		ping := fixture[[]Event](t, "events-valid.json")[0]
		ping.V = 3
		b, _ := json.Marshal(ping)
		fmt.Fprintf(w, "data: %s\n\n", b)
		if connects.Add(1) > 1 {
			e := wrappedTool(t)
			e.Seq = 1
			b, _ = json.Marshal(e)
			fmt.Fprintf(w, "data: %s\n\n", b)
		}
	})
	c.config.NativeTasks = true
	s, err := c.Events(context.Background(), wrappedTool(t).ChatID, StreamOptions{MaxReconnects: 1, ReconnectDelay: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, want := range []string{"ping", "ping", "task.event"} {
		e, err := s.Next()
		if err != nil || e.Type != want {
			t.Fatal(e, err)
		}
	}
	if connects.Load() != 2 {
		t.Fatal(connects.Load())
	}
}
func TestDuplicateDelegationResumesRunningAndTerminal(t *testing.T) {
	for _, state := range []TaskStatus{TaskRunning, TaskCompleted, TaskFailed, TaskCancelled} {
		t.Run(string(state), func(t *testing.T) {
			old := signedEvent(t, `{}`)
			old.V = 3
			old.Seq = 1
			target := old.ChatID
			var polls atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/capabilities"):
					fmt.Fprint(w, backgroundCapabilities)
				case strings.HasSuffix(r.URL.Path, "/events"):
					w.Header().Set("Content-Type", "text/event-stream")
					b, _ := json.Marshal(old)
					fmt.Fprintf(w, "data: %s\n\n", b)
					ping := fixture[[]Event](t, "events-valid.json")[0]
					ping.V = 3
					b, _ = json.Marshal(ping)
					fmt.Fprintf(w, "data: %s\n\n", b)
					w.(http.Flusher).Flush()
					<-r.Context().Done()
				case strings.HasSuffix(r.URL.Path, "/delegations"):
					fmt.Fprintf(w, `{"task_id":"existing","target_chat_id":%q,"turn_id":%q}`, target, *old.TurnID)
				case strings.Contains(r.URL.Path, "/tasks/"):
					next := state
					if polls.Add(1) > 1 {
						next = TaskCompleted
					}
					fmt.Fprintf(w, `{"task_id":"existing","origin_turn_id":"origin","status":%q,"output":"previous final","result_turn_id":"final"}`, next)
				default:
					t.Error("unexpected request", r.URL.Path)
				}
			})
			c.config.NativeTasks = true
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			options := BackgroundOptions{}
			if state != TaskRunning {
				registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(context.Context, ToolInvocation) (ToolResult, error) {
					t.Error("terminal duplicate executed historical tool")
					return ToolResult{Output: "bad"}, nil
				}})
				if err != nil {
					t.Fatal(err)
				}
				store, err := NewFileCallStore(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				options.Tools, err = NewToolRunner(c, registry, store, RunnerOptions{Namespace: "terminal-duplicate"})
				if err != nil {
					t.Fatal(err)
				}
			}
			result, err := c.RunDelegatedJob(ctx, "origin", NativeDelegationRequest{TargetChatID: target, EventID: "duplicate"}, options)
			want := state
			if want == TaskRunning {
				want = TaskCompleted
			}
			if err != nil || result.Output != "previous final" || result.Status != want {
				t.Fatal(result, err)
			}
		})
	}
}
func TestNativeResponseAcknowledgmentFailsClosed(t *testing.T) {
	for _, ack := range []string{"", "other"} {
		t.Run(ack, func(t *testing.T) {
			var posts atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					fmt.Fprint(w, backgroundCapabilities)
					return
				}
				posts.Add(1)
				w.Header().Del("X-Envai-Extensions")
				if ack != "" {
					w.Header().Set("X-Envai-Extensions", ack)
				}
				fmt.Fprint(w, `{"turn_id":"accepted","status":"accepted"}`)
			})
			c.config.NativeTasks = true
			c.config.RetryAttempts = 3
			_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{Message: "native"}})
			if err == nil || !strings.Contains(err.Error(), "uncertain") || posts.Load() != 1 {
				t.Fatal(err, posts.Load())
			}
		})
	}
}
