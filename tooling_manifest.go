package envai

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"go.yaml.in/yaml/v3"
)

// DefinitionPayload preserves absent/null and JSON object/array distinctions.
// Handlers belong in Registry, never in a manifest or synchronized state.
type DefinitionPayload = map[string]any
type DefinitionManifest struct {
	FormatVersion string                       `json:"format_version"`
	Tools         map[string]DefinitionPayload `json:"tools"`
	Agents        map[string]DefinitionPayload `json:"agents"`
	Teams         map[string]DefinitionPayload `json:"teams"`
}

func EmptyDefinitionManifest() DefinitionManifest {
	return DefinitionManifest{"1", map[string]DefinitionPayload{}, map[string]DefinitionPayload{}, map[string]DefinitionPayload{}}
}
func (m DefinitionManifest) resources(kind string) map[string]DefinitionPayload {
	switch kind {
	case "tool":
		return m.Tools
	case "agent":
		return m.Agents
	case "team":
		return m.Teams
	}
	return nil
}
func definitionClone(p DefinitionPayload) DefinitionPayload {
	if p == nil {
		return nil
	}
	b, _ := json.Marshal(p)
	var q DefinitionPayload
	_ = json.Unmarshal(b, &q)
	return q
}
func definitionString(p DefinitionPayload, k string) string { s, _ := p[k].(string); return s }
func definitionKeys[T any](m map[string]T) []string {
	r := make([]string, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	sort.Strings(r)
	return r
}
func definitionSet(v any) []any {
	by := map[string]any{}
	if a, ok := v.([]any); ok {
		for _, x := range a {
			if p, ok := x.(map[string]any); ok {
				if name := definitionString(p, "name"); name != "" {
					if _, exists := by[name]; !exists {
						by[name] = p
					}
				}
			}
		}
	}
	r := []any{}
	for _, k := range definitionKeys(by) {
		r = append(r, by[k])
	}
	return r
}
func definitionHash(v any) (string, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	b, e = jsoncanonicalizer.Transform(b)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(b)), nil
}
func (m DefinitionManifest) Hash() (string, error) {
	copy := EmptyDefinitionManifest()
	copy.FormatVersion = m.FormatVersion
	for _, kind := range []string{"tool", "agent", "team"} {
		for key, p := range m.resources(kind) {
			q := definitionClone(p)
			if kind == "agent" {
				for _, field := range []string{"tools_enabled", "external_tools_enabled"} {
					if _, ok := q[field]; ok && q[field] != nil {
						q[field] = definitionSet(q[field])
					}
				}
			}
			copy.resources(kind)[key] = q
		}
	}
	return definitionHash(copy)
}
func definitionResourceHash(kind, key string, p DefinitionPayload) string {
	m := EmptyDefinitionManifest()
	m.resources(kind)[key] = p
	h, _ := m.Hash()
	return h
}

// CompileDefinitions is offline and never takes a store lock or contacts a server.
func CompileDefinitions(data []byte) (DefinitionManifest, error) {
	var m DefinitionManifest
	if _, err := jsoncanonicalizer.Transform(data); err != nil {
		return m, err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return m, err
	}
	if m.FormatVersion != "1" || m.Tools == nil || m.Agents == nil || m.Teams == nil {
		return m, errors.New("envai: definition format 1 requires tools, agents and teams objects")
	}
	if err := m.Validate(nil); err != nil {
		return m, err
	}
	return m, nil
}
func containsMaxTokens(v any) bool {
	switch x := v.(type) {
	case map[string]any:
		for k, v := range x {
			if k == "max_tokens" || containsMaxTokens(v) {
				return true
			}
		}
	case []any:
		for _, v := range x {
			if containsMaxTokens(v) {
				return true
			}
		}
	}
	return false
}

// Validate optionally checks model availability against an authoritative catalog.
// A nil catalog selects local-only validation.
func (m DefinitionManifest) Validate(models map[string]bool) error {
	if m.FormatVersion != "1" || m.Tools == nil || m.Agents == nil || m.Teams == nil {
		return errors.New("envai: invalid definition manifest")
	}
	for _, kind := range []string{"tool", "agent", "team"} {
		names := map[string]bool{}
		for key, p := range m.resources(kind) {
			name := definitionString(p, "name")
			if key == "" || name == "" || names[name] {
				return fmt.Errorf("envai: %s:%s empty or duplicate name", kind, key)
			}
			names[name] = true
			switch kind {
			case "tool":
				if name != key || !toolNamePattern.MatchString(name) || p["kind"] != "client" {
					return fmt.Errorf("envai: invalid client tool %s", key)
				}
				schema, ok := p["parameters"].(map[string]any)
				if !ok || schema["type"] != "object" {
					return fmt.Errorf("envai: %s parameters must be object schema", key)
				}
				compiler := jsonschema.NewCompiler()
				if err := compiler.AddResource("urn:envai:definition", schema); err != nil {
					return err
				}
				if _, err := compiler.Compile("urn:envai:definition"); err != nil {
					return err
				}
				n, ok := p["timeout_seconds"].(float64)
				if !ok {
					if n2, ok2 := p["timeout_seconds"].(int); ok2 {
						n = float64(n2)
						ok = true
					}
				}
				if !ok || n < 1 || n > 600 || n != float64(int(n)) {
					return fmt.Errorf("envai: invalid timeout for %s", key)
				}
			case "agent":
				if strings.TrimSpace(definitionString(p, "system_prompt")) == "" || (p["mode"] != "primary" && p["mode"] != "subagent") {
					return fmt.Errorf("envai: invalid agent %s", key)
				}
				for _, field := range []string{"can_write_knowledge", "can_write_skills"} {
					if _, ok := p[field].(bool); !ok {
						return fmt.Errorf("envai: %s requires boolean %s", key, field)
					}
				}
				if containsMaxTokens(p["config"]) {
					return fmt.Errorf("envai: %s config contains forbidden max_tokens", key)
				}
				for _, field := range []string{"config", "permissions"} {
					if v, exists := p[field]; exists && v != nil {
						if _, ok := v.(map[string]any); !ok {
							return fmt.Errorf("envai: %s must be object", field)
						}
					}
				}
				if model := definitionString(p, "model"); model != "" && models != nil && !models[model] {
					return fmt.Errorf("envai: model unavailable: %s", model)
				}
				for _, field := range []string{"tools_enabled", "external_tools_enabled"} {
					if v, exists := p[field]; exists && v != nil {
						a, ok := v.([]any)
						if !ok {
							return fmt.Errorf("envai: %s must be an array", field)
						}
						for _, x := range a {
							entry, ok := x.(map[string]any)
							if !ok || definitionString(entry, "name") == "" {
								return fmt.Errorf("envai: invalid %s reference", field)
							}
							if field == "external_tools_enabled" {
								if _, ok := m.Tools[definitionString(entry, "name")]; !ok {
									return fmt.Errorf("envai: unresolved client tool %s", definitionString(entry, "name"))
								}
							}
						}
					}
				}
			case "team":
				primary := definitionString(p, "orchestrator_id")
				if m.Agents[primary] == nil || m.Agents[primary]["mode"] != "primary" {
					return fmt.Errorf("envai: team %s needs primary orchestrator", key)
				}
				members, ok := p["members"].([]any)
				if !ok {
					return fmt.Errorf("envai: team %s needs members array", key)
				}
				seen := map[string]bool{primary: true}
				for _, v := range members {
					member, ok := v.(map[string]any)
					if !ok {
						return errors.New("envai: invalid member")
					}
					id := definitionString(member, "agent_definition_id")
					if seen[id] || m.Agents[id] == nil || m.Agents[id]["mode"] != "subagent" {
						return fmt.Errorf("envai: invalid member %s", id)
					}
					seen[id] = true
				}
			}
		}
	}
	return nil
}

// LoadDefinitions reads a portable JSON document, or a directory containing
// manifest.json and agents/*.md with YAML frontmatter. Markdown agents replace
// no existing key: collisions fail closed.
func LoadDefinitions(path string) (DefinitionManifest, error) {
	info, err := os.Stat(path)
	if err != nil {
		return DefinitionManifest{}, err
	}
	if !info.IsDir() {
		b, e := os.ReadFile(path)
		if e != nil {
			return DefinitionManifest{}, e
		}
		return CompileDefinitions(b)
	}
	m := EmptyDefinitionManifest()
	b, err := os.ReadFile(filepath.Join(path, "manifest.json"))
	if err == nil {
		if err = json.Unmarshal(b, &m); err != nil {
			return m, err
		}
	} else if !os.IsNotExist(err) {
		return m, err
	}
	files, err := filepath.Glob(filepath.Join(path, "agents", "*.md"))
	if err != nil {
		return m, err
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return m, err
		}
		text := strings.ReplaceAll(string(b), "\r\n", "\n")
		parts := strings.SplitN(text, "---\n", 3)
		if len(parts) != 3 || parts[0] != "" {
			return m, fmt.Errorf("envai: %s requires YAML frontmatter", file)
		}
		var p DefinitionPayload
		if err = yaml.Unmarshal([]byte(parts[1]), &p); err != nil {
			return m, err
		}
		p["system_prompt"] = strings.TrimSpace(parts[2])
		for _, k := range []string{"can_write_knowledge", "can_write_skills"} {
			if _, ok := p[k]; !ok {
				p[k] = false
			}
		}
		key := strings.TrimSuffix(filepath.Base(file), ".md")
		if _, ok := m.Agents[key]; ok {
			return m, fmt.Errorf("envai: duplicate agent %s", key)
		}
		m.Agents[key] = p
	}
	// JSON roundtrip normalizes YAML integer/container representations.
	b, err = json.Marshal(m)
	if err != nil {
		return m, err
	}
	return CompileDefinitions(b)
}
