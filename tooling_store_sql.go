package envai

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SQLDefinitionStore supports caller-supplied database/sql SQLite and MySQL
// drivers. No driver is imported or connection created by this store. Lock rows
// are committed ownership records, so a crash fails closed rather than allowing
// a second deploy to guess whether a remote mutation landed. After inspecting
// remote effects, an operator can remove an abandoned row by exact owner token.
type SQLDefinitionStore struct {
	DB          *sql.DB
	Dialect     string
	LockTimeout time.Duration
}

// DefinitionSQLSchema returns explicit caller-owned migrations. Never run them
// implicitly at import, construction, CLI startup or synchronization time.
func DefinitionSQLSchema(dialect string) ([]string, error) {
	switch dialect {
	case "sqlite":
		return []string{`CREATE TABLE envai_definition_state (scope TEXT PRIMARY KEY, snapshot TEXT NOT NULL)`, `CREATE TABLE envai_definition_locks (scope TEXT PRIMARY KEY, owner TEXT NOT NULL)`}, nil
	case "mysql":
		return []string{`CREATE TABLE envai_definition_state (scope VARBINARY(191) PRIMARY KEY, snapshot LONGTEXT NOT NULL) ENGINE=InnoDB`, `CREATE TABLE envai_definition_locks (scope VARBINARY(191) PRIMARY KEY, owner VARCHAR(64) NOT NULL) ENGINE=InnoDB`}, nil
	}
	return nil, errors.New("envai: definition SQL store supports sqlite/mysql")
}
func (s *SQLDefinitionStore) WithLock(ctx context.Context, scope string, fn func(DefinitionSession) error) (result error) {
	if s.DB == nil || scope == "" {
		return errors.New("envai: SQL definitions store requires database and scope")
	}
	if _, e := DefinitionSQLSchema(s.Dialect); e != nil {
		return e
	}
	if len(scope) > 191 {
		return errors.New("envai: SQL definition scope exceeds 191 bytes")
	}
	if s.Dialect == "mysql" {
		var count int
		if e := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.columns WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ('envai_definition_state','envai_definition_locks') AND COLUMN_NAME='scope' AND DATA_TYPE='varbinary'").Scan(&count); e != nil {
			return e
		}
		if count != 2 {
			return errors.New("envai: MySQL definition scope columns require VARBINARY; apply an explicit schema migration")
		}
	}
	token := make([]byte, 24)
	if _, e := rand.Read(token); e != nil {
		return e
	}
	owner := hex.EncodeToString(token)
	timeout := s.LockTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	acquire, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		_, e := s.DB.ExecContext(acquire, "INSERT INTO envai_definition_locks(scope,owner) VALUES(?,?)", scope, owner)
		if e == nil {
			break
		}
		if acquire.Err() != nil {
			return acquire.Err()
		}
		var held string
		readErr := s.DB.QueryRowContext(acquire, "SELECT owner FROM envai_definition_locks WHERE scope=?", scope).Scan(&held)
		if readErr != nil && !errors.Is(readErr, sql.ErrNoRows) {
			return fmt.Errorf("envai: cannot read definitions lock: %w", readErr)
		}
		if errors.Is(readErr, sql.ErrNoRows) {
			return e
		}
		select {
		case <-acquire.Done():
			return acquire.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
	defer func() {
		release, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		_, e := s.DB.ExecContext(release, "DELETE FROM envai_definition_locks WHERE scope=? AND owner=?", scope, owner)
		result = errors.Join(result, e)
	}()
	return fn(sqlDefinitionSession{ctx, s.DB, s.Dialect, scope, owner})
}

type sqlDefinitionSession struct {
	ctx                   context.Context
	db                    *sql.DB
	dialect, scope, owner string
}

func (s sqlDefinitionSession) Load() (DefinitionSnapshot, error) {
	var b []byte
	e := s.db.QueryRowContext(s.ctx, "SELECT snapshot FROM envai_definition_state WHERE scope=?", s.scope).Scan(&b)
	if errors.Is(e, sql.ErrNoRows) {
		return emptyDefinitionSnapshot(), nil
	}
	if e != nil {
		return DefinitionSnapshot{}, e
	}
	return decodeDefinitionSnapshot(b)
}
func (s sqlDefinitionSession) Save(v DefinitionSnapshot) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	tx, e := s.db.BeginTx(s.ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var owner string
	if e = tx.QueryRowContext(s.ctx, "SELECT owner FROM envai_definition_locks WHERE scope=?", s.scope).Scan(&owner); e != nil {
		return e
	}
	if owner != s.owner {
		return errors.New("envai: definitions lock ownership changed")
	}
	query := "INSERT INTO envai_definition_state(scope,snapshot) VALUES(?,?) ON CONFLICT(scope) DO UPDATE SET snapshot=excluded.snapshot"
	if s.dialect == "mysql" {
		query = "INSERT INTO envai_definition_state(scope,snapshot) VALUES(?,?) ON DUPLICATE KEY UPDATE snapshot=VALUES(snapshot)"
	}
	if _, e = tx.ExecContext(s.ctx, query, s.scope, string(b)); e != nil {
		return e
	}
	return tx.Commit()
}
