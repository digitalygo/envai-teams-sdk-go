package envai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestMessageModelValidationAndWire(t *testing.T) {
	for _, strict := range []bool{false, true} {
		for _, tc := range []struct {
			name  string
			model Optional[string]
			extra map[string]json.RawMessage
			valid bool
		}{
			{name: "absent", valid: true}, {name: "null", model: Null[string]()}, {name: "empty", model: Present("")}, {name: "space", model: Present(" x")}, {name: "unicode space", model: Present("x\u2003")}, {name: "valid", model: Present("Vendor/模型"), valid: true}, {name: "raw number", extra: map[string]json.RawMessage{"model": json.RawMessage(`3`)}}, {name: "raw null", extra: map[string]json.RawMessage{"model": json.RawMessage(`null`)}},
		} {
			t.Run(fmt.Sprintf("strict=%t/%s", strict, tc.name), func(t *testing.T) {
				posts := 0
				c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
					posts++
					var body map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if tc.name == "valid" && string(body["model"]) != `"Vendor/模型"` {
						t.Error(string(body["model"]))
					}
					w.WriteHeader(202)
					fmt.Fprint(w, `{"status":"accepted","turn_id":"cmpl_01j5abcdefghjkmnpqrstvwxyz"}`)
				})
				c.config.StrictValidation = strict
				_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "cht_01j5abcdefghjkmnpqrstvwxyz", Body: ChatMessageRequest{Message: "hello", Model: tc.model, Extra: tc.extra}})
				if tc.valid && (err != nil || posts != 1) {
					t.Fatalf("err=%v posts=%d", err, posts)
				}
				if !tc.valid && (err == nil || posts != 0) {
					t.Fatalf("err=%v posts=%d", err, posts)
				}
			})
		}
	}
}

func TestInvalidRawMessageBody(t *testing.T) {
	c := negotiatedTestClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected request") })
	for _, body := range []json.RawMessage{json.RawMessage(`{`), json.RawMessage(`[1]`)} {
		if _, err := c.do(context.Background(), "sendChatMessage", map[string]any{"chat_id": "chat"}, body, "application/json"); err == nil {
			t.Fatal("invalid body accepted")
		}
	}
}

func TestExplicitModelWithStrictModelPolicies(t *testing.T) {
	for _, policy := range []Optional[bool]{{}, Present(false), Present(true)} {
		t.Run(fmt.Sprint(policy), func(t *testing.T) {
			posts := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					fmt.Fprint(w, backgroundCapabilities)
					return
				}
				posts++
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["model"] != "vendor/model" {
					t.Error(body, err)
				}
				w.WriteHeader(202)
				fmt.Fprint(w, `{"status":"accepted","turn_id":"cmpl_01j5abcdefghjkmnpqrstvwxyz"}`)
			})
			c.config.NativeTasks = policy.Set && policy.Value
			_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", Body: ChatMessageRequest{Message: "hello", Model: Present("vendor/model"), StrictModel: policy}})
			if err != nil || posts != 1 {
				t.Fatal(err, posts)
			}
		})
	}
}

func TestModelErrorsAndNoRetry(t *testing.T) {
	for _, tc := range []struct {
		code   string
		status int
	}{{"invalid_model", 422}, {"unknown_model", 422}, {"model_unavailable", 422}, {"model_incompatible", 422}, {"model_catalog_unavailable", 503}, {"idempotency_conflict", 409}} {
		t.Run(tc.code, func(t *testing.T) {
			posts := 0
			c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				posts++
				w.WriteHeader(tc.status)
				fmt.Fprintf(w, `{"error":{"type":"validation_error","code":%q,"message":"failure"}}`, tc.code)
			})
			_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "chat", IdempotencyKey: Present("key"), Body: ChatMessageRequest{Message: "hello", Model: Present("vendor/model")}})
			var api *APIError
			if !errors.As(err, &api) || api.Status != tc.status || api.Code != tc.code || posts != 1 {
				t.Fatalf("error=%v posts=%d", err, posts)
			}
		})
	}
}

func TestModelSelectableFalseRetained(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"object":"list","data":[{"id":"vendor/model","object":"model","created":0,"owned_by":"vendor","type":"llm","deprecated":false,"selectable":false,"sunset_at":null}],"defaults":{}}`)
	})
	list, err := c.ListModels(context.Background(), ListModelsParams{})
	if err != nil || len(list.Data) != 1 {
		t.Fatal(list, err)
	}
	b, err := json.Marshal(list.Data[0])
	if err != nil {
		t.Fatal(err)
	}
	var model LLMModelEntry
	if err := json.Unmarshal(b, &model); err != nil || model.ID != "vendor/model" || model.Selectable {
		t.Fatal(model, err)
	}
}

func TestEventRequestedAndObservedModelsRemainDistinct(t *testing.T) {
	var event ATPAgentDone
	if err := json.Unmarshal([]byte(`{"status":"completed","summary":"","usage":{},"duration_ms":1,"requested_model":"vendor/requested","model":"vendor/actual"}`), &event); err != nil {
		t.Fatal(err)
	}
	if !event.RequestedModel.Set || event.RequestedModel.Value != "vendor/requested" || !event.Model.Set || event.Model.Value != "vendor/actual" {
		t.Fatal(event)
	}
}

func TestRunTurnModelIsPerTurn(t *testing.T) {
	events := fixture[[]Event](t, "events-valid.json")
	sent := make(chan struct{}, 2)
	posts := 0
	c := attachmentClient(t, true, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/messages") {
			posts++
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if posts == 1 {
				if string(body["model"]) != `"vendor/model"` || string(body["attachment_ids"]) != `["a"]` || r.Header.Get("Idempotency-Key") != "key" {
					t.Error(body, r.Header)
				}
			} else if _, ok := body["model"]; ok {
				t.Error("model persisted", body)
			}
			w.WriteHeader(202)
			if posts == 1 {
				fmt.Fprintf(w, `{"status":"accepted","turn_id":%q,"attachment_ids":["a"]}`, *events[1].TurnID)
			} else {
				fmt.Fprintf(w, `{"status":"accepted","turn_id":%q}`, *events[1].TurnID)
			}
			sent <- struct{}{}
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		writeEvent(w, events[0])
		w.(http.Flusher).Flush()
		select {
		case <-sent:
		case <-r.Context().Done():
			return
		}
		terminal := events[7]
		terminal.Seq = 1
		writeEvent(w, terminal)
	})
	for _, opts := range []TurnOptions{{Model: Present("vendor/model"), AttachmentIDs: []string{"a"}, IdempotencyKey: "key"}, {}} {
		if _, err := c.RunTurn(context.Background(), events[0].ChatID, "hello", opts); err != nil {
			t.Fatal(err)
		}
	}
	if posts != 2 {
		t.Fatal(posts)
	}
}
