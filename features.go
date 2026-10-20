package envai

import (
	"bytes"
	"encoding/json"
	"errors"
)

func normalizeFeatureNames(raw json.RawMessage) ([]string, error) {
	value := bytes.TrimSpace(raw)
	if len(value) == 0 || bytes.Equal(value, []byte("null")) {
		return nil, errors.New("envai: capabilities features must be a string list or a strict boolean map")
	}
	switch value[0] {
	case '[':
		var entries []json.RawMessage
		if err := json.Unmarshal(value, &entries); err != nil {
			return nil, errors.New("envai: capabilities features must be a string list or a strict boolean map")
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			entry = bytes.TrimSpace(entry)
			var name string
			if len(entry) == 0 || entry[0] != '"' || json.Unmarshal(entry, &name) != nil {
				return nil, errors.New("envai: capabilities features list must contain only strings")
			}
			names = append(names, name)
		}
		return names, nil
	case '{':
		decoder := json.NewDecoder(bytes.NewReader(value))
		if _, err := decoder.Token(); err != nil {
			return nil, errors.New("envai: capabilities features must be a string list or a strict boolean map")
		}
		names := []string{}
		seen := map[string]struct{}{}
		for decoder.More() {
			token, err := decoder.Token()
			name, ok := token.(string)
			if err != nil || !ok {
				return nil, errors.New("envai: capabilities feature map must contain boolean values")
			}
			if _, duplicate := seen[name]; duplicate {
				return nil, errors.New("envai: capabilities feature map must not repeat a feature name")
			}
			seen[name] = struct{}{}
			var enabled json.RawMessage
			if err = decoder.Decode(&enabled); err != nil {
				return nil, errors.New("envai: capabilities feature map must contain boolean values")
			}
			switch string(bytes.TrimSpace(enabled)) {
			case "true":
				names = append(names, name)
			case "false":
			default:
				return nil, errors.New("envai: capabilities feature map must contain boolean values")
			}
		}
		if _, err := decoder.Token(); err != nil {
			return nil, errors.New("envai: capabilities features must be a string list or a strict boolean map")
		}
		return names, nil
	default:
		return nil, errors.New("envai: capabilities features must be a string list or a strict boolean map")
	}
}
