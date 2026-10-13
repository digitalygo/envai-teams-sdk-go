package envai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	envai "github.com/digitalygo/envai-teams-sdk-go"
)

func TestIndependentClientFormattingRedactsCredentials(t *testing.T) {
	client, err := envai.New(envai.Config{BaseURL: "https://example.test/api/v1", APIKey: "independent-secret"})
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if strings.Contains(fmt.Sprintf(format, client), "independent-secret") {
			t.Fatalf("client formatting leaks credentials for %s", format)
		}
	}
}

func TestIndependentWhitespaceArgumentsFailClosed(t *testing.T) {
	err := (&envai.Verifier{}).Verify(context.Background(), envai.ToolRequest{
		Kind: "client", Args: json.RawMessage(" \n\t"), Signature: envai.Signature{Alg: "EdDSA", Kid: "fixture"},
	})
	if err == nil {
		t.Fatal("whitespace arguments were accepted")
	}
}

func TestIndependentSigningFixturesThroughPublicVerifier(t *testing.T) {
	data, err := os.ReadFile("testdata/signatures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		JWKS    json.RawMessage `json:"jwks"`
		Vectors []struct {
			Name      string          `json:"name"`
			Args      string          `json:"args_json"`
			CallID    string          `json:"call_id"`
			ToolName  string          `json:"tool_name"`
			Signature envai.Signature `json:"signature"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Envai-Contract-Version", "3")
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"protocol_version":3,"contract_version":3,"supported_contract_versions":[2,3],"legacy_v2_mode":"historical_reads_replay_only"}`)
			return
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("JWKS request disclosed bearer token")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture.JWKS)
	}))
	defer server.Close()
	client, err := envai.New(envai.Config{BaseURL: server.URL + "/api/v1", APIKey: "secret", AllowHTTP: true})
	if err != nil {
		t.Fatal(err)
	}
	verifier := &envai.Verifier{Client: client}
	for _, vector := range fixture.Vectors {
		t.Run(vector.Name, func(t *testing.T) {
			request := envai.ToolRequest{CallID: vector.CallID, ToolName: vector.ToolName, Kind: "client", Args: json.RawMessage(vector.Args), Signature: vector.Signature}
			if err := verifier.Verify(context.Background(), request); err != nil {
				t.Fatal(err)
			}
			request.Args = json.RawMessage(`{"tampered":true}`)
			if err := verifier.Verify(context.Background(), request); err == nil {
				t.Fatal("tampered arguments accepted")
			}
		})
	}
}

func TestIndependentSSEEveryChunkBoundary(t *testing.T) {
	data, err := os.ReadFile("testdata/sse-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Name string   `json:"name"`
		Wire string   `json:"wire"`
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		for split := 0; split <= len(fixture.Wire); split++ {
			decoder := envai.NewSSEDecoder(io.MultiReader(strings.NewReader(fixture.Wire[:split]), strings.NewReader(fixture.Wire[split:])), 4096)
			for _, expected := range fixture.Data {
				actual, err := decoder.Next()
				if err != nil || actual != expected {
					t.Fatalf("%s split %d: %q %v", fixture.Name, split, actual, err)
				}
			}
			if _, err := decoder.Next(); err != io.EOF {
				t.Fatalf("%s split %d: expected EOF, got %v", fixture.Name, split, err)
			}
		}
	}
}

func TestIndependentInvalidUTF8EventRejected(t *testing.T) {
	data, err := os.ReadFile("testdata/events-valid.json")
	if err != nil {
		t.Fatal(err)
	}
	var events []json.RawMessage
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatal(err)
	}
	invalid := bytes.Replace(events[3], []byte("hello"), []byte{0xff}, 1)
	if _, err := envai.DecodeEvent(invalid); err == nil {
		t.Fatal("invalid UTF-8 wire event accepted")
	}
}

func TestIndependentHandlerReceivesVerifiedCanonicalNumbers(t *testing.T) {
	data, err := os.ReadFile("testdata/signatures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		JWKS    json.RawMessage `json:"jwks"`
		Vectors []struct {
			Name      string          `json:"name"`
			Args      string          `json:"args_json"`
			Canonical string          `json:"canonical"`
			CallID    string          `json:"call_id"`
			ToolName  string          `json:"tool_name"`
			Signature envai.Signature `json:"signature"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	vector := fixture.Vectors[3]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Envai-Contract-Version", "3")
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"protocol_version":3,"contract_version":3,"supported_contract_versions":[2,3],"legacy_v2_mode":"historical_reads_replay_only"}`)
			return
		}
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(fixture.JWKS)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()
	client, err := envai.New(envai.Config{BaseURL: server.URL + "/api/v1", APIKey: "secret", AllowHTTP: true})
	if err != nil {
		t.Fatal(err)
	}
	var arguments string
	registry, err := envai.NewRegistry(envai.Tool{Name: vector.ToolName, Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(ctx context.Context, invocation envai.ToolInvocation) (envai.ToolResult, error) {
		arguments = string(invocation.Arguments)
		return envai.ToolResult{Output: "ok"}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := envai.NewToolRunner(client, registry, envai.NewMemoryCallStore(), envai.RunnerOptions{Namespace: "independent-review"})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := os.ReadFile("testdata/events-valid.json")
	if err != nil {
		t.Fatal(err)
	}
	var events []json.RawMessage
	if err := json.Unmarshal(wire, &events); err != nil {
		t.Fatal(err)
	}
	event, err := envai.DecodeEvent(events[4])
	if err != nil {
		t.Fatal(err)
	}
	altered := strings.Replace(vector.Args, "1e20", "100000000000000000001", 1)
	event.Data, err = json.Marshal(envai.ToolRequest{CallID: vector.CallID, ToolName: vector.ToolName, Kind: "client", Args: json.RawMessage(altered), Signature: vector.Signature})
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Handle(context.Background(), event, event.ChatID); err != nil {
		t.Fatal(err)
	}
	if arguments != vector.Canonical {
		t.Fatalf("handler received unverified numeric spelling: %s", arguments)
	}
}
