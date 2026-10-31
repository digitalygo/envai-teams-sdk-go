package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestStrictValidationSharedAgentResponses(t *testing.T) {
	fixtures := fixture[struct {
		AgentID string                     `json:"agent_id"`
		Cases   map[string]json.RawMessage `json:"response_cases"`
	}](t, "parity/schema-fixtures.json")
	for name, body := range fixtures.Cases {
		t.Run(name, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.Write(body) })
			c.config.StrictValidation = true
			agent, err := c.GetAgent(context.Background(), GetAgentParams{AgentID: fixtures.AgentID})
			if name == "valid" {
				if err != nil || agent.ID == "" {
					t.Fatal(agent, err)
				}
			} else {
				var protocol *ProtocolError
				if !errors.As(err, &protocol) {
					t.Fatal("invalid canonical response accepted", name, err)
				}
			}
		})
	}
	// Preserve the existing additive default for callers who have not opted in.
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"data":{}}`) })
	if _, err := c.GetAgent(context.Background(), GetAgentParams{AgentID: fixtures.AgentID}); err != nil {
		t.Fatal("default decoding changed", err)
	}
}
func TestStrictValidationRejectsMutationBeforeAnyHTTP(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		t.Error("invalid request reached HTTP", r.URL.Path)
	})
	c.config.StrictValidation = true
	c.config.NativeTasks = true
	if _, err := c.CreateAgent(context.Background(), CreateAgentParams{Body: AgentCreateRequest{}}); err == nil {
		t.Fatal("missing body fields accepted")
	}
	if _, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{}}); err == nil {
		t.Fatal("invalid mutation reached capability preflight")
	}
	if _, err := c.ListCompletions(context.Background(), ListCompletionsParams{WorkspaceID: "workspace", ChatID: "chat", Page: Present(int64(-1))}); err == nil {
		t.Fatal("negative page accepted")
	}
	if calls.Load() != 0 {
		t.Fatal(calls.Load())
	}
}
func TestStrictValidationCanonicalOperationCoverage(t *testing.T) {
	c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("schema lookup made network call") })
	if len(c.operations) != 76 {
		t.Fatal(len(c.operations))
	}
	for id, op := range c.operations {
		t.Run(id, func(t *testing.T) {
			pointer := strictOperationPointer(op)
			node, err := canonicalObject(pointer)
			if err != nil || node == nil || node["operationId"] != id {
				t.Fatal(id, node, err)
			}
			validate := func(pointer string) {
				t.Helper()
				err := validateSchema("openapi.json", pointer, []byte("null"))
				if _, compiled := schemaCache.Load("openapi.json" + pointer); !compiled {
					t.Fatal("unresolvable canonical schema", pointer, err)
				}
			}
			for _, media := range []string{"application/json", "multipart/form-data"} {
				p := pointer + "/requestBody/content/" + pointerToken(media) + "/schema"
				schema, err := canonicalNode(p)
				if err != nil {
					t.Fatal(err)
				}
				if schema != nil {
					validate(p)
				}
			}
			responses, _ := node["responses"].(map[string]any)
			for status := range responses {
				if !strings.HasPrefix(status, "2") {
					continue
				}
				p := pointer + "/responses/" + status + "/content/application~1json/schema"
				schema, err := canonicalNode(p)
				if err != nil {
					t.Fatal(err)
				}
				if schema != nil {
					validate(p)
				}
			}
		})
	}
}

func TestStrictValidationValidMutationAndUnexpectedStatus(t *testing.T) {
	fixtures := fixture[struct {
		Cases map[string]json.RawMessage `json:"response_cases"`
	}](t, "parity/schema-fixtures.json")
	calls := 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost {
			t.Fatal(r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(fixtures.Cases["valid"])
	})
	c.config.StrictValidation = true
	if _, err := c.CreateAgent(context.Background(), CreateAgentParams{Body: AgentCreateRequest{Name: "agent", SystemPrompt: "Help", Mode: "primary"}}); err != nil || calls != 1 {
		t.Fatal(calls, err)
	}
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write(fixtures.Cases["valid"])
	})
	c.config.StrictValidation = true
	var protocol *ProtocolError
	if _, err := c.GetAgent(context.Background(), GetAgentParams{AgentID: "agent"}); !errors.As(err, &protocol) {
		t.Fatal("unexpected success status accepted", err)
	}
}

func TestStrictValidationCLI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "definitions.json")
	raw, err := json.Marshal(toolingManifest())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		env, flag string
		valid     bool
	}{{"true", "", true}, {"false", "", true}, {"bad", "", false}, {"", "true", true}, {"", "bad", false}} {
		args := []string{"definitions:validate", "--path", path}
		if tc.flag != "" {
			args = append(args, "--strict-validation="+tc.flag)
		}
		var out bytes.Buffer
		code := RunDefinitionsCLI(context.Background(), args, &out, &out, func(k string) string {
			if k == "ENVAI_STRICT_VALIDATION" {
				return tc.env
			}
			return ""
		})
		if (code == 0) != tc.valid {
			t.Fatal(tc, code, out.String())
		}
	}
}

func TestStrictValidationCapabilitiesRemainAdditive(t *testing.T) {
	for _, features := range []string{`["future_feature"]`, `{"future_feature":true,"disabled":false}`} {
		raw := strings.Replace(testCapabilities, `"families":[]`, `"families":["future_family"]`, 1)
		raw = strings.Replace(raw, `"features":[]`, `"features":`+features, 1)
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, raw) })
		c.config.StrictValidation = true
		if err := c.checkMutation(context.Background(), c.operations["sendChatMessage"], map[string]any{"chatId": "chat"}, ChatMessageRequest{}); err != nil {
			t.Fatal("capability preflight rejected additive discovery", err)
		}
		result, err := c.GetCapabilities(context.Background(), GetCapabilitiesParams{})
		if err != nil || !featureAdvertised(result.Features, "future_feature") || !featureAdvertised(result.Families, "future_family") {
			t.Fatal(result, err)
		}
	}
}

func TestStrictValidationLegacyCapabilities(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"protocol_version":2,"families":["future_family"],"features":{"future_feature":true},"event_types":[],"tool_kinds":[]}`)
	})
	c.config.StrictValidation = true
	c.config.LegacyReadOnly = true
	result, err := c.GetCapabilities(context.Background(), GetCapabilitiesParams{})
	if err != nil || result.ProtocolVersion != 2 {
		t.Fatal(result, err)
	}
}
