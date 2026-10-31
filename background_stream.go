package envai

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// NativeTaskEvent retains the signed payload exactly as supplied by the server.
type NativeTaskEvent struct {
	TaskID       string          `json:"task_id"`
	OriginTurnID string          `json:"origin_turn_id"`
	EventType    string          `json:"event_type"`
	EventPayload json.RawMessage `json:"event_payload"`
}
type NativeDeliveryEvent struct {
	EventID        string `json:"event_id"`
	OriginTurnID   string `json:"origin_turn_id"`
	TaskID         string `json:"task_id,omitempty"`
	DeliveryTurnID string `json:"delivery_turn_id,omitempty"`
}

// backgroundEvent validates the outer wire event first, then validates the inner
// event using its task identity as the local cancellation scope. This scope is
// never sent to the server and never replaces the signed call ID or arguments.
func backgroundEvent(e Event) (Event, error) {
	wire, err := json.Marshal(e)
	if err != nil {
		return Event{}, err
	}
	e, err = DecodeEvent(wire)
	if err != nil {
		return Event{}, err
	}
	switch e.Type {
	case "task.event":
		var task NativeTaskEvent
		if err = json.Unmarshal(e.Data, &task); err != nil {
			return Event{}, err
		}
		if task.TaskID == "" || task.OriginTurnID == "" || task.EventType == "" {
			return Event{}, &ProtocolError{"invalid task wrapper"}
		}
		scope := "task:" + task.TaskID
		e.TurnID = &task.OriginTurnID
		e.Type = task.EventType
		e.Data = task.EventPayload
		wire, err = json.Marshal(e)
		if err != nil {
			return Event{}, err
		}
		e, err = DecodeEvent(wire)
		if err != nil {
			return Event{}, err
		}
		e.TurnID = &scope
		e.taskID = task.TaskID
		e.originTurnID = task.OriginTurnID
	case "task.completed", "task.failed", "task.cancelled":
		var task NativeTask
		if err = json.Unmarshal(e.Data, &task); err != nil {
			return Event{}, err
		}
		if task.TaskID == "" {
			return Event{}, &ProtocolError{"missing task identity"}
		}
		scope := "task:" + task.TaskID
		e.TurnID = &scope
		e.Type = "turn.cancelled"
	}
	return e, nil
}

// HandleBackground accepts trusted live task wrappers and ordinary tool requests.
// The existing signed-tool verification and account/chat/call-ID store apply.
func (r *ToolRunner) HandleBackground(ctx context.Context, e Event, chatID string) error {
	if !r.client.config.NativeTasks || r.client.config.LegacyReadOnly {
		return &CompatibilityError{Reason: "background tool dispatch requires Config.NativeTasks"}
	}
	if e.ChatID != chatID {
		return &ProtocolError{"tool request does not match execution context"}
	}
	inner, err := backgroundEvent(e)
	if err != nil {
		return err
	}
	return r.handle(ctx, inner, chatID)
}
func (r *ToolRunner) ObserveBackground(ctx context.Context, e Event) error {
	inner, err := backgroundEvent(e)
	if err != nil {
		return err
	}
	return r.observeValidated(ctx, inner)
}

type BackgroundOptions struct {
	Stream      StreamOptions
	LiveTimeout time.Duration
	Tools       *ToolRunner
	OnEvent     func(Event) error
}

// BackgroundSession consumes new live work on one explicitly authorized execution chat until
// Close, context cancellation, or error. Foreground turn terminals do not close it.
type BackgroundSession struct {
	stream      *EventStream
	cancel      context.CancelFunc
	done        chan struct{}
	mu          sync.Mutex
	err         error
	replayTurns map[string]bool
}

func (c *Client) OpenBackgroundSession(ctx context.Context, chatID string, options BackgroundOptions) (*BackgroundSession, error) {
	if options.Tools != nil && (options.Tools.client != c || !options.Tools.client.config.NativeTasks) {
		return nil, &CompatibilityError{Reason: "background runner requires a native-enabled client identical to the session client"}
	}
	if err := c.requireFeatures(ctx, "detached_background"); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	stream, err := c.Events(ctx, chatID, options.Stream)
	if err != nil {
		cancel()
		return nil, err
	}
	timeout := options.LiveTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	live, stop := context.WithTimeout(ctx, timeout)
	replayTurns := map[string]bool{}
	replayBytes := 0
	err = stream.waitLive(live, func(e Event) error {
		replayBytes += len(e.Raw)
		if replayBytes > 16*1024*1024 {
			return &ProtocolError{"background replay exceeds reconciliation limit"}
		}
		if e.TurnID != nil {
			replayTurns[*e.TurnID] = true
		}
		if e.Type == "task.event" {
			var wrapper NativeTaskEvent
			if json.Unmarshal(e.Data, &wrapper) == nil {
				replayTurns[wrapper.OriginTurnID] = true
			}
		}

		if options.Tools != nil {
			return options.Tools.ObserveBackground(live, e)
		}
		return nil
	})
	stop()
	if err != nil {
		stream.Close()
		cancel()
		return nil, err
	}
	session := &BackgroundSession{stream: stream, cancel: cancel, done: make(chan struct{}), replayTurns: replayTurns}
	go func() {
		var workers sync.WaitGroup
		defer func() { cancel(); stream.Close(); workers.Wait(); close(session.done) }()
		fail := func(err error) {
			session.mu.Lock()
			if session.err == nil {
				session.err = err
			}
			session.mu.Unlock()
			cancel()
			stream.Close()
		}
		for {
			e, err := stream.Next()
			if err != nil {
				if ctx.Err() == nil {
					fail(err)
				}
				return
			}
			if options.Tools != nil {
				if err = options.Tools.ObserveBackground(ctx, e); err != nil {
					fail(err)
					return
				}
				tool := e.Type == "agent.tool.request"
				if e.Type == "task.event" {
					var wrapper NativeTaskEvent
					if json.Unmarshal(e.Data, &wrapper) == nil {
						tool = wrapper.EventType == "agent.tool.request"
					}
				}
				if tool {
					workers.Add(1)
					go func(e Event) {
						defer workers.Done()
						if err := options.Tools.HandleBackground(ctx, e, chatID); err != nil && ctx.Err() == nil && !errors.Is(err, context.Canceled) {
							fail(err)
						}
					}(e)
				}
			}
			if options.OnEvent != nil {
				if err = options.OnEvent(e); err != nil {
					fail(err)
					return
				}
			}
		}
	}()
	return session, nil
}
func (s *BackgroundSession) Close() error { s.cancel(); s.stream.Close(); return s.Wait() }
func (s *BackgroundSession) Wait() error  { <-s.done; s.mu.Lock(); defer s.mu.Unlock(); return s.err }

// RunDelegatedJob opens target tool execution before admitting the job and keeps
// it alive until the origin task reaches its final outcome. For phase jobs pass
// NotifyOrigin pointing to false. Cancellation of ctx does not cancel the job.
func (c *Client) RunDelegatedJob(ctx context.Context, originChatID string, request NativeDelegationRequest, options BackgroundOptions) (NativeTask, error) {
	if options.Stream.After != 0 {
		return NativeTask{}, &ProtocolError{"RunDelegatedJob requires a fresh stream; reconcile existing jobs separately"}
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	session, err := c.OpenBackgroundSession(ctx, request.TargetChatID, options)
	if err != nil {
		return NativeTask{}, err
	}
	defer session.Close()
	go func() { <-session.done; cancel() }()
	receipt, err := c.DelegateTeam(ctx, originChatID, request)
	if err != nil {
		return NativeTask{}, err
	}
	if receipt.TargetChatID != request.TargetChatID {
		return NativeTask{}, &ProtocolError{"delegation receipt target mismatch"}
	}
	if session.replayTurns[receipt.TurnID] {
		task, err := c.GetTask(ctx, originChatID, receipt.TaskID)
		if err != nil {
			return task, err
		}
		if task.TaskID != receipt.TaskID {
			return task, &ProtocolError{"task response identity mismatch"}
		}
		if task.Terminal() {
			return task, nil
		}
		return task, &ProtocolError{"existing delegated job requires reconciliation of pending tools; no historical tool executed"}
	}
	task, err := c.WaitTask(ctx, originChatID, receipt.TaskID, 0)
	if err != nil {
		session.mu.Lock()
		streamErr := session.err
		session.mu.Unlock()
		if streamErr != nil {
			return task, streamErr
		}
	}
	return task, err
}
