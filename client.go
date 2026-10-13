package envai

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//go:embed contract/*.json
var contracts embed.FS

type Config struct {
	// LegacyReadOnly selects strict ATP v2 historical reads; all mutations are rejected.
	LegacyReadOnly   bool
	BaseURL          string
	APIKey           string
	JWKSURL          string
	HTTPClient       *http.Client
	RequestTimeout   time.Duration
	MaxResponseBytes int64
	RetryAttempts    int
	RetryDelay       time.Duration
	AllowHTTP        bool
	Observer         func(RequestObservation)
}

type RequestObservation struct {
	Operation string
	Status    int
	Attempt   int
	Duration  time.Duration
}
type operation struct {
	ID            string      `json:"id"`
	Method        string      `json:"method"`
	Path          string      `json:"path"`
	Authenticated bool        `json:"authenticated"`
	Kind          string      `json:"kind"`
	Body          string      `json:"body"`
	Unwrap        bool        `json:"unwrap"`
	Parameters    []parameter `json:"parameters"`
}
type parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
}
type Client struct {
	config     Config
	base       *url.URL
	http       *http.Client
	operations map[string]operation
}
type APIError struct {
	Status     int
	Type       string
	Code       string
	Message    string
	Details    json.RawMessage
	RequestID  string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("envai: HTTP %d (%s/%s)", e.Status, e.Type, e.Code)
}

type ProtocolError struct{ Reason string }

func (e *ProtocolError) Error() string { return "envai: protocol violation: " + e.Reason }

type MultipartPart struct {
	Name        string
	Filename    string
	ContentType string
	Data        []byte
	// Reader and Size enable bounded streaming instead of buffering Data.
	Reader io.Reader
	Size   int64
}
type Download struct {
	Body   io.ReadCloser
	Header http.Header
	Status int
}

func (d *Download) Close() error { return d.Body.Close() }

type cancelBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelBody) Close() error { defer b.cancel(); return b.ReadCloser.Close() }

func New(config Config) (*Client, error) {
	u, err := url.Parse(config.BaseURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && config.AllowHTTP)) {
		return nil, errors.New("envai: base URL must be HTTPS, without credentials, query or fragment; HTTP requires AllowHTTP")
	}
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "" {
		u.Path = "/api/v1"
	}
	if !strings.HasSuffix(u.Path, "/api/v1") {
		return nil, errors.New("envai: base URL must end with /api/v1")
	}
	if strings.ContainsAny(config.APIKey, "\r\n") {
		return nil, errors.New("envai: invalid API key")
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 60 * time.Second
	}
	if config.RequestTimeout < 0 || config.RetryAttempts < 0 || config.RetryAttempts > 10 {
		return nil, errors.New("envai: invalid timeout or retry budget")
	}
	if config.MaxResponseBytes == 0 {
		config.MaxResponseBytes = 8 << 20
	}
	if config.MaxResponseBytes < 1 {
		return nil, errors.New("envai: invalid response limit")
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 250 * time.Millisecond
	}
	if config.RetryDelay < 0 {
		return nil, errors.New("envai: invalid retry delay")
	}
	hc := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, ResponseHeaderTimeout: 10 * time.Second}}
	if config.HTTPClient != nil {
		*hc = *config.HTTPClient
	}
	hc.Timeout = 0
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	if config.JWKSURL == "" {
		j := *u
		j.Path = "/.well-known/jwks.json"
		j.RawPath = ""
		config.JWKSURL = j.String()
	}
	j, err := url.Parse(config.JWKSURL)
	if err != nil || (j.Scheme != "https" && !(j.Scheme == "http" && config.AllowHTTP)) || j.Host == "" || j.User != nil || j.Fragment != "" || j.RawQuery != "" {
		return nil, errors.New("envai: JWKS URL must use an allowed HTTP(S) origin without credentials or query")
	}
	data, _ := contracts.ReadFile("contract/operations.json")
	var list []operation
	if err = json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	c := &Client{config: config, base: u, http: hc, operations: map[string]operation{}}
	for _, op := range list {
		c.operations[op.ID] = op
	}
	return c, nil
}

func (c *Client) OperationIDs() []string {
	list := make([]string, 0, len(c.operations))
	for id := range c.operations {
		list = append(list, id)
	}
	return list
}
func (c *Client) requestURL(op operation, params map[string]any) (string, error) {
	path := op.Path
	if atpOperation(op) {
		path = strings.TrimPrefix(path, "/atp/v3")
		if !c.config.LegacyReadOnly && !strings.HasSuffix(op.ID, "Legacy") {
			path = "/atp/v3" + path
		}
	}
	q := url.Values{}
	allowed := map[string]bool{}
	for _, p := range op.Parameters {
		allowed[p.Name] = true
		v, exists := params[p.Name]
		if p.Required && (!exists || v == nil || fmt.Sprint(v) == "") {
			return "", fmt.Errorf("envai: required parameter %s missing", p.Name)
		}
		if !exists || v == nil {
			continue
		}
		value := fmt.Sprint(v)
		if p.In == "path" {
			if value == "." || value == ".." || strings.ContainsAny(value, "/\\\x00") {
				return "", errors.New("envai: invalid path parameter")
			}
			path = strings.ReplaceAll(path, "{"+p.Name+"}", url.PathEscape(value))
		} else if p.In == "query" {
			q.Set(p.Name, value)
		}
	}
	for k := range params {
		if !allowed[k] {
			return "", fmt.Errorf("envai: unknown parameter %s", k)
		}
	}
	u := *c.base
	if op.ID == "getWikiWebhookJwks" {
		u.Path = "/.well-known/wiki-webhook-jwks.json"
		u.RawPath = ""
		u.RawQuery = ""
		return u.String(), nil
	}
	if op.ID == "getJwks" {
		return c.config.JWKSURL, nil
	}
	u.Path = c.base.Path + path
	u.RawPath = ""
	if strings.Contains(path, "%") {
		raw := c.base.EscapedPath() + path
		decoded, err := url.PathUnescape(raw)
		if err != nil {
			return "", err
		}
		u.Path = decoded
		u.RawPath = raw
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *Client) makeBody(op operation, body any) ([]byte, string, error) {
	if body == nil {
		return nil, "", nil
	}
	if op.Body == "multipart" {
		parts, ok := body.([]MultipartPart)
		if !ok {
			return nil, "", errors.New("envai: multipart parts required")
		}
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		for _, p := range parts {
			if p.Name == "" || strings.ContainsAny(p.Name+p.Filename+p.ContentType, "\r\n") {
				return nil, "", errors.New("envai: invalid multipart metadata")
			}
			h := textproto.MIMEHeader{}
			disp := `form-data; name="` + strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(p.Name) + `"`
			if p.Filename != "" {
				disp += `; filename="` + strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(p.Filename) + `"`
			}
			h.Set("Content-Disposition", disp)
			if p.ContentType != "" {
				h.Set("Content-Type", p.ContentType)
			}
			part, err := w.CreatePart(h)
			if err != nil {
				return nil, "", err
			}
			if _, err = part.Write(p.Data); err != nil {
				return nil, "", err
			}
		}
		if err := w.Close(); err != nil {
			return nil, "", err
		}
		return b.Bytes(), w.FormDataContentType(), nil
	}
	data, err := json.Marshal(body)
	return data, "application/json", err
}

func (c *Client) do(ctx context.Context, id string, params map[string]any, body any, accept string) (*http.Response, error) {
	op, ok := c.operations[id]
	if !ok {
		return nil, errors.New("envai: unknown operation")
	}
	uri, err := c.requestURL(op, params)
	if err != nil {
		return nil, err
	}
	if op.Authenticated && c.config.APIKey == "" {
		return nil, errors.New("envai: API key is required")
	}
	if err = c.validateOperationParams(op, params); err != nil {
		return nil, err
	}
	if err = validateWikiBody(op, body); err != nil {
		return nil, err
	}
	if err = c.checkMutation(ctx, op); err != nil {
		return nil, err
	}
	var data []byte
	var ct string
	var streaming io.Reader
	if op.Body == "multipart" {
		if parts, ok := body.([]MultipartPart); ok {
			streaming, ct, err = multipartReader(parts)
		} else {
			err = errors.New("envai: multipart parts required")
		}
	} else {
		data, ct, err = c.makeBody(op, body)
	}
	if err != nil {
		return nil, err
	}
	for attempt := 0; ; attempt++ {
		var reader io.Reader = bytes.NewReader(data)
		if streaming != nil {
			reader = streaming
		}
		req, err := http.NewRequestWithContext(ctx, op.Method, uri, reader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", accept)
		if negotiated(op) {
			req.Header.Set(ContractHeader, c.versionFor(op))
		}
		for _, p := range op.Parameters {
			if p.In == "header" {
				if v, ok := params[p.Name]; ok {
					req.Header.Set(p.Name, fmt.Sprint(v))
				}
			}
		}
		req.Header.Set("User-Agent", "envai-go/0.1.0")
		if ct != "" {
			req.Header.Set("Content-Type", ct)
		}
		if op.Authenticated {
			req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		}
		start := time.Now()
		resp, err := c.http.Do(req)
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		if c.config.Observer != nil {
			c.config.Observer(RequestObservation{id, status, attempt, time.Since(start)})
		}
		if err != nil {
			return nil, fmt.Errorf("envai: transport request failed: %w", sanitizeTransportError(err))
		}
		if op.Method == "GET" && op.Kind == "json" && attempt < c.config.RetryAttempts && (status == 429 || status == 502 || status == 503 || status == 504) {
			delay := time.Duration(float64(c.config.RetryDelay*time.Duration(1<<attempt)) * (0.8 + 0.4*rand.Float64()))
			if d := retryAfter(resp.Header.Get("Retry-After")); d > delay {
				delay = d
			}
			resp.Body.Close()
			if err := wait(ctx, delay); err != nil {
				return nil, err
			}
			continue
		}
		if status < 200 || status >= 300 {
			apiErr := c.apiError(resp)
			var api *APIError
			if negotiated(op) && errors.As(apiErr, &api) && (status == 426 || (status == 400 && api.Code == "unsupported_contract_version") || (status == 404 && atpOperation(op) && !c.config.LegacyReadOnly)) {
				return nil, &CompatibilityError{Reason: "selected contract endpoint is unavailable or requires an upgrade", Cause: apiErr}
			}
			return nil, apiErr
		}
		if err = c.checkAcknowledgment(op, resp); err != nil {
			resp.Body.Close()
			return nil, err
		}
		return resp, nil
	}
}
func sanitizeTransportError(err error) error {
	var u *url.Error
	if errors.As(err, &u) {
		return u.Err
	}
	return err
}
func retryAfter(value string) time.Duration {
	if n, err := strconv.ParseInt(value, 10, 64); err == nil && n >= 0 && n < 86400*365 {
		return time.Duration(n) * time.Second
	}
	if t, err := http.ParseTime(value); err == nil && time.Until(t) > 0 {
		return time.Until(t)
	}
	return 0
}
func wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
func (c *Client) apiError(resp *http.Response) error {
	defer resp.Body.Close()
	e := &APIError{Status: resp.StatusCode, RequestID: resp.Header.Get("X-Request-ID"), RetryAfter: retryAfter(resp.Header.Get("Retry-After"))}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, c.config.MaxResponseBytes))
	var wire struct {
		Error struct {
			Type    string          `json:"type"`
			Code    string          `json:"code"`
			Message string          `json:"message"`
			Details json.RawMessage `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(b, &wire) == nil {
		e.Type = wire.Error.Type
		e.Code = wire.Error.Code
		e.Message = wire.Error.Message
		e.Details = wire.Error.Details
	}
	return e
}
func (c *Client) call(ctx context.Context, id string, params map[string]any, body any, out any) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()
	resp, err := c.do(ctx, id, params, body, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		return nil
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, c.config.MaxResponseBytes+1))
	if err != nil {
		return err
	}
	if int64(len(b)) > c.config.MaxResponseBytes {
		return errors.New("envai: response exceeds configured limit")
	}
	if c.operations[id].Unwrap {
		var envelope map[string]json.RawMessage
		if err = json.Unmarshal(b, &envelope); err != nil {
			return &ProtocolError{"invalid JSON response"}
		}
		var ok bool
		b, ok = envelope["data"]
		if !ok {
			return &ProtocolError{"response is missing data"}
		}
	}
	if err = c.validateNegotiatedJSON(id, b); err != nil {
		return err
	}
	if err = json.Unmarshal(b, out); err != nil {
		return &ProtocolError{"invalid JSON response shape"}
	}
	return nil
}
func (c *Client) download(ctx context.Context, id string, params map[string]any, body any) (*Download, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	resp, err := c.do(ctx, id, params, body, "*/*")
	if err != nil {
		cancel()
		return nil, err
	}
	return &Download{Body: &cancelBody{ReadCloser: resp.Body, cancel: cancel}, Header: resp.Header, Status: resp.StatusCode}, nil
}

func (c Config) Format(state fmt.State, verb rune) {
	fmt.Fprint(state, "envai.Config{APIKey:[REDACTED]}")
}
func (c *Client) Format(state fmt.State, verb rune) {
	fmt.Fprintf(state, "envai.Client{base_url:%q, credentials:[REDACTED]}", c.base.String())
}
