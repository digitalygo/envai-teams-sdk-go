package envai

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestIndependentMalformedCompletionGuard(t *testing.T) {
	for _, s := range []string{`{"cost_usd":-1}`, `{"cost_usd":1e400}`, `{"cost_usd":0,"usage":{"cost_usd":null}}`} {
		if err := validateCompletionJSON([]byte(s), true); err == nil {
			t.Errorf("accepted legacy %s", s)
		}
	}
}
func TestIndependentCompletionUsageUnknownFields(t *testing.T) {
	var c Completion
	input := []byte(`{"cost_usd":0,"usage":{"future_metric":7}}`)
	if err := json.Unmarshal(input, &c); err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(c)
	var row map[string]any
	_ = json.Unmarshal(out, &row)
	if _, ok := row["usage"].(map[string]any)["future_metric"]; !ok {
		t.Errorf("lost nested allowed unknown usage: %s", out)
	}
}

func TestNestedMonetaryValidationAndExtensions(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		for _, s := range []string{`{"cost_usd":-1}`, `{"cost_usd":1e400}`, `{"cost_usd":0,"usage":{"cost_usd":-1}}`, `{"cost_usd":0,"usage":{"cost_usd":1e400}}`, `{"cost_usd":0,"usage":{"cost_usd":"free"}}`} {
			if err := validateCompletionJSON([]byte(s), legacy); err == nil {
				t.Fatal("invalid monetary value", legacy, s)
			}
		}
	}
	if err := validateCompletionJSON([]byte(`{"cost_usd":null,"usage":{"cost_usd":null},"future":{"cost":"free","data":"opaque"}}`), false); err != nil {
		t.Fatal("valid future fields rejected", err)
	}
	all := fixture[[]json.RawMessage](t, "events-valid.json")
	for _, raw := range all {
		var event map[string]json.RawMessage
		json.Unmarshal(raw, &event)
		var typ string
		json.Unmarshal(event["type"], &typ)
		if typ != "agent.done" {
			continue
		}
		var data map[string]json.RawMessage
		json.Unmarshal(event["data"], &data)
		var usage map[string]json.RawMessage
		json.Unmarshal(data["usage"], &usage)
		usage["cost_usd"] = json.RawMessage(`1e400`)
		data["usage"], _ = json.Marshal(usage)
		event["data"], _ = json.Marshal(data)
		event["v"] = json.RawMessage(`3`)
		modified, _ := json.Marshal(event)
		if _, err := DecodeEvent(modified); err == nil {
			t.Fatal("overflowing ATP cost accepted")
		}
	}
}

func TestVerifiedBackendHTTPFixtures(t *testing.T) {
	for _, kind := range []string{"source", "completion-pending", "legacy-completion"} {
		t.Run(kind, func(t *testing.T) {
			body, err := os.ReadFile("testdata/v3-http/" + kind + ".json")
			if err != nil {
				t.Fatal(err)
			}
			headers := map[string][]string{}
			headerFile := "completion-headers"
			if kind == "source" {
				headerFile = "source-headers"
			}
			rawHeaders, _ := os.ReadFile("testdata/v3-http/" + headerFile + ".json")
			json.Unmarshal(rawHeaders, &headers)
			c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
				for name, values := range headers {
					for _, value := range values {
						w.Header().Add(name, value)
					}
				}
				if kind == "legacy-completion" {
					w.Header().Set(ContractHeader, "2")
				}
				w.Write(body)
			})
			if kind == "source" {
				source, err := c.GetWikiSource(context.Background(), GetWikiSourceParams{WorkspaceID: "w", FileID: "f"})
				if err != nil || source.Conversion.ErrorCode.Value != "provider_limit_exceeded" || !source.Conversion.Recoverable.Value || source.UsageCost.Totals == nil || len(source.UsageCost.Totals) != 0 {
					t.Fatal(source, err)
				}
				return
			}
			if kind == "legacy-completion" {
				c.config.LegacyReadOnly = true
			}
			completion, err := c.GetCompletion(context.Background(), GetCompletionParams{WorkspaceID: "w", ChatID: "c", CompletionID: "t"})
			if err != nil {
				t.Fatal(err)
			}
			if kind == "legacy-completion" {
				if completion.CostUSD == nil || *completion.CostUSD != .25 || completion.UsageCost.Set || completion.TerminalSnapshot.Set {
					t.Fatal(completion)
				}
				return
			}
			if completion.CostUSD != nil || completion.UsageCost.Value.Status != "pending" || !completion.TerminalSnapshot.Set || completion.TerminalSnapshot.Value.AcceptedProtocolVersion.Value != 3 {
				t.Fatal(completion)
			}
			immutable, _ := json.Marshal(completion.TerminalSnapshot)
			completion.UsageCost.Value.Revision++
			projection := .25
			completion.CostUSD = &projection
			after, _ := json.Marshal(completion.TerminalSnapshot)
			if string(immutable) != string(after) {
				t.Fatal("mutable terminal snapshot")
			}
		})
	}
}

func TestNestedLegacyUsageCostDescendants(t *testing.T) {
	for _, raw := range []string{`{"cost_usd":0,"usage":{"agent":{"cost_usd":null}}}`, `{"cost_usd":0,"usage":{"agents":[{"cost_usd":null}]}}`, `{"cost_usd":0,"usage":{"agents":[{"cost":-1}]}}`, `{"cost_usd":0,"usage":{"agents":[{"cost_usd":1e400}]}}`} {
		if validateCompletionJSON([]byte(raw), true) == nil {
			t.Fatal("legacy descendant accepted", raw)
		}
	}
	if err := validateCompletionJSON([]byte(`{"cost_usd":null,"usage":{"agents":[{"cost_usd":null,"opaque":"future"}]},"extension":{"cost":"opaque"}}`), false); err != nil {
		t.Fatal(err)
	}
}
