package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	envai "github.com/digitalygo/envai-teams-sdk-go"
)

func TestQuickstartOffline(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/events-valid.json")
	if err != nil {
		t.Fatal(err)
	}
	var events []envai.Event
	if err = json.Unmarshal(raw, &events); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile("../../testdata/api-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		ID       string          `json:"id"`
		Response json.RawMessage `json:"response"`
	}
	if err = json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	responses := map[string]json.RawMessage{}
	for _, c := range cases {
		responses[c.ID] = c.Response
	}
	var chatEnvelope map[string]json.RawMessage
	if err = json.Unmarshal(responses["createChat"], &chatEnvelope); err != nil {
		t.Fatal(err)
	}
	var chat map[string]any
	if err = json.Unmarshal(chatEnvelope["data"], &chat); err != nil {
		t.Fatal(err)
	}
	chat["id"] = events[0].ChatID
	chatEnvelope["data"], _ = json.Marshal(chat)
	responses["createChat"], _ = json.Marshal(chatEnvelope)
	sent := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fixture-only" {
			t.Error("missing authentication")
		}
		switch {
		case r.Method == "POST" && r.URL.Path == "/api/v1/workspaces":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			w.Write(responses["createWorkspace"])
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/chats"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			w.Write(responses["createChat"])
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/messages"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			fmt.Fprintf(w, `{"status":"accepted","turn_id":%q}`, *events[1].TurnID)
			close(sent)
		case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/events"):
			w.Header().Set("Content-Type", "text/event-stream")
			write := func(e envai.Event) {
				b, _ := json.Marshal(e)
				fmt.Fprintf(w, "data: %s\n\n", b)
				w.(http.Flusher).Flush()
			}
			write(events[0])
			select {
			case <-sent:
			case <-r.Context().Done():
				return
			}
			write(events[1])
			text := events[3]
			text.Seq = 2
			write(text)
			terminal := events[7]
			terminal.Seq = 3
			write(terminal)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := envai.New(envai.Config{BaseURL: server.URL, APIKey: "fixture-only", AllowHTTP: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var out bytes.Buffer
	if err = run(ctx, client, "team-example", &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "hello\nterminal: turn.completed") {
		t.Fatal(out.String())
	}
	if err = run(ctx, client, "", &out); err == nil {
		t.Fatal("missing team accepted")
	}
}
