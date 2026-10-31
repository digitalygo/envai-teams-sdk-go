package envai

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"
)

type DefinitionState struct {
	Type          string            `json:"type"`
	LocalKey      string            `json:"local_key"`
	RemoteID      string            `json:"remote_id"`
	RemoteName    string            `json:"remote_name"`
	DesiredHash   string            `json:"desired_hash"`
	LastApplied   DefinitionPayload `json:"last_applied"`
	RemoteMarker  string            `json:"remote_marker"`
	Status        string            `json:"status"`
	Intended      DefinitionPayload `json:"intended,omitempty"`
	PendingAction string            `json:"pending_action,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
}
type DefinitionSnapshot struct {
	FormatVersion string                     `json:"format_version"`
	Resources     map[string]DefinitionState `json:"resources"`
	ReadyHash     string                     `json:"ready_hash"`
	ReadyAt       time.Time                  `json:"ready_at"`
}

func emptyDefinitionSnapshot() DefinitionSnapshot {
	return DefinitionSnapshot{FormatVersion: "1", Resources: map[string]DefinitionState{}}
}

// WithLock must serialize the full callback across processes for durable stores.
// Save persists checkpoints immediately; successful remote mutations are not rolled
// back when a later mutation fails. No store implicitly executes schema migrations.
type DefinitionStore interface {
	WithLock(context.Context, string, func(DefinitionSession) error) error
}
type DefinitionSession interface {
	Load() (DefinitionSnapshot, error)
	Save(DefinitionSnapshot) error
}
type DefinitionRemote interface {
	Models(context.Context) (map[string]bool, error)
	List(context.Context, string) ([]DefinitionPayload, error)
	Mutate(context.Context, string, string, string, DefinitionPayload) (DefinitionPayload, error)
}
type DefinitionOptions struct {
	Adopt           bool   `json:"adopt"`
	OverwriteDrift  bool   `json:"overwrite_drift"`
	RecreateMissing bool   `json:"recreate_missing"`
	Prune           bool   `json:"prune"`
	FailOnDrift     bool   `json:"fail_on_drift"`
	DryRun          bool   `json:"dry_run"`
	Scope           string `json:"scope"`
}
type DefinitionChange struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}
type DefinitionOperation struct {
	ID             string             `json:"id"`
	Type           string             `json:"type"`
	LocalKey       string             `json:"local_key"`
	Classification string             `json:"classification"`
	Action         string             `json:"action"`
	Blocking       bool               `json:"blocking"`
	ReasonCode     string             `json:"reason_code,omitempty"`
	RemoteID       string             `json:"remote_id,omitempty"`
	RemoteName     string             `json:"remote_name,omitempty"`
	RemoteMarker   string             `json:"remote_marker,omitempty"`
	Destructive    bool               `json:"destructive"`
	Changes        []DefinitionChange `json:"changes"`
	Desired        DefinitionPayload  `json:"-"`
	Observed       DefinitionPayload  `json:"-"`
}
type DefinitionPlan struct {
	Hash         string                `json:"hash"`
	ManifestHash string                `json:"manifest_hash"`
	Options      DefinitionOptions     `json:"options"`
	Operations   []DefinitionOperation `json:"operations"`
}
type DefinitionOutcome struct {
	Cause      error  `json:"-"` // Original typed SDK error for errors.As; omitted from CLI output.
	ErrorCode  string `json:"error_code,omitempty"`
	HTTPStatus int    `json:"http_status,omitempty"`
	RequestID  string `json:"request_id,omitempty"`

	ID      string `json:"id"`
	Outcome string `json:"outcome"`
	Error   string `json:"error,omitempty"`
}
type DefinitionReport struct {
	FormatVersion string              `json:"format_version"`
	Status        string              `json:"status"`
	ManifestReady bool                `json:"manifest_ready"`
	Plan          DefinitionPlan      `json:"plan"`
	Results       []DefinitionOutcome `json:"results"`
}

func definitionNormalize(kind string, p DefinitionPayload, reverse map[string]string) DefinitionPayload {
	q := definitionClone(p)
	if q == nil {
		return nil
	}
	for _, k := range []string{"id", "user_id", "created_at", "updated_at"} {
		delete(q, k)
	}
	optional := []string{"description"}
	if kind == "agent" {
		optional = []string{"description", "model", "tools_enabled", "external_tools_enabled", "permissions", "config"}
		for _, k := range []string{"tools_enabled", "external_tools_enabled"} {
			if _, ok := q[k]; ok {
				q[k] = definitionSet(q[k])
			}
		}
	}
	for _, k := range optional {
		v := q[k]
		if v == nil {
			delete(q, k)
		} else if kind == "agent" {
			switch x := v.(type) {
			case []any:
				if len(x) == 0 {
					delete(q, k)
				}
			case map[string]any:
				if len(x) == 0 {
					delete(q, k)
				}
			}
		}
	}
	if kind == "team" {
		if local, ok := reverse[definitionString(q, "orchestrator_id")]; ok {
			q["orchestrator_id"] = local
		}
		if members, ok := q["members"].([]any); ok {
			for _, v := range members {
				if m, ok := v.(map[string]any); ok {
					if local, ok := reverse[definitionString(m, "agent_definition_id")]; ok {
						m["agent_definition_id"] = local
					}
					delete(m, "name")
					delete(m, "position")
				}
			}
		}
	}
	return q
}
func definitionDiff(a, b DefinitionPayload, prefix string) []DefinitionChange {
	out := []DefinitionChange{}
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	for _, k := range definitionKeys(keys) {
		av, ae := a[k]
		bv, be := b[k]
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		switch {
		case !ae:
			out = append(out, DefinitionChange{path, "added"})
		case !be:
			out = append(out, DefinitionChange{path, "removed"})
		case !reflect.DeepEqual(av, bv):
			am, aok := av.(map[string]any)
			bm, bok := bv.(map[string]any)
			if aok && bok {
				out = append(out, definitionDiff(am, bm, path)...)
			} else {
				out = append(out, DefinitionChange{path, "changed"})
			}
		}
	}
	return out
}
func planDefinitions(m DefinitionManifest, s DefinitionSnapshot, remote map[string][]DefinitionPayload, o DefinitionOptions) (DefinitionPlan, error) {
	if o.Scope == "" {
		o.Scope = "all"
	}
	if o.Scope != "all" && o.Scope != "tools" && o.Scope != "agents" && o.Scope != "teams" {
		return DefinitionPlan{}, errors.New("envai: invalid definition scope")
	}
	hash, err := m.Hash()
	if err != nil {
		return DefinitionPlan{}, err
	}
	plan := DefinitionPlan{ManifestHash: hash, Options: o, Operations: []DefinitionOperation{}}
	reverse := map[string]string{}
	for _, v := range s.Resources {
		if v.Type == "agent" {
			reverse[v.RemoteID] = v.LocalKey
		}
	}
	for _, kind := range []string{"tool", "agent", "team"} {
		if o.Scope != "all" && o.Scope != kind+"s" {
			continue
		}
		keys := map[string]bool{}
		for k := range m.resources(kind) {
			keys[k] = true
		}
		for _, v := range s.Resources {
			if v.Type == kind {
				keys[v.LocalKey] = true
			}
		}
		for _, key := range definitionKeys(keys) {
			id := kind + ":" + key
			old, managed := s.Resources[id]
			desired := m.resources(kind)[key]
			var observed DefinitionPayload
			for _, r := range remote[kind] {
				match := false
				if managed {
					if kind == "tool" || (old.Status == "ambiguous" && old.RemoteID == "") {
						match = definitionString(r, "name") == old.RemoteName
					} else {
						match = definitionString(r, "id") == old.RemoteID
					}
				} else {
					match = desired != nil && definitionString(r, "name") == definitionString(desired, "name")
				}
				if match {
					if observed != nil {
						return plan, fmt.Errorf("envai: ambiguous remote identity %s", id)
					}
					observed = r
				}
			}
			d := definitionNormalize(kind, desired, reverse)
			a := definitionNormalize(kind, old.LastApplied, reverse)
			r := definitionNormalize(kind, observed, reverse)
			op := DefinitionOperation{ID: id, Type: kind, LocalKey: key, Action: "none", Desired: desired, Observed: observed, Changes: definitionDiff(r, d, "")}
			if managed {
				op.RemoteID = old.RemoteID
				op.RemoteName = old.RemoteName
				op.RemoteMarker = old.RemoteMarker
			}
			if observed != nil {
				op.RemoteID = definitionString(observed, "id")
				op.RemoteName = definitionString(observed, "name")
				op.RemoteMarker = definitionString(observed, "updated_at")
			}
			switch {
			case managed && old.Status == "ambiguous":
				if old.PendingAction == "delete" && observed == nil {
					op.Classification = "orphan_managed"
					op.Action = "forget"
				} else if observed != nil && old.Intended != nil && reflect.DeepEqual(definitionNormalize(kind, old.Intended, reverse), r) {
					op.Classification = "unchanged"
					op.Action = "adopt"
				} else {
					op.Classification = "conflict"
					op.Blocking = true
					op.ReasonCode = "ambiguous_mutation_requires_reconciliation"
				}
			case desired == nil:
				op.Classification = "orphan_managed"
				if o.Prune {
					op.Action = "delete"
					op.Destructive = true
				}
			case !managed && observed == nil:
				op.Classification = "create"
				op.Action = "create"
			case !managed:
				op.Classification = "adoption_required"
				if o.Adopt {
					op.Action = "adopt"
				} else {
					op.Blocking = true
				}
			case observed == nil:
				op.Classification = "missing_remote"
				if o.RecreateMissing {
					op.Action = "recreate"
				} else {
					op.Blocking = true
				}
			case reflect.DeepEqual(d, r):
				op.Classification = "unchanged"
			case reflect.DeepEqual(a, r):
				op.Classification = "update"
				op.Action = "update"
			case reflect.DeepEqual(d, a):
				op.Classification = "drift"
				if o.OverwriteDrift {
					op.Action = "update"
				} else {
					op.Blocking = true
				}
			default:
				op.Classification = "conflict"
				op.Blocking = true
			}
			if op.Blocking {
				if op.ReasonCode == "" {
					op.ReasonCode = op.Classification
				}
			}
			plan.Operations = append(plan.Operations, op)
		}
	}
	// Deletion follows every upsert and reverses dependencies.
	rank := func(op DefinitionOperation) int {
		n := map[string]int{"tool": 0, "agent": 1, "team": 2}[op.Type]
		if op.Action == "delete" {
			return 5 - n
		}
		return n
	}
	sort.SliceStable(plan.Operations, func(i, j int) bool { return rank(plan.Operations[i]) < rank(plan.Operations[j]) })
	plan.Hash, err = definitionHash(struct {
		ManifestHash string
		Options      DefinitionOptions
		Operations   []DefinitionOperation
	}{hash, o, plan.Operations})
	return plan, err
}

// Definitions coordinates explicit deploy operations. Scope is a stable public
// account/deployment identity, never an API key; credential rotation keeps mapping.
type Definitions struct {
	Remote DefinitionRemote
	Store  DefinitionStore
	Scope  string
}

func (d *Definitions) check() error {
	if d.Store == nil || d.Scope == "" {
		return errors.New("envai: definitions require store and stable connection scope")
	}
	return nil
}
func (d *Definitions) prepare(ctx context.Context, m DefinitionManifest, session DefinitionSession, o DefinitionOptions) (DefinitionPlan, DefinitionSnapshot, error) {
	s, err := session.Load()
	if err != nil {
		return DefinitionPlan{}, s, err
	}
	if d.Remote == nil {
		return DefinitionPlan{}, s, errors.New("envai: definitions remote required")
	}
	models, err := d.Remote.Models(ctx)
	if err != nil || models == nil {
		if err == nil {
			err = errors.New("envai: unavailable model catalog")
		}
		return DefinitionPlan{}, s, err
	}
	if err = m.Validate(models); err != nil {
		return DefinitionPlan{}, s, err
	}
	remote := map[string][]DefinitionPayload{}
	for _, kind := range []string{"tool", "agent", "team"} {
		remote[kind], err = d.Remote.List(ctx, kind)
		if err != nil {
			return DefinitionPlan{}, s, err
		}
	}
	plan, err := planDefinitions(m, s, remote, o)
	return plan, s, err
}
func (d *Definitions) Plan(ctx context.Context, m DefinitionManifest, o DefinitionOptions) (DefinitionPlan, error) {
	var p DefinitionPlan
	if e := d.check(); e != nil {
		return p, e
	}
	err := d.Store.WithLock(ctx, d.Scope, func(s DefinitionSession) error { var e error; p, _, e = d.prepare(ctx, m, s, o); return e })
	return p, err
}
func (d *Definitions) State(ctx context.Context) (DefinitionSnapshot, error) {
	var out DefinitionSnapshot
	if e := d.check(); e != nil {
		return out, e
	}
	err := d.Store.WithLock(ctx, d.Scope, func(s DefinitionSession) error { var e error; out, e = s.Load(); return e })
	return out, err
}
func (d *Definitions) RequireSynchronized(ctx context.Context, m DefinitionManifest) error {
	s, err := d.State(ctx)
	if err != nil {
		return err
	}
	h, err := m.Hash()
	if err != nil {
		return err
	}
	if h != s.ReadyHash {
		return errors.New("envai: definitions are not synchronized")
	}
	return nil
}
func (d *Definitions) Lookup(ctx context.Context, m DefinitionManifest, kind, key string) (DefinitionState, error) {
	s, err := d.State(ctx)
	if err != nil {
		return DefinitionState{}, err
	}
	h, err := m.Hash()
	if err != nil {
		return DefinitionState{}, err
	}
	if h != s.ReadyHash {
		return DefinitionState{}, errors.New("envai: definitions are not synchronized")
	}
	v, ok := s.Resources[kind+":"+key]
	if !ok {
		return v, errors.New("envai: unknown managed definition")
	}
	return v, nil
}
func (d *Definitions) Sync(ctx context.Context, m DefinitionManifest, o DefinitionOptions) (DefinitionReport, error) {
	report := DefinitionReport{FormatVersion: "1", Status: "failed", Results: []DefinitionOutcome{}}
	if e := d.check(); e != nil {
		return report, e
	}
	err := d.Store.WithLock(ctx, d.Scope, func(session DefinitionSession) error {
		p, s, err := d.prepare(ctx, m, session, o)
		if err != nil {
			return err
		}
		report.Plan = p
		blocked := false
		for _, op := range p.Operations {
			blocked = blocked || op.Blocking
		}
		if blocked || o.DryRun {
			if blocked && !o.DryRun {
				s.ReadyHash = ""
				s.ReadyAt = time.Time{}
				if err = session.Save(s); err != nil {
					return err
				}
			}
			report.Status = "blocked"
			outcome := "skipped"
			if o.DryRun && !blocked {
				report.Status = "dry_run"
				outcome = "pending"
			}
			for _, op := range p.Operations {
				report.Results = append(report.Results, DefinitionOutcome{ID: op.ID, Outcome: outcome})
			}
			return nil
		}
		report.Status = "succeeded"
		// Invalidate readiness before any write. A failed deploy cannot retain an old
		// ready flag for an unchanged desired hash whose remote has drifted.
		s.ReadyHash = ""
		s.ReadyAt = time.Time{}
		if err = session.Save(s); err != nil {
			return err
		}
		stop := false
		allDesiredApplied := true
		for _, op := range p.Operations {
			out := DefinitionOutcome{ID: op.ID, Outcome: "skipped"}
			if stop || op.Action == "none" {
				if !stop && op.Classification == "unchanged" && op.Observed != nil {
					record := s.Resources[op.ID]
					record.LastApplied = definitionNormalize(op.Type, op.Observed, nil)
					record.DesiredHash = definitionResourceHash(op.Type, op.LocalKey, op.Desired)
					record.Status = "applied"
					record.RemoteMarker = op.RemoteMarker
					record.UpdatedAt = time.Now().UTC()
					s.Resources[op.ID] = record
					if err = session.Save(s); err != nil {
						return err
					}
				}
				report.Results = append(report.Results, out)
				continue
			}
			intended := definitionClone(op.Desired)
			if op.Type == "team" && intended != nil {
				primary := definitionString(intended, "orchestrator_id")
				agent, ok := s.Resources["agent:"+primary]
				if !ok || agent.RemoteID == "" {
					return fmt.Errorf("envai: unresolved applied agent %s", primary)
				}
				intended["orchestrator_id"] = agent.RemoteID
				if members, ok := intended["members"].([]any); ok {
					for _, v := range members {
						member := v.(map[string]any)
						key := definitionString(member, "agent_definition_id")
						agent, ok := s.Resources["agent:"+key]
						if !ok || agent.RemoteID == "" {
							return fmt.Errorf("envai: unresolved applied agent %s", key)
						}
						member["agent_definition_id"] = agent.RemoteID
					}
				}
			}
			action := op.Action
			target := op.RemoteID
			if op.Type == "tool" {
				target = op.RemoteName
			}
			if action == "create" || action == "recreate" {
				action = "create"
				target = definitionString(intended, "name")
			}
			if action == "forget" {
				delete(s.Resources, op.ID)
				if err = session.Save(s); err != nil {
					return err
				}
				out.Outcome = "succeeded"
				report.Results = append(report.Results, out)
				continue
			}
			actual := op.Observed
			if action == "adopt" && (!o.OverwriteDrift || s.Resources[op.ID].Status == "ambiguous") && !reflect.DeepEqual(definitionNormalize(op.Type, actual, nil), definitionNormalize(op.Type, intended, nil)) {
				allDesiredApplied = false
			}
			if action == "adopt" {
				if previousState := s.Resources[op.ID]; previousState.Status != "ambiguous" && o.OverwriteDrift && !reflect.DeepEqual(definitionNormalize(op.Type, actual, nil), definitionNormalize(op.Type, intended, nil)) {
					action = "update"
				}
			} else if action == "recreate" {
				action = "create"
			}
			var mutationErr error
			previous, hadPrevious := s.Resources[op.ID]
			if action != "adopt" {
				pending := previous
				pending.Type = op.Type
				pending.LocalKey = op.LocalKey
				pending.RemoteID = op.RemoteID
				pending.RemoteName = target
				if intended != nil {
					pending.RemoteName = definitionString(intended, "name")
				}
				pending.DesiredHash = definitionResourceHash(op.Type, op.LocalKey, op.Desired)
				pending.Status = "ambiguous"
				pending.Intended = intended
				pending.PendingAction = action
				pending.UpdatedAt = time.Now().UTC()
				s.Resources[op.ID] = pending
				if err = session.Save(s); err != nil {
					return err
				}
				actual, mutationErr = d.Remote.Mutate(ctx, op.Type, action, target, intended)
				var notFound *APIError
				if action == "delete" && errors.As(mutationErr, &notFound) && notFound.Status == 404 {
					mutationErr = nil
				}
			}
			if mutationErr != nil {
				var api *APIError
				if errors.As(mutationErr, &api) && api.Status >= 400 && api.Status < 500 && api.Status != 408 {
					report.Status = "failed"
					out.Outcome = "failed"
				} else {
					// Only authoritative readback resolves an uncertain write; never resend.
					found, readErr := d.Remote.List(ctx, op.Type)
					var recovered DefinitionPayload
					if readErr == nil {
						for _, v := range found {
							match := definitionString(v, "id") == target
							if action == "create" || op.Type == "tool" {
								match = definitionString(v, "name") == target
							}
							if match {
								if recovered != nil {
									readErr = errors.New("envai: duplicate readback identity")
									break
								}
								recovered = v
							}
						}
					}
					if readErr == nil && ((action == "delete" && recovered == nil) || (action != "delete" && recovered != nil && reflect.DeepEqual(definitionNormalize(op.Type, recovered, nil), definitionNormalize(op.Type, intended, nil)))) {
						actual = recovered
						mutationErr = nil
					} else {
						report.Status = "ambiguous"
						out.Outcome = "ambiguous"
					}
				}
				if mutationErr != nil {
					out.Error = "remote mutation failed; inspect server error and reconcile before retry"
					out.Cause = mutationErr
					var apiFailure *APIError
					if errors.As(mutationErr, &apiFailure) {
						out.ErrorCode = apiFailure.Code
						out.HTTPStatus = apiFailure.Status
						out.RequestID = apiFailure.RequestID
					}
					stop = true
					if out.Outcome == "failed" {
						if hadPrevious {
							s.Resources[op.ID] = previous
						} else {
							delete(s.Resources, op.ID)
						}
					}
					if err = session.Save(s); err != nil {
						return err
					}
					report.Results = append(report.Results, out)
					continue
				}
			}
			if action == "delete" {
				delete(s.Resources, op.ID)
			} else {
				if actual == nil {
					return errors.New("envai: mutation returned no authoritative resource")
				}
				if definitionString(actual, "name") == "" {
					return errors.New("envai: mutation returned no resource name")
				}
				if !reflect.DeepEqual(definitionNormalize(op.Type, actual, nil), definitionNormalize(op.Type, intended, nil)) {
					allDesiredApplied = false
				}
				remoteID := definitionString(actual, "id")
				if op.Type != "tool" && remoteID == "" {
					return errors.New("envai: mutation returned no resource id")
				}
				s.Resources[op.ID] = DefinitionState{Type: op.Type, LocalKey: op.LocalKey, RemoteID: remoteID, RemoteName: definitionString(actual, "name"), DesiredHash: definitionResourceHash(op.Type, op.LocalKey, op.Desired), LastApplied: definitionNormalize(op.Type, actual, nil), RemoteMarker: definitionString(actual, "updated_at"), Status: "applied", UpdatedAt: time.Now().UTC()}
			}
			if err = session.Save(s); err != nil {
				return err
			}
			out.Outcome = "succeeded"
			report.Results = append(report.Results, out)
		}
		if report.Status == "succeeded" && p.Options.Scope == "all" && allDesiredApplied {
			s.ReadyHash = p.ManifestHash
			s.ReadyAt = time.Now().UTC()
			if err = session.Save(s); err != nil {
				return err
			}
			report.ManifestReady = true
		}
		return nil
	})
	return report, err
}
