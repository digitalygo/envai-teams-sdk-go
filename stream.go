package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type StreamOptions struct {
	After          int64
	MaxReconnects  int
	ReconnectDelay time.Duration
	MaxEventBytes  int
	ConnectTimeout time.Duration
	// MaxLegacySnapshotBytes bounds a complete finite v2 replay before any yield.
	MaxLegacySnapshotBytes int64
}
type ReplayGapError struct {
	From int64
	To   int64
}

func (e *ReplayGapError) Error() string { return fmt.Sprintf("envai: replay gap %d..%d", e.From, e.To) }

type EventStream struct {
	client       *Client
	chatID       string
	options      StreamOptions
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	body         io.ReadCloser
	decoder      *SSEDecoder
	cursor       atomic.Int64
	readerMu     sync.Mutex
	reconnects   int
	recovering   bool
	closed       bool
	legacyEvents []Event
	legacyIndex  int
}

func (c *Client) Events(ctx context.Context, chatID string, options StreamOptions) (*EventStream, error) {
	if options.After < 0 || options.MaxReconnects < 0 || options.ReconnectDelay < 0 || options.MaxLegacySnapshotBytes < 0 || options.MaxLegacySnapshotBytes == int64(^uint64(0)>>1) {
		return nil, errors.New("envai: invalid stream options")
	}
	if options.ReconnectDelay == 0 {
		options.ReconnectDelay = 250 * time.Millisecond
	}
	if options.ConnectTimeout == 0 {
		options.ConnectTimeout = 10 * time.Second
	}
	if options.ConnectTimeout < 0 {
		return nil, errors.New("envai: invalid stream connect timeout")
	}
	ctx, cancel := context.WithCancel(ctx)
	s := &EventStream{client: c, chatID: chatID, options: options, ctx: ctx, cancel: cancel}
	s.cursor.Store(options.After)
	if err := s.connect(); err != nil {
		cancel()
		return nil, err
	}
	return s, nil
}
func (s *EventStream) connect() error {
	timer := time.AfterFunc(s.options.ConnectTimeout, s.cancel)
	defer timer.Stop()
	resp, err := s.client.do(s.ctx, "streamChatEvents", map[string]any{"chat_id": s.chatID, "after": s.Cursor()}, nil, "text/event-stream")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToLower(resp.Header.Get("Content-Type")), "text/event-stream") {
		resp.Body.Close()
		return &ProtocolError{"stream response is not text/event-stream"}
	}
	if s.client.config.LegacyReadOnly {
		timer.Stop()
		return s.bufferLegacy(resp.Body)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		resp.Body.Close()
		return context.Canceled
	}
	s.body = resp.Body
	s.decoder = NewSSEDecoder(resp.Body, s.options.MaxEventBytes)
	return nil
}
func (s *EventStream) Cursor() int64 { return s.cursor.Load() }
func (s *EventStream) Close() error {
	s.cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.body != nil {
		return s.body.Close()
	}
	return nil
}
func (s *EventStream) reconnect() error {
	if s.reconnects >= s.options.MaxReconnects {
		return errors.New("envai: stream disconnected; reconnect budget exhausted")
	}
	s.reconnects++
	s.mu.Lock()
	if s.body != nil {
		s.body.Close()
	}
	s.mu.Unlock()
	if err := wait(s.ctx, s.options.ReconnectDelay*time.Duration(s.reconnects)); err != nil {
		return err
	}
	return s.connect()
}
func (s *EventStream) Next() (Event, error) {
	if !s.readerMu.TryLock() {
		return Event{}, errors.New("envai: stream permits only one reader")
	}
	defer s.readerMu.Unlock()
	if s.client.config.LegacyReadOnly {
		if err := s.ctx.Err(); err != nil {
			return Event{}, err
		}
		if s.legacyIndex >= len(s.legacyEvents) {
			return Event{}, io.EOF
		}
		event := s.legacyEvents[s.legacyIndex]
		s.legacyIndex++
		if event.Type != "ping" {
			s.cursor.Store(event.Seq)
		}
		return event, nil
	}
	for {
		if err := s.ctx.Err(); err != nil {
			return Event{}, err
		}
		data, err := s.decoder.Next()
		if err != nil {
			if e := s.ctx.Err(); e != nil {
				return Event{}, e
			}
			var p *ProtocolError
			if errors.As(err, &p) {
				return Event{}, err
			}
			if err = s.reconnect(); err != nil {
				return Event{}, err
			}
			continue
		}
		e, err := DecodeEvent([]byte(data))
		if err != nil {
			return Event{}, err
		}
		if e.ChatID != s.chatID {
			return Event{}, &ProtocolError{"event chat does not match stream"}
		}
		if e.Type == "ping" {
			return e, nil
		}
		if e.Seq <= s.Cursor() {
			continue
		}
		if e.Seq > s.Cursor()+1 {
			if s.recovering {
				return Event{}, &ReplayGapError{From: s.Cursor() + 1, To: e.Seq - 1}
			}
			s.recovering = true
			if err = s.reconnect(); err != nil {
				return Event{}, &ReplayGapError{From: s.Cursor() + 1, To: e.Seq - 1}
			}
			continue
		}
		s.recovering = false
		s.cursor.Store(e.Seq)
		return e, nil
	}
}
func (s *EventStream) WaitLive(ctx context.Context) error {
	done := make(chan struct{})
	stopped := make(chan struct{})
	defer func() { close(done); <-stopped }()
	go func() {
		defer close(stopped)
		select {
		case <-ctx.Done():
			s.Close()
		case <-done:
		}
	}()
	for {
		e, err := s.Next()
		if err != nil {
			return err
		}
		if e.Type == "ping" {
			return nil
		}
	}
}

type TurnResult struct {
	ChatID           string
	TurnID           string
	LastSeq          int64
	OrchestratorText string
	Terminal         Event
}
type TurnOptions struct {
	Stream      StreamOptions
	LiveTimeout time.Duration
	OnEvent     func(Event) error
	Tools       *ToolRunner
}

func (c *Client) RunTurn(ctx context.Context, chatID, message string, options TurnOptions) (TurnResult, error) {
	if c.config.LegacyReadOnly {
		return TurnResult{}, &CompatibilityError{Reason: "legacy mode is read-only"}
	}
	s, err := c.Events(ctx, chatID, options.Stream)
	if err != nil {
		return TurnResult{}, err
	}
	defer s.Close()
	timeout := options.LiveTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	live, cancel := context.WithTimeout(ctx, timeout)
	err = s.WaitLive(live)
	cancel()
	if err != nil {
		return TurnResult{}, err
	}
	accepted, err := c.SendChatMessage(ctx, SendChatMessageParams{ChatID: chatID, Body: ChatMessageRequest{Message: message}})
	if err != nil {
		return TurnResult{}, err
	}
	result := TurnResult{ChatID: chatID, TurnID: accepted.TurnID}
	var text strings.Builder
	toolCtx, cancelTools := context.WithCancel(ctx)
	var work sync.WaitGroup
	defer func() { cancelTools(); work.Wait() }()
	toolErrors := make(chan error, 1)
	for {
		e, err := s.Next()
		result.LastSeq = s.Cursor()
		if err != nil {
			select {
			case toolErr := <-toolErrors:
				return result, toolErr
			default:
			}
			return result, err
		}
		if e.TurnID == nil || *e.TurnID != accepted.TurnID {
			continue
		}
		if options.Tools != nil {
			if err = options.Tools.ObserveContext(ctx, e); err != nil {
				return result, err
			}
			if e.Type == "agent.tool.request" {
				work.Add(1)
				go func(e Event) {
					defer work.Done()
					if toolErr := options.Tools.Handle(toolCtx, e, chatID); toolErr != nil && toolCtx.Err() == nil && !errors.Is(toolErr, context.Canceled) {
						select {
						case toolErrors <- toolErr:
						default:
						}
						s.Close()
					}
				}(e)
			}
		}
		if options.OnEvent != nil {
			if err = options.OnEvent(e); err != nil {
				return result, err
			}
		}
		if e.Agent != nil && e.Agent.ParentID == nil && e.Agent.Role == "orchestrator" {
			var data struct {
				Delta string `json:"delta"`
				Text  string `json:"text"`
			}
			_ = json.Unmarshal(e.Data, &data)
			if e.Type == "agent.text.delta" {
				text.WriteString(data.Delta)
			}
			if e.Type == "agent.text.done" {
				text.Reset()
				text.WriteString(data.Text)
			}
		}
		result.OrchestratorText = text.String()
		if e.Terminal() {
			result.Terminal = e
			cancelTools()
			return result, nil
		}
	}
}

// LegacyEvents reads a finite historical numeric-v2 snapshot without live reconnects.
func (c *Client) LegacyEvents(ctx context.Context, chatID string, options StreamOptions) (*EventStream, error) {
	clone := *c
	clone.config.LegacyReadOnly = true
	return clone.Events(ctx, chatID, options)
}

// bufferLegacy validates the entire finite snapshot before making a prefix
// visible. Unlike v3, legacy reads cannot reconnect or follow a live stream.
func (s *EventStream) bufferLegacy(body io.ReadCloser) error {
	defer body.Close()
	deadline := time.AfterFunc(s.client.config.RequestTimeout, s.cancel)
	defer deadline.Stop()
	limit := s.options.MaxLegacySnapshotBytes
	if limit == 0 {
		limit = 64 << 20
	}
	raw, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return err
	}
	if int64(len(raw)) > limit {
		return &ProtocolError{"legacy replay exceeds configured snapshot limit"}
	}
	decoder := NewSSEDecoder(bytes.NewReader(raw), s.options.MaxEventBytes)
	decoder.strictEOF = true
	cursor := s.Cursor()
	var events []Event
	for {
		data, err := decoder.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		event, err := DecodeEvent([]byte(data))
		if err != nil {
			return err
		}
		if event.V != 2 || event.ChatID != s.chatID {
			return &ProtocolError{"legacy snapshot version or chat mismatch"}
		}
		if event.Type != "ping" {
			if event.Seq <= cursor {
				return &ProtocolError{"legacy replay sequence must strictly increase beyond the requested cursor"}
			}
			if event.Seq != cursor+1 {
				return &ReplayGapError{From: cursor + 1, To: event.Seq - 1}
			}
			cursor = event.Seq
		}
		events = append(events, event)
	}
	if err := s.ctx.Err(); err != nil {
		return err
	}
	s.legacyEvents = events
	return nil
}
