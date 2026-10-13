package envai

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

// WikiInboxSchema is an explicit PostgreSQL migration. Constructors never execute
// DDL. Apply it under your own deployment process using transactional storage.
const WikiInboxSchema = `CREATE TABLE envai_wiki_webhook_inbox (
 deduplication_key CHAR(64) PRIMARY KEY,
 event_id TEXT NOT NULL,
 workspace_id TEXT NOT NULL,
 payload_hash CHAR(64) NOT NULL,
 event_type TEXT NOT NULL,
 received_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE envai_wiki_application_outbox (
 deduplication_key CHAR(64) PRIMARY KEY REFERENCES envai_wiki_webhook_inbox(deduplication_key),
 event_id TEXT NOT NULL,
 workspace_id TEXT NOT NULL,
 payload BYTEA NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

type WikiReceiptOutcome string

const (
	WikiReceiptAccepted  WikiReceiptOutcome = "accepted"
	WikiReceiptDuplicate WikiReceiptOutcome = "duplicate"
	WikiReceiptCollision WikiReceiptOutcome = "collision"
)

// WikiTransaction intentionally exposes no Commit/Rollback or underlying DB.
// Persist only durable local application work here, never remote side effects.
type WikiTransaction interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}
type wikiTransaction struct{ tx *sql.Tx }

func (t wikiTransaction) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

type WikiWebhookInbox struct {
	DB       *sql.DB
	Verifier *WikiWebhookVerifier
}

func wikiReceiptKey(connection, workspace, event string) string {
	h := sha256.Sum256([]byte(connection + "\x00" + workspace + "\x00" + event))
	return hex.EncodeToString(h[:])
}

// Receive verifies first and atomically commits the receipt and local callback.
// Only Accepted or Duplicate may be acknowledged with 2xx. Collision is a 409;
// verification/store errors must produce non-2xx. A failed callback rolls back.
func (i *WikiWebhookInbox) Receive(ctx context.Context, connection, method string, headers http.Header, body []byte, persist func(context.Context, WikiTransaction, VerifiedWikiEvent) error) (WikiReceiptOutcome, error) {
	if i.DB == nil || i.Verifier == nil || persist == nil || connection == "" || strings.IndexFunc(connection, func(r rune) bool { return r < 32 || r == 127 }) >= 0 {
		return "", errors.New("envai: invalid wiki inbox configuration or connection scope")
	}
	event, err := i.Verifier.Verify(ctx, method, headers, body)
	if err != nil {
		return "", err
	}
	// All identity components must exclude NUL, so delimiter-based keys are unambiguous.
	if strings.ContainsRune(event.WorkspaceID, '\x00') {
		return "", &ProtocolError{"invalid wiki workspace identity"}
	}
	identity := wikiReceiptKey(connection, event.WorkspaceID, event.EventID)
	hash := sha256.Sum256(body)
	digest := hex.EncodeToString(hash[:])
	tx, err := i.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	inserted, err := tx.ExecContext(ctx, `INSERT INTO envai_wiki_webhook_inbox (deduplication_key,event_id,workspace_id,payload_hash,event_type) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (deduplication_key) DO NOTHING`, identity, event.EventID, event.WorkspaceID, digest, event.EventType)
	if err != nil {
		return "", err
	}
	count, err := inserted.RowsAffected()
	if err != nil {
		return "", err
	}
	var stored string
	if err = tx.QueryRowContext(ctx, `SELECT payload_hash FROM envai_wiki_webhook_inbox WHERE deduplication_key=$1 FOR UPDATE`, identity).Scan(&stored); err != nil {
		return "", err
	}
	if subtle.ConstantTimeCompare([]byte(stored), []byte(digest)) != 1 {
		return WikiReceiptCollision, nil
	}
	outcome := WikiReceiptDuplicate
	if count == 1 {
		if err = persist(ctx, wikiTransaction{tx}, event); err != nil {
			return "", err
		}
		outcome = WikiReceiptAccepted
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return outcome, nil
}

// ReceiveToOutbox stores the exact verified body in the included durable outbox.
// Dispatch later with your own idempotent effects policy and current-source gates.
func (i *WikiWebhookInbox) ReceiveToOutbox(ctx context.Context, connection, method string, headers http.Header, body []byte) (WikiReceiptOutcome, error) {
	return i.Receive(ctx, connection, method, headers, body, func(ctx context.Context, tx WikiTransaction, event VerifiedWikiEvent) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO envai_wiki_application_outbox (deduplication_key,event_id,workspace_id,payload) VALUES ($1,$2,$3,$4)`, wikiReceiptKey(connection, event.WorkspaceID, event.EventID), event.EventID, event.WorkspaceID, []byte(event.Raw))
		return err
	})
}
