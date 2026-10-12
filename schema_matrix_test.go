package envai

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

// These paired cases were independently checked against their declared drafts.
// Run each through real Ed25519 verification, registry validation and delivery.
func TestSignedToolSchemaMatrix(t *testing.T) {
	cases := fixture[[]struct {
		Name   string          `json:"name"`
		Schema json.RawMessage `json:"schema"`
		Good   json.RawMessage `json:"good"`
		Bad    json.RawMessage `json:"bad"`
	}](t, "schema-matrix.json")
	for _, tc := range cases {
		for _, valid := range []bool{true, false} {
			label := "invalid"
			args := tc.Bad
			if valid {
				label, args = "valid", tc.Good
			}
			t.Run(tc.Name+"/"+label, func(t *testing.T) {
				var executed atomic.Int32
				runner, submitted, output := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
					executed.Add(1)
					return ToolResult{Output: "ok"}, nil
				}, string(tc.Schema), 204)
				event := signedEvent(t, string(args))
				if err := runner.Handle(context.Background(), event, event.ChatID); err != nil {
					t.Fatal(err)
				}
				if submitted.Load() != 1 {
					t.Fatalf("submissions: %d", submitted.Load())
				}
				if valid {
					if executed.Load() != 1 || output.IsError.Value || output.Output != "ok" {
						t.Fatalf("valid arguments: executed=%d output=%+v", executed.Load(), output)
					}
				} else if executed.Load() != 0 || !output.IsError.Value || !strings.Contains(output.Output, "invalid_arguments") {
					t.Fatalf("invalid arguments: executed=%d output=%+v", executed.Load(), output)
				}
				if err := runner.Handle(context.Background(), event, event.ChatID); err != nil || submitted.Load() != 1 {
					t.Fatalf("replay: %v submissions=%d", err, submitted.Load())
				}
			})
		}
	}
}

func TestExplicitToolSchemaDrafts(t *testing.T) {
	for _, draft := range []string{"http://json-schema.org/draft-04/schema#", "http://json-schema.org/draft-06/schema#", "http://json-schema.org/draft-07/schema#", "https://json-schema.org/draft/2019-09/schema", "https://json-schema.org/draft/2020-12/schema"} {
		t.Run(draft, func(t *testing.T) {
			schema := `{"$schema":` + strconv.Quote(draft) + `,"type":"object","properties":{"value":{"type":"integer"}},"required":["value"],"additionalProperties":false}`
			for _, valid := range []bool{true, false} {
				var effects atomic.Int32
				runner, submissions, out := toolHarness(t, func(context.Context, ToolInvocation) (ToolResult, error) {
					effects.Add(1)
					return ToolResult{Output: "ok"}, nil
				}, schema, 204)
				args := `{"value":1}`
				if !valid {
					args = `{"value":"1"}`
				}
				event := signedEvent(t, args)
				if err := runner.Handle(context.Background(), event, event.ChatID); err != nil {
					t.Fatal(err)
				}
				if submissions.Load() != 1 || (valid && (effects.Load() != 1 || out.IsError.Value)) || (!valid && (effects.Load() != 0 || !out.IsError.Value)) {
					t.Fatal(valid, effects.Load(), out)
				}
			}
		})
	}
}
