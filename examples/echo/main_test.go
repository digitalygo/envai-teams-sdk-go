package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	envai "github.com/digitalygo/envai-teams-sdk-go"
)

// Exercise the real example with a local TLS server and an ephemeral test key.
func TestEchoExampleOffline(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/events-valid.json")
	if err != nil {
		t.Fatal(err)
	}
	var events []envai.Event
	if err = json.Unmarshal(raw, &events); err != nil {
		t.Fatal(err)
	}
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		CallID string `json:"call_id"`
	}
	if err = json.Unmarshal(events[4].Data, &original); err != nil {
		t.Fatal(err)
	}
	args := `{"text":"hello"}`
	signed := ed25519.Sign(key, []byte(original.CallID+"\necho\n"+args))
	events[4].Data, _ = json.Marshal(envai.ToolRequest{
		CallID: original.CallID, ToolName: "echo", Kind: "client", Args: json.RawMessage(args),
		Signature: envai.Signature{Alg: "EdDSA", Kid: "example-test", Sig: base64.RawURLEncoding.EncodeToString(signed)},
	})
	sent, delivered := make(chan struct{}), make(chan struct{})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Envai-Contract-Version", "3")
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"protocol_version":3,"contract_version":3,"supported_contract_versions":[2,3],"legacy_v2_mode":"historical_reads_replay_only"}`)
			return
		}
		switch {
		case r.URL.Path == "/.well-known/jwks.json":
			if r.Header.Get("Authorization") != "" {
				t.Error("JWKS received bearer token")
			}
			json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kty": "OKP", "crv": "Ed25519", "use": "sig", "alg": "EdDSA", "kid": "example-test", "x": base64.RawURLEncoding.EncodeToString(pub)}}})
		case strings.HasSuffix(r.URL.Path, "/messages"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			fmt.Fprintf(w, `{"status":"accepted","turn_id":%q}`, *events[1].TurnID)
			close(sent)
		case strings.HasSuffix(r.URL.Path, "/tool-output"):
			var result envai.ToolOutputRequest
			if err := json.NewDecoder(r.Body).Decode(&result); err != nil || result.Output != `{"text":"hello"}` {
				t.Errorf("unexpected output: %+v %v", result, err)
			}
			w.WriteHeader(204)
			close(delivered)
		case strings.HasSuffix(r.URL.Path, "/events"):
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
			request := events[4]
			request.Seq = 2
			write(request)
			select {
			case <-delivered:
			case <-r.Context().Done():
				return
			}
			result := events[5]
			result.Seq = 3
			write(result)
			terminal := events[7]
			terminal.Seq = 4
			write(terminal)
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	caPath := filepath.Join(t.TempDir(), "test-ca.pem")
	if err = os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SSL_CERT_FILE", caPath)
	t.Setenv("ENVAI_BASE_URL", server.URL)
	t.Setenv("ENVAI_API_KEY", "example-fixture")
	t.Setenv("ENVAI_CHAT_ID", events[0].ChatID)
	t.Setenv("ENVAI_STATE_DIR", t.TempDir())
	t.Setenv("ENVAI_ACCOUNT_SCOPE", "example/account")
	if err = run(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-delivered:
	default:
		t.Fatal("example did not deliver output")
	}
}
