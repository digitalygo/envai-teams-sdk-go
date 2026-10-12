package envai

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Signature struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Sig string `json:"sig"`
}
type ToolRequest struct {
	CallID    string          `json:"call_id"`
	ToolName  string          `json:"tool_name"`
	Kind      string          `json:"kind"`
	Args      json.RawMessage `json:"args"`
	Signature Signature       `json:"signature"`
}
type ToolInvocation struct {
	ChatID    string
	TurnID    string
	CallID    string
	Arguments json.RawMessage
}
type ToolResult struct {
	Output  string
	IsError bool
}

func JSONResult(value any) (ToolResult, error) {
	b, err := json.Marshal(value)
	return ToolResult{Output: string(b)}, err
}

type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Timeout     time.Duration
	Handler     func(context.Context, ToolInvocation) (ToolResult, error)
	schema      *jsonschema.Schema
}
type Registry struct{ tools map[string]Tool }

var toolNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,254}$`)

func NewRegistry(definitions ...Tool) (*Registry, error) {
	r := &Registry{tools: map[string]Tool{}}
	for _, t := range definitions {
		if !toolNamePattern.MatchString(t.Name) || t.Handler == nil {
			return nil, errors.New("envai: invalid tool definition")
		}
		if _, ok := r.tools[t.Name]; ok {
			return nil, errors.New("envai: duplicate local tool")
		}
		if t.Timeout == 0 {
			t.Timeout = 120 * time.Second
		}
		if t.Timeout < time.Second || t.Timeout > 600*time.Second {
			return nil, errors.New("envai: tool timeout must be 1..600 seconds")
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(t.Parameters))
		if err != nil {
			return nil, errors.New("envai: invalid tool schema")
		}
		root, ok := doc.(map[string]any)
		if !ok || root["type"] != "object" {
			return nil, errors.New("envai: tool schema root must be object")
		}
		compiler := jsonschema.NewCompiler()
		compiler.AssertFormat()
		compiler.UseLoader(denySchemaLoader{})
		if err = compiler.AddResource("https://sdk.invalid/tool.json", doc); err != nil {
			return nil, err
		}
		t.schema, err = compiler.Compile("https://sdk.invalid/tool.json")
		if err != nil {
			return nil, errors.New("envai: unsupported or invalid local tool schema")
		}
		t.Parameters = append(json.RawMessage(nil), t.Parameters...)
		r.tools[t.Name] = t
	}
	return r, nil
}

type denySchemaLoader struct{}

func (denySchemaLoader) Load(string) (any, error) {
	return nil, errors.New("envai: external schema references are disabled")
}
func (r *Registry) Definitions() []UserToolCreateRequest {
	out := make([]UserToolCreateRequest, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, UserToolCreateRequest{Name: t.Name, Description: Present(t.Description), Parameters: map[string]json.RawMessage{}, TimeoutSeconds: Present(int64(t.Timeout / time.Second))})
		_ = json.Unmarshal(t.Parameters, &out[len(out)-1].Parameters)
	}
	return out
}

type Verifier struct {
	Client  *Client
	TTL     time.Duration
	mu      sync.Mutex
	keys    map[string]JWK
	expires time.Time
}

func (v *Verifier) Verify(ctx context.Context, r ToolRequest) error {
	if r.Kind != "client" || r.Signature.Alg != "EdDSA" || r.Signature.Kid == "" {
		return &ProtocolError{"invalid client tool signature metadata"}
	}
	if !utf8.Valid(r.Args) || len(bytes.TrimSpace(r.Args)) == 0 || bytes.TrimSpace(r.Args)[0] != '{' {
		return &ProtocolError{"tool arguments must be an object"}
	}
	canonical, err := jsoncanonicalizer.Transform(r.Args)
	if err != nil {
		return &ProtocolError{"tool arguments cannot be canonicalized"}
	}
	sig, err := base64.RawURLEncoding.Strict().DecodeString(r.Signature.Sig)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return &ProtocolError{"invalid tool signature encoding"}
	}
	key, err := v.key(ctx, r.Signature.Kid)
	if err != nil {
		return err
	}
	pub, err := base64.RawURLEncoding.Strict().DecodeString(key.X)
	if err != nil || len(pub) != ed25519.PublicKeySize || key.Kty != "OKP" || key.Crv != "Ed25519" || key.Use != "sig" || (key.Alg != "" && key.Alg != "EdDSA") {
		return &ProtocolError{"invalid signing key"}
	}
	message := r.CallID + "\n" + r.ToolName + "\n" + string(canonical)
	if !ed25519.Verify(pub, []byte(message), sig) {
		return &ProtocolError{"client tool signature verification failed"}
	}
	return nil
}
func (v *Verifier) key(ctx context.Context, kid string) (JWK, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.Client == nil {
		return JWK{}, errors.New("envai: verifier has no client")
	}
	fresh := v.keys != nil && time.Now().Before(v.expires)
	if fresh {
		if key, ok := v.keys[kid]; ok {
			return key, nil
		}
	}
	doc, err := v.Client.GetJWKS(ctx, GetJWKSParams{})
	if err != nil {
		return JWK{}, err
	}
	keys := map[string]JWK{}
	for _, key := range doc.Keys {
		if _, ok := keys[key.Kid]; ok {
			return JWK{}, &ProtocolError{"ambiguous signing key id"}
		}
		keys[key.Kid] = key
	}
	v.keys = keys
	ttl := v.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	v.expires = time.Now().Add(ttl)
	if key, ok := keys[kid]; ok {
		return key, nil
	}
	return JWK{}, &ProtocolError{"unknown signing key id"}
}

type activeCall struct {
	turnID string
	cancel context.CancelFunc
}
type RunnerOptions struct {
	Namespace        string
	MaxConcurrent    int
	SubmissionMargin time.Duration
	Authorize        func(context.Context, ToolInvocation, string) error
}
type ToolRunner struct {
	client   *Client
	registry *Registry
	verifier *Verifier
	store    CallStore
	options  RunnerOptions
	slots    chan struct{}
	mu       sync.Mutex
	active   map[string]activeCall
	settled  map[string]CallState
}

func NewToolRunner(c *Client, r *Registry, store CallStore, options RunnerOptions) (*ToolRunner, error) {
	if c == nil || r == nil || store == nil || options.Namespace == "" {
		return nil, errors.New("envai: client, registry, call store and stable account namespace are required")
	}
	if options.MaxConcurrent == 0 {
		options.MaxConcurrent = 4
	}
	if options.MaxConcurrent < 1 || options.MaxConcurrent > 256 {
		return nil, errors.New("envai: invalid tool concurrency")
	}
	if options.SubmissionMargin == 0 {
		options.SubmissionMargin = time.Second
	}
	if options.SubmissionMargin < 0 {
		return nil, errors.New("envai: invalid submission margin")
	}
	origin := url.URL{Scheme: c.base.Scheme, Host: c.base.Host}
	options.Namespace = origin.String() + "\n" + options.Namespace
	return &ToolRunner{client: c, registry: r, store: store, verifier: &Verifier{Client: c}, options: options, slots: make(chan struct{}, options.MaxConcurrent), active: map[string]activeCall{}, settled: map[string]CallState{}}, nil
}
func (r *ToolRunner) Observe(e Event) error { return r.ObserveContext(context.Background(), e) }
func (r *ToolRunner) ObserveContext(ctx context.Context, e Event) error {
	wire, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	e, err = DecodeEvent(wire)
	if err != nil || e.TurnID == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if e.Terminal() {
		r.settled["turn\n"+e.ChatID+"\n"+*e.TurnID] = CallAmbiguous
		for key, call := range r.active {
			if len(key) > len(e.ChatID) && key[:len(e.ChatID)+1] == e.ChatID+"\n" && call.turnID == *e.TurnID {
				call.cancel()
			}
		}
		id := CallIdentity{Namespace: r.options.Namespace, ChatID: e.ChatID, TurnID: *e.TurnID, CallID: "terminal:" + *e.TurnID}
		if _, err := r.store.Claim(ctx, id); err != nil {
			return err
		}
		return r.store.Record(ctx, id, CallAmbiguous)
	}
	if e.Type == "agent.tool.result" {
		var out struct {
			CallID string `json:"call_id"`
		}
		if json.Unmarshal(e.Data, &out) == nil {
			r.settled["call\n"+e.ChatID+"\n"+*e.TurnID+"\n"+out.CallID] = CallResolved
			if call, ok := r.active[e.ChatID+"\n"+out.CallID]; ok && call.turnID == *e.TurnID {
				call.cancel()
			}
			id := CallIdentity{Namespace: r.options.Namespace, ChatID: e.ChatID, TurnID: *e.TurnID, CallID: out.CallID}
			if _, err := r.store.Claim(ctx, id); err != nil {
				return err
			}
			return r.store.Record(ctx, id, CallResolved)
		}
	}
	return nil
}
func (r *ToolRunner) Handle(ctx context.Context, event Event, chatID string) error {
	admitted := time.Now()
	if event.ChatID != chatID || event.Type != "agent.tool.request" || event.TurnID == nil {
		return &ProtocolError{"tool request does not match execution context"}
	}
	wire, err := json.Marshal(event)
	if err != nil {
		return err
	}
	event, err = DecodeEvent(wire)
	if err != nil {
		return err
	}
	var req ToolRequest
	if err = json.Unmarshal(event.Data, &req); err != nil {
		return err
	}
	if err = r.verifier.Verify(ctx, req); err != nil {
		return err
	}
	req.Args, err = jsoncanonicalizer.Transform(req.Args)
	if err != nil {
		return &ProtocolError{"verified tool arguments cannot be canonicalized"}
	}
	id := CallIdentity{Namespace: r.options.Namespace, ChatID: chatID, CallID: req.CallID, TurnID: *event.TurnID, ToolName: req.ToolName}
	barrier := CallIdentity{Namespace: r.options.Namespace, ChatID: chatID, CallID: "terminal:" + *event.TurnID}
	if _, exists, err := r.store.State(ctx, barrier); err != nil {
		return err
	} else if exists {
		return nil
	}
	claimed, err := r.store.Claim(ctx, id)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	controlled := func(code string) error {
		b, _ := json.Marshal(map[string]string{"code": code})
		return r.deliver(ctx, id, ToolResult{Output: string(b), IsError: true})
	}
	t, ok := r.registry.tools[req.ToolName]
	if !ok {
		return controlled("tool_not_found")
	}
	args, err := jsonschema.UnmarshalJSON(bytes.NewReader(req.Args))
	if err != nil || t.schema.Validate(args) != nil {
		return controlled("invalid_arguments")
	}
	invocation := ToolInvocation{ChatID: chatID, TurnID: *event.TurnID, CallID: req.CallID, Arguments: append(json.RawMessage(nil), req.Args...)}
	if r.options.Authorize != nil {
		if err = r.options.Authorize(ctx, invocation, req.ToolName); err != nil {
			return controlled("tool_not_authorized")
		}
	}
	select {
	case r.slots <- struct{}{}:
	default:
		return controlled("tool_busy")
	}
	timeout := t.Timeout - r.options.SubmissionMargin - time.Since(admitted)
	if timeout <= 0 {
		<-r.slots
		_ = r.store.Record(ctx, id, CallAmbiguous)
		return errors.New("envai: tool deadline leaves no execution window")
	}
	toolCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	activeKey := chatID + "\n" + req.CallID
	r.mu.Lock()
	settled := r.settled["call\n"+chatID+"\n"+*event.TurnID+"\n"+req.CallID]
	if settled == "" {
		settled = r.settled["turn\n"+chatID+"\n"+*event.TurnID]
	}
	if settled != "" {
		r.mu.Unlock()
		<-r.slots
		return r.store.Record(ctx, id, settled)
	}
	r.active[activeKey] = activeCall{turnID: *event.TurnID, cancel: cancel}
	r.mu.Unlock()
	defer func() { r.mu.Lock(); delete(r.active, activeKey); r.mu.Unlock() }()
	type execution struct {
		result ToolResult
		err    error
	}
	done := make(chan execution, 1)
	go func() {
		defer func() { <-r.slots }()
		out := execution{}
		defer func() {
			if recover() != nil {
				out.err = errors.New("handler panic")
			}
			done <- out
		}()
		if out.err = toolCtx.Err(); out.err != nil {
			return
		}
		out.result, out.err = t.Handler(toolCtx, invocation)
	}()
	select {
	case <-toolCtx.Done():
		_ = r.store.Record(context.WithoutCancel(ctx), id, CallAmbiguous)
		return toolCtx.Err()
	case out := <-done:
		if toolCtx.Err() != nil {
			_ = r.store.Record(context.WithoutCancel(ctx), id, CallAmbiguous)
			return toolCtx.Err()
		}
		if out.err != nil {
			return controlled("handler_error")
		}
		if !utf8.ValidString(out.result.Output) {
			return controlled("invalid_output")
		}
		if len(out.result.Output) > 131072 {
			return controlled("output_too_large")
		}
		if out.result.Output == "" {
			return controlled("empty_output")
		}
		return r.deliver(ctx, id, out.result)
	}
}
func (r *ToolRunner) deliver(ctx context.Context, id CallIdentity, result ToolResult) error {
	if err := r.store.Record(ctx, id, CallSubmitted); err != nil {
		return err
	}
	err := r.client.ResolveChatTool(ctx, ResolveChatToolParams{ChatID: id.ChatID, Body: ToolOutputRequest{CallID: id.CallID, Output: result.Output, IsError: Present(result.IsError)}})
	state := CallResolved
	if err != nil {
		state = CallAmbiguous
		var api *APIError
		if errors.As(err, &api) && api.Status == 409 && api.Code == "tool_already_resolved" {
			state = CallResolved
		}
	}
	if storeErr := r.store.Record(context.WithoutCancel(ctx), id, state); storeErr != nil {
		return fmt.Errorf("envai: output delivery state could not be recorded: %w", storeErr)
	}
	return err
}
