package envai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
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

var eventSchemaOnce sync.Once
var eventSchema *jsonschema.Schema
var eventSchemaErr error

func DecodeEvent(b []byte) (Event, error) {
	if !utf8.Valid(b) {
		return Event{}, &ProtocolError{"event bytes are not valid UTF-8"}
	}
	eventSchemaOnce.Do(func() {
		raw, _ := contracts.ReadFile("contract/events.json")
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
		if err != nil {
			eventSchemaErr = err
			return
		}
		c := jsonschema.NewCompiler()
		c.AssertFormat()
		if err = c.AddResource("https://sdk.invalid/events.json", doc); err != nil {
			eventSchemaErr = err
			return
		}
		eventSchema, eventSchemaErr = c.Compile("https://sdk.invalid/events.json#/$defs/Event")
	})
	if eventSchemaErr != nil {
		return Event{}, eventSchemaErr
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return Event{}, &ProtocolError{"invalid event JSON"}
	}
	if err = eventSchema.Validate(doc); err != nil {
		return Event{}, &ProtocolError{"event does not conform to ATP v2"}
	}
	var e Event
	if err = json.Unmarshal(b, &e); err != nil {
		return e, &ProtocolError{"invalid event envelope"}
	}
	e.Raw = append(json.RawMessage(nil), b...)
	return e, nil
}

type SSEDecoder struct {
	reader *bufio.Reader
	limit  int
	first  bool
	skipLF bool
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
