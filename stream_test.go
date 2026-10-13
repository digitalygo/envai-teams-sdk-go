package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type byteReader struct{ io.Reader }

func (r byteReader) Read(b []byte) (int, error) {
	if len(b) > 1 {
		b = b[:1]
	}
	return r.Reader.Read(b)
}
func writeEvent(w io.Writer, e Event) { b, _ := json.Marshal(e); fmt.Fprintf(w, "data: %s\n\n", b) }
func TestSharedEvents(t *testing.T) {
	for _, raw := range fixture[[]json.RawMessage](t, "events-valid.json") {
		e, err := DecodeEvent(raw)
		if err != nil {
			t.Fatal(string(raw), err)
		}
		if !bytes.Equal(e.Raw, raw) {
			t.Fatal("raw not retained")
		}
		if e.Type == "future.additive" && string(e.Data) == "" {
			t.Fatal("lost unknown payload")
		}
	}
	for _, tc := range fixture[[]struct {
		Name  string          `json:"name"`
		Event json.RawMessage `json:"event"`
	}](t, "events-invalid.json") {
		if _, err := DecodeEvent(tc.Event); err == nil {
			t.Fatal("accepted", tc.Name)
		}
	}
	for _, b := range []string{`{`, `null`, `{"v":2}`} {
		if _, err := DecodeEvent([]byte(b)); err == nil {
			t.Fatal("accepted invalid json")
		}
	}
	e := fixture[[]Event](t, "events-valid.json")[1]
	e.TS = "no timestamp"
	b, _ := json.Marshal(e)
	if _, err := DecodeEvent(b); err == nil {
		t.Fatal("date format ignored")
	}
}
func TestSSESharedFrames(t *testing.T) {
	cases := fixture[[]struct {
		Name string   `json:"name"`
		Wire string   `json:"wire"`
		Data []string `json:"data"`
	}](t, "sse-cases.json")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			d := NewSSEDecoder(byteReader{strings.NewReader(tc.Wire)}, 0)
			var got []string
			for {
				value, err := d.Next()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, value)
			}
			if len(got) != len(tc.Data) {
				t.Fatal(got, tc.Data)
			}
			for i := range got {
				if got[i] != tc.Data[i] {
					t.Fatal(got, tc.Data)
				}
			}
		})
	}
	d := NewSSEDecoder(strings.NewReader("data: a\r\rdata: b\r\r"), 100)
	for _, want := range []string{"a", "b"} {
		got, err := d.Next()
		if err != nil || got != want {
			t.Fatal(got, err)
		}
	}
	d = NewSSEDecoder(strings.NewReader("data: "+strings.Repeat("x", 100)+"\n\n"), 20)
	if _, err := d.Next(); err == nil {
		t.Fatal("unbounded line")
	}
	d = NewSSEDecoder(strings.NewReader(strings.Repeat("data: abc\n", 10)+"\n"), 40)
	if _, err := d.Next(); err == nil {
		t.Fatal("unbounded frame")
	}
}
func TestStreamReplayAndFailures(t *testing.T) {
	ev := fixture[[]Event](t, "events-valid.json")
	chat := ev[0].ChatID
	t.Run("replay_dedup_ping", func(t *testing.T) {
		var n atomic.Int32
		c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			if n.Add(1) == 1 {
				writeEvent(w, ev[0])
				writeEvent(w, ev[1])
				writeEvent(w, ev[3])
			} else {
				if r.URL.Query().Get("after") != "1" {
					t.Error(r.URL.RawQuery)
				}
				writeEvent(w, ev[1])
				writeEvent(w, ev[2])
				writeEvent(w, ev[3])
			}
		})
		s, err := c.Events(context.Background(), chat, StreamOptions{MaxReconnects: 1, ReconnectDelay: time.Millisecond})
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		for _, seq := range []int64{0, 1, 2, 3} {
			e, err := s.Next()
			if err != nil || e.Seq != seq {
				t.Fatal(e, err)
			}
			if s.Cursor() != seq && seq != 0 {
				t.Fatal("cursor")
			}
		}
		if n.Load() != 2 {
			t.Fatal(n.Load())
		}
	})
	t.Run("unrecoverable_gap", func(t *testing.T) {
		c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			writeEvent(w, ev[3])
		})
		s, err := c.Events(context.Background(), chat, StreamOptions{MaxReconnects: 1, ReconnectDelay: time.Millisecond})
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		_, err = s.Next()
		var gap *ReplayGapError
		if !errors.As(err, &gap) || gap.From != 1 || gap.To != 2 {
			t.Fatal(err)
		}
	})
	t.Run("wrong_chat", func(t *testing.T) {
		c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			writeEvent(w, ev[0])
		})
		s, err := c.Events(context.Background(), "cht_01ARZ3NDEKTSV4RRFFQ69G5FA0", StreamOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		if _, err = s.Next(); err == nil {
			t.Fatal("accepted foreign event")
		}
	})
	t.Run("bad_content_type", func(t *testing.T) {
		c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "html") })
		if _, err := c.Events(context.Background(), chat, StreamOptions{}); err == nil {
			t.Fatal("html accepted")
		}
	})
	t.Run("close_cancels_read", func(t *testing.T) {
		c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		})
		s, err := c.Events(context.Background(), chat, StreamOptions{})
		if err != nil {
			t.Fatal(err)
		}
		go func() { time.Sleep(10 * time.Millisecond); s.Close() }()
		if _, err = s.Next(); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if err = s.Close(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("invalid_options", func(t *testing.T) {
		c := negotiatedTestClient(t, func(http.ResponseWriter, *http.Request) {})
		for _, o := range []StreamOptions{{After: -1}, {MaxReconnects: -1}, {ConnectTimeout: -1}} {
			if _, err := c.Events(context.Background(), chat, o); err == nil {
				t.Fatal(o)
			}
		}
	})
}
func TestRunTurnWaitsLiveAndRequiresTerminal(t *testing.T) {
	ev := fixture[[]Event](t, "events-valid.json")
	chat := ev[0].ChatID
	for _, terminal := range []bool{true, false} {
		t.Run(fmt.Sprint(terminal), func(t *testing.T) {
			sent := make(chan struct{})
			var live atomic.Bool
			c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/messages") {
					if !live.Load() {
						t.Error("message preceded live ping")
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(202)
					fmt.Fprintf(w, `{"status":"accepted","turn_id":%q}`, *ev[1].TurnID)
					close(sent)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				live.Store(true)
				writeEvent(w, ev[0])
				w.(http.Flusher).Flush()
				select {
				case <-sent:
				case <-r.Context().Done():
					return
				}
				writeEvent(w, ev[1])
				text := ev[3]
				text.Seq = 2
				writeEvent(w, text)
				if terminal {
					end := ev[7]
					end.Seq = 3
					writeEvent(w, end)
				}
			})
			result, err := c.RunTurn(context.Background(), chat, "hello", TurnOptions{})
			if terminal {
				if err != nil || result.OrchestratorText != "hello" || !result.Terminal.Terminal() || result.LastSeq != 3 {
					t.Fatal(result, err)
				}
			} else if err == nil {
				t.Fatal("incomplete turn silently completed")
			}
		})
	}
}
func TestRunTurnHandshakeFailureDoesNotSend(t *testing.T) {
	ev := fixture[[]Event](t, "events-valid.json")
	calls := 0
	c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "text/event-stream")
	})
	_, err := c.RunTurn(context.Background(), ev[0].ChatID, "hello", TurnOptions{LiveTimeout: time.Millisecond})
	if err == nil || calls != 1 {
		t.Fatal(err, calls)
	}
}

func TestTypedEventPayloads(t *testing.T) {
	raw, _ := contracts.ReadFile("contract/events.json")
	var schema struct {
		Defs map[string]json.RawMessage `json:"$defs"`
	}
	json.Unmarshal(raw, &schema)
	for name := range schema.Defs {
		if !strings.Contains(name, ".") && name != "ping" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			e := Event{Type: name, Data: json.RawMessage(`{}`)}
			payload, err := e.TypedPayload()
			if err != nil || payload == nil {
				t.Fatal(payload, err)
			}
			e.Data = []byte("invalid")
			if _, err = e.TypedPayload(); err == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
	e := Event{Type: "future.event", Data: json.RawMessage(`{"future":true}`)}
	payload, err := e.TypedPayload()
	if err != nil || string(payload.(json.RawMessage)) != string(e.Data) {
		t.Fatal(payload, err)
	}
	p, err := Payload[map[string]bool](e)
	if err != nil || !p["future"] {
		t.Fatal(p, err)
	}
}
func TestConcurrentStreamReaderRejectedAndLiveCancellationStable(t *testing.T) {
	ev := fixture[[]Event](t, "events-valid.json")
	c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeEvent(w, ev[0])
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	})
	for range 25 {
		s, err := c.Events(context.Background(), ev[0].ChatID, StreamOptions{})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		if err = s.WaitLive(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
		if s.ctx.Err() != nil {
			t.Fatal("live timeout watcher closed successful stream")
		}
		s.Close()
	}
	s, err := c.Events(context.Background(), ev[0].ChatID, StreamOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.readerMu.Lock()
	if _, err = s.Next(); err == nil {
		t.Fatal("concurrent reader accepted")
	}
	s.readerMu.Unlock()
}

func TestRunTurnWithSignedToolEndToEnd(t *testing.T) {
	ev := fixture[[]Event](t, "events-valid.json")
	signatures := fixture[signatureFixtures](t, "signatures.json")
	message := make(chan struct{})
	output := make(chan struct{})
	var executed atomic.Int32
	c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/.well-known/jwks.json":
			json.NewEncoder(w).Encode(signatures.JWKS)
		case strings.HasSuffix(r.URL.Path, "/messages"):
			w.WriteHeader(202)
			fmt.Fprintf(w, `{"status":"accepted","turn_id":%q}`, *ev[1].TurnID)
			close(message)
		case strings.HasSuffix(r.URL.Path, "/tool-output"):
			var result ToolOutputRequest
			if err := json.NewDecoder(r.Body).Decode(&result); err != nil || result.Output != "ok" {
				t.Error(result, err)
			}
			w.WriteHeader(204)
			close(output)
		default:
			w.Header().Set("Content-Type", "text/event-stream")
			writeEvent(w, ev[0])
			w.(http.Flusher).Flush()
			select {
			case <-message:
			case <-r.Context().Done():
				return
			}
			writeEvent(w, ev[1])
			request := ev[4]
			request.Seq = 2
			writeEvent(w, request)
			w.(http.Flusher).Flush()
			select {
			case <-output:
			case <-r.Context().Done():
				return
			}
			result := ev[5]
			result.Seq = 3
			writeEvent(w, result)
			terminal := ev[7]
			terminal.Seq = 4
			writeEvent(w, terminal)
		}
	})
	registry, err := NewRegistry(Tool{Name: "fixture_echo", Parameters: json.RawMessage(`{"type":"object"}`), Handler: func(context.Context, ToolInvocation) (ToolResult, error) {
		executed.Add(1)
		return ToolResult{Output: "ok"}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewToolRunner(c, registry, NewMemoryCallStore(), RunnerOptions{Namespace: "account"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := c.RunTurn(ctx, ev[0].ChatID, "hello", TurnOptions{Tools: runner})
	if err != nil || result.Terminal.Type != "turn.completed" || executed.Load() != 1 {
		t.Fatal(result, err, executed.Load())
	}
}
