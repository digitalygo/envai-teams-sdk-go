package envai

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

type EventAgent struct {
	InstanceID string  `json:"instance_id"`
	Name       string  `json:"name"`
	Role       string  `json:"role"`
	ParentID   *string `json:"parent_id"`
	Model      string  `json:"model,omitempty"`
}
type Event struct {
	V      int             `json:"v"`
	ID     string          `json:"id"`
	Seq    int64           `json:"seq"`
	TS     string          `json:"ts"`
	ChatID string          `json:"chat_id"`
	TurnID *string         `json:"turn_id"`
	Agent  *EventAgent     `json:"agent"`
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data"`
	Raw    json.RawMessage `json:"-"`
}

func (e Event) Terminal() bool {
	return e.Type == "turn.completed" || e.Type == "turn.cancelled" || e.Type == "turn.failed"
}
func Payload[T any](e Event) (T, error) { var v T; err := json.Unmarshal(e.Data, &v); return v, err }

// DecodeEvent validates before hydration. A v3 stream can replay untouched numeric v2
// records, but null-bearing v2 events always fail the immutable legacy schema.
func DecodeEvent(b []byte) (Event, error) {
	if !utf8.Valid(b) {
		return Event{}, &ProtocolError{"event bytes are not valid UTF-8"}
	}
	var envelope struct {
		V    int             `json:"v"`
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(b, &envelope) != nil {
		return Event{}, &ProtocolError{"invalid event JSON"}
	}
	file := "events.json"
	if envelope.V == 2 {
		file = "events-v2.json"
	} else if envelope.V != 3 {
		return Event{}, &ProtocolError{"unsupported event version"}
	}
	if err := validateDefinition(file, "Event", b); err != nil {
		return Event{}, err
	}
	switch envelope.Type {
	case "turn.completed", "turn.cancelled", "turn.failed", "agent.tool.complete", "agent.done":
		if err := validateCostFields(envelope.Data, envelope.V == 3); err != nil {
			return Event{}, err
		}
	}
	var event Event
	if err := json.Unmarshal(b, &event); err != nil {
		return Event{}, &ProtocolError{"invalid event envelope"}
	}
	event.Raw = append(json.RawMessage(nil), b...)
	return event, nil
}

type SSEDecoder struct {
	reader    *bufio.Reader
	limit     int
	first     bool
	skipLF    bool
	strictEOF bool
}

func NewSSEDecoder(r io.Reader, maxBytes int) *SSEDecoder {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	return &SSEDecoder{reader: bufio.NewReader(r), limit: maxBytes, first: true}
}
func (d *SSEDecoder) Next() (string, error) {
	var data []string
	size := 0
	for {
		line, err := d.line()
		if err != nil {
			if d.strictEOF && err == io.EOF && size > 0 {
				return "", io.ErrUnexpectedEOF
			}
			return "", err
		}
		size += len(line) + 1
		if size > d.limit {
			return "", &ProtocolError{"SSE frame exceeds configured limit"}
		}
		if d.first {
			line = strings.TrimPrefix(line, "\xef\xbb\xbf")
			d.first = false
		}
		if line == "" {
			if len(data) > 0 {
				return strings.Join(data, "\n"), nil
			}
			size = 0
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		if field == "data" {
			data = append(data, value)
		}
	}
}
func (d *SSEDecoder) line() (string, error) {
	var b []byte
	for {
		c, err := d.reader.ReadByte()
		if err != nil {
			if d.strictEOF && err == io.EOF && len(b) > 0 {
				return "", io.ErrUnexpectedEOF
			}
			return "", err
		}
		if d.skipLF {
			d.skipLF = false
			if c == '\n' {
				continue
			}
		}
		if c == '\r' {
			d.skipLF = true
			return string(b), nil
		}
		if c == '\n' {
			return string(b), nil
		}
		b = append(b, c)
		if len(b) > d.limit {
			return "", &ProtocolError{"SSE line exceeds configured limit"}
		}
	}
}
