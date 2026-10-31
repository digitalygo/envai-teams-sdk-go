package envai

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MemoryDefinitionStore is ephemeral; use only for tests or intentionally
// process-local state. Durable stores are required for deployment automation.
type MemoryDefinitionStore struct {
	mu     sync.Mutex
	scopes map[string]DefinitionSnapshot
}
type memoryDefinitionSession struct {
	store *MemoryDefinitionStore
	scope string
}

func (s *MemoryDefinitionStore) WithLock(ctx context.Context, scope string, fn func(DefinitionSession) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.scopes == nil {
		s.scopes = map[string]DefinitionSnapshot{}
	}
	return fn(memoryDefinitionSession{s, scope})
}
func cloneDefinitionSnapshot(s DefinitionSnapshot) (DefinitionSnapshot, error) {
	b, e := json.Marshal(s)
	if e != nil {
		return s, e
	}
	var out DefinitionSnapshot
	e = json.Unmarshal(b, &out)
	return out, e
}
func (s memoryDefinitionSession) Load() (DefinitionSnapshot, error) {
	v, ok := s.store.scopes[s.scope]
	if !ok {
		v = emptyDefinitionSnapshot()
	}
	return cloneDefinitionSnapshot(v)
}
func (s memoryDefinitionSession) Save(v DefinitionSnapshot) error {
	copy, e := cloneDefinitionSnapshot(v)
	if e == nil {
		s.store.scopes[s.scope] = copy
	}
	return e
}

// FileDefinitionStore stores one private, atomically replaced snapshot per public
// connection scope. A separate OS lock protects the full read/plan/apply sequence.
// Directory must be private and trusted (never a shared world-writable directory).
type FileDefinitionStore struct{ Directory string }
type fileDefinitionSession struct{ path string }

func (s *FileDefinitionStore) WithLock(ctx context.Context, scope string, fn func(DefinitionSession) error) error {
	if s.Directory == "" || scope == "" {
		return errors.New("envai: state directory and scope required")
	}
	if err := os.MkdirAll(s.Directory, 0700); err != nil {
		return err
	}
	base := filepath.Join(s.Directory, fmt.Sprintf("%x", sha256.Sum256([]byte(scope))))
	release, err := definitionFileLock(ctx, base+".lock")
	if err != nil {
		return err
	}
	defer release()
	return fn(fileDefinitionSession{base + ".json"})
}
func decodeDefinitionSnapshot(b []byte) (DefinitionSnapshot, error) {
	var s DefinitionSnapshot
	if _, err := jsoncanonicalizer.Transform(b); err != nil {
		return s, fmt.Errorf("envai: corrupt definitions state: %w", err)
	}
	if e := json.Unmarshal(b, &s); e != nil {
		return s, fmt.Errorf("envai: corrupt definitions state: %w", e)
	}
	if s.FormatVersion != "1" || s.Resources == nil {
		return s, errors.New("envai: unsupported or corrupt definitions state")
	}
	for key, v := range s.Resources {
		if key != v.Type+":"+v.LocalKey || (v.Type != "tool" && v.Type != "agent" && v.Type != "team") {
			return s, errors.New("envai: invalid definitions state identity")
		}
		switch v.Status {
		case "applied", "drifted", "conflicted", "missing":
			if v.RemoteName == "" || (v.Type != "tool" && v.RemoteID == "") || v.LastApplied == nil {
				return s, errors.New("envai: incomplete managed definitions state")
			}
		case "ambiguous":
			if v.PendingAction == "" || v.RemoteName == "" || (v.PendingAction != "delete" && v.Intended == nil) {
				return s, errors.New("envai: incomplete ambiguous definitions state")
			}
		default:
			return s, errors.New("envai: invalid definitions state status")
		}
	}
	return s, nil
}
func (s fileDefinitionSession) Load() (DefinitionSnapshot, error) {
	b, e := os.ReadFile(s.path)
	if os.IsNotExist(e) {
		return emptyDefinitionSnapshot(), nil
	}
	if e != nil {
		return DefinitionSnapshot{}, e
	}
	return decodeDefinitionSnapshot(b)
}
func (s fileDefinitionSession) Save(v DefinitionSnapshot) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(s.path), ".definitions-*")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if e = f.Chmod(0600); e == nil {
		_, e = f.Write(b)
	}
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	if e = os.Rename(name, s.path); e != nil {
		return e
	}
	dir, e := os.Open(filepath.Dir(s.path))
	if e != nil {
		return e
	}
	defer dir.Close()
	return dir.Sync()
}

// PostgresDefinitionStoreSchema is caller-owned migration SQL. Constructors and
// synchronization never execute DDL or silently create missing state tables.
const PostgresDefinitionStoreSchema = `CREATE TABLE envai_definition_state (scope TEXT PRIMARY KEY, snapshot JSONB NOT NULL);`

type PostgresDefinitionStore struct{ DB *sql.DB }
type postgresDefinitionSession struct {
	ctx   context.Context
	conn  *sql.Conn
	scope string
}

func (s *PostgresDefinitionStore) WithLock(ctx context.Context, scope string, fn func(DefinitionSession) error) error {
	if s.DB == nil || scope == "" {
		return errors.New("envai: database and scope required")
	}
	conn, e := s.DB.Conn(ctx)
	if e != nil {
		return e
	}
	defer conn.Close()
	digest := sha256.Sum256([]byte("envai-definitions:" + scope))
	key := int64(binary.BigEndian.Uint64(digest[:8]))
	if _, e = conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", key); e != nil {
		_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		return e
	}
	defer func() {
		release, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := conn.ExecContext(release, "SELECT pg_advisory_unlock($1)", key); err != nil {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
	}()
	return fn(postgresDefinitionSession{ctx, conn, scope})
}
func (s postgresDefinitionSession) Load() (DefinitionSnapshot, error) {
	var b []byte
	e := s.conn.QueryRowContext(s.ctx, "SELECT snapshot FROM envai_definition_state WHERE scope=$1", s.scope).Scan(&b)
	if errors.Is(e, sql.ErrNoRows) {
		return emptyDefinitionSnapshot(), nil
	}
	if e != nil {
		return DefinitionSnapshot{}, e
	}
	return decodeDefinitionSnapshot(b)
}
func (s postgresDefinitionSession) Save(v DefinitionSnapshot) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	tx, e := s.conn.BeginTx(s.ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(s.ctx, "INSERT INTO envai_definition_state(scope,snapshot) VALUES($1,$2) ON CONFLICT(scope) DO UPDATE SET snapshot=excluded.snapshot", s.scope, string(b))
	if e != nil {
		return e
	}
	return tx.Commit()
}
