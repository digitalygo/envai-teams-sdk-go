package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestFileCallStorePreservesCorruptRecords(t *testing.T) {
	id := CallIdentity{Namespace: "account", ChatID: "chat", CallID: "call", TurnID: "turn", ToolName: "echo"}
	valid, _ := json.Marshal(callRecord{id, CallStarted})
	badState := bytes.Replace(valid, []byte(`"started"`), []byte(`"unknown"`), 1)
	wrongIdentity := bytes.Replace(valid, []byte(`"chat"`), []byte(`"other"`), 1)
	missingTool := bytes.Replace(valid, []byte(`,"tool_name":"echo"`), nil, 1)
	nullTool := bytes.Replace(valid, []byte(`"tool_name":"echo"`), []byte(`"tool_name":null`), 1)
	wrongTurn := bytes.Replace(valid, []byte(`"turn"`), []byte(`"other"`), 1)
	wrongTool := bytes.Replace(valid, []byte(`"echo"`), []byte(`"other"`), 1)
	nullState := bytes.Replace(valid, []byte(`"started"`), []byte(`null`), 1)
	numericState := bytes.Replace(valid, []byte(`"started"`), []byte(`42`), 1)
	wrongCase := bytes.Replace(valid, []byte(`"state"`), []byte(`"State"`), 1)
	duplicate := append([]byte(`{"state":"resolved",`), valid[1:]...)
	for name, raw := range map[string][]byte{"null": []byte("null"), "array": []byte("[]"), "empty": []byte("{}"), "scalar": []byte("1"), "broken": []byte("{"), "unknown_state": badState, "wrong_identity": wrongIdentity, "missing_tool": missingTool, "null_tool": nullTool, "wrong_turn": wrongTurn, "wrong_tool": wrongTool, "null_state": nullState, "numeric_state": numericState, "wrong_case": wrongCase, "duplicate_key": duplicate} {
		t.Run(name, func(t *testing.T) {
			store, err := NewFileCallStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if claimed, err := store.Claim(context.Background(), id); err != nil || !claimed {
				t.Fatal(claimed, err)
			}
			path := filepath.Join(store.root, identityKey(id), "state.json")
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			if claimed, _ := store.Claim(context.Background(), id); claimed {
				t.Fatal("corrupt claim reclaimed")
			}
			if _, exists, err := store.State(context.Background(), id); err == nil || !exists {
				t.Fatalf("corrupt state accepted: exists=%v err=%v", exists, err)
			}
			if err := store.Record(context.Background(), id, CallResolved); err == nil {
				t.Error("corrupt record overwritten")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(raw, after) {
				t.Fatalf("record bytes changed: %v", err)
			}
		})
	}
}

func TestCorruptStoreCannotExecuteSignedRequest(t *testing.T) {
	for _, barrier := range []bool{false, true} {
		t.Run(map[bool]string{false: "call", true: "turn"}[barrier], func(t *testing.T) {
			var executed atomic.Int32
			runner, submitted, _ := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
				executed.Add(1)
				return ToolResult{Output: "ok"}, nil
			}, `{"type":"object"}`, 204)
			store, err := NewFileCallStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			runner.store = store
			event := signedEvent(t, `{}`)
			var req ToolRequest
			json.Unmarshal(event.Data, &req)
			id := CallIdentity{Namespace: runner.options.Namespace, ChatID: event.ChatID, CallID: req.CallID, TurnID: *event.TurnID, ToolName: req.ToolName}
			if barrier {
				id.CallID = "terminal:" + *event.TurnID
				id.ToolName = ""
			}
			if claimed, err := store.Claim(context.Background(), id); err != nil || !claimed {
				t.Fatal(claimed, err)
			}
			path := filepath.Join(store.root, identityKey(id), "state.json")
			os.WriteFile(path, []byte("null"), 0600)
			_ = runner.Handle(context.Background(), event, event.ChatID)
			if executed.Load() != 0 || submitted.Load() != 0 {
				t.Fatal("corruption permitted execution or submission")
			}
			observed := fixture[[]Event](t, "events-valid.json")[5]
			if barrier {
				observed = fixture[[]Event](t, "events-valid.json")[7]
			}
			if err := runner.Observe(observed); err == nil {
				t.Error("observation overwrote corrupt state")
			}
			after, _ := os.ReadFile(path)
			if string(after) != "null" {
				t.Fatal("corrupt bytes changed")
			}
		})
	}
}

func TestCallStoreRejectsInvalidTransitions(t *testing.T) {
	id := CallIdentity{Namespace: "account", ChatID: "chat", CallID: "call", TurnID: "turn", ToolName: "echo"}
	file, err := NewFileCallStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, store := range []CallStore{NewMemoryCallStore(), file} {
		if claimed, err := store.Claim(context.Background(), id); !claimed || err != nil {
			t.Fatal(claimed, err)
		}
		for _, state := range []CallState{"", "unknown"} {
			if err := store.Record(context.Background(), id, state); err == nil {
				t.Fatal("invalid state accepted")
			}
			if current, exists, err := store.State(context.Background(), id); err != nil || !exists || current != CallStarted {
				t.Fatal(current, exists, err)
			}
		}
	}
	memory := NewMemoryCallStore()
	memory.calls[identityKey(id)] = "damaged"
	if claimed, err := memory.Claim(context.Background(), id); claimed || err == nil {
		t.Fatal(claimed, err)
	}
	if err := memory.Record(context.Background(), id, CallResolved); err == nil {
		t.Fatal("invalid memory state overwritten")
	}
	if _, exists, err := memory.State(context.Background(), id); !exists || err == nil {
		t.Fatal(exists, err)
	}
	if memory.calls[identityKey(id)] != "damaged" {
		t.Fatal("memory corruption changed")
	}
}

func TestFileCallStoreMissingAndNonregularRecords(t *testing.T) {
	id := CallIdentity{Namespace: "account", ChatID: "chat", CallID: "call", TurnID: "turn", ToolName: "echo"}
	for _, kind := range []string{"missing", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			store, err := NewFileCallStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if state, exists, err := store.State(context.Background(), id); err != nil || exists || state != "" {
				t.Fatal(state, exists, err)
			}
			if claimed, err := store.Claim(context.Background(), id); !claimed || err != nil {
				t.Fatal(claimed, err)
			}
			path := filepath.Join(store.root, identityKey(id), "state.json")
			os.Remove(path)
			if kind == "directory" {
				os.Mkdir(path, 0700)
			}
			target := filepath.Join(t.TempDir(), "target")
			os.WriteFile(target, []byte("null"), 0600)
			if kind == "symlink" {
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			}
			state, exists, err := store.State(context.Background(), id)
			if !exists || (kind == "missing" && (err != nil || state != CallAmbiguous)) || (kind != "missing" && err == nil) {
				t.Fatal(state, exists, err)
			}
			if claimed, _ := store.Claim(context.Background(), id); claimed {
				t.Fatal("record reclaimed")
			}
			if err := store.Record(context.Background(), id, CallResolved); err == nil {
				t.Fatal("record repaired automatically")
			}
			data, _ := os.ReadFile(target)
			if string(data) != "null" {
				t.Fatal("external target changed")
			}
		})
	}
}

func TestFileRecordPreservesIdentityOnObservedOutcome(t *testing.T) {
	store, err := NewFileCallStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id := CallIdentity{Namespace: "account", ChatID: "chat", CallID: "call", TurnID: "turn", ToolName: "echo"}
	store.Claim(context.Background(), id)
	observed := id
	observed.ToolName = ""
	if err := store.Record(context.Background(), observed, CallResolved); err != nil {
		t.Fatal(err)
	}
	state, exists, err := store.State(context.Background(), id)
	if err != nil || !exists || state != CallResolved {
		t.Fatal(state, exists, err)
	}
}

// Local state disappearing after a side effect cannot justify sending an
// unrecorded output. This is failure injection, not a remote attack model.
func TestSignedRunnerMissingClaimBeforeDelivery(t *testing.T) {
	for _, kind := range []string{"memory", "file_state", "file_directory"} {
		t.Run(kind, func(t *testing.T) {
			var runner *ToolRunner
			var effects atomic.Int32
			var store CallStore
			if kind == "memory" {
				store = NewMemoryCallStore()
			} else {
				var err error
				store, err = NewFileCallStore(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
			}
			handler := func(ctx context.Context, inv ToolInvocation) (ToolResult, error) {
				effects.Add(1)
				id := CallIdentity{Namespace: runner.options.Namespace, ChatID: inv.ChatID, CallID: inv.CallID, TurnID: inv.TurnID, ToolName: "fixture_echo"}
				switch s := store.(type) {
				case *MemoryCallStore:
					s.mu.Lock()
					delete(s.calls, identityKey(id))
					s.mu.Unlock()
				case *FileCallStore:
					dir := filepath.Join(s.root, identityKey(id))
					if err := os.Remove(filepath.Join(dir, "state.json")); err != nil {
						return ToolResult{}, err
					}
					if kind == "file_directory" {
						if err := os.Remove(dir); err != nil {
							return ToolResult{}, err
						}
					}
				}
				return ToolResult{Output: "ok"}, nil
			}
			var submissions *atomic.Int32
			runner, submissions, _ = toolHarness(t, handler, `{"type":"object"}`, 204)
			runner.store = store
			event := signedEvent(t, `{}`)
			if err := runner.Handle(context.Background(), event, event.ChatID); err == nil {
				t.Fatal("missing claim did not block delivery")
			}
			if effects.Load() != 1 || submissions.Load() != 0 {
				t.Fatalf("effects=%d submissions=%d", effects.Load(), submissions.Load())
			}
		})
	}
}
