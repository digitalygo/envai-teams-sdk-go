package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var projectLogViewSchemas = map[string]string{
	"events":    "ProjectLogCollection",
	"analytics": "ProjectActivityAnalyticsResource",
	"filters":   "ProjectEventTypeFiltersResource",
}

func projectLogParameterSchema(op operation, name string) (string, error) {
	for _, base := range []string{"#/paths/" + pointerToken(op.Path), strictOperationPointer(op)} {
		node, err := canonicalNode(base + "/parameters")
		if err != nil {
			return "", err
		}
		parameters, _ := node.([]any)
		for index := range parameters {
			pointer := fmt.Sprintf("%s/parameters/%d", base, index)
			parameter, err := canonicalObject(pointer)
			if err != nil {
				return "", err
			}
			if ref, ok := parameter["$ref"].(string); ok {
				pointer = ref
				parameter, err = canonicalObject(ref)
				if err != nil {
					return "", err
				}
			}
			if parameter["name"] == name {
				return pointer + "/schema", nil
			}
		}
	}
	return "", strictContractError(op, "parameter "+name)
}

func validateProjectLogQuery(op operation, params map[string]any) error {
	view := "events"
	if raw, exists := params["view"]; exists {
		value, ok := raw.(string)
		if !ok {
			return strictContractError(op, "parameter view")
		}
		pointer, err := projectLogParameterSchema(op, "view")
		if err != nil {
			return err
		}
		if err = validateStrictValue(op, "parameter view", pointer, value); err != nil {
			return err
		}
		view = value
	}
	repeated := map[string]bool{}
	for _, parameter := range op.Parameters {
		if parameter.Repeated {
			repeated[parameter.Name] = true
		}
	}
	for name, value := range params {
		pointer, err := projectLogParameterSchema(op, name)
		if err != nil {
			return err
		}
		if repeated[name] {
			list, ok := value.([]string)
			if !ok {
				return strictContractError(op, "parameter "+name)
			}
			if err = validateStrictValue(op, "parameter "+name, pointer, list); err != nil {
				return err
			}
			continue
		}
		if err = validateStrictValue(op, "parameter "+name, pointer, value); err != nil {
			return err
		}
	}
	has := func(name string) bool {
		_, exists := params[name]
		return exists
	}
	if has("type") && has("types[]") || has("api_key_id") && has("api_key_ids[]") {
		return strictContractError(op, "filter conflict")
	}
	var prohibited []string
	switch view {
	case "analytics":
		prohibited = []string{"type", "types[]", "workspace_ids[]", "api_key_ids[]", "page", "per_page"}
	case "filters":
		prohibited = []string{"type", "api_key_id", "chat_id", "types[]", "workspace_ids[]", "api_key_ids[]", "page", "per_page"}
	}
	for _, name := range prohibited {
		if has(name) {
			return strictContractError(op, "parameter "+name)
		}
	}
	return nil
}

func normalizeProjectLogEntryMetadata(raw []byte) ([]byte, error) {
	var root map[string]json.RawMessage
	if json.Unmarshal(raw, &root) != nil {
		return nil, &ProtocolError{"invalid project log response"}
	}
	data := bytes.TrimSpace(root["data"])
	if len(data) == 0 || data[0] != '[' {
		return raw, nil
	}
	var entries []json.RawMessage
	if json.Unmarshal(data, &entries) != nil {
		return nil, &ProtocolError{"invalid project log entries"}
	}
	changed := false
	for index, entryRaw := range entries {
		var entry map[string]json.RawMessage
		if json.Unmarshal(entryRaw, &entry) != nil {
			return nil, &ProtocolError{"invalid project log entry"}
		}
		entryChanged := false
		for _, key := range []string{"agent", "data"} {
			value := bytes.TrimSpace(entry[key])
			if len(value) == 0 || value[0] != '[' {
				continue
			}
			var items []json.RawMessage
			if json.Unmarshal(value, &items) != nil || len(items) != 0 {
				return nil, &ProtocolError{"invalid project log entry metadata"}
			}
			entry[key] = json.RawMessage("{}")
			entryChanged = true
		}
		if !entryChanged {
			continue
		}
		normalized, err := json.Marshal(entry)
		if err != nil {
			return nil, &ProtocolError{"invalid project log entry"}
		}
		entries[index] = normalized
		changed = true
	}
	if !changed {
		return raw, nil
	}
	normalizedData, err := json.Marshal(entries)
	if err != nil {
		return nil, &ProtocolError{"invalid project log entries"}
	}
	root["data"] = normalizedData
	out, err := json.Marshal(root)
	if err != nil {
		return nil, &ProtocolError{"invalid project log response"}
	}
	return out, nil
}

func validateProjectLogsViewResponse(view string, raw []byte) error {
	name, ok := projectLogViewSchemas[view]
	if !ok {
		return &ProtocolError{"unknown project log view"}
	}
	if err := validateSchema("openapi.json", "#/components/schemas/"+name, raw); err != nil {
		return err
	}
	return nil
}

func projectLogRequestView(request GetProjectLogsParams, view string) (GetProjectLogsParams, error) {
	if request.View.Set && (request.View.Null || request.View.Value != view) {
		return request, errors.New("envai: project log view does not match helper")
	}
	request.View = Present(view)
	return request, nil
}

func (c *Client) GetProjectLogsEvents(ctx context.Context, request GetProjectLogsParams) (ProjectLogCollection, error) {
	request, err := projectLogRequestView(request, "events")
	if err != nil {
		return ProjectLogCollection{}, err
	}
	raw, err := c.GetProjectLogs(ctx, request)
	if err != nil {
		return ProjectLogCollection{}, err
	}
	if err = validateProjectLogsViewResponse("events", raw); err != nil {
		return ProjectLogCollection{}, err
	}
	var result ProjectLogCollection
	if json.Unmarshal(raw, &result) != nil {
		return result, &ProtocolError{"invalid project log events response"}
	}
	return result, nil
}

func decodeProjectLogData[T any](view string, raw []byte) (T, error) {
	var result T
	if err := validateProjectLogsViewResponse(view, raw); err != nil {
		return result, err
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(raw, &envelope) != nil || len(envelope.Data) == 0 {
		return result, &ProtocolError{"invalid project log response"}
	}
	if json.Unmarshal(envelope.Data, &result) != nil {
		return result, &ProtocolError{"invalid project log response"}
	}
	return result, nil
}

func (c *Client) GetProjectLogsAnalytics(ctx context.Context, request GetProjectLogsParams) (ProjectActivityAnalytics, error) {
	request, err := projectLogRequestView(request, "analytics")
	if err != nil {
		return ProjectActivityAnalytics{}, err
	}
	raw, err := c.GetProjectLogs(ctx, request)
	if err != nil {
		return ProjectActivityAnalytics{}, err
	}
	return decodeProjectLogData[ProjectActivityAnalytics]("analytics", raw)
}

func (c *Client) GetProjectLogsFilterCatalog(ctx context.Context, request GetProjectLogsParams) (ProjectEventTypeFilters, error) {
	request, err := projectLogRequestView(request, "filters")
	if err != nil {
		return ProjectEventTypeFilters{}, err
	}
	raw, err := c.GetProjectLogs(ctx, request)
	if err != nil {
		return ProjectEventTypeFilters{}, err
	}
	return decodeProjectLogData[ProjectEventTypeFilters]("filters", raw)
}
