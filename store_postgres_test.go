package envai

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"sync/atomic"
	"testing"
)

func TestPostgresCallStoreIndependentClaimsAndBarriers(t *testing.T) {
	dsn := os.Getenv("ENVAI_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("ENVAI_TEST_POSTGRES_DSN not set")
	}
	a, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if _, err = a.Exec(`DROP TABLE IF EXISTS envai_client_tool_calls`); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Exec(PostgresCallStoreSchema); err != nil {
		t.Fatal(err)
	}
	stores := []*PostgresCallStore{{a}, {b}}
	ctx := context.Background()
	id := CallIdentity{"account", "chat", "call", "turn", "tool"}
	var claims atomic.Int32
	var work sync.WaitGroup
	for i := range 20 {
		work.Add(1)
		go func(i int) {
			defer work.Done()
			claimed, err := stores[i%2].Claim(ctx, id)
			if err != nil {
				t.Error(err)
			}
			if claimed {
				claims.Add(1)
			}
		}(i)
	}
	work.Wait()
	if claims.Load() != 1 {
		t.Fatal(claims.Load())
	}
	if err = stores[0].Record(ctx, id, CallResolved); err != nil {
		t.Fatal(err)
	}
	if err = stores[1].Record(ctx, id, CallSubmitted); err != nil {
		t.Fatal(err)
	}
	state, exists, err := stores[1].State(ctx, id)
	if err != nil || !exists || state != CallResolved {
		t.Fatal(state, exists, err)
	}
	other := id
	other.Namespace = "other-account"
	if ok, err := stores[1].Claim(ctx, other); err != nil || !ok {
		t.Fatal(ok, err)
	}
	wrong := id
	wrong.TurnID = "different-task"
	if _, err := stores[1].Claim(ctx, wrong); err == nil {
		t.Fatal("ownership conflict accepted")
	}
	terminal := CallIdentity{"account", "chat", "terminal:task:one", "task:one", ""}
	if _, err = stores[0].Claim(ctx, terminal); err != nil {
		t.Fatal(err)
	}
	if err = stores[0].Record(ctx, terminal, CallAmbiguous); err != nil {
		t.Fatal(err)
	}
	if state, ok, err := stores[1].State(ctx, terminal); err != nil || !ok || state != CallAmbiguous {
		t.Fatal(state, ok, err)
	}
	// A fresh worker must honor a persisted task barrier before executing a
	// valid signed request, without relying on the first runner's memory.
	runner, posts, _ := backgroundHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
		t.Fatal("executed after durable task cancellation")
		return ToolResult{}, nil
	}, `{"type":"object"}`, 204)
	runner.store = stores[0]
	if err = runner.ObserveBackground(ctx, taskTerminal(t, "task.cancelled")); err != nil {
		t.Fatal(err)
	}
	restarted, err := NewToolRunner(runner.client, runner.registry, stores[1], RunnerOptions{Namespace: "background-harness"})
	if err != nil {
		t.Fatal(err)
	}
	event := wrappedTool(t)
	if err = restarted.HandleBackground(ctx, event, event.ChatID); err != nil {
		t.Fatal(err)
	}
	if posts.Load() != 0 {
		t.Fatal("submitted after durable cancellation")
	}
	// Corruption must fail closed even if a database administrator removed the check.
	if _, err = a.Exec(`ALTER TABLE envai_client_tool_calls DROP CONSTRAINT envai_client_tool_calls_state_check`); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Exec(`UPDATE envai_client_tool_calls SET state='corrupt' WHERE identity_key=$1`, identityKey(id)); err != nil {
		t.Fatal(err)
	}
	if _, err = stores[0].Claim(ctx, id); err == nil {
		t.Fatal("corrupt claim accepted")
	}
	if err = stores[0].Record(ctx, id, CallResolved); err == nil {
		t.Fatal("corrupt record overwritten")
	}
	if _, _, err = stores[1].State(ctx, id); err == nil {
		t.Fatal("corrupt state accepted")
	}
	if _, err = a.Exec(`DROP TABLE envai_client_tool_calls`); err != nil {
		t.Fatal(err)
	}
}
