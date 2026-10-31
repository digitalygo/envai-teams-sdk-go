package tooling_test

import (
	"context"
	"database/sql"
	"errors"
	envai "github.com/digitalygo/envai-teams-sdk-go"
	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLDefinitionStoreEngines(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			dsn := filepath.Join(t.TempDir(), "state.db")
			if dialect == "mysql" {
				dsn = os.Getenv("ENVAI_TEST_MYSQL_DSN")
				if dsn == "" {
					t.Skip("set disposable MySQL DSN")
				}
			}
			db, err := sql.Open(dialect, dsn)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			schema, err := envai.DefinitionSQLSchema(dialect)
			if err != nil {
				t.Fatal(err)
			}
			for _, ddl := range schema {
				if _, err = db.Exec(ddl); err != nil {
					t.Fatal(err)
				}
			}
			defer db.Exec("DROP TABLE envai_definition_state")
			defer db.Exec("DROP TABLE envai_definition_locks")
			ctx := context.Background()
			store := &envai.SQLDefinitionStore{DB: db, Dialect: dialect, LockTimeout: 50 * time.Millisecond}
			if err = store.WithLock(ctx, "one", func(s envai.DefinitionSession) error {
				v, e := s.Load()
				if e != nil {
					return e
				}
				v.ReadyHash = "ready"
				return s.Save(v)
			}); err != nil {
				t.Fatal(err)
			}
			identities := []string{"Account", "account", "cafe", "café", "account "}
			for _, scope := range identities {
				if e := store.WithLock(ctx, scope, func(s envai.DefinitionSession) error {
					v, e := s.Load()
					if e != nil {
						return e
					}
					v.ReadyHash = scope
					return s.Save(v)
				}); e != nil {
					t.Fatal(e)
				}
			}
			for _, scope := range identities {
				if e := store.WithLock(ctx, scope, func(s envai.DefinitionSession) error {
					v, e := s.Load()
					if v.ReadyHash != scope {
						t.Errorf("scope alias %q: %+v", scope, v)
					}
					return e
				}); e != nil {
					t.Fatal(e)
				}
			}
			if e := store.WithLock(ctx, "Account", func(envai.DefinitionSession) error {
				return store.WithLock(ctx, "account ", func(envai.DefinitionSession) error { return nil })
			}); e != nil {
				t.Fatal(e)
			}
			reopened := &envai.SQLDefinitionStore{DB: db, Dialect: dialect, LockTimeout: 40 * time.Millisecond}
			if err = reopened.WithLock(ctx, "one", func(s envai.DefinitionSession) error {
				v, e := s.Load()
				if v.ReadyHash != "ready" {
					t.Error(v)
				}
				return e
			}); err != nil {
				t.Fatal(err)
			}
			if err = store.WithLock(ctx, "one", func(s envai.DefinitionSession) error {
				e := reopened.WithLock(ctx, "one", func(envai.DefinitionSession) error { t.Error("stole lock"); return nil })
				if !errors.Is(e, context.DeadlineExceeded) {
					t.Errorf("wait returned %v", e)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			// Callback errors release locks without undoing already committed checkpoints.
			sentinel := errors.New("later operation failed")
			if err = store.WithLock(ctx, "one", func(s envai.DefinitionSession) error {
				v, e := s.Load()
				if e != nil {
					return e
				}
				v.ReadyHash = ""
				if e = s.Save(v); e != nil {
					return e
				}
				return sentinel
			}); !errors.Is(err, sentinel) {
				t.Fatal(err)
			}
			if err = store.WithLock(ctx, "one", func(s envai.DefinitionSession) error {
				v, e := s.Load()
				if v.ReadyHash != "" {
					t.Error("stale ready")
				}
				return e
			}); err != nil {
				t.Fatal(err)
			}
			// Corruption fails closed; no recreation or overwriting by read.
			if _, err = db.Exec("UPDATE envai_definition_state SET snapshot=? WHERE scope=?", "not-json", "one"); err != nil {
				t.Fatal(err)
			}
			if err = store.WithLock(ctx, "one", func(s envai.DefinitionSession) error { _, e := s.Load(); return e }); err == nil {
				t.Fatal("corruption accepted")
			}
			// Lost owner prevents checkpoint mutation; transaction rolls back.
			if err = store.WithLock(ctx, "two", func(s envai.DefinitionSession) error {
				v, e := s.Load()
				if e != nil {
					return e
				}
				if _, e = db.Exec("UPDATE envai_definition_locks SET owner=? WHERE scope=?", "other", "two"); e != nil {
					return e
				}
				if e = s.Save(v); e == nil {
					t.Error("owner loss accepted")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			var count int
			if err = db.QueryRow("SELECT COUNT(*) FROM envai_definition_state WHERE scope=?", "two").Scan(&count); err != nil || count != 0 {
				t.Fatal(count, err)
			}
			if dialect == "mysql" {
				if _, e := db.Exec("DELETE FROM envai_definition_state"); e != nil {
					t.Fatal(e)
				}
				if _, e := db.Exec("DELETE FROM envai_definition_locks"); e != nil {
					t.Fatal(e)
				}
				if _, e := db.Exec("ALTER TABLE envai_definition_state MODIFY scope VARCHAR(191)"); e != nil {
					t.Fatal(e)
				}
				if e := store.WithLock(ctx, "unsafe-schema", func(envai.DefinitionSession) error { t.Fatal("unsafe schema admitted"); return nil }); e == nil {
					t.Fatal("unsafe schema accepted")
				}
			}

		})
	}
}
