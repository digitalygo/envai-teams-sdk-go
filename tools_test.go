package envai

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
)

type signatureFixtures struct {
	JWKS    JWKSDocument `json:"jwks"`
	Vectors []struct {
		Name      string    `json:"name"`
		ArgsJSON  string    `json:"args_json"`
		Canonical string    `json:"canonical"`
		CallID    string    `json:"call_id"`
		ToolName  string    `json:"tool_name"`
		Signature Signature `json:"signature"`
	} `json:"vectors"`
}

func signedEvent(t *testing.T, args string) Event {
	t.Helper()
	ev := fixture[[]Event](t, "events-valid.json")[4]
	seed, _ := hex.DecodeString("9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60")
	canonical, err := jsoncanonicalizer.Transform([]byte(args))
	if err != nil {
		t.Fatal(err)
	}
	var req ToolRequest
	json.Unmarshal(ev.Data, &req)
	req.Args = json.RawMessage(args)
	req.Signature.Sig = base64.RawURLEncoding.EncodeToString(ed25519.Sign(ed25519.NewKeyFromSeed(seed), []byte(req.CallID+"\n"+req.ToolName+"\n"+string(canonical))))
	ev.Data, _ = json.Marshal(req)
	return ev
}
func TestSharedSignaturesAndStrictVerification(t *testing.T) {
	f := fixture[signatureFixtures](t, "signatures.json")
	var calls atomic.Int32
	c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("JWKS leaked auth")
		}
		calls.Add(1)
		json.NewEncoder(w).Encode(f.JWKS)
	})
	v := &Verifier{Client: c}
	for _, tc := range f.Vectors {
		got, err := jsoncanonicalizer.Transform([]byte(tc.ArgsJSON))
		if err != nil || string(got) != tc.Canonical {
			t.Fatal(tc.Name, string(got), err)
		}
		req := ToolRequest{CallID: tc.CallID, ToolName: tc.ToolName, Kind: "client", Args: json.RawMessage(tc.ArgsJSON), Signature: tc.Signature}
		if err := v.Verify(context.Background(), req); err != nil {
			t.Fatal(tc.Name, err)
		}
		req.Args = []byte(`{"tampered":true}`)
		if err := v.Verify(context.Background(), req); err == nil {
			t.Fatal("tampered args accepted")
		}
	}
	if calls.Load() != 1 {
		t.Fatal("cache not used", calls.Load())
	}
	vector := f.Vectors[0]
	req := ToolRequest{CallID: vector.CallID, ToolName: vector.ToolName, Kind: "client", Args: json.RawMessage(vector.ArgsJSON), Signature: vector.Signature}
	for _, mutate := range []func(*ToolRequest){func(r *ToolRequest) { r.Args = []byte("   ") }, func(r *ToolRequest) { r.Args = []byte("[]") }, func(r *ToolRequest) { r.Signature.Alg = "none" }, func(r *ToolRequest) { r.Signature.Sig = "not a signature" }, func(r *ToolRequest) { r.Signature.Kid = "unknown" }, func(r *ToolRequest) { r.Kind = "server" }, func(r *ToolRequest) { r.Args = []byte(`{"a":1,"a":2}`) }} {
		bad := req
		mutate(&bad)
		if err := v.Verify(context.Background(), bad); err == nil {
			t.Fatal("invalid request accepted")
		}
	}
	if calls.Load() != 2 {
		t.Fatal("unknown kid must refresh once", calls.Load())
	}
	f.JWKS.Keys[0].Use = "enc"
	badVerifier := &Verifier{Client: c}
	if err := badVerifier.Verify(context.Background(), req); err == nil {
		t.Fatal("wrong key use")
	}
}
func TestRegistryValidation(t *testing.T) {
	valid := Tool{Name: "fixture_echo", Description: "echo", Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(context.Context, ToolInvocation) (ToolResult, error) {
		return JSONResult(map[string]bool{"ok": true})
	}}
	for _, mutation := range []func(*Tool){func(v *Tool) { v.Name = "bad name" }, func(v *Tool) { v.Handler = nil }, func(v *Tool) { v.Timeout = time.Millisecond }, func(v *Tool) { v.Parameters = []byte(`[]`) }, func(v *Tool) { v.Parameters = []byte(`{`) }, func(v *Tool) {
		v.Parameters = []byte(`{"type":"object","properties":{"value":{"$ref":"file:///etc/passwd"}}}`)
	}} {
		tool := valid
		mutation(&tool)
		if _, err := NewRegistry(tool); err == nil {
			t.Fatal("invalid registry accepted")
		}
	}
	if _, err := NewRegistry(valid, valid); err == nil {
		t.Fatal("duplicates accepted")
	}
	r, err := NewRegistry(valid)
	if err != nil {
		t.Fatal(err)
	}
	defs := r.Definitions()
	if len(defs) != 1 || defs[0].Name != "fixture_echo" || defs[0].TimeoutSeconds.Value != 120 || string(defs[0].Parameters["type"]) != `"object"` {
		t.Fatal(defs)
	}
	if _, err = JSONResult(make(chan int)); err == nil {
		t.Fatal("nonserializable result")
	}
}
func toolHarness(t *testing.T, handler func(context.Context, ToolInvocation) (ToolResult, error), schema string, status int) (*ToolRunner, *atomic.Int32, *ToolOutputRequest) {
	t.Helper()
	f := fixture[signatureFixtures](t, "signatures.json")
	calls := &atomic.Int32{}
	output := &ToolOutputRequest{}
	c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			json.NewEncoder(w).Encode(f.JWKS)
			return
		}
		calls.Add(1)
		json.NewDecoder(r.Body).Decode(output)
		w.WriteHeader(status)
		if status == 409 {
			io.WriteString(w, `{"error":{"type":"conflict","code":"tool_timed_out"}}`)
		}
	})
	registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(schema), Handler: handler})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewToolRunner(c, registry, NewMemoryCallStore(), RunnerOptions{Namespace: "fixture-account"})
	if err != nil {
		t.Fatal(err)
	}
	return runner, calls, output
}
func TestToolRunnerGuardrails(t *testing.T) {
	cases := []struct {
		name    string
		args    string
		schema  string
		handler func(context.Context, ToolInvocation) (ToolResult, error)
		code    string
	}{
		{"success", `{}`, `{"type":"object"}`, func(ctx context.Context, i ToolInvocation) (ToolResult, error) {
			return JSONResult(map[string]bool{"ok": true})
		}, ""},
		{"handler_error", `{}`, `{"type":"object"}`, func(context.Context, ToolInvocation) (ToolResult, error) {
			return ToolResult{}, errors.New("secret must not leak")
		}, "handler_error"},
		{"handler_panic", `{}`, `{"type":"object"}`, func(context.Context, ToolInvocation) (ToolResult, error) { panic("secret panic") }, "handler_error"},
		{"output_limit", `{}`, `{"type":"object"}`, func(context.Context, ToolInvocation) (ToolResult, error) {
			return ToolResult{Output: strings.Repeat("a", 131073)}, nil
		}, "output_too_large"},
		{"empty_output", `{}`, `{"type":"object"}`, func(context.Context, ToolInvocation) (ToolResult, error) { return ToolResult{}, nil }, "empty_output"},
		{"invalid_utf8", `{}`, `{"type":"object"}`, func(context.Context, ToolInvocation) (ToolResult, error) {
			return ToolResult{Output: string([]byte{255})}, nil
		}, "invalid_output"},
		{"nested_schema", `{"outer":{"value":"no"}}`, `{"type":"object","properties":{"outer":{"type":"object","properties":{"value":{"type":"integer"}},"required":["value"],"additionalProperties":false}},"required":["outer"],"additionalProperties":false}`, func(context.Context, ToolInvocation) (ToolResult, error) { panic("must not execute") }, "invalid_arguments"},
		{"enum", `{"value":"no"}`, `{"type":"object","properties":{"value":{"enum":["yes"]}},"required":["value"]}`, func(context.Context, ToolInvocation) (ToolResult, error) { panic("must not execute") }, "invalid_arguments"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, calls, output := toolHarness(t, tc.handler, tc.schema, 204)
			event := signedEvent(t, tc.args)
			if err := runner.Handle(context.Background(), event, event.ChatID); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 1 {
				t.Fatal(calls.Load())
			}
			if tc.code != "" {
				if !output.IsError.Value || !strings.Contains(output.Output, tc.code) || strings.Contains(output.Output, "secret") {
					t.Fatal(output)
				}
			} else if output.IsError.Value || output.Output != `{"ok":true}` {
				t.Fatal(output)
			}
			if err := runner.Handle(context.Background(), event, event.ChatID); err != nil || calls.Load() != 1 {
				t.Fatal("replayed side effect", err, calls.Load())
			}
		})
	}
}
func TestToolRunnerContextSignatureAndAuthorization(t *testing.T) {
	var executions atomic.Int32
	runner, calls, output := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
		executions.Add(1)
		return ToolResult{Output: "ok"}, nil
	}, `{"type":"object"}`, 204)
	event := signedEvent(t, `{}`)
	if err := runner.Handle(context.Background(), event, "other-chat"); err == nil || calls.Load() != 0 || executions.Load() != 0 {
		t.Fatal("wrong-chat executed")
	}
	var req ToolRequest
	json.Unmarshal(event.Data, &req)
	req.Args = []byte(`{"tampered":true}`)
	bad := event
	bad.Data, _ = json.Marshal(req)
	if err := runner.Handle(context.Background(), bad, event.ChatID); err == nil || calls.Load() != 0 || executions.Load() != 0 {
		t.Fatal("bad signature executed")
	}
	runner.options.Authorize = func(context.Context, ToolInvocation, string) error { return errors.New("denied") }
	if err := runner.Handle(context.Background(), event, event.ChatID); err != nil || executions.Load() != 0 || !strings.Contains(output.Output, "tool_not_authorized") {
		t.Fatal(err, output)
	}
	empty, _ := NewRegistry()
	runner.registry = empty
	runner.store = NewMemoryCallStore()
	if err := runner.Handle(context.Background(), event, event.ChatID); err != nil || !strings.Contains(output.Output, "tool_not_found") {
		t.Fatal(err, output)
	}
}
func TestToolRunnerCancellationAndBoundedConcurrency(t *testing.T) {
	started := make(chan struct{})
	runner, calls, _ := toolHarness(t, func(ctx context.Context, i ToolInvocation) (ToolResult, error) {
		close(started)
		<-ctx.Done()
		return ToolResult{}, ctx.Err()
	}, `{"type":"object"}`, 204)
	event := signedEvent(t, `{}`)
	done := make(chan error, 1)
	go func() { done <- runner.Handle(context.Background(), event, event.ChatID) }()
	<-started
	old := fixture[[]Event](t, "events-valid.json")[7]
	otherTurn := "cmpl_01ARZ3NDEKTSV4RRFFQ69G5FA0"
	old.TurnID = &otherTurn
	var data map[string]any
	json.Unmarshal(old.Data, &data)
	data["turn_id"] = otherTurn
	old.Data, _ = json.Marshal(data)
	runner.Observe(old)
	select {
	case err := <-done:
		t.Fatal("old turn cancelled current tool", err)
	case <-time.After(5 * time.Millisecond):
	}
	runner.Observe(fixture[[]Event](t, "events-valid.json")[7])
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("tool not cancelled")
	}
	if calls.Load() != 0 {
		t.Fatal("late output submitted")
	}
}
func TestToolOutputFailureNotRetried(t *testing.T) {
	runner, calls, _ := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) { return ToolResult{Output: "ok"}, nil }, `{"type":"object"}`, 409)
	event := signedEvent(t, `{}`)
	if err := runner.Handle(context.Background(), event, event.ChatID); err == nil {
		t.Fatal("missing delivery error")
	}
	if err := runner.Handle(context.Background(), event, event.ChatID); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatal("retried mutation")
	}
}
func TestDurableAndMemoryCallStores(t *testing.T) {
	id := CallIdentity{Namespace: "fixture-account", ChatID: "chat", CallID: "call", TurnID: "turn", ToolName: "echo"}
	for _, kind := range []string{"memory", "file"} {
		t.Run(kind, func(t *testing.T) {
			var store CallStore
			var err error
			root := t.TempDir()
			if kind == "memory" {
				store = NewMemoryCallStore()
			} else {
				store, err = NewFileCallStore(root)
				if err != nil {
					t.Fatal(err)
				}
			}
			var won atomic.Int32
			var wg sync.WaitGroup
			for range 20 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ok, err := store.Claim(context.Background(), id)
					if err != nil {
						t.Error(err)
					}
					if ok {
						won.Add(1)
					}
				}()
			}
			wg.Wait()
			if won.Load() != 1 {
				t.Fatal("duplicate claims", won.Load())
			}
			for _, state := range []CallState{CallSubmitted, CallResolved, CallAmbiguous} {
				if err = store.Record(context.Background(), id, state); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "file" {
				store, err = NewFileCallStore(root)
				if err != nil {
					t.Fatal(err)
				}
				dir := filepath.Join(root, identityKey(id))
				if err = os.WriteFile(filepath.Join(dir, "state.json"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if ok, err := store.Claim(context.Background(), id); err != nil || ok {
				t.Fatal("restart/corruption replayed", ok, err)
			}
			foreign := id
			foreign.Namespace = "other-account"
			if ok, err := store.Claim(context.Background(), foreign); err != nil || !ok {
				t.Fatal(ok, err)
			}
			missing := id
			missing.CallID = "missing"
			if err = store.Record(context.Background(), missing, CallResolved); err == nil {
				t.Fatal("record without claim")
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err = store.Claim(ctx, id); err == nil {
				t.Fatal("ignored cancellation")
			}
			if err = store.Record(ctx, id, CallResolved); err == nil {
				t.Fatal("ignored cancellation")
			}
		})
	}
	target := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileCallStore(link); err == nil {
		t.Fatal("symlink store accepted")
	}
}
func TestRunnerConfiguration(t *testing.T) {
	c := negotiatedTestClient(t, func(http.ResponseWriter, *http.Request) {})
	r, _ := NewRegistry()
	for _, o := range []RunnerOptions{{}, {Namespace: "scope", MaxConcurrent: -1}, {Namespace: "scope", SubmissionMargin: -1}} {
		if _, err := NewToolRunner(c, r, NewMemoryCallStore(), o); err == nil {
			t.Fatal(fmt.Sprint(o))
		}
	}
}

func TestToolDeadlineIncludesVerificationAndAuthorization(t *testing.T) {
	var ran atomic.Bool
	runner, calls, _ := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
		ran.Store(true)
		return ToolResult{Output: "ok"}, nil
	}, `{"type":"object"}`, 204)
	tool := runner.registry.tools["fixture_echo"]
	tool.Timeout = time.Second
	runner.registry.tools["fixture_echo"] = tool
	runner.options.SubmissionMargin = 990 * time.Millisecond
	runner.options.Authorize = func(context.Context, ToolInvocation, string) error { time.Sleep(20 * time.Millisecond); return nil }
	event := signedEvent(t, `{}`)
	if err := runner.Handle(context.Background(), event, event.ChatID); err == nil || ran.Load() || calls.Load() != 0 {
		t.Fatal("expired deadline ran handler", err, ran.Load(), calls.Load())
	}
}
func TestBoundedToolExecutionAndTimedOutHandler(t *testing.T) {
	runner, calls, output := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) { return ToolResult{Output: "ok"}, nil }, `{"type":"object"}`, 204)
	event := signedEvent(t, `{}`)
	for range cap(runner.slots) {
		runner.slots <- struct{}{}
	}
	if err := runner.Handle(context.Background(), event, event.ChatID); err != nil || calls.Load() != 1 || !strings.Contains(output.Output, "tool_busy") {
		t.Fatal(err, output)
	}
	for range cap(runner.slots) {
		<-runner.slots
	}
	runner.store = NewMemoryCallStore()
	tool := runner.registry.tools["fixture_echo"]
	tool.Timeout = time.Second
	tool.Handler = func(ctx context.Context, _ ToolInvocation) (ToolResult, error) {
		<-ctx.Done()
		return ToolResult{}, ctx.Err()
	}
	runner.registry.tools["fixture_echo"] = tool
	runner.options.SubmissionMargin = 900 * time.Millisecond
	if err := runner.Handle(context.Background(), event, event.ChatID); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestVerifiedCanonicalNumbersAreExecuted(t *testing.T) {
	var got string
	runner, _, _ := toolHarness(t, func(ctx context.Context, inv ToolInvocation) (ToolResult, error) {
		got = string(inv.Arguments)
		return ToolResult{Output: "ok"}, nil
	}, `{"type":"object","properties":{"number":{"type":"number"}},"required":["number"]}`, 204)
	event := signedEvent(t, `{"number":9007199254740992}`)
	var req ToolRequest
	json.Unmarshal(event.Data, &req)
	req.Args = []byte(`{"number":9007199254740993}`)
	event.Data, _ = json.Marshal(req)
	if err := runner.Handle(context.Background(), event, event.ChatID); err != nil {
		t.Fatal(err)
	}
	if got != `{"number":9007199254740992}` {
		t.Fatal("executed noncanonical numeric payload", got)
	}
}

func TestObservedOutcomeBeforeHandlerAdmissionPreventsSideEffect(t *testing.T) {
	for _, kind := range []string{"result", "terminal"} {
		t.Run(kind, func(t *testing.T) {
			var effects atomic.Int32
			runner, calls, _ := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
				effects.Add(1)
				return ToolResult{Output: "ok"}, nil
			}, `{"type":"object"}`, 204)
			event := signedEvent(t, `{}`)
			gate := make(chan struct{})
			entered := make(chan struct{})
			runner.options.Authorize = func(context.Context, ToolInvocation, string) error { close(entered); <-gate; return nil }
			done := make(chan error, 1)
			go func() { done <- runner.Handle(context.Background(), event, event.ChatID) }()
			<-entered
			observed := fixture[[]Event](t, "events-valid.json")[5]
			if kind == "terminal" {
				observed = fixture[[]Event](t, "events-valid.json")[7]
			}
			runner.Observe(observed)
			close(gate)
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			if effects.Load() != 0 || calls.Load() != 0 {
				t.Fatal("settled call executed", effects.Load(), calls.Load())
			}
		})
	}
}

func TestObservedOutcomesPersistBeforeRequest(t *testing.T) {
	for _, kind := range []string{"result", "terminal"} {
		t.Run(kind, func(t *testing.T) {
			var effects atomic.Int32
			runner, calls, _ := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
				effects.Add(1)
				return ToolResult{Output: "ok"}, nil
			}, `{"type":"object"}`, 204)
			root := t.TempDir()
			store, err := NewFileCallStore(root)
			if err != nil {
				t.Fatal(err)
			}
			runner.store = store
			observed := fixture[[]Event](t, "events-valid.json")[5]
			if kind == "terminal" {
				observed = fixture[[]Event](t, "events-valid.json")[7]
			}
			if err = runner.Observe(observed); err != nil {
				t.Fatal(err)
			}
			reopened, err := NewFileCallStore(root)
			if err != nil {
				t.Fatal(err)
			}
			options := runner.options
			options.Namespace = "fixture-account"
			fresh, err := NewToolRunner(runner.client, runner.registry, reopened, options)
			if err != nil {
				t.Fatal(err)
			}
			event := signedEvent(t, `{}`)
			if err = fresh.Handle(context.Background(), event, event.ChatID); err != nil {
				t.Fatal(err)
			}
			if effects.Load() != 0 || calls.Load() != 0 {
				t.Fatal("persisted outcome lost", effects.Load(), calls.Load())
			}
		})
	}
}
