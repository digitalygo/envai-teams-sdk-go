package envai

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type toolingFake struct {
	resources map[string][]DefinitionPayload
	calls     []string
	fail      string
	landed    bool
	different bool
	modelsErr error
	reads     int
	failure   error
}

func (f *toolingFake) Models(context.Context) (map[string]bool, error) {
	return map[string]bool{"test-model": true}, f.modelsErr
}
func (f *toolingFake) List(_ context.Context, kind string) ([]DefinitionPayload, error) {
	f.reads++
	return f.resources[kind], nil
}
func (f *toolingFake) Mutate(_ context.Context, kind, action, target string, p DefinitionPayload) (DefinitionPayload, error) {
	f.calls = append(f.calls, kind+":"+action+":"+target)
	failure := f.fail == kind+":"+action
	if failure && !f.landed {
		if f.failure != nil {
			return nil, f.failure
		}
		return nil, errors.New("lost connection")
	}
	actual := definitionClone(p)
	if actual != nil {
		actual["id"] = kind + "-" + definitionString(p, "name")
		if action == "update" && kind != "tool" {
			actual["id"] = target
		}
		actual["updated_at"] = "2026-01-01T00:00:00Z"
		if f.different {
			actual["description"] = "other"
		}
	}
	old := f.resources[kind]
	next := []DefinitionPayload{}
	for _, v := range old {
		match := definitionString(v, "id") == target
		if kind == "tool" || action == "create" {
			match = definitionString(v, "name") == target
		}
		if !match {
			next = append(next, v)
		}
	}
	if action != "delete" {
		next = append(next, actual)
	}
	f.resources[kind] = next
	if failure {
		return nil, errors.New("response lost")
	}
	return actual, nil
}
func toolingManifest() DefinitionManifest {
	m := EmptyDefinitionManifest()
	m.Tools["echo"] = DefinitionPayload{"name": "echo", "kind": "client", "description": "before", "parameters": map[string]any{"type": "object", "properties": map[string]any{"text": map[string]any{"type": "string"}}}, "timeout_seconds": float64(30)}
	return m
}
func toolingFixture(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	b, e := os.ReadFile("testdata/parity/tooling-golden.json")
	if e != nil {
		t.Fatal(e)
	}
	var v map[string]json.RawMessage
	if e = json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func TestToolingGoldenHashes(t *testing.T) {
	var vectors []struct {
		Document DefinitionManifest
		SHA      string `json:"sha256"`
	}
	if e := json.Unmarshal(toolingFixture(t)["hash_vectors"], &vectors); e != nil {
		t.Fatal(e)
	}
	for _, v := range vectors {
		h, e := v.Document.Hash()
		if e != nil || h != v.SHA {
			t.Fatalf("%s want %s: %v", h, v.SHA, e)
		}
	}
}
func TestToolingGoldenPlansAndReconciliation(t *testing.T) {
	overridesBytes, e := os.ReadFile("testdata/parity/safety-overrides.json")
	if e != nil {
		t.Fatal(e)
	}
	var safety struct {
		Overrides []struct {
			CaseID   string `json:"case_id"`
			Expected struct {
				Ready bool `json:"manifest_ready"`
			}
		}
	}
	if e = json.Unmarshal(overridesBytes, &safety); e != nil {
		t.Fatal(e)
	}
	overrides := map[string]bool{}
	for _, v := range safety.Overrides {
		overrides[v.CaseID] = v.Expected.Ready
	}

	var cases []struct {
		ID                           string
		Desired, LastApplied, Remote DefinitionPayload
		Options                      json.RawMessage
		Expected                     struct {
			Operations []DefinitionOperation
			Status     string `json:"report_status"`
			Ready      bool   `json:"manifest_ready"`
			Outcomes   []string
		}
	}
	var raw []map[string]json.RawMessage
	if e := json.Unmarshal(toolingFixture(t)["planner_reconciler_cases"], &raw); e != nil {
		t.Fatal(e)
	}
	for _, r := range raw {
		b, _ := json.Marshal(r)
		var c struct {
			ID          string
			Desired     DefinitionPayload
			LastApplied DefinitionPayload `json:"last_applied"`
			Remote      DefinitionPayload
			Options     json.RawMessage
			Expected    struct {
				Operations []DefinitionOperation
				Status     string `json:"report_status"`
				Ready      bool   `json:"manifest_ready"`
				Outcomes   []string
			}
		}
		_ = json.Unmarshal(b, &c)
		cases = append(cases, struct {
			ID                           string
			Desired, LastApplied, Remote DefinitionPayload
			Options                      json.RawMessage
			Expected                     struct {
				Operations []DefinitionOperation
				Status     string `json:"report_status"`
				Ready      bool   `json:"manifest_ready"`
				Outcomes   []string
			}
		}{c.ID, c.Desired, c.LastApplied, c.Remote, c.Options, c.Expected})
	}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			m := EmptyDefinitionManifest()
			if c.Desired != nil {
				m.Tools["echo"] = c.Desired
			}
			store := &MemoryDefinitionStore{}
			if c.LastApplied != nil {
				_ = store.WithLock(context.Background(), "test", func(s DefinitionSession) error {
					state := emptyDefinitionSnapshot()
					state.Resources["tool:echo"] = DefinitionState{Type: "tool", LocalKey: "echo", RemoteID: "twl_echo", RemoteName: "echo", LastApplied: c.LastApplied, Status: "applied"}
					return s.Save(state)
				})
			}
			remote := &toolingFake{resources: map[string][]DefinitionPayload{}}
			if c.Remote != nil {
				r := definitionClone(c.Remote)
				r["id"] = "twl_echo"
				r["updated_at"] = "2026-01-01T00:00:00Z"
				remote.resources["tool"] = []DefinitionPayload{r}
			}
			var opts map[string]any
			if len(c.Options) > 0 && c.Options[0] == '{' {
				_ = json.Unmarshal(c.Options, &opts)
			}
			option := DefinitionOptions{}
			if opts != nil {
				option.Adopt, _ = opts["adopt"].(bool)
				option.OverwriteDrift, _ = opts["overwriteDrift"].(bool)
				option.RecreateMissing, _ = opts["recreateMissing"].(bool)
				option.Prune, _ = opts["prune"].(bool)
				option.DryRun, _ = opts["dryRun"].(bool)
				option.Scope, _ = opts["scope"].(string)
			}
			d := Definitions{Remote: remote, Store: store, Scope: "test"}
			report, e := d.Sync(context.Background(), m, option)
			if e != nil {
				t.Fatal(e)
			}
			expectedReady := c.Expected.Ready
			if value, ok := overrides[c.ID]; ok {
				expectedReady = value
			}
			if report.Status != c.Expected.Status || report.ManifestReady != expectedReady {
				t.Fatalf("%s ready %v; want %s %v", report.Status, report.ManifestReady, c.Expected.Status, c.Expected.Ready)
			}
			if len(report.Plan.Operations) != len(c.Expected.Operations) {
				t.Fatal("operation length")
			}
			for i, op := range report.Plan.Operations {
				want := c.Expected.Operations[i]
				if op.Classification != want.Classification || op.Action != want.Action || op.Blocking != want.Blocking {
					t.Fatalf("operation %+v want %+v", op, want)
				}
			}
			for i, result := range report.Results {
				if result.Outcome != c.Expected.Outcomes[i] {
					t.Fatalf("outcome %s want %s", result.Outcome, c.Expected.Outcomes[i])
				}
			}
		})
	}
}
func toolingFullManifest() DefinitionManifest {
	m := toolingManifest()
	m.Agents["coordinator"] = DefinitionPayload{"name": "Coordinator", "system_prompt": "Coordinate.", "mode": "primary", "can_write_knowledge": false, "can_write_skills": false}
	m.Agents["worker"] = DefinitionPayload{"name": "Worker", "system_prompt": "Work.", "mode": "subagent", "can_write_knowledge": false, "can_write_skills": false}
	m.Teams["team"] = DefinitionPayload{"name": "Example team", "orchestrator_id": "coordinator", "members": []any{map[string]any{"agent_definition_id": "worker", "role": "research"}}}
	return m
}
func TestToolingDependencyRenamePruneAndReadiness(t *testing.T) {
	ctx := context.Background()
	m := toolingFullManifest()
	remote := &toolingFake{resources: map[string][]DefinitionPayload{}}
	d := Definitions{Remote: remote, Store: &MemoryDefinitionStore{}, Scope: "test"}
	r, e := d.Sync(ctx, m, DefinitionOptions{})
	if e != nil || !r.ManifestReady {
		t.Fatalf("%+v %v", r, e)
	}
	want := []string{"tool:create:echo", "agent:create:Coordinator", "agent:create:Worker", "team:create:Example team"}
	if !reflect.DeepEqual(remote.calls, want) {
		t.Fatal(remote.calls)
	}
	team := remote.resources["team"][0]
	if team["orchestrator_id"] != "agent-Coordinator" {
		t.Fatal(team)
	}
	if e = d.RequireSynchronized(ctx, m); e != nil {
		t.Fatal(e)
	}
	lookup, e := d.Lookup(ctx, m, "team", "team")
	if e != nil || lookup.RemoteID != "team-Example team" {
		t.Fatal(lookup, e)
	}
	remote.calls = nil
	r, e = d.Sync(ctx, m, DefinitionOptions{})
	if e != nil || !r.ManifestReady || len(remote.calls) != 0 {
		t.Fatalf("rerun %+v %v calls %v", r, e, remote.calls)
	}
	m.Agents["worker"]["name"] = "Worker renamed"
	r, e = d.Sync(ctx, m, DefinitionOptions{})
	if e != nil || !r.ManifestReady || !reflect.DeepEqual(remote.calls, []string{"agent:update:agent-Worker"}) {
		t.Fatal(r, e, remote.calls)
	}
	remote.calls = nil
	r, e = d.Sync(ctx, EmptyDefinitionManifest(), DefinitionOptions{Prune: true})
	if e != nil || !r.ManifestReady {
		t.Fatal(r, e)
	}
	if len(remote.calls) != 4 || !strings.HasPrefix(remote.calls[0], "team:delete") || !strings.HasPrefix(remote.calls[3], "tool:delete") {
		t.Fatal(remote.calls)
	}
	if _, e = d.Lookup(ctx, m, "team", "team"); e == nil {
		t.Fatal("stale lookup")
	}
}
func TestToolingAmbiguousWriteAndPartialFailure(t *testing.T) {
	for _, tc := range []struct {
		name              string
		landed, different bool
		status            string
	}{{"landed", true, false, "succeeded"}, {"before", false, false, "ambiguous"}, {"different", true, true, "ambiguous"}} {
		t.Run(tc.name, func(t *testing.T) {
			remote := &toolingFake{resources: map[string][]DefinitionPayload{}, fail: "tool:create", landed: tc.landed, different: tc.different}
			d := Definitions{Remote: remote, Store: &MemoryDefinitionStore{}, Scope: "test"}
			r, e := d.Sync(context.Background(), toolingManifest(), DefinitionOptions{})
			if e != nil || r.Status != tc.status || len(remote.calls) != 1 {
				t.Fatal(r, e, remote.calls)
			}
		})
	}
	remote := &toolingFake{resources: map[string][]DefinitionPayload{}, fail: "agent:create", failure: &APIError{Status: 422}}
	d := Definitions{Remote: remote, Store: &MemoryDefinitionStore{}, Scope: "test"}
	r, e := d.Sync(context.Background(), toolingFullManifest(), DefinitionOptions{})
	if e != nil || r.Status != "failed" || r.ManifestReady || len(remote.calls) != 2 {
		t.Fatal(r, e, remote.calls)
	}
	s, e := d.State(context.Background())
	if e != nil || s.Resources["tool:echo"].Status != "applied" || s.ReadyHash != "" {
		t.Fatal(s, e)
	}
}
func TestToolingCompileAndValidation(t *testing.T) {
	m := toolingFullManifest()
	b, _ := json.Marshal(m)
	got, e := CompileDefinitions(b)
	if e != nil {
		t.Fatal(e)
	}
	h, _ := m.Hash()
	h2, _ := got.Hash()
	if h != h2 {
		t.Fatal(h, h2)
	}
	tests := []func(*DefinitionManifest){func(m *DefinitionManifest) { m.FormatVersion = "2" }, func(m *DefinitionManifest) { m.Tools["echo"]["name"] = "bad" }, func(m *DefinitionManifest) { m.Tools["echo"]["parameters"] = []any{} }, func(m *DefinitionManifest) { m.Tools["echo"]["timeout_seconds"] = 0 }, func(m *DefinitionManifest) { m.Agents["worker"]["system_prompt"] = "" }, func(m *DefinitionManifest) { m.Agents["worker"]["mode"] = "wrong" }, func(m *DefinitionManifest) {
		m.Agents["worker"]["config"] = map[string]any{"nested": []any{map[string]any{"max_tokens": 4}}}
	}, func(m *DefinitionManifest) { m.Teams["team"]["orchestrator_id"] = "worker" }, func(m *DefinitionManifest) {
		m.Teams["team"]["members"] = []any{map[string]any{"agent_definition_id": "coordinator"}}
	}, func(m *DefinitionManifest) { m.Agents["worker"]["name"] = "Coordinator" }, func(m *DefinitionManifest) {
		m.Agents["worker"]["external_tools_enabled"] = []any{map[string]any{"name": "missing"}}
	}, func(m *DefinitionManifest) { m.Agents["worker"]["permissions"] = []any{} }, func(m *DefinitionManifest) { m.Agents["worker"]["can_write_skills"] = "yes" }}
	for i, mutate := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			m := toolingFullManifest()
			mutate(&m)
			if e := m.Validate(nil); e == nil {
				t.Fatal("invalid accepted")
			}
		})
	}
	m = toolingFullManifest()
	m.Agents["worker"]["model"] = "missing"
	if e = m.Validate(map[string]bool{}); e == nil {
		t.Fatal("model accepted")
	}
	if e = m.Validate(nil); e != nil {
		t.Fatal("offline rejected", e)
	}
	dir := t.TempDir()
	if e = os.Mkdir(filepath.Join(dir, "agents"), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, "agents", "worker.md"), []byte("---\r\nname: Worker\r\nmode: subagent\r\n---\r\nWork.\r\n"), 0600); e != nil {
		t.Fatal(e)
	}
	loaded, e := LoadDefinitions(dir)
	if e != nil || loaded.Agents["worker"]["system_prompt"] != "Work." {
		t.Fatal(loaded, e)
	}
	// Null, empty object and empty array remain distinct canonical values.
	a := toolingManifest()
	a.Tools["echo"]["parameters"].(map[string]any)["enum"] = []any{nil, map[string]any{}, []any{}}
	b, _ = json.Marshal(a)
	c, e := CompileDefinitions(b)
	if e != nil {
		t.Fatal(e)
	}
	x, _ := a.Hash()
	y, _ := c.Hash()
	if x != y {
		t.Fatal("schema shape lost")
	}
}
func TestToolingFileLockDurabilityIsolationAndCorruption(t *testing.T) {
	ctx := context.Background()
	store := &FileDefinitionStore{Directory: t.TempDir()}
	for _, scope := range []string{"a", "b"} {
		if e := store.WithLock(ctx, scope, func(s DefinitionSession) error {
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
	reopened := &FileDefinitionStore{Directory: store.Directory}
	for _, scope := range []string{"a", "b"} {
		if e := reopened.WithLock(ctx, scope, func(s DefinitionSession) error {
			v, e := s.Load()
			if e == nil && v.ReadyHash != scope {
				t.Fatal(v)
			}
			return e
		}); e != nil {
			t.Fatal(e)
		}
	}
	locked := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- store.WithLock(ctx, "a", func(DefinitionSession) error { close(locked); <-release; return nil })
	}()
	<-locked
	short, cancel := context.WithTimeout(ctx, 40*time.Millisecond)
	defer cancel()
	if e := reopened.WithLock(short, "a", func(DefinitionSession) error { return nil }); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
	close(release)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
	if e := store.WithLock(ctx, "a", func(s DefinitionSession) error {
		return os.WriteFile(s.(fileDefinitionSession).path, []byte("{corrupt"), 0600)
	}); e != nil {
		t.Fatal(e)
	}
	if e := store.WithLock(ctx, "a", func(s DefinitionSession) error { _, e := s.Load(); return e }); e == nil {
		t.Fatal("corrupt accepted")
	}
	if _, e := decodeDefinitionSnapshot([]byte(`{"format_version":"2","resources":{}}`)); e == nil {
		t.Fatal("unknown version accepted")
	}
	// Race test exercises separate store objects and OS locks, not shared mutex.
	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := &FileDefinitionStore{Directory: store.Directory}
			if e := s.WithLock(ctx, "counter", func(session DefinitionSession) error {
				v, e := session.Load()
				if e != nil {
					return e
				}
				v.ReadyAt = v.ReadyAt.Add(time.Second)
				return session.Save(v)
			}); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	if e := store.WithLock(ctx, "counter", func(s DefinitionSession) error {
		v, e := s.Load()
		if v.ReadyAt.Sub(time.Time{}) != 4*time.Second {
			t.Error(v)
		}
		return e
	}); e != nil {
		t.Fatal(e)
	}
}
func TestToolingPostgresStore(t *testing.T) {
	dsn := os.Getenv("ENVAI_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL")
	}
	db, e := sql.Open("postgres", dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	schema := fmt.Sprintf("tooling_%d", time.Now().UnixNano())
	if _, e = db.Exec(`CREATE SCHEMA ` + schema); e != nil {
		t.Fatal(e)
	}
	defer db.Exec(`DROP SCHEMA ` + schema + ` CASCADE`)
	db.SetMaxOpenConns(1)
	if _, e = db.Exec(`SET search_path TO ` + schema); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(PostgresDefinitionStoreSchema); e != nil {
		t.Fatal(e)
	}
	store := &PostgresDefinitionStore{DB: db}
	ctx := context.Background()
	if e = store.WithLock(ctx, "a", func(s DefinitionSession) error {
		v, e := s.Load()
		if e != nil {
			return e
		}
		v.ReadyHash = "saved"
		return s.Save(v)
	}); e != nil {
		t.Fatal(e)
	}
	if e = store.WithLock(ctx, "a", func(s DefinitionSession) error {
		v, e := s.Load()
		if e == nil && v.ReadyHash != "saved" {
			t.Fatal(v)
		}
		return e
	}); e != nil {
		t.Fatal(e)
	}
}
func TestToolingCLIAndLazyGate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	b, _ := json.Marshal(toolingManifest())
	_ = os.WriteFile(path, b, 0600)
	var out, errOut bytes.Buffer
	env := func(string) string { return "" }
	if exit := RunDefinitionsCLI(context.Background(), []string{"definitions:validate", "--path", path, "--json"}, &out, &errOut, env); exit != 0 || !strings.Contains(out.String(), `"valid":true`) {
		t.Fatal(exit, out.String(), errOut.String())
	}
	for _, args := range [][]string{nil, {"unknown"}, {"plan", "--path", path}, {"validate", "--path", "missing"}, {"state", "--connection", "x", "--store", "bad"}, {"validate", "--nonsense"}} {
		if exit := RunDefinitionsCLI(context.Background(), args, &out, &errOut, env); exit == 0 {
			t.Fatal(args)
		}
	}
	args := []string{"state", "--connection", "x", "--state-path", filepath.Join(dir, "state"), "--json"}
	if exit := RunDefinitionsCLI(context.Background(), args, &out, &errOut, env); exit != 0 {
		t.Fatal(exit)
	}
	remote := &toolingFake{resources: map[string][]DefinitionPayload{}}
	d := Definitions{Remote: remote, Store: &MemoryDefinitionStore{}, Scope: "test"}
	m := toolingManifest()
	for _, mode := range []string{"off", "manual", "deploy"} {
		if e := d.Ensure(context.Background(), m, mode, time.Hour); e != nil {
			t.Fatal(e)
		}
	}
	if remote.reads != 0 {
		t.Fatal("implicit calls")
	}
	if e := d.Ensure(context.Background(), m, "invalid", time.Hour); e == nil {
		t.Fatal("invalid mode")
	}
	if e := d.Ensure(context.Background(), m, "lazy", time.Hour); e != nil {
		t.Fatal(e)
	}
	reads := remote.reads
	if e := d.Ensure(context.Background(), m, "lazy", time.Hour); e != nil || reads != remote.reads {
		t.Fatal(e)
	}
}
func TestToolingClientAdapterAndNoMutationRetry(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("auth missing")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/models"):
			fmt.Fprint(w, `{"object":"list","data":[{"id":"test-model"}],"defaults":{}}`)
		case r.Method == "GET":
			fmt.Fprint(w, `{"data":[]}`)
		default:
			w.WriteHeader(503)
			fmt.Fprint(w, `{"error":{"code":"unavailable","message":"test"}}`)
		}
	}))
	defer server.Close()
	client, e := New(Config{BaseURL: server.URL, APIKey: "secret", AllowHTTP: true, RetryAttempts: 3})
	if e != nil {
		t.Fatal(e)
	}
	remote := ClientDefinitionRemote{Client: client}
	if _, e = remote.Models(context.Background()); e != nil {
		t.Fatal(e)
	}
	for _, kind := range []string{"tool", "agent", "team"} {
		if _, e = remote.List(context.Background(), kind); e != nil {
			t.Fatal(e)
		}
	}
	before := len(calls)
	if _, e = remote.Mutate(context.Background(), "tool", "create", "echo", toolingManifest().Tools["echo"]); e == nil {
		t.Fatal("expected error")
	}
	if len(calls) != before+1 {
		t.Fatal("mutation retried", calls)
	}
}

func TestToolingDurableAmbiguityBlocksAndRecovers(t *testing.T) {
	ctx := context.Background()
	remote := &toolingFake{resources: map[string][]DefinitionPayload{}, fail: "tool:create"}
	store := &FileDefinitionStore{Directory: t.TempDir()}
	d := Definitions{Remote: remote, Store: store, Scope: "test"}
	m := toolingManifest()
	first, e := d.Sync(ctx, m, DefinitionOptions{})
	if e != nil || first.Status != "ambiguous" {
		t.Fatal(first, e)
	}
	before := len(remote.calls)
	again, e := d.Sync(ctx, m, DefinitionOptions{Adopt: true, OverwriteDrift: true, RecreateMissing: true, Prune: true})
	if e != nil || again.Status != "blocked" || len(remote.calls) != before {
		t.Fatal(again, e, remote.calls)
	}
	// Simulate authoritative discovery of the previously uncertain mutation.
	actual := definitionClone(m.Tools["echo"])
	actual["id"] = "remote-id"
	remote.resources["tool"] = []DefinitionPayload{actual}
	recovered, e := d.Sync(ctx, m, DefinitionOptions{})
	if e != nil || !recovered.ManifestReady || len(remote.calls) != before {
		t.Fatal(recovered, e, remote.calls)
	}
	s, e := d.State(ctx)
	if e != nil || s.Resources["tool:echo"].Status != "applied" {
		t.Fatal(s, e)
	}
	remote.resources["tool"][0]["description"] = "drift"
	blocked, e := d.Sync(ctx, m, DefinitionOptions{})
	if e != nil || blocked.Status != "blocked" {
		t.Fatal(blocked, e)
	}
	if e = d.RequireSynchronized(ctx, m); e == nil {
		t.Fatal("stale readiness survived blocked deploy")
	}
}
func TestToolingCanonicalSetsAndCLIStatus(t *testing.T) {
	m := toolingFullManifest()
	m.Agents["worker"]["tools_enabled"] = []any{map[string]any{"name": "z"}, map[string]any{"name": "a"}, map[string]any{"name": "z"}}
	a, e := m.Hash()
	if e != nil {
		t.Fatal(e)
	}
	m.Agents["worker"]["tools_enabled"] = []any{map[string]any{"name": "a"}, map[string]any{"name": "z"}}
	b, e := m.Hash()
	if e != nil || a != b {
		t.Fatal(a, b, e)
	}
	m.Agents["worker"]["config"] = map[string]any{}
	m.Agents["worker"]["description"] = nil
	n := definitionNormalize("agent", m.Agents["worker"], nil)
	if _, ok := n["config"]; ok {
		t.Fatal(n)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	raw, _ := json.Marshal(m)
	_ = os.WriteFile(path, raw, 0600)
	var out bytes.Buffer
	args := []string{"status", "--connection", "test", "--path", path, "--state-path", filepath.Join(dir, "state"), "--json"}
	if code := RunDefinitionsCLI(context.Background(), args, &out, &out, func(string) string { return "" }); code != 2 {
		t.Fatal(code, out.String())
	}
	store := &FileDefinitionStore{Directory: filepath.Join(dir, "state")}
	if e = store.WithLock(context.Background(), "test", func(s DefinitionSession) error {
		v, e := s.Load()
		v.ReadyHash, _ = m.Hash()
		if e != nil {
			return e
		}
		return s.Save(v)
	}); e != nil {
		t.Fatal(e)
	}
	if code := RunDefinitionsCLI(context.Background(), args, &out, &out, func(string) string { return "" }); code != 0 {
		t.Fatal(code, out.String())
	}
}

type toolingCrashRemote struct{ toolingFake }

func (r *toolingCrashRemote) Mutate(context.Context, string, string, string, DefinitionPayload) (DefinitionPayload, error) {
	os.Exit(19)
	return nil, nil
}
func TestToolingProcessCrashPersistsIntent(t *testing.T) {
	if dir := os.Getenv("ENVAI_TOOLING_CRASH_DIR"); dir != "" {
		d := Definitions{Remote: &toolingCrashRemote{toolingFake{resources: map[string][]DefinitionPayload{}}}, Store: &FileDefinitionStore{Directory: dir}, Scope: "crash-test"}
		_, _ = d.Sync(context.Background(), toolingManifest(), DefinitionOptions{})
		t.Fatal("did not terminate during remote write")
	}
	dir := t.TempDir()
	binary, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command(binary, "-test.run=^TestToolingProcessCrashPersistsIntent$")
	cmd.Env = append(os.Environ(), "ENVAI_TOOLING_CRASH_DIR="+dir)
	e = cmd.Run()
	var exit *exec.ExitError
	if !errors.As(e, &exit) || exit.ExitCode() != 19 {
		t.Fatal(e)
	}
	remote := &toolingFake{resources: map[string][]DefinitionPayload{}}
	d := Definitions{Remote: remote, Store: &FileDefinitionStore{Directory: dir}, Scope: "crash-test"}
	s, e := d.State(context.Background())
	if e != nil || s.Resources["tool:echo"].Status != "ambiguous" || s.Resources["tool:echo"].Intended == nil {
		t.Fatal(s, e)
	}
	report, e := d.Sync(context.Background(), toolingManifest(), DefinitionOptions{RecreateMissing: true, OverwriteDrift: true, Adopt: true})
	if e != nil || report.Status != "blocked" || len(remote.calls) != 0 {
		t.Fatal(report, e, remote.calls)
	}
}
func TestToolingMalformedJSONFailsClosed(t *testing.T) {
	for _, raw := range []string{`{"format_version":"1","tools":{},"agents":{},"teams":{}} {}`, `{"format_version":"1","format_version":"1","tools":{},"agents":{},"teams":{}}`, `{"format_version":"1","tools":[],"agents":{},"teams":{}}`} {
		if _, e := CompileDefinitions([]byte(raw)); e == nil {
			t.Fatal("malformed accepted", raw)
		}
	}
}

func TestToolingCLIRemotePlanAndDryRun(t *testing.T) {
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "GET" {
			mutations++
		}
		if strings.HasSuffix(r.URL.Path, "/models") {
			fmt.Fprint(w, `{"object":"list","data":[],"defaults":{}}`)
		} else {
			fmt.Fprint(w, `{"data":[]}`)
		}
	}))
	defer server.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	b, _ := json.Marshal(toolingManifest())
	if e := os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	env := func(key string) string {
		return map[string]string{"ENVAI_BASE_URL": server.URL, "ENVAI_API_KEY": "secret", "ENVAI_ALLOW_HTTP": "true"}[key]
	}
	for _, cmd := range []string{"plan", "sync"} {
		args := []string{cmd, "--path", path, "--connection", "test", "--state-path", filepath.Join(dir, "state"), "--json"}
		if cmd == "sync" {
			args = append(args, "--dry-run")
		}
		var out, stderr bytes.Buffer
		if code := RunDefinitionsCLI(context.Background(), args, &out, &stderr, env); code != 0 {
			t.Fatal(code, out.String(), stderr.String())
		}
		if strings.Contains(out.String(), "secret") {
			t.Fatal("credential leak")
		}
		var result map[string]any
		if e := json.Unmarshal(out.Bytes(), &result); e != nil {
			t.Fatal(e)
		}
	}
	if mutations != 0 {
		t.Fatal("dry-run mutated", mutations)
	}
}
func TestToolingDeleteGoneAndDependencyRejection(t *testing.T) {
	for _, status := range []int{404, 422} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			remote := &toolingFake{resources: map[string][]DefinitionPayload{}}
			d := Definitions{Remote: remote, Store: &MemoryDefinitionStore{}, Scope: "test"}
			if r, e := d.Sync(context.Background(), toolingFullManifest(), DefinitionOptions{}); e != nil || !r.ManifestReady {
				t.Fatal(r, e)
			}
			remote.calls = nil
			remote.fail = "team:delete"
			remote.failure = &APIError{Status: status}
			r, e := d.Sync(context.Background(), EmptyDefinitionManifest(), DefinitionOptions{Prune: true})
			if e != nil {
				t.Fatal(e)
			}
			if status == 404 {
				if !r.ManifestReady || len(remote.calls) != 4 {
					t.Fatal(r, remote.calls)
				}
			} else if r.Status != "failed" || r.ManifestReady || len(remote.calls) != 1 {
				t.Fatal(r, remote.calls)
			}
		})
	}
}

func TestToolingLibraryFailOnDriftAndPrepareFailureInvalidateReady(t *testing.T) {
	ctx := context.Background()
	m := toolingManifest()
	remote := &toolingFake{resources: map[string][]DefinitionPayload{}}
	d := Definitions{Remote: remote, Store: &MemoryDefinitionStore{}, Scope: "test"}
	if r, e := d.Sync(ctx, m, DefinitionOptions{}); e != nil || !r.ManifestReady {
		t.Fatal(r, e)
	}
	remote.resources["tool"][0]["description"] = "outside edit"
	before := len(remote.calls)
	r, e := d.Sync(ctx, m, DefinitionOptions{OverwriteDrift: true, FailOnDrift: true})
	if e != nil || r.Status != "blocked" || len(remote.calls) != before {
		t.Fatal(r, e, remote.calls)
	}
	if e = d.RequireSynchronized(ctx, m); e == nil {
		t.Fatal("fail-on-drift kept ready")
	}
	remote.resources["tool"][0]["description"] = "before"
	if r, e = d.Sync(ctx, m, DefinitionOptions{}); e != nil || !r.ManifestReady {
		t.Fatal(r, e)
	}
	remote.modelsErr = errors.New("catalog unavailable")
	if _, e = d.Plan(ctx, m, DefinitionOptions{}); e == nil {
		t.Fatal("plan should fail")
	}
	if e = d.RequireSynchronized(ctx, m); e != nil {
		t.Fatal("plan mutated ready", e)
	}
	if _, e = d.Sync(ctx, m, DefinitionOptions{DryRun: true}); e == nil {
		t.Fatal("dry-run should fail")
	}
	if e = d.RequireSynchronized(ctx, m); e != nil {
		t.Fatal("dry-run mutated ready", e)
	}
	if _, e = d.Sync(ctx, m, DefinitionOptions{}); e == nil {
		t.Fatal("real sync should fail")
	}
	if e = d.RequireSynchronized(ctx, m); e == nil {
		t.Fatal("catalog error kept stale readiness")
	}
}
