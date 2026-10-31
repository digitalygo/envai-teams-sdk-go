package envai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var canonicalHTTP = sync.OnceValues(func() (map[string]any, error) {
	raw, err := contracts.ReadFile("contract/openapi.json")
	if err != nil {
		return nil, err
	}
	var root map[string]any
	err = json.Unmarshal(raw, &root)
	return root, err
})

func pointerToken(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "~", "~0"), "/", "~1")
}
func canonicalNode(pointer string) (any, error) {
	root, err := canonicalHTTP()
	if err != nil {
		return nil, err
	}
	var node any = root
	for _, token := range strings.Split(strings.TrimPrefix(pointer, "#/"), "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch current := node.(type) {
		case map[string]any:
			node = current[token]
		case []any:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(current) {
				return nil, nil
			}
			node = current[index]
		default:
			return nil, nil
		}
	}
	return node, nil
}
func canonicalObject(pointer string) (map[string]any, error) {
	node, err := canonicalNode(pointer)
	if err != nil {
		return nil, err
	}
	object, _ := node.(map[string]any)
	return object, nil
}
func strictOperationPointer(op operation) string {
	return "#/paths/" + pointerToken(op.Path) + "/" + strings.ToLower(op.Method)
}
func strictContractError(op operation, part string) error {
	return &ProtocolError{Reason: part + " for " + op.ID + " violates the canonical HTTP contract"}
}
func validateStrictValue(op operation, part, pointer string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return strictContractError(op, part)
	}
	if err = validateSchema("openapi.json", pointer, raw); err != nil {
		return strictContractError(op, part)
	}
	return nil
}
func (c *Client) validateStrictRequest(op operation, params map[string]any, body any) error {
	pointer := strictOperationPointer(op)
	operationSpec, err := canonicalObject(pointer)
	if err != nil {
		return err
	}
	if operationSpec == nil {
		return strictContractError(op, "operation")
	}
	for _, base := range []string{"#/paths/" + pointerToken(op.Path), pointer} {
		node, err := canonicalNode(base + "/parameters")
		if err != nil {
			return err
		}
		parameters, _ := node.([]any)
		for index := range parameters {
			parameterPointer := fmt.Sprintf("%s/parameters/%d", base, index)
			parameter, err := canonicalObject(parameterPointer)
			if err != nil {
				return err
			}
			if ref, ok := parameter["$ref"].(string); ok {
				parameterPointer = ref
				parameter, err = canonicalObject(ref)
				if err != nil {
					return err
				}
			}
			name, _ := parameter["name"].(string)
			// These headers are owned and validated by the negotiation layer.
			if name == ContractHeader || name == "X-Envai-Extensions" {
				continue
			}
			value, exists := params[name]
			if !exists {
				if required, _ := parameter["required"].(bool); required {
					return strictContractError(op, "required parameter")
				}
				continue
			}
			if err := validateStrictValue(op, "parameter "+name, parameterPointer+"/schema", value); err != nil {
				return err
			}
		}
	}
	requestBody, err := canonicalObject(pointer + "/requestBody")
	if err != nil {
		return err
	}
	if requestBody == nil {
		return nil
	}
	if body == nil {
		if required, _ := requestBody["required"].(bool); required {
			return strictContractError(op, "required body")
		}
		return nil
	}
	media := "application/json"
	value := body
	if op.Body == "multipart" {
		media = "multipart/form-data"
		parts, ok := body.([]MultipartPart)
		if !ok {
			return strictContractError(op, "multipart body")
		}
		fields := map[string]any{}
		for _, part := range parts {
			field := strings.TrimSuffix(part.Name, "[]")
			data := string(part.Data)
			if part.Filename != "" {
				data = part.Filename
			}
			if field != part.Name {
				values, _ := fields[field].([]string)
				fields[field] = append(values, data)
			} else {
				if _, duplicate := fields[field]; duplicate {
					return strictContractError(op, "duplicate multipart field")
				}
				fields[field] = data
			}
		}
		value = fields
	}
	schemaPointer := pointer + "/requestBody/content/" + pointerToken(media) + "/schema"
	if schema, err := canonicalNode(schemaPointer); err != nil {
		return err
	} else if schema == nil {
		return strictContractError(op, "body media type")
	}
	return validateStrictValue(op, "request body", schemaPointer, value)
}
func (c *Client) validateStrictResponse(op operation, status int, raw []byte) error {
	if c.config.LegacyReadOnly && op.ID == "getCapabilities" {
		op = c.operations["getCapabilitiesLegacy"]
	}
	pointer := strictOperationPointer(op) + "/responses/" + strconv.Itoa(status)
	response, err := canonicalObject(pointer)
	if err != nil {
		return err
	}
	if response == nil {
		pointer = strictOperationPointer(op) + "/responses/default"
		response, err = canonicalObject(pointer)
		if err != nil {
			return err
		}
	}
	if response == nil {
		return strictContractError(op, "success status")
	}
	schemaPointer := pointer + "/content/application~1json/schema"
	schema, err := canonicalNode(schemaPointer)
	if err != nil {
		return err
	}
	if schema == nil {
		return strictContractError(op, "JSON success body")
	}
	// Discovery lists are additive and features also accept the existing strict
	// boolean-map wire representation. Validate their types, then project only
	// those fields for canonical validation without changing decoded discovery.
	if op.ID == "getCapabilities" || op.ID == "getCapabilitiesLegacy" {
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			return strictContractError(op, "capabilities")
		}
		families := bytes.TrimSpace(fields["families"])
		if len(families) == 0 || families[0] != '[' {
			return strictContractError(op, "families")
		}
		if _, err := normalizeFeatureNames(families); err != nil {
			return strictContractError(op, "families")
		}
		if _, err := normalizeFeatureNames(fields["features"]); err != nil {
			return strictContractError(op, "features")
		}
		fields["families"] = json.RawMessage(`[]`)
		fields["features"] = json.RawMessage(`[]`)
		raw, err = json.Marshal(fields)
		if err != nil {
			return strictContractError(op, "capabilities")
		}
	}
	if err = validateSchema("openapi.json", schemaPointer, raw); err != nil {
		return strictContractError(op, "JSON success body")
	}
	return nil
}

func validateStrictStatus(op operation, status int) error {
	response, err := canonicalObject(strictOperationPointer(op) + "/responses/" + strconv.Itoa(status))
	if err != nil {
		return err
	}
	if response != nil {
		return nil
	}
	response, err = canonicalObject(strictOperationPointer(op) + "/responses/default")
	if err != nil {
		return err
	}
	if response == nil {
		return strictContractError(op, "success status")
	}
	return nil
}
