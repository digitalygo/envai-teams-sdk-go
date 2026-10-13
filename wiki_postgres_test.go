package envai

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func postgresWiki(t *testing.T, reset bool) (*WikiWebhookInbox, http.Header, []byte) {
	t.Helper()
	dsn := os.Getenv("ENVAI_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("ENVAI_TEST_POSTGRES_DSN not set; disposable PostgreSQL integration only")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	if reset {
		if _, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS envai_wiki_application_outbox,envai_wiki_webhook_inbox CASCADE`); err != nil {
			t.Fatal(err)
		}
		if _, err = db.ExecContext(ctx, WikiInboxSchema); err != nil {
			t.Fatal(err)
		}
	}
	verifier, headers, body, _ := wikiHarness(t)
	return &WikiWebhookInbox{DB: db, Verifier: verifier}, headers, body
}
func wikiRows(t *testing.T, db *sql.DB) (int, int) {
	t.Helper()
	var receipts, jobs int
	if err := db.QueryRow(`SELECT COUNT(*) FROM envai_wiki_webhook_inbox`).Scan(&receipts); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM envai_wiki_application_outbox`).Scan(&jobs); err != nil {
		t.Fatal(err)
	}
	return receipts, jobs
}
func TestWikiPostgresAtomicRollbackCollisionAndScopes(t *testing.T) {
	inbox, headers, body := postgresWiki(t, true)
	ctx := context.Background()
	failure := errors.New("application failure")
	outcome, err := inbox.Receive(ctx, "production", "POST", headers, body, func(ctx context.Context, tx WikiTransaction, event VerifiedWikiEvent) error {
		if _, ok := tx.(interface{ Commit() error }); ok {
			t.Fatal("callback exposes commit")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO envai_wiki_application_outbox (deduplication_key,event_id,workspace_id,payload) VALUES($1,$2,$3,$4)`, wikiReceiptKey("production", event.WorkspaceID, event.EventID), event.EventID, event.WorkspaceID, event.Raw); err != nil {
			t.Fatal(err)
		}
		return failure
	})
	if !errors.Is(err, failure) || outcome != "" {
		t.Fatal(outcome, err)
	}
	if receipts, jobs := wikiRows(t, inbox.DB); receipts != 0 || jobs != 0 {
		t.Fatal("rollback lost atomicity", receipts, jobs)
	}
	for _, want := range []WikiReceiptOutcome{WikiReceiptAccepted, WikiReceiptDuplicate} {
		outcome, err = inbox.ReceiveToOutbox(ctx, "production", "POST", headers, body)
		if err != nil || outcome != want {
			t.Fatal(outcome, err)
		}
	}
	changed := []byte(strings.Replace(string(body), `"source_revision":3`, `"source_revision":4`, 1))
	signWiki(t, headers, changed)
	if outcome, err = inbox.ReceiveToOutbox(ctx, "production", "POST", headers, changed); err != nil || outcome != WikiReceiptCollision {
		t.Fatal(outcome, err)
	}
	if receipts, jobs := wikiRows(t, inbox.DB); receipts != 1 || jobs != 1 {
		t.Fatal(receipts, jobs)
	}
	signWiki(t, headers, body)
	if outcome, err = inbox.ReceiveToOutbox(ctx, "staging", "POST", headers, body); err != nil || outcome != WikiReceiptAccepted {
		t.Fatal(outcome, err)
	}
	var persisted []byte
	if err = inbox.DB.QueryRow(`SELECT payload FROM envai_wiki_application_outbox LIMIT 1`).Scan(&persisted); err != nil || string(persisted) != string(body) {
		t.Fatal(err, "raw body changed")
	}
	if receipts, jobs := wikiRows(t, inbox.DB); receipts != 2 || jobs != 2 {
		t.Fatal(receipts, jobs)
	}
	for _, scope := range []string{"", "bad\x00scope", "bad\nscope"} {
		if _, err = inbox.ReceiveToOutbox(ctx, scope, "POST", headers, body); err == nil {
			t.Fatal("bad connection scope")
		}
	}
	if _, err = inbox.Receive(ctx, "production", "POST", headers, body, nil); err == nil {
		t.Fatal("nil callback")
	}
	if _, err = inbox.DB.Exec(`DROP TABLE envai_wiki_application_outbox`); err != nil {
		t.Fatal(err)
	}
	if _, err = inbox.ReceiveToOutbox(ctx, "third", "POST", headers, body); err == nil {
		t.Fatal("store failure acknowledged")
	}
	var count int
	inbox.DB.QueryRow(`SELECT COUNT(*) FROM envai_wiki_webhook_inbox`).Scan(&count)
	if count != 2 {
		t.Fatal("failed outbox leaked receipt", count)
	}
	if _, err = inbox.DB.Exec(`DROP TABLE envai_wiki_webhook_inbox`); err != nil {
		t.Fatal(err)
	}
	if _, err = inbox.ReceiveToOutbox(ctx, "third", "POST", headers, body); err == nil {
		t.Fatal("missing table acknowledged")
	}
}
func TestWikiPostgresVerificationBeforeTransaction(t *testing.T) {
	inbox, headers, body := postgresWiki(t, true)
	inbox.DB.Close()
	headers.Set("Envai-Webhook-Signature", "bad")
	if _, err := inbox.ReceiveToOutbox(context.Background(), "scope", "POST", headers, body); err == nil {
		t.Fatal("bad signature accepted")
	} else {
		var protocol *ProtocolError
		if !errors.As(err, &protocol) {
			t.Fatal("database reached before verification", err)
		}
	}
	signWiki(t, headers, body)
	if _, err := inbox.ReceiveToOutbox(context.Background(), "scope", "POST", headers, body); err == nil {
		t.Fatal("closed database acknowledged")
	}
}
func TestWikiPostgresConcurrentProcesses(t *testing.T) {
	inbox, _, _ := postgresWiki(t, true)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	const workers = 6
	results := make(chan string, workers)
	var wg sync.WaitGroup
	for n := 0; n < workers; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWikiPostgresChild$", "-test.v")
			cmd.Env = append(os.Environ(), "ENVAI_WIKI_RECEIPT_CHILD=1")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("process: %v\n%s", err, out)
			}
			results <- string(out)
		}()
	}
	wg.Wait()
	close(results)
	accepted, duplicate := 0, 0
	for output := range results {
		if strings.Contains(output, "receipt-result:accepted") {
			accepted++
		}
		if strings.Contains(output, "receipt-result:duplicate") {
			duplicate++
		}
	}
	if accepted != 1 || duplicate != workers-1 {
		t.Fatal(accepted, duplicate)
	}
	if receipts, jobs := wikiRows(t, inbox.DB); receipts != 1 || jobs != 1 {
		t.Fatal(receipts, jobs)
	}
}
func TestWikiPostgresChild(t *testing.T) {
	if os.Getenv("ENVAI_WIKI_RECEIPT_CHILD") != "1" {
		t.Skip("subprocess-only helper")
	}
	inbox, h, b := postgresWiki(t, false)
	outcome, err := inbox.ReceiveToOutbox(context.Background(), "concurrency", "POST", h, b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("receipt-result:" + string(outcome))
}
func TestWikiInboxConfiguration(t *testing.T) {
	if _, err := (&WikiWebhookInbox{}).ReceiveToOutbox(context.Background(), "scope", "POST", nil, nil); err == nil {
		t.Fatal("nil database")
	}
}
