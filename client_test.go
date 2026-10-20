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
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func fixture[T any](t *testing.T, name string) T {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	var out T
	if err = json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func testClient(t *testing.T, fn http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(ContractHeader) != "" {
			w.Header().Set(ContractHeader, r.Header.Get(ContractHeader))
		}
		fn(w, r)
	}))
	t.Cleanup(srv.Close)
	c, err := New(Config{BaseURL: srv.URL, APIKey: "fixture-key", AllowHTTP: true, RetryDelay: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestGeneratedOperationInventoryAndTypedMethods(t *testing.T) {
	cases := fixture[[]struct {
		ID         string          `json:"id"`
		GoName     string          `json:"go_name"`
		Parameters map[string]any  `json:"parameters"`
		Body       json.RawMessage `json:"body"`
		Status     int             `json:"status"`
		Response   json.RawMessage `json:"response"`
	}](t, "api-cases.json")
	inventory := fixture[[]struct {
		ID            string `json:"id"`
		Method        string `json:"method"`
		Path          string `json:"path"`
		Authenticated bool   `json:"authenticated"`
	}](t, "operations.json")
	if len(cases) != 71 || len(inventory) != 71 {
		t.Fatal("contract inventory changed")
	}
	for _, tc := range cases {
		t.Run(tc.ID, func(t *testing.T) {
			calls := 0
			var client *Client
			client = testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") && tc.ID != "getCapabilities" && tc.ID != "getCapabilitiesLegacy" {
					if strings.Contains(tc.ID, "ChatAttachment") {
						fmt.Fprint(w, strings.Replace(testCapabilities, `"features":[]`, `"features":["chat_attachments","chat_attachment_originals"]`, 1))
					} else {
						writeCapabilities(w)
					}
					return
				}
				calls++
				op := client.operations[tc.ID]
				if r.Method != op.Method {
					t.Errorf("method %s", r.Method)
				}
				wantAuth := ""
				if op.Authenticated {
					wantAuth = "Bearer fixture-key"
				}
				if r.Header.Get("Authorization") != wantAuth {
					t.Errorf("auth %q", r.Header.Get("Authorization"))
				}
				if strings.Contains(r.URL.Path, "{") {
					t.Error("unresolved path")
				}
				if tc.ID == "downloadArtifact" && r.URL.RawQuery != "artifact=art_01ARZ3NDEKTSV4RRFFQ69G5FAV&expires=1772355600&generation=igr_01ARZ3NDEKTSV4RRFFQ69G5FAV&signature=abc123def456" {
					t.Logf("artifact query %s", r.URL.RawQuery)
				}
				if op.Kind == "sse" {
					w.Header().Set("Content-Type", "text/event-stream")
					events := fixture[[]json.RawMessage](t, "events-valid.json")
					compact := new(bytes.Buffer)
					_ = json.Compact(compact, events[0])
					fmt.Fprintf(w, "data: %s\n\n", compact.Bytes())
					return
				}
				if op.Body == "multipart" {
					if err := r.ParseMultipartForm(1024); err != nil {
						t.Error(err)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.Status)
				if tc.Response != nil && string(tc.Response) != "null" {
					w.Write(tc.Response)
				} else if op.Kind == "binary" {
					w.Write([]byte("binary"))
				}
			})
			found := false
			for _, op := range inventory {
				if op.ID == tc.ID {
					found = true
					got := client.operations[tc.ID]
					if op.Method != got.Method || op.Path != got.Path || op.Authenticated != got.Authenticated {
						t.Fatalf("inventory differs %+v %+v", op, got)
					}
				}
			}
			if !found {
				t.Fatal("missing contract method")
			}
			method := reflect.ValueOf(client).MethodByName(tc.GoName)
			if !method.IsValid() {
				t.Fatal("missing named method", tc.GoName)
			}
			req := reflect.New(method.Type().In(1)).Elem()
			for key, value := range tc.Parameters {
				field := req.FieldByName(key)
				if !field.IsValid() {
					t.Fatalf("missing parameter %s", key)
				}
				if strings.HasPrefix(tc.ID, "streamChatEvents") && key == "ChatID" {
					value = "cht_01ARZ3NDEKTSV4RRFFQ69G5FAV"
				}
				raw, _ := json.Marshal(value)
				if field.Kind() == reflect.Struct && field.FieldByName("Set").IsValid() {
					field.FieldByName("Set").SetBool(true)
					if err := json.Unmarshal(raw, field.FieldByName("Value").Addr().Interface()); err != nil {
						t.Fatal(err)
					}
				} else if err := json.Unmarshal(raw, field.Addr().Interface()); err != nil {
					t.Fatal(err)
				}
			}
			if body := req.FieldByName("Body"); body.IsValid() {
				if body.Type() == reflect.TypeOf([]MultipartPart{}) {
					body.Set(reflect.ValueOf([]MultipartPart{{Name: "file", Filename: "fixture.txt", Data: []byte("hello")}}))
				} else if string(tc.Body) != "null" {
					if err := json.Unmarshal(tc.Body, body.Addr().Interface()); err != nil {
						t.Fatal(err)
					}
				}
			}
			results := method.Call([]reflect.Value{reflect.ValueOf(context.Background()), req})
			last := results[len(results)-1]
			if tc.Status == 409 {
				if last.IsNil() {
					t.Fatal("stub silently succeeded")
				}
			} else if !last.IsNil() {
				t.Fatal(last.Interface(), string(tc.Response))
			}
			if strings.HasPrefix(tc.ID, "streamChatEvents") && !results[0].IsNil() {
				stream := results[0].Interface().(*EventStream)
				defer stream.Close()
				event, err := stream.Next()
				if err != nil || event.Type != "ping" {
					t.Fatal(event, err)
				}
			}
			if client.operations[tc.ID].Kind == "binary" && !results[0].IsNil() {
				d := results[0].Interface().(*Download)
				b, err := io.ReadAll(d.Body)
				d.Close()
				if err != nil || string(b) != "binary" {
					t.Fatal(string(b), err)
				}
			}
			if calls != 1 {
				t.Fatal("unexpected calls", calls)
			}
		})
	}
}
func TestOptionalTriState(t *testing.T) {
	type model struct {
		Value Optional[string] `json:"value,omitzero"`
	}
	for _, tc := range []struct {
		v    model
		json string
	}{{model{}, `{}`}, {model{Present("x")}, `{"value":"x"}`}, {model{Null[string]()}, `{"value":null}`}} {
		b, err := json.Marshal(tc.v)
		if err != nil || string(b) != tc.json {
			t.Fatal(string(b), err)
		}
		var out model
		if err = json.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(out, tc.v) {
			t.Fatal(out, tc.v)
		}
	}
	v := Present("old")
	if err := json.Unmarshal([]byte("null"), &v); err != nil || v.Value != "" {
		t.Fatal(v, err)
	}
}
func TestHTTPPolicy(t *testing.T) {
	t.Run("retry_get_only", func(t *testing.T) {
		var calls atomic.Int32
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(503)
				return
			}
			io.WriteString(w, `{"data":[]}`)
		})
		c.config.RetryAttempts = 2
		if _, err := c.ListWorkspaces(context.Background(), ListWorkspacesParams{}); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != 2 {
			t.Fatal(calls.Load())
		}
		calls.Store(0)
		_, err := c.CreateWorkspace(context.Background(), CreateWorkspaceParams{Body: WorkspaceCreateRequest{Name: "hello"}})
		var api *APIError
		if !errors.As(err, &api) || api.Status != 503 || calls.Load() != 1 {
			t.Fatal(err, calls.Load())
		}
	})
	t.Run("redirect_no_follow", func(t *testing.T) {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://example.invalid/secret", 302)
		})
		_, err := c.ListWorkspaces(context.Background(), ListWorkspacesParams{})
		var api *APIError
		if !errors.As(err, &api) || api.Status != 302 {
			t.Fatal(err)
		}
	})
	t.Run("structured_error", func(t *testing.T) {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-ID", "request1")
			w.Header().Set("Retry-After", "3")
			w.WriteHeader(429)
			io.WriteString(w, `{"error":{"type":"too_many_requests","code":"slow","message":"safe","details":{"limit":1}}}`)
		})
		_, err := c.GetCapabilities(context.Background(), GetCapabilitiesParams{})
		var api *APIError
		if !errors.As(err, &api) || api.Code != "slow" || api.Type != "too_many_requests" || api.RequestID != "request1" || api.RetryAfter != 3*time.Second || !strings.Contains(string(api.Details), "limit") {
			t.Fatal(err)
		}
	})
	t.Run("response_limit_and_shape", func(t *testing.T) {
		response := `{"data":{"name":"long"}}`
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, response) })
		c.config.MaxResponseBytes = 4
		if _, err := c.GetWorkspace(context.Background(), GetWorkspaceParams{WorkspaceID: "wsp"}); err == nil {
			t.Fatal("limit ignored")
		}
		c.config.MaxResponseBytes = 100
		response = `{}`
		if _, err := c.GetWorkspace(context.Background(), GetWorkspaceParams{WorkspaceID: "wsp"}); err == nil {
			t.Fatal("missing envelope")
		}
		response = `no json`
		if _, err := c.GetCapabilities(context.Background(), GetCapabilitiesParams{}); err == nil {
			t.Fatal("bad json")
		}
	})
	t.Run("cancellation", func(t *testing.T) {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() })
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		_, err := c.GetCapabilities(ctx, GetCapabilitiesParams{})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	})
}
func TestConfigurationAndEncoding(t *testing.T) {
	for _, cfg := range []Config{{BaseURL: "ftp://x/api/v1"}, {BaseURL: "https://u:p@x/api/v1"}, {BaseURL: "https://x/no"}, {BaseURL: "https://x/api/v1?q=x"}, {BaseURL: "https://x", RetryAttempts: -1}, {BaseURL: "https://x", MaxResponseBytes: -1}, {BaseURL: "https://x", APIKey: "x\ny"}, {BaseURL: "https://x", JWKSURL: "http://x/keys"}} {
		if _, err := New(cfg); err == nil {
			t.Fatalf("accepted %+v", cfg)
		}
	}
	c, err := New(Config{BaseURL: "https://api.test", JWKSURL: "https://control.test/keys"})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.OperationIDs()) != 71 {
		t.Fatal("inventory")
	}
	if _, err = c.requestURL(c.operations["getUserTool"], map[string]any{"name": "../x"}); err == nil {
		t.Fatal("traversal")
	}
	if _, err = c.requestURL(c.operations["getWorkspace"], nil); err == nil {
		t.Fatal("missing")
	}
	if _, err = c.requestURL(c.operations["getWorkspace"], map[string]any{"workspace_id": "id", "extra": 1}); err == nil {
		t.Fatal("unknown")
	}
	u, err := c.requestURL(c.operations["getUserTool"], map[string]any{"name": "a b%"})
	if err != nil || !strings.HasSuffix(u, "a%20b%25") {
		t.Fatal(u, err)
	}
	if _, err = c.GetWorkspace(context.Background(), GetWorkspaceParams{WorkspaceID: "id"}); err == nil {
		t.Fatal("missing auth")
	}
	if retryAfter("invalid") != 0 || retryAfter("-1") != 0 || retryAfter(time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)) <= 0 {
		t.Fatal("retry after")
	}
	if _, _, err = c.makeBody(operation{Body: "multipart"}, []MultipartPart{{Name: "x\ny"}}); err == nil {
		t.Fatal("multipart injection")
	}
	_, _, err = c.makeBody(operation{}, make(chan int))
	if err == nil {
		t.Fatal("invalid json")
	}
	if _, err := NewSSEDecoder(bytes.NewReader(nil), 1).Next(); !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
}

func TestSecretRepresentationsAndObservation(t *testing.T) {
	var observed RequestObservation
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, testCapabilities) })
	c.config.Observer = func(o RequestObservation) { observed = o }
	for _, format := range []string{"%v", "%+v", "%#v"} {
		for _, value := range []any{c, c.config, &c.config} {
			text := fmt.Sprintf(format, value)
			if strings.Contains(text, "fixture-key") {
				t.Fatal("credential leak", text)
			}
		}
	}
	if _, err := c.GetCapabilities(context.Background(), GetCapabilitiesParams{}); err != nil {
		t.Fatal(err)
	}
	if observed.Operation != "getCapabilities" || observed.Status != 200 {
		t.Fatal(observed)
	}
}

const testCapabilities = `{"protocol_version":3,"contract_version":3,"supported_contract_versions":[2,3],"legacy_v2_mode":"historical_reads_replay_only","families":[],"event_types":[],"tool_kinds":[],"features":[]}`

func writeCapabilities(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(ContractHeader, "3")
	io.WriteString(w, testCapabilities)
}
func negotiatedTestClient(t *testing.T, fn http.HandlerFunc) *Client {
	return testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			writeCapabilities(w)
			return
		}
		fn(w, r)
	})
}
