package envai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ClientDefinitionRemote reuses the SDK operation catalog, authentication,
// timeouts and error envelopes. Mutations disable retries so uncertain effects
// are reconciled by authoritative readback rather than silently repeated.
type ClientDefinitionRemote struct{ Client *Client }

func (r ClientDefinitionRemote) Models(ctx context.Context) (map[string]bool, error) {
	if r.Client == nil {
		return nil, errors.New("envai: nil definitions client")
	}
	models, e := r.Client.ListModels(ctx, ListModelsParams{})
	if e != nil {
		return nil, e
	}
	out := map[string]bool{}
	for _, v := range models.Data {
		var p DefinitionPayload
		if e = definitionConvert(v, &p); e != nil {
			return nil, e
		}
		if id := definitionString(p, "id"); id != "" {
			out[id] = true
		}
	}
	return out, nil
}
func (r ClientDefinitionRemote) List(ctx context.Context, kind string) ([]DefinitionPayload, error) {
	if r.Client == nil {
		return nil, errors.New("envai: nil definitions client")
	}
	operation := map[string]string{"tool": "listUserTools", "agent": "listAgents", "team": "listTeams"}[kind]
	if operation == "" {
		return nil, fmt.Errorf("envai: invalid definition type %s", kind)
	}
	var out []DefinitionPayload
	e := r.Client.call(ctx, operation, map[string]any{}, nil, &out)
	return out, e
}
func (r ClientDefinitionRemote) Mutate(ctx context.Context, kind, action, target string, p DefinitionPayload) (DefinitionPayload, error) {
	if r.Client == nil {
		return nil, errors.New("envai: nil definitions client")
	}
	resource := map[string]string{"tool": "UserTool", "agent": "Agent", "team": "Team"}[kind]
	if resource == "" || (action != "create" && action != "update" && action != "delete") {
		return nil, errors.New("envai: invalid definitions mutation")
	}
	params := map[string]any{}
	if action != "create" {
		param := kind + "_id"
		if kind == "tool" {
			param = "name"
		}
		params[param] = target
	}
	clone := *r.Client
	clone.config.RetryAttempts = 0
	var out DefinitionPayload
	var e error
	if action == "delete" {
		e = clone.call(ctx, action+resource, params, nil, nil)
	} else {
		e = clone.call(ctx, action+resource, params, p, &out)
	}
	return out, e
}

func definitionConvert(v any, out any) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, out)
}
