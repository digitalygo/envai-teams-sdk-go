package envai

import (
	"context"
	"database/sql"
	"errors"
)

// PostgresCallStoreSchema is an explicit, caller-owned migration. No constructor
// or method executes DDL. Use a dedicated database or schema/search_path.
const PostgresCallStoreSchema = `CREATE TABLE envai_client_tool_calls (
 identity_key CHAR(64) PRIMARY KEY,
 namespace TEXT NOT NULL,
 chat_id TEXT NOT NULL,
 call_id TEXT NOT NULL,
 turn_id TEXT NOT NULL,
 tool_name TEXT NOT NULL,
 state TEXT NOT NULL CHECK (state IN ('started','submitted','resolved','ambiguous')),
 version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0)
);`

// PostgresCallStore coordinates multiple processes using atomic claims and
// locked compare-and-swap transitions. Existing claims are never reclaimed.
// The caller owns DB lifetime and namespace selection on the ToolRunner.
type PostgresCallStore struct{ DB *sql.DB }

func (s *PostgresCallStore) Claim(ctx context.Context, id CallIdentity) (bool, error) {
	if s.DB == nil {
		return false, errors.New("envai: PostgreSQL call store requires DB")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO envai_client_tool_calls(identity_key,namespace,chat_id,call_id,turn_id,tool_name,state) VALUES($1,$2,$3,$4,$5,$6,'started') ON CONFLICT(identity_key) DO NOTHING`, identityKey(id), id.Namespace, id.ChatID, id.CallID, id.TurnID, id.ToolName)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	_, _, err = readPostgresCall(tx.QueryRowContext(ctx, `SELECT namespace,chat_id,call_id,turn_id,tool_name,state,version FROM envai_client_tool_calls WHERE identity_key=$1 FOR UPDATE`, identityKey(id)), id)
	if err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return count == 1, nil
}
func (s *PostgresCallStore) State(ctx context.Context, id CallIdentity) (CallState, bool, error) {
	if s.DB == nil {
		return "", false, errors.New("envai: PostgreSQL call store requires DB")
	}
	state, _, err := readPostgresCall(s.DB.QueryRowContext(ctx, `SELECT namespace,chat_id,call_id,turn_id,tool_name,state,version FROM envai_client_tool_calls WHERE identity_key=$1`, identityKey(id)), id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return state, err == nil, err
}
func (s *PostgresCallStore) Record(ctx context.Context, id CallIdentity, state CallState) error {
	if s.DB == nil || !validCallState(state) {
		return errors.New("envai: invalid PostgreSQL call store or state")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	old, version, err := readPostgresCall(tx.QueryRowContext(ctx, `SELECT namespace,chat_id,call_id,turn_id,tool_name,state,version FROM envai_client_tool_calls WHERE identity_key=$1 FOR UPDATE`, identityKey(id)), id)
	if err != nil {
		return err
	}
	// Results observed by another worker cannot be overwritten by a late local
	// submission. Ambiguous claims never become executable again.
	if old == CallResolved || old == CallAmbiguous && state != CallResolved || old == CallSubmitted && state == CallStarted {
		return tx.Commit()
	}
	result, err := tx.ExecContext(ctx, `UPDATE envai_client_tool_calls SET state=$1,version=version+1 WHERE identity_key=$2 AND version=$3`, state, identityKey(id), version)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("envai: concurrent call state transition")
	}
	return tx.Commit()
}
func readPostgresCall(row *sql.Row, expected CallIdentity) (CallState, int64, error) {
	var actual CallIdentity
	var state CallState
	var version int64
	if err := row.Scan(&actual.Namespace, &actual.ChatID, &actual.CallID, &actual.TurnID, &actual.ToolName, &state, &version); err != nil {
		return "", 0, err
	}
	if !validCallState(state) || version < 1 {
		return "", 0, errors.New("envai: corrupt call state; reconciliation required")
	}
	if identityKey(actual) != identityKey(expected) || (expected.TurnID != "" && actual.TurnID != expected.TurnID) || (expected.ToolName != "" && actual.ToolName != expected.ToolName) {
		return "", 0, errors.New("envai: mismatched call state")
	}
	return state, version, nil
}
