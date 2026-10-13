package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"math"
	"net/http"
	"strings"
	"sync"
)

const ContractVersion = 3
const ContractHeader = "X-Envai-Contract-Version"

// CompatibilityError means the server cannot safely honor the selected contract.
// A mutation is never resubmitted or silently downgraded in response to this error.
type CompatibilityError struct {
	Reason string
	Cause  error
}

func (e *CompatibilityError) Error() string { return "envai: contract compatibility: " + e.Reason }
func (e *CompatibilityError) Unwrap() error { return e.Cause }

func (c *Client) selectedVersion() string {
	if c.config.LegacyReadOnly {
		return "2"
	}
	return "3"
}
func negotiated(op operation) bool {
	return op.ID == "getCapabilities" || op.ID == "getCapabilitiesLegacy" || strings.HasPrefix(op.Path, "/chats/") || strings.HasPrefix(op.Path, "/atp/v3/") || strings.Contains(op.Path, "/completions") || (op.Authenticated && (strings.Contains(op.Path, "/wiki/") || strings.Contains(op.Path, "/files/sources")))
}
func atpOperation(op operation) bool {
	return op.ID == "getCapabilities" || op.ID == "getCapabilitiesLegacy" || strings.HasPrefix(op.Path, "/chats/") || strings.HasPrefix(op.Path, "/atp/v3/")
}
func (c *Client) checkMutation(ctx context.Context, op operation) error {
	if c.config.LegacyReadOnly && op.Method != "GET" {
		return &CompatibilityError{Reason: "legacy mode is read-only"}
	}
	if !negotiated(op) || op.Method == "GET" {
		return nil
	}
	_, err := c.GetCapabilities(ctx, GetCapabilitiesParams{})
	if err != nil {
		return &CompatibilityError{Reason: "contract 3 capabilities preflight failed before mutation", Cause: err}
	}
	return nil
}
func (c *Client) checkAcknowledgment(op operation, resp *http.Response) error {
	if !negotiated(op) {
		return nil
	}
	values := resp.Header.Values(ContractHeader)
	// Explicit legacy reads remain usable against old servers that predate negotiation.
	if (c.config.LegacyReadOnly || strings.HasSuffix(op.ID, "Legacy")) && len(values) == 0 {
		return nil
	}
	if len(values) != 1 || values[0] != c.versionFor(op) {
		return &CompatibilityError{Reason: "missing or mismatched response acknowledgment"}
	}
	return nil
}
func (c *Client) validateNegotiatedJSON(id string, b []byte) error {
	if id == "getCapabilitiesLegacy" || (id == "getCapabilities" && c.config.LegacyReadOnly) {
		var capability struct {
			Version int `json:"protocol_version"`
		}
		if json.Unmarshal(b, &capability) != nil || capability.Version != 2 {
			return &CompatibilityError{Reason: "legacy capabilities must advertise protocol 2"}
		}
	}
	if id == "getCapabilities" && !c.config.LegacyReadOnly {
		var v struct {
			ProtocolVersion int    `json:"protocol_version"`
			ContractVersion int    `json:"contract_version"`
			Supported       []int  `json:"supported_contract_versions"`
			LegacyMode      string `json:"legacy_v2_mode"`
		}
		if json.Unmarshal(b, &v) != nil || v.ProtocolVersion != 3 || v.ContractVersion != 3 || v.LegacyMode != "historical_reads_replay_only" {
			return &CompatibilityError{Reason: "server does not advertise contract 3"}
		}
		supported := false
		for _, n := range v.Supported {
			if n == 3 {
				supported = true
			}
		}
		if !supported {
			return &CompatibilityError{Reason: "server does not support contract 3"}
		}
	}
	if id == "getCompletion" || id == "listCompletions" {
		var root map[string]json.RawMessage
		if json.Unmarshal(b, &root) != nil {
			return &ProtocolError{"invalid completion response"}
		}
		if id == "getCompletion" {
			return validateCompletionJSON(b, c.config.LegacyReadOnly)
		}
		if data := bytes.TrimSpace(root["data"]); len(data) == 0 || data[0] != '[' {
			return &ProtocolError{"completion page data must be an array"}
		}
		var rows []json.RawMessage
		if err := json.Unmarshal(root["data"], &rows); err != nil {
			return &ProtocolError{"invalid completion page"}
		}
		for _, row := range rows {
			if err := validateCompletionJSON(row, c.config.LegacyReadOnly); err != nil {
				return err
			}
		}
	}
	return nil
}
func validateCompletionJSON(b []byte, legacy bool) error {
	var row map[string]json.RawMessage
	if json.Unmarshal(b, &row) != nil {
		return &ProtocolError{"invalid completion"}
	}
	cost, ok := row["cost_usd"]
	if !ok || (legacy && bytes.Equal(bytes.TrimSpace(cost), []byte("null"))) {
		return &ProtocolError{"completion cost is missing or not representable in legacy contract"}
	}
	if err := validateCostFields(b, !legacy); err != nil {
		return err
	}

	if snapshot := row["terminal_snapshot"]; len(snapshot) > 0 && !bytes.Equal(bytes.TrimSpace(snapshot), []byte("null")) {
		var snapshotFields map[string]json.RawMessage
		if json.Unmarshal(snapshot, &snapshotFields) != nil {
			return &ProtocolError{"invalid terminal snapshot"}
		}
		if err := validateCostFields(snapshotFields["data"], !legacy); err != nil {
			return err
		}
		if err := validateDefinition("terminal-snapshot.json", "CompletionTerminalSnapshot", snapshot); err != nil {
			return err
		}
		var identity struct {
			Data struct {
				TurnID string `json:"turn_id"`
			} `json:"data"`
		}
		var id string
		_ = json.Unmarshal(snapshot, &identity)
		_ = json.Unmarshal(row["id"], &id)
		if identity.Data.TurnID != id {
			return &ProtocolError{"terminal snapshot turn does not match completion"}
		}
	}
	return nil
}

// Validate monetary numeric fields before any float64 hydration. Only contract
// containers are traversed; unrelated future extension values remain opaque.
func validateCostFields(b []byte, allowNull bool) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil {
		return &ProtocolError{"invalid cost-bearing object"}
	}
	if _, ok := value.(map[string]any); !ok {
		return &ProtocolError{"invalid cost-bearing object"}
	}
	return walkCostFields(value, allowNull, false)
}
func walkCostFields(value any, allowNull, descend bool) error {
	switch object := value.(type) {
	case map[string]any:
		for name, child := range object {
			switch name {
			case "cost_usd", "cost", "chat_cost":
				if child == nil {
					if allowNull {
						continue
					}
					return &ProtocolError{"legacy cost cannot be null"}
				}
				number, ok := child.(json.Number)
				if !ok {
					return &ProtocolError{"cost must be numeric or null"}
				}
				amount, err := number.Float64()
				if err != nil || math.IsInf(amount, 0) || math.IsNaN(amount) || amount < 0 {
					return &ProtocolError{"cost must be finite and nonnegative"}
				}
			default:
				if descend || name == "usage" || name == "chat_usage" {
					if err := walkCostFields(child, allowNull, true); err != nil {
						return err
					}
				}
			}
		}
	case []any:
		if descend {
			for _, child := range object {
				if err := walkCostFields(child, allowNull, true); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

var schemaCache sync.Map

func validateDefinition(file, definition string, b []byte) error {
	key := file + "#/$defs/" + definition
	cached, ok := schemaCache.Load(key)
	if !ok {
		compiler := jsonschema.NewCompiler()
		compiler.AssertFormat()
		compiler.UseLoader(denySchemaLoader{})
		for _, f := range []string{"events.json", "events-v2.json", "wiki.json", "terminal-snapshot.json"} {
			raw, err := contracts.ReadFile("contract/" + f)
			if err != nil {
				return err
			}
			doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
			if err != nil {
				return err
			}
			if err = compiler.AddResource("https://sdk.invalid/"+f, doc); err != nil {
				return err
			}
		}
		schema, err := compiler.Compile("https://sdk.invalid/" + key)
		if err != nil {
			return err
		}
		cached, _ = schemaCache.LoadOrStore(key, schema)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return &ProtocolError{"invalid JSON"}
	}
	if err = cached.(*jsonschema.Schema).Validate(doc); err != nil {
		return &ProtocolError{fmt.Sprintf("invalid %s representation", definition)}
	}
	return nil
}
func unknownFields(b []byte, keys []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return nil, err
	}
	if fields == nil {
		return nil, errors.New("envai: object cannot be null")
	}
	for _, key := range keys {
		delete(fields, key)
	}
	if len(fields) == 0 {
		return nil, nil
	}
	return fields, nil
}
func marshalUnknown(value any, extra map[string]json.RawMessage) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(extra) == 0 {
		return b, nil
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(b, &fields); err != nil {
		return nil, err
	}
	for k, v := range extra {
		if _, exists := fields[k]; exists {
			return nil, errors.New("envai: extra field conflicts with typed field")
		}
		fields[k] = v
	}
	return json.Marshal(fields)
}

func (c *Client) versionFor(op operation) string {
	if strings.HasSuffix(op.ID, "Legacy") {
		return "2"
	}
	return c.selectedVersion()
}
