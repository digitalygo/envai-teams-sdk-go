package envai

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// RunDefinitionsCLI executes explicit deployment tooling. getenv is injectable
// for offline automation tests; credential values are never included in output.
// Exit 2 means a blocked/drifted plan; exit 1 means configuration/runtime failure.
func RunDefinitionsCLI(ctx context.Context, args []string, out, errOut io.Writer, getenv func(string) string) int {
	if getenv == nil {
		getenv = os.Getenv
	}
	if len(args) == 0 {
		fmt.Fprintln(errOut, "usage: envai definitions:{validate,plan,sync,status,state} [flags]")
		return 1
	}
	command := strings.TrimPrefix(args[0], "definitions:")
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(errOut)
	envDefault := func(k, fallback string) string {
		if v := getenv(k); v != "" {
			return v
		}
		return fallback
	}
	strictDefault := false
	if value := getenv("ENVAI_STRICT_VALIDATION"); value != "" {
		var err error
		strictDefault, err = strconv.ParseBool(value)
		if err != nil {
			fmt.Fprintln(errOut, "ENVAI_STRICT_VALIDATION must be a boolean")
			return 1
		}
	}
	strictValidation := flags.Bool("strict-validation", strictDefault, "validate canonical HTTP request and response schemas")
	path := flags.String("path", envDefault("ENVAI_DEFINITIONS_PATH", "definitions"), "JSON manifest or definition directory")
	scope := flags.String("connection", getenv("ENVAI_CONNECTION"), "stable public account/deployment scope")
	driver := flags.String("store", envDefault("ENVAI_STATE_DRIVER", "file"), "file, postgres, sqlite or mysql (SQL drivers must be linked)")
	statePath := flags.String("state-path", envDefault("ENVAI_STATE_PATH", ".envai-state"), "private file state directory")
	asJSON := flags.Bool("json", false, "machine-readable output")
	o := DefinitionOptions{}
	flags.BoolVar(&o.DryRun, "dry-run", false, "plan without writes")
	flags.BoolVar(&o.Adopt, "adopt", false, "adopt existing named resources")
	flags.BoolVar(&o.OverwriteDrift, "overwrite-drift", false, "overwrite remote-only drift")
	flags.BoolVar(&o.RecreateMissing, "recreate-missing", false, "recreate missing managed resources")
	flags.BoolVar(&o.Prune, "prune", false, "delete orphan managed resources")
	flags.BoolVar(&o.FailOnDrift, "fail-on-drift", false, "exit 2 on drift")
	flags.StringVar(&o.Scope, "only", "all", "all/tools/agents/teams")
	if err := flags.Parse(args[1:]); err != nil {
		return 1
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(errOut, "unexpected positional arguments")
		return 1
	}
	emit := func(v any) {
		if *asJSON {
			_ = json.NewEncoder(out).Encode(v)
		} else {
			b, _ := json.MarshalIndent(v, "", "  ")
			fmt.Fprintln(out, string(b))
		}
	}
	fail := func(err error) int { // Do not echo transport error bodies, URLs or credentials.
		_ = err
		fmt.Fprintln(errOut, "envai: definitions command failed; check configuration, input validation and server availability")
		return 1
	}
	switch command {
	case "validate", "plan", "sync", "status", "state":
	default:
		fmt.Fprintln(errOut, "unknown definitions command")
		return 1
	}
	var m DefinitionManifest
	if command != "state" {
		var err error
		m, err = LoadDefinitions(*path)
		if err != nil {
			return fail(err)
		}
	}
	if command == "validate" {
		hash, err := m.Hash()
		if err != nil {
			return fail(err)
		}
		emit(map[string]any{"format_version": "1", "valid": true, "manifest_hash": hash})
		return 0
	}
	if *scope == "" {
		return fail(fmt.Errorf("missing connection scope"))
	}
	var store DefinitionStore
	switch *driver {
	case "file":
		store = &FileDefinitionStore{Directory: *statePath}
	case "postgres":
		db, err := sql.Open("postgres", getenv("ENVAI_STATE_DSN"))
		if err != nil {
			return fail(err)
		}
		defer db.Close()
		store = &PostgresDefinitionStore{DB: db}
	case "sqlite", "mysql":
		db, err := sql.Open(*driver, getenv("ENVAI_STATE_DSN"))
		if err != nil {
			return fail(err)
		}
		defer db.Close()
		store = &SQLDefinitionStore{DB: db, Dialect: *driver}
	default:
		return fail(fmt.Errorf("unsupported store"))
	}
	d := Definitions{Store: store, Scope: *scope}
	if command == "state" {
		s, err := d.State(ctx)
		if err != nil {
			return fail(err)
		}
		emit(s)
		return 0
	}
	if command == "status" {
		s, err := d.State(ctx)
		if err != nil {
			return fail(err)
		}
		h, err := m.Hash()
		if err != nil {
			return fail(err)
		}
		ready := h == s.ReadyHash
		emit(map[string]any{"format_version": "1", "manifest_hash": h, "manifest_ready": ready})
		if !ready {
			return 2
		}
		return 0
	}
	client, err := New(Config{StrictValidation: *strictValidation, BaseURL: getenv("ENVAI_BASE_URL"), APIKey: getenv("ENVAI_API_KEY"), AllowHTTP: getenv("ENVAI_ALLOW_HTTP") == "true"})
	if err != nil {
		return fail(err)
	}
	d.Remote = ClientDefinitionRemote{Client: client}
	if command == "plan" {
		p, err := d.Plan(ctx, m, o)
		if err != nil {
			return fail(err)
		}
		emit(p)
		for _, op := range p.Operations {
			if op.Blocking || (o.FailOnDrift && (op.Classification == "drift" || op.Classification == "conflict")) {
				return 2
			}
		}
		return 0
	}
	report, err := d.Sync(ctx, m, o)
	if err != nil {
		return fail(err)
	}
	emit(report)
	if report.Status == "blocked" {
		return 2
	}
	if report.Status != "succeeded" && report.Status != "dry_run" {
		return 1
	}
	if o.FailOnDrift {
		for _, op := range report.Plan.Operations {
			if op.Classification == "drift" || op.Classification == "conflict" {
				return 2
			}
		}
	}
	return 0
}

// Ensure lazily synchronizes only in explicit "lazy" mode. deploy/manual/off
// never make network calls. A second check occurs under the same exclusive lock
// as the subsequent plan/apply, preventing duplicate deployments.
func (d *Definitions) Ensure(ctx context.Context, m DefinitionManifest, mode string, ttl time.Duration) error {
	switch mode {
	case "deploy", "manual", "off":
		return nil
	case "lazy":
	default:
		return fmt.Errorf("envai: unknown definitions gate mode %s", mode)
	}
	s, e := d.State(ctx)
	if e != nil {
		return e
	}
	h, e := m.Hash()
	if e != nil {
		return e
	}
	if s.ReadyHash == h && time.Since(s.ReadyAt) < ttl {
		return nil
	}
	// Sync serializes and replans under its lock, so a competing completed apply
	// becomes an unchanged plan and cannot duplicate creates.
	report, e := d.Sync(ctx, m, DefinitionOptions{})
	if e != nil {
		return e
	}
	if !report.ManifestReady {
		return fmt.Errorf("envai: definitions synchronization %s", report.Status)
	}
	return nil
}
