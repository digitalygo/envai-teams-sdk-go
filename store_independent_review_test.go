package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIndependentReviewObservationIdentityAndReload(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	s, err := NewFileCallStore(root)
	if err != nil {
		t.Fatal(err)
	}
	original := CallIdentity{Namespace: "account-A", ChatID: "chat-A", CallID: "call-A", TurnID: "turn-A", ToolName: "local.echo"}
	if ok, err := s.Claim(ctx, original); !ok || err != nil {
		t.Fatal(ok, err)
	}
	// Both public query/observation metadata fields can be omitted.
	observed := original
	observed.ToolName = ""
	observed.TurnID = ""
	if err := s.Record(ctx, observed, CallResolved); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, identityKey(original), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var persisted callRecord
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.Identity != original {
		t.Fatalf("identity changed: %#v", persisted.Identity)
	}
	reopened, err := NewFileCallStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if state, exists, err := reopened.State(ctx, original); state != CallResolved || !exists || err != nil {
		t.Fatal(state, exists, err)
	}
	if ok, err := reopened.Claim(ctx, original); ok || err != nil {
		t.Fatal(ok, err)
	}
	// Supplying contradictory nonempty metadata must not overwrite the old row.
	for _, field := range []string{"tool", "turn"} {
		bad := original
		if field == "tool" {
			bad.ToolName = "other"
		} else {
			bad.TurnID = "other"
		}
		if err := reopened.Record(ctx, bad, CallAmbiguous); err == nil {
			t.Fatal("contradictory metadata accepted", field)
		}
		after, _ := os.ReadFile(filepath.Join(root, identityKey(original), "state.json"))
		if !bytes.Equal(raw, after) {
			t.Fatal("contradictory metadata overwrote record")
		}
	}
}

func TestIndependentReviewUnfinishedClaimStaysReserved(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	s, err := NewFileCallStore(root)
	if err != nil {
		t.Fatal(err)
	}
	id := CallIdentity{Namespace: "a", ChatID: "b", CallID: "c", TurnID: "t", ToolName: "echo"}
	dir := filepath.Join(root, identityKey(id))
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 3; n++ {
		if state, exists, err := s.State(ctx, id); state != CallAmbiguous || !exists || err != nil {
			t.Fatal(state, exists, err)
		}
		if ok, err := s.Claim(ctx, id); ok || err != nil {
			t.Fatal(ok, err)
		}
		if err := s.Record(ctx, id, CallResolved); err == nil {
			t.Fatal("unfinished claim repaired")
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 0 {
			t.Fatal(entries, err)
		}
	}
}
