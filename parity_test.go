package envai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestForegroundReplayResolvedToolNeverExecutes(t *testing.T) {
	request := fixture[struct {
		Request Event `json:"request"`
	}](t, "parity/fixtures.json").Request
	request.V = 3
	request.Seq = 1
	var payload ToolRequest
	json.Unmarshal(request.Data, &payload)
	resolved := request
	resolved.ID = "evt_01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	resolved.Seq = 2
	resolved.Type = "agent.tool.result"
	resolved.Data, _ = json.Marshal(map[string]any{"call_id": payload.CallID, "output": `{"code":"tool_timeout","message":"expired while client offline"}`, "is_error": true})
	keys := fixture[signatureFixtures](t, "signatures.json")
	var effects atomic.Int32
	executed := make(chan struct{})
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/capabilities"):
			fmt.Fprint(w, testCapabilities)
		case r.URL.Path == "/.well-known/jwks.json":
			json.NewEncoder(w).Encode(keys.JWKS)
		case strings.HasSuffix(r.URL.Path, "/events"):
			w.Header().Set("Content-Type", "text/event-stream")
			for _, event := range []Event{request, resolved} {
				b, _ := json.Marshal(event)
				fmt.Fprintf(w, "data: %s\n\n", b)
			}
			ping := fixture[[]Event](t, "events-valid.json")[0]
			ping.V = 3
			b, _ := json.Marshal(ping)
			fmt.Fprintf(w, "data: %s\n\n", b)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case strings.HasSuffix(r.URL.Path, "/messages"):
			fmt.Fprintf(w, `{"status":"accepted","turn_id":%q}`, *request.TurnID)
		case strings.HasSuffix(r.URL.Path, "/tool-output"):
			w.WriteHeader(409)
			fmt.Fprint(w, `{"error":{"type":"conflict","code":"tool_timed_out"}}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(context.Context, ToolInvocation) (ToolResult, error) {
		effects.Add(1)
		close(executed)
		return ToolResult{Output: "external effect performed"}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewFileCallStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewToolRunner(c, registry, store, RunnerOptions{Namespace: "audit-recovery"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = c.RunTurn(ctx, request.ChatID, "original", TurnOptions{IdempotencyKey: "previously-accepted", Tools: runner, OnEvent: func(e Event) error {

		if e.Type == "agent.tool.result" {
			return errors.New("audit observed recorded timeout")
		}
		return nil
	}})
	t.Logf("ALREADY TIMED OUT replay: external effects=%d RunTurn error=%v", effects.Load(), err)
	if effects.Load() != 0 {
		t.Fatalf("obsolete replay executed: %v", err)
	}
}

func TestRunningDuplicateResumesPendingSignedTool(t *testing.T) {
	old := wrappedTool(t)
	old.Seq = 1
	var wrapper NativeTaskEvent
	json.Unmarshal(old.Data, &wrapper)
	var executed atomic.Int32
	var submitted atomic.Bool
	keys := fixture[signatureFixtures](t, "signatures.json")
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/.well-known/jwks.json":
			json.NewEncoder(w).Encode(keys.JWKS)
		case strings.HasSuffix(r.URL.Path, "/tool-output"):
			submitted.Store(true)
			fmt.Fprint(w, `{}`)
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
			fmt.Fprintf(w, `{"task_id":"existing","target_chat_id":%q,"turn_id":%q}`, old.ChatID, wrapper.OriginTurnID)
		case strings.Contains(r.URL.Path, "/tasks/"):
			if submitted.Load() {
				fmt.Fprint(w, `{"task_id":"existing","origin_turn_id":"origin","status":"completed","output":"final result","result_turn_id":"final"}`)
			} else {
				fmt.Fprint(w, `{"task_id":"existing","origin_turn_id":"origin","status":"running"}`)
			}
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	})
	c.config.NativeTasks = true
	registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(context.Context, ToolInvocation) (ToolResult, error) {
		executed.Add(1)
		return ToolResult{Output: "ok"}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewFileCallStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewToolRunner(c, registry, store, RunnerOptions{Namespace: "audit"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	task, err := c.RunDelegatedJob(ctx, "origin", NativeDelegationRequest{WorkspaceID: "ws", EventID: "existing-event", OriginTurnID: "origin-turn", TargetChatID: old.ChatID, Prompt: "work"}, BackgroundOptions{Tools: runner})
	if err != nil || task.Status != TaskCompleted || task.Output != "final result" || task.ResultTurnID != "final" || executed.Load() != 1 || !submitted.Load() {
		t.Fatalf("unexpected result %+v %v effects=%d", task, err, executed.Load())
	}
	t.Logf("RUNNING DUPLICATE: status=%s executed=%d error=%v", task.Status, executed.Load(), err)
}

func TestNativeResponsesRejectInvalidShapesAndCorrelation(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"missing_origin", `{"task_id":"job","status":"completed"}`},
		{"invalid_state", `{"task_id":"job","origin_turn_id":"origin","status":"invented"}`},
		{"wrong_task", `{"task_id":"other","origin_turn_id":"origin","status":"completed"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					fmt.Fprint(w, backgroundCapabilities)
				} else {
					fmt.Fprint(w, tc.body)
				}
			})
			c.config.NativeTasks = true
			if _, err := c.GetDetachedTask(context.Background(), GetDetachedTaskParams{ChatID: "chat", TaskID: "job"}); err == nil {
				t.Fatal("invalid task accepted")
			}
		})
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, backgroundCapabilities)
			return
		}
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":{"type":"not_found","code":"task_unknown","message":"unknown"}}`)
	})
	c.config.NativeTasks = true
	_, err := c.GetTask(context.Background(), "chat", "job")
	var api *APIError
	var compatibility *CompatibilityError
	if !errors.As(err, &api) || api.Code != "task_unknown" || errors.As(err, &compatibility) {
		t.Fatal(err)
	}
}

func TestCompletionPagesUsesNumericSameOriginAndStops(t *testing.T) {
	var pages []string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing authorization")
		}
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		if !strings.HasPrefix(r.URL.Path, "/api/v1/workspaces/") {
			t.Error(r.URL.Path)
		}
		fmt.Fprintf(w, `{"data":[],"links":{"next":"https://foreign.invalid/steal"},"meta":{"current_page":%s,"last_page":2}}`, page)
	})
	for _, err := range c.CompletionPages(context.Background(), ListCompletionsParams{WorkspaceID: "ws", ChatID: "chat"}) {
		if err != nil {
			t.Fatal(err)
		}
	}
	if strings.Join(pages, ",") != "1,2" {
		t.Fatal(pages)
	}
	pages = nil
	for _, err := range c.CompletionPages(context.Background(), ListCompletionsParams{WorkspaceID: "ws", ChatID: "chat"}) {
		if err != nil {
			t.Fatal(err)
		}
		break
	}
	if len(pages) != 1 {
		t.Fatal(pages)
	}
}

func TestImageWaiterAdmitsOnceAndPreservesTerminalFailures(t *testing.T) {
	for _, status := range []string{"succeeded", "failed", "cancelled"} {
		t.Run(status, func(t *testing.T) {
			var posts, gets int
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" {
					posts++
					fmt.Fprint(w, `{"data":{"id":"image","status":"accepted"}}`)
					return
				}
				gets++
				state := "processing"
				if gets > 1 {
					state = status
				}
				fmt.Fprintf(w, `{"data":{"id":"image","status":%q}}`, state)
			})
			out, err := c.GenerateAndWait(context.Background(), CreateImageGenerationParams{}, PollOptions{time.Millisecond, time.Second})
			if err != nil || string(out.Status) != status || posts != 1 || gets != 2 {
				t.Fatal(out, err, posts, gets)
			}
		})
	}
}

func TestReadOnlySourceAndOperationWaiters(t *testing.T) {
	source := fixture[struct {
		Data WikiSource `json:"data"`
	}](t, "wiki/source-conversion-retry.json").Data
	operation := fixture[struct {
		Data WikiOperation `json:"data"`
	}](t, "wiki/operation-completed.json").Data
	var reads atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("waiter mutated", r.Method)
		}
		if strings.Contains(r.URL.Path, "operations") {
			json.NewEncoder(w).Encode(map[string]any{"data": operation})
			return
		}
		out := source
		if reads.Add(1) > 1 {
			out.Conversion.Recoverable = Present(false)
		}
		json.NewEncoder(w).Encode(map[string]any{"data": out})
	})
	out, err := c.WaitForSource(context.Background(), source.WorkspaceID, source.FileID, PollOptions{time.Millisecond, time.Second})
	if err != nil || out.Conversion.Status != "failed" || out.Conversion.Recoverable.Value || reads.Load() != 2 {
		t.Fatal(out, err, reads.Load())
	}
	op, err := c.WaitForWikiOperation(context.Background(), operation.WorkspaceID, operation.OperationID, PollOptions{time.Millisecond, time.Second})
	if err != nil || op.Status != "completed" {
		t.Fatal(op, err)
	}
}
func TestPollingDeadlineIncludesAdmission(t *testing.T) {
	var posts atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		select {
		case <-r.Context().Done():
		case <-time.After(100 * time.Millisecond):
		}
	})
	start := time.Now()
	_, err := c.GenerateAndWait(context.Background(), CreateImageGenerationParams{}, PollOptions{time.Millisecond, 20 * time.Millisecond})
	if err == nil || posts.Load() != 1 || time.Since(start) > time.Second {
		t.Fatal(err, posts.Load(), time.Since(start))
	}
}

func TestTaskCancellationDuringRealSignatureVerification(t *testing.T) {
	shared := fixture[struct {
		Request   Event           `json:"task_request"`
		Cancelled Event           `json:"task_cancelled"`
		JWKS      json.RawMessage `json:"jwks"`
	}](t, "parity/fixtures.json")
	reached, release := make(chan struct{}), make(chan struct{})
	var effects, posts atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			close(reached)
			<-release
			w.Write(shared.JWKS)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, backgroundCapabilities)
			return
		}
		posts.Add(1)
	})
	c.config.NativeTasks = true
	registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(context.Context, ToolInvocation) (ToolResult, error) {
		effects.Add(1)
		return ToolResult{Output: "effect"}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewFileCallStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewToolRunner(c, registry, store, RunnerOptions{Namespace: "cancel-race"})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- runner.HandleBackground(context.Background(), shared.Request, shared.Request.ChatID) }()
	<-reached
	if err = runner.ObserveBackground(context.Background(), shared.Cancelled); err != nil {
		close(release)
		t.Fatal(err)
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if effects.Load() != 0 || posts.Load() != 0 {
		t.Fatal(effects.Load(), posts.Load())
	}
}

func TestSharedForegroundReplayConformance(t *testing.T) {
	shared := fixture[struct {
		Request  Event           `json:"request"`
		Terminal Event           `json:"turn_terminal"`
		JWKS     json.RawMessage `json:"jwks"`
	}](t, "parity/fixtures.json")
	scenarios := fixture[struct {
		Cases map[string]struct {
			Events      []Event `json:"events"`
			Effects     int32   `json:"expected_effects"`
			Submissions int32   `json:"expected_submissions"`
		} `json:"cases"`
	}](t, "parity/scenarios.json")
	for _, id := range []string{"REPLAY-01", "REPLAY-02", "REPLAY-03"} {
		t.Run(id, func(t *testing.T) {
			scenario := scenarios.Cases[id]
			admitted, submitted := make(chan struct{}), make(chan struct{})
			var effects, posts atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/capabilities"):
					fmt.Fprint(w, testCapabilities)
				case r.URL.Path == "/.well-known/jwks.json":
					w.Write(shared.JWKS)
				case strings.HasSuffix(r.URL.Path, "/events"):
					w.Header().Set("Content-Type", "text/event-stream")
					for _, e := range scenario.Events {
						b, _ := json.Marshal(e)
						fmt.Fprintf(w, "data: %s\n\n", b)
					}
					w.(http.Flusher).Flush()
					select {
					case <-admitted:
					case <-r.Context().Done():
						return
					}
					if scenario.Effects > 0 {
						select {
						case <-submitted:
						case <-r.Context().Done():
							return
						}
					}
					if id != "REPLAY-02" {
						terminal := shared.Terminal
						terminal.Seq = int64(len(scenario.Events))
						b, _ := json.Marshal(terminal)
						fmt.Fprintf(w, "data: %s\n\n", b)
						w.(http.Flusher).Flush()
					}
					<-r.Context().Done()
				case strings.HasSuffix(r.URL.Path, "/messages"):
					fmt.Fprintf(w, `{"status":"accepted","turn_id":%q}`, *shared.Request.TurnID)
					close(admitted)
				case strings.HasSuffix(r.URL.Path, "/tool-output"):
					if posts.Add(1) == 1 {
						close(submitted)
					}
					w.WriteHeader(204)
				default:
					t.Error(r.URL.Path)
				}
			})
			registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(context.Context, ToolInvocation) (ToolResult, error) {
				effects.Add(1)
				return ToolResult{Output: "ok"}, nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			store, err := NewFileCallStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			runner, err := NewToolRunner(c, registry, store, RunnerOptions{Namespace: "shared-replay"})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			result, err := c.RunTurn(ctx, shared.Request.ChatID, "original", TurnOptions{IdempotencyKey: "stable", Tools: runner})
			if err != nil || result.TurnID != *shared.Request.TurnID || effects.Load() != scenario.Effects || posts.Load() != scenario.Submissions {
				t.Fatal(result, err, effects.Load(), posts.Load())
			}
		})
	}
}

func TestNativePhaseAndCancellationRejectInvalidResponse(t *testing.T) {
	for _, response := range []string{`{"decision":"MAYBE"}`, `{"decision":"BLOCKED","model":"model","request_id":"other","invocation_id":"inv"}`} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/capabilities") {
				fmt.Fprint(w, backgroundCapabilities)
			} else {
				fmt.Fprint(w, response)
			}
		})
		c.config.NativeTasks = true
		_, err := c.ClassifyNativePhase(context.Background(), "chat", NativePhaseRequest{RequestID: "request", TaskContext: json.RawMessage(`{}`), PhaseID: "phase", Attempt: 1, FinalOutput: "done"})
		if err == nil {
			t.Fatal("invalid classifier accepted", response)
		}
	}
	for _, response := range []string{`{"task_id":"job","status":"cancelled"}`, `{"task_id":"job","status":"cancelled","subtree":null}`, `{"task_id":"other","status":"cancelled","subtree":true}`, `{"task_id":"job","status":"completed","subtree":true}`, `{"task_id":"job","status":"cancelled","subtree":false}`} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/capabilities") {
				fmt.Fprint(w, backgroundCapabilities)
			} else {
				fmt.Fprint(w, response)
			}
		})
		c.config.NativeTasks = true
		if _, err := c.CancelTask(context.Background(), "chat", "job", true); err == nil {
			t.Fatal("invalid cancellation accepted", response)
		}
	}
}

func TestCallOwnershipChangesNeverExecuteTwice(t *testing.T) {
	for _, mode := range []string{"direct-wrapped", "wrapped-direct", "task-task"} {
		t.Run(mode, func(t *testing.T) {
			var effects atomic.Int32
			runner, posts, _ := backgroundHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
				effects.Add(1)
				return ToolResult{Output: "ok"}, nil
			}, `{"type":"object"}`, 204)
			store, err := NewFileCallStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			runner.store = store
			direct := signedEvent(t, `{}`)
			direct.V = 3
			wrapped := wrappedTool(t)
			first, second := direct, wrapped
			if mode == "wrapped-direct" {
				first, second = wrapped, direct
			}
			if mode == "task-task" {
				first, second = wrapped, wrapped
				var payload NativeTaskEvent
				json.Unmarshal(second.Data, &payload)
				payload.TaskID = "task_2"
				second.Data, _ = json.Marshal(payload)
			}
			if err := runner.HandleBackground(context.Background(), first, first.ChatID); err != nil {
				t.Fatal(err)
			}
			_ = runner.HandleBackground(context.Background(), second, second.ChatID) // ownership conflict or safe suppression
			if effects.Load() != 1 || posts.Load() != 1 {
				t.Fatal(effects.Load(), posts.Load())
			}
		})
	}
}
func TestActiveTaskCancellationStopsSubmission(t *testing.T) {
	started, stopped := make(chan struct{}), make(chan struct{})
	runner, posts, _ := backgroundHarness(t, func(ctx context.Context, _ ToolInvocation) (ToolResult, error) {
		close(started)
		<-ctx.Done()
		close(stopped)
		return ToolResult{Output: "late"}, nil
	}, `{"type":"object"}`, 204)
	event := wrappedTool(t)
	done := make(chan error, 1)
	go func() { done <- runner.HandleBackground(context.Background(), event, event.ChatID) }()
	<-started
	if err := runner.ObserveBackground(context.Background(), taskTerminal(t, "task.cancelled")); err != nil {
		t.Fatal(err)
	}
	<-stopped
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if posts.Load() != 0 {
		t.Fatal(posts.Load())
	}
}
func TestMultipartArrayFieldsPreserveElements(t *testing.T) {
	for _, count := range []int{0, 1, 3} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseMultipartForm(4096); err != nil {
					t.Error(err)
					return
				}
				files := r.MultipartForm.File["files[]"]
				if len(files) != count {
					t.Errorf("files=%d expected=%d", len(files), count)
				}
				for i, file := range files {
					if file.Filename != fmt.Sprintf("image%d.png", i) {
						t.Error(file.Filename)
					}
				}
				fmt.Fprint(w, `{"data":{}}`)
			})
			parts := []MultipartPart{{Name: "prompt", Data: []byte("test")}, {Name: "model_code", Data: []byte("fixture")}}
			for i := range count {
				parts = append(parts, MultipartPart{Name: "files[]", Filename: fmt.Sprintf("image%d.png", i), ContentType: "image/png", Data: []byte("synthetic")})
			}
			if _, err := c.TestGenerateImage(context.Background(), TestGenerateImageParams{Body: parts}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAPIErrorStatusesAndRequestIdentity(t *testing.T) {
	for _, status := range []int{401, 403, 404, 409, 422, 429, 502, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var calls atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.Header().Set("X-Request-ID", "request-parity")
				w.WriteHeader(status)
				fmt.Fprint(w, `{"error":{"type":"fixture","code":"fixture_code","message":"message","details":{"field":"value"}}}`)
			})
			c.config.RetryAttempts = 2
			_, err := c.CreateWorkspace(context.Background(), CreateWorkspaceParams{Body: WorkspaceCreateRequest{Name: "test"}})
			var api *APIError
			if !errors.As(err, &api) || api.Status != status || api.Code != "fixture_code" || api.RequestID != "request-parity" || string(api.Details) != `{"field":"value"}` || calls.Load() != 1 {
				t.Fatal(err, calls.Load())
			}
		})
	}
}

func TestStructuredTransportErrorsRetainCause(t *testing.T) {
	sentinel := errors.New("offline")
	c, err := New(Config{BaseURL: "https://api.example", APIKey: "synthetic", HTTPClient: &http.Client{Transport: wikiRoundTripper(func(*http.Request) (*http.Response, error) { return nil, sentinel })}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.CreateWorkspace(context.Background(), CreateWorkspaceParams{Body: WorkspaceCreateRequest{Name: "test"}})
	var transport *TransportError
	if !errors.As(err, &transport) || !errors.Is(err, sentinel) || transport.Operation != "createWorkspace" || strings.Contains(err.Error(), "api.example") {
		t.Fatal(err)
	}
}

func TestConvenienceWaitersPreserveErrors(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "waiter-request")
		w.WriteHeader(403)
		fmt.Fprint(w, `{"error":{"type":"access_denied","code":"fixture_denied","message":"denied"}}`)
	})
	options := PollOptions{time.Millisecond, time.Second}
	_, sourceErr := c.WaitForSource(context.Background(), "workspace", "source", options)
	_, operationErr := c.WaitForWikiOperation(context.Background(), "workspace", "operation", options)
	_, imageErr := c.WaitForImage(context.Background(), "image", options)
	for _, err := range []error{sourceErr, operationErr, imageErr} {
		var api *APIError
		if !errors.As(err, &api) || api.Status != 403 || api.Code != "fixture_denied" || api.RequestID != "waiter-request" {
			t.Fatal(err)
		}
	}
}
