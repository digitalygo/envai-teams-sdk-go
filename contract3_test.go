package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func rawClient(t *testing.T, fn http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(fn)
	t.Cleanup(server.Close)
	c, err := New(Config{BaseURL: server.URL + "/api/v1", APIKey: "test", AllowHTTP: true})
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestContract3NegotiationAndFailClosed(t *testing.T) {
	for _, failure := range []string{"old404", "noack", "badack", "duplicateack", "oldprotocol", "unsupported", "missingSupported", "mutation404", "mutationNoack", "mutation503", "success"} {
		t.Run(failure, func(t *testing.T) {
			var mutations atomic.Int32
			var preflights atomic.Int32
			c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get(ContractHeader) != "3" {
					t.Error("missing automatic selector")
				}
				if r.URL.Path == "/api/v1/atp/v3/capabilities" {
					preflights.Add(1)
					if failure == "old404" {
						w.WriteHeader(404)
						return
					}
					if failure == "unsupported" {
						w.WriteHeader(400)
						fmt.Fprint(w, `{"error":{"code":"unsupported_contract_version"}}`)
						return
					}
					if failure != "noack" {
						w.Header().Set(ContractHeader, "3")
					}
					if failure == "badack" {
						w.Header().Set(ContractHeader, "2")
					}
					if failure == "duplicateack" {
						w.Header().Add(ContractHeader, "3")
					}
					response := testCapabilities
					if failure == "oldprotocol" {
						response = `{"protocol_version":2}`
					}
					if failure == "missingSupported" {
						response = strings.Replace(response, `[2,3]`, `[2]`, 1)
					}
					fmt.Fprint(w, response)
					return
				}
				if r.URL.Path != "/api/v1/atp/v3/chats/chat/messages" {
					t.Errorf("unsafe mutation path: %s", r.URL.Path)
				}
				mutations.Add(1)
				if failure == "mutation404" {
					w.WriteHeader(404)
					return
				}
				if failure == "mutation503" {
					w.WriteHeader(503)
					return
				}
				if failure != "mutationNoack" {
					w.Header().Set(ContractHeader, "3")
				}
				w.WriteHeader(202)
				fmt.Fprint(w, `{"status":"accepted","turn_id":"turn"}`)
			})
			c.config.RetryAttempts = 3
			_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{Message: "hello"}})
			if failure == "success" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("unsafe success")
			}
			expected := int32(0)
			if strings.HasPrefix(failure, "mutation") || failure == "success" {
				expected = 1
			}
			if mutations.Load() != expected || preflights.Load() != 1 {
				t.Fatal(mutations.Load(), preflights.Load())
			}
		})
	}
}
func TestLegacyReadOnlyAndCosts(t *testing.T) {
	var count atomic.Int32
	c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		if r.Header.Get(ContractHeader) != "2" {
			t.Error("not v2")
		}
		if strings.Contains(r.URL.Path, "atp/v3") {
			t.Error("not legacy path")
		}
		fmt.Fprint(w, `{"protocol_version":2}`)
	})
	c.config.LegacyReadOnly = true
	if _, err := c.GetCapabilities(context.Background(), GetCapabilitiesParams{}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat"}); err == nil {
		t.Fatal("legacy mutation allowed")
	}
	if _, err := c.CreateWorkspace(context.Background(), CreateWorkspaceParams{}); err == nil {
		t.Fatal("legacy CRUD mutation allowed")
	}
	if _, err := c.RunTurn(context.Background(), "chat", "hello", TurnOptions{}); err == nil {
		t.Fatal("legacy RunTurn allowed")
	}
	if count.Load() != 1 {
		t.Fatal(count.Load())
	}
	for _, raw := range []string{`{}`, `{"cost_usd":null}`, `{"cost_usd":"0"}`, `{"cost_usd":false}`} {
		if validateCompletionJSON([]byte(raw), true) == nil {
			t.Fatal(raw)
		}
	}
	for _, cost := range []string{"0", "0.25", "123.456"} {
		raw := []byte(`{"cost_usd":` + cost + `}`)
		if err := validateCompletionJSON(raw, true); err != nil {
			t.Fatal(err)
		}
	}
	if err := validateCompletionJSON([]byte(`{"cost_usd":null}`), false); err != nil {
		t.Fatal(err)
	}
}
func TestMixedReplayStrictBeforeHydration(t *testing.T) {
	all := fixture[[]json.RawMessage](t, "events-valid.json")
	for _, raw := range all {
		var event map[string]any
		json.Unmarshal(raw, &event)
		for _, version := range []int{2, 3} {
			event["v"] = version
			b, _ := json.Marshal(event)
			e, err := DecodeEvent(b)
			if err != nil {
				t.Fatalf("numeric v%d %s: %v", version, b, err)
			}
			if !bytes.Equal(e.Raw, b) {
				t.Fatal("rewrote raw")
			}
		}
		data := event["data"].(map[string]any)
		_, cost := data["cost"]
		_, usage := data["usage"]
		if !cost && !usage {
			continue
		}
		if cost {
			data["cost"] = nil
			data["chat_cost"] = nil
		}
		if usage {
			data["usage"].(map[string]any)["cost_usd"] = nil
		}
		if u, ok := data["chat_usage"].(map[string]any); ok {
			u["cost_usd"] = nil
		}
		event["v"] = 2
		b, _ := json.Marshal(event)
		if _, err := DecodeEvent(b); err == nil {
			t.Fatal("v2 null decoded")
		}
		event["v"] = 3
		b, _ = json.Marshal(event)
		e, err := DecodeEvent(b)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = e.TypedPayload(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := DecodeEvent([]byte(`{"v":9}`)); err == nil {
		t.Fatal("unknown version")
	}
}
func TestCompletionNegotiatedReadsAndImmutableSnapshot(t *testing.T) {
	legacy, _ := os.ReadFile("testdata/control-plane-http/completion-legacy.json")
	var fixtureEnvelope map[string]json.RawMessage
	json.Unmarshal(legacy, &fixtureEnvelope)
	// The source fixture is an actual HTTP response, retained byte-for-byte.
	body := legacy
	if response := fixtureEnvelope["response"]; len(response) > 0 {
		body = response
	}
	var completionEnvelope map[string]json.RawMessage
	if json.Unmarshal(body, &completionEnvelope) != nil || len(completionEnvelope["data"]) == 0 {
		t.Fatal("missing completion fixture data", string(body))
	}
	var row map[string]json.RawMessage
	json.Unmarshal(completionEnvelope["data"], &row)
	row["cost_usd"] = json.RawMessage("null")
	id := row["id"]
	snapshot := map[string]any{"event_id": "evt_fixture", "event_type": "turn.completed", "occurred_at": "2026-10-13T16:00:00Z", "data": map[string]any{"turn_id": json.RawMessage(id), "cost": nil}}
	row["terminal_snapshot"], _ = json.Marshal(snapshot)
	newRow, _ := json.Marshal(row)
	response, _ := json.Marshal(map[string]any{"data": json.RawMessage(newRow)})
	c := rawClient(t, func(w http.ResponseWriter, r *http.Request) { w.Header().Set(ContractHeader, "3"); w.Write(response) })
	out, err := c.GetCompletion(context.Background(), GetCompletionParams{WorkspaceID: "w", ChatID: "c", CompletionID: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if out.CostUSD != nil || !out.TerminalSnapshot.Set {
		t.Fatal(out)
	}
	snapshot["data"].(map[string]any)["turn_id"] = "wrong"
	row["terminal_snapshot"], _ = json.Marshal(snapshot)
	newRow, _ = json.Marshal(row)
	if validateCompletionJSON(newRow, false) == nil {
		t.Fatal("cross-turn snapshot")
	}
}
func TestUnknownFieldsAndCurrencyProvenance(t *testing.T) {
	var fixture struct{ Data json.RawMessage }
	b, _ := os.ReadFile("testdata/wiki/usage-pending.json")
	json.Unmarshal(b, &fixture)
	var raw map[string]json.RawMessage
	json.Unmarshal(fixture.Data, &raw)
	raw["extension"] = json.RawMessage(`{"n":9007199254740993}`)
	input, _ := json.Marshal(raw)
	var usage UsageCost
	if err := json.Unmarshal(input, &usage); err != nil {
		t.Fatal(err)
	}
	if usage.Totals != nil || usage.Status != "pending" {
		t.Fatal(usage)
	}
	output, err := json.Marshal(usage)
	if err != nil {
		t.Fatal(err)
	}
	var round map[string]json.RawMessage
	json.Unmarshal(output, &round)
	if !bytes.Equal(input, output) {
		t.Fatal(string(input), string(output))
	}
	var amount CurrencyAmount
	if err = json.Unmarshal([]byte(`{"currency":"EUR","amount":"123456789.00000000000000000001"}`), &amount); err != nil {
		t.Fatal(err)
	}
	if amount.Amount != "123456789.00000000000000000001" || amount.Currency != "EUR" {
		t.Fatal(amount)
	}
	raw["totals"] = json.RawMessage(`[]`)
	input, _ = json.Marshal(raw)
	if json.Unmarshal(input, &usage) == nil {
		t.Fatal("premature total accepted")
	}
	if json.Unmarshal([]byte(`{"currency":"EUR","amount":0.1}`), &amount) == nil {
		t.Fatal("float amount accepted")
	}
}
func TestV3ReconnectAndLegacyFiniteEOF(t *testing.T) {
	events := fixture[[]Event](t, "events-valid.json")
	var connects atomic.Int32
	c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
		v := r.Header.Get(ContractHeader)
		w.Header().Set(ContractHeader, v)
		w.Header().Set("Content-Type", "text/event-stream")
		if v == "3" && !strings.Contains(r.URL.Path, "/atp/v3/") {
			t.Error(r.URL.Path)
		}
		count := connects.Add(1)
		event := events[0]
		if count > 1 {
			event = events[1]
		}
		if v == "3" {
			event.V = 3
		}
		writeEvent(w, event)
	})
	stream, err := c.Events(context.Background(), events[0].ChatID, StreamOptions{MaxReconnects: 1, ReconnectDelay: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if e, err := stream.Next(); err != nil || e.V != 3 {
		t.Fatal(e, err)
	}
	if e, err := stream.Next(); err != nil || e.Seq != 1 {
		t.Fatal(e, err)
	}
	connects.Store(0)
	old, err := c.LegacyEvents(context.Background(), events[0].ChatID, StreamOptions{MaxReconnects: 5})
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	old.Next()
	if _, err = old.Next(); !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if connects.Load() != 1 {
		t.Fatal("legacy reconnected")
	}
}

func TestExplicitLegacyMethodsBoundToVersionTwo(t *testing.T) {
	for _, wireVersion := range []int{2, 3} {
		t.Run(fmt.Sprint(wireVersion), func(t *testing.T) {
			c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/capabilities" || r.Header.Get(ContractHeader) != "2" {
					t.Error(r.URL.Path, r.Header)
				}
				w.Header().Set(ContractHeader, "2")
				fmt.Fprintf(w, `{"protocol_version":%d}`, wireVersion)
			})
			_, err := c.GetCapabilitiesLegacy(context.Background(), GetCapabilitiesLegacyParams{})
			if (err == nil) != (wireVersion == 2) {
				t.Fatal(err)
			}
		})
	}
	events := fixture[[]Event](t, "events-valid.json")
	c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/atp/v3/") || r.Header.Get(ContractHeader) != "2" {
			t.Error(r.URL.Path, r.Header)
		}
		w.Header().Set(ContractHeader, "2")
		w.Header().Set("Content-Type", "text/event-stream")
		e := events[0]
		e.V = 3
		writeEvent(w, e)
	})
	s, err := c.StreamChatEventsLegacy(context.Background(), StreamChatEventsLegacyParams{ChatID: events[0].ChatID})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.Next(); err == nil {
		t.Fatal("legacy accepted v3")
	}
}
func TestCompletionPageRejectsAtomicallyAndSchemaHelpers(t *testing.T) {
	c := &Client{config: Config{LegacyReadOnly: true}}
	for _, bad := range []string{`[]`, `{"data":{}}`, `{"data":[{"cost_usd":1},{"cost_usd":null}]}`} {
		if c.validateNegotiatedJSON("listCompletions", []byte(bad)) == nil {
			t.Fatal("invalid legacy page", bad)
		}
	}
	if err := c.validateNegotiatedJSON("listCompletions", []byte(`{"data":[{"cost_usd":0},{"cost_usd":0.25}]}`)); err != nil {
		t.Fatal(err)
	}
	cause := errors.New("test")
	e := &CompatibilityError{Reason: "test reason", Cause: cause}
	if !errors.Is(e, cause) || !strings.Contains(e.Error(), "test reason") {
		t.Fatal(e)
	}
	if _, err := unknownFields([]byte("invalid"), nil); err == nil {
		t.Fatal("invalid extra JSON")
	}
	if _, err := marshalUnknown(make(chan int), nil); err == nil {
		t.Fatal("invalid model marshal")
	}
	if _, err := marshalUnknown([]string{"x"}, map[string]json.RawMessage{"x": json.RawMessage(`1`)}); err == nil {
		t.Fatal("non-object extra")
	}
	if _, err := marshalUnknown(map[string]string{"x": "typed"}, map[string]json.RawMessage{"x": json.RawMessage(`1`)}); err == nil {
		t.Fatal("extra shadows typed")
	}
	if err := validateDefinition("wiki.json", "missing", []byte(`{}`)); err == nil {
		t.Fatal("missing schema")
	}
	if err := validateDefinition("wiki.json", "UsageCost", []byte("invalid")); err == nil {
		t.Fatal("invalid JSON")
	}
	v := &WikiWebhookVerifier{}
	if time.Since(v.now()) > time.Second {
		t.Fatal("default clock")
	}
}

func TestATPTypedUsageAndChatTitleCoverage(t *testing.T) {
	for _, name := range []string{"usage-pending", "usage-empty"} {
		data := wikiFixtureData(t, name)
		var usage ATPUsageCost
		if err := json.Unmarshal(data, &usage); err != nil {
			t.Fatal(err)
		}
		if _, err := json.Marshal(usage); err != nil {
			t.Fatal(err)
		}
	}
	var billing ATPBillingState
	if err := json.Unmarshal([]byte(`{"status":"inactive","extension":true}`), &billing); err != nil {
		t.Fatal(err)
	}
	if _, err := json.Marshal(billing); err != nil {
		t.Fatal(err)
	}
	var amount ATPCurrencyAmount
	if err := json.Unmarshal([]byte(`{"currency":"EUR","amount":"1.0000000000001","source":"reported"}`), &amount); err != nil {
		t.Fatal(err)
	}
	if _, err := json.Marshal(amount); err != nil {
		t.Fatal(err)
	}
	all := fixture[[]Event](t, "events-valid.json")
	event := all[1]
	event.V = 3
	event.Type = "chat.title.updated"
	event.Agent = nil
	event.Data = json.RawMessage(`{"title":"New chat title"}`)
	raw, _ := json.Marshal(event)
	decoded, err := DecodeEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	typed, err := decoded.TypedPayload()
	if err != nil {
		t.Fatal(err)
	}
	if typed.(ATPChatTitleUpdated).Title != "New chat title" {
		t.Fatal(typed)
	}
}

func TestCompletionUnknownFieldsPreserved(t *testing.T) {
	var completion Completion
	if err := json.Unmarshal([]byte(`{"cost_usd":null,"future":{"count":9007199254740993}}`), &completion); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(completion)
	if err != nil || !strings.Contains(string(raw), `"future":{"count":9007199254740993}`) {
		t.Fatal(string(raw), err)
	}
	for _, invalid := range []string{"null", "[]", `{"cost_usd":"unknown"}`} {
		if json.Unmarshal([]byte(invalid), &completion) == nil {
			t.Fatal("invalid completion", invalid)
		}
	}
}
