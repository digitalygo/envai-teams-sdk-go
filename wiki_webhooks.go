package envai

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// VerifiedWikiEvent is a discriminated outgoing Envai event, not a provider callback.
// Raw is the exact signed body. File events intentionally have no terminal cost.
type VerifiedWikiEvent struct {
	EventID       string
	EventType     string
	WorkspaceID   string
	OperationID   string
	OccurredAt    string
	FileProcessed *WikiFileProcessedEvent
	FileDeleted   *WikiFileDeletedEvent
	Operation     *WikiOperationEvent
	Raw           json.RawMessage
}

func DecodeWikiWebhook(b []byte) (VerifiedWikiEvent, error) {
	if !utf8.Valid(b) {
		return VerifiedWikiEvent{}, &ProtocolError{"wiki body is not UTF-8"}
	}
	if err := validateDefinition("wiki.json", "WikiWebhookEvent", b); err != nil {
		return VerifiedWikiEvent{}, err
	}
	var common struct {
		EventID     string `json:"event_id"`
		EventType   string `json:"event_type"`
		WorkspaceID string `json:"workspace_id"`
		OperationID string `json:"operation_id"`
		OccurredAt  string `json:"occurred_at"`
	}
	if err := json.Unmarshal(b, &common); err != nil {
		return VerifiedWikiEvent{}, err
	}
	e := VerifiedWikiEvent{EventID: common.EventID, EventType: common.EventType, WorkspaceID: common.WorkspaceID, OperationID: common.OperationID, OccurredAt: common.OccurredAt, Raw: append(json.RawMessage(nil), b...)}
	switch e.EventType {
	case "wiki.inbox.file.processed":
		var v WikiFileProcessedEvent
		if err := json.Unmarshal(b, &v); err != nil {
			return VerifiedWikiEvent{}, err
		}
		e.FileProcessed = &v
	case "wiki.outbox.file.deleted":
		var v WikiFileDeletedEvent
		if err := json.Unmarshal(b, &v); err != nil {
			return VerifiedWikiEvent{}, err
		}
		e.FileDeleted = &v
	default:
		var v WikiOperationEvent
		if err := json.Unmarshal(b, &v); err != nil {
			return VerifiedWikiEvent{}, err
		}
		e.Operation = &v
	}
	return e, nil
}

// WikiWebhookVerifier uses only the client's configured origin public wiki JWKS.
// It is safe for concurrent handlers. Configure fields before first use.
type WikiWebhookVerifier struct {
	Client       *Client
	WorkspaceID  string
	Tolerance    time.Duration
	MaxBodyBytes int64
	Clock        func() time.Time
	mu           sync.Mutex
	keys         map[string]JWK
	expires      time.Time
}

var wikiIdentityPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,255}$`)
var wikiTimestampPattern = regexp.MustCompile(`^[1-9][0-9]{0,10}$`)

func WikiSigningInput(kid, eventID, timestamp string, body []byte) []byte {
	hash := sha256.Sum256(body)
	return []byte("envai-webhook-v1\n" + kid + "\n" + eventID + "\n" + timestamp + "\n" + hex.EncodeToString(hash[:]))
}
func singleWikiHeader(headers http.Header, name string) (string, error) {
	var values []string
	for k, v := range headers {
		if strings.EqualFold(k, name) {
			values = append(values, v...)
		}
	}
	if len(values) != 1 || values[0] == "" {
		return "", &ProtocolError{"missing or duplicate wiki signing header"}
	}
	return values[0], nil
}
func canonicalBase64(value string, length int) ([]byte, error) {
	b, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(b) != length || base64.RawURLEncoding.EncodeToString(b) != value {
		return nil, &ProtocolError{"malformed wiki signature or key encoding"}
	}
	return b, nil
}
func (v *WikiWebhookVerifier) now() time.Time {
	if v.Clock != nil {
		return v.Clock()
	}
	return time.Now()
}
func (v *WikiWebhookVerifier) Verify(ctx context.Context, method string, headers http.Header, body []byte) (VerifiedWikiEvent, error) {
	if v.Client == nil || v.WorkspaceID == "" || v.Tolerance < 0 || v.MaxBodyBytes < 0 {
		return VerifiedWikiEvent{}, errors.New("envai: invalid wiki verifier configuration")
	}
	limit := v.MaxBodyBytes
	if limit == 0 {
		limit = 1 << 20
	}
	if int64(len(body)) > limit {
		return VerifiedWikiEvent{}, &ProtocolError{"wiki body exceeds configured limit"}
	}
	if method != "POST" {
		return VerifiedWikiEvent{}, &ProtocolError{"wiki webhooks require POST"}
	}
	names := []string{"Envai-Webhook-Id", "Envai-Webhook-Key-Id", "Envai-Webhook-Timestamp", "Envai-Webhook-Signature"}
	values := make([]string, len(names))
	for i, name := range names {
		var err error
		values[i], err = singleWikiHeader(headers, name)
		if err != nil {
			return VerifiedWikiEvent{}, err
		}
	}
	id, kid, ts := values[0], values[1], values[2]
	if !wikiIdentityPattern.MatchString(id) || !wikiIdentityPattern.MatchString(kid) || !wikiTimestampPattern.MatchString(ts) {
		return VerifiedWikiEvent{}, &ProtocolError{"malformed wiki signing header"}
	}
	timestamp, _ := strconv.ParseInt(ts, 10, 64)
	tolerance := v.Tolerance
	if tolerance == 0 {
		tolerance = 300 * time.Second
	}
	difference := v.now().Unix() - timestamp
	if difference > int64(tolerance/time.Second) || difference < -int64(tolerance/time.Second) {
		return VerifiedWikiEvent{}, &ProtocolError{"wiki timestamp outside replay window"}
	}
	signature, err := canonicalBase64(values[3], ed25519.SignatureSize)
	if err != nil {
		return VerifiedWikiEvent{}, err
	}
	key, err := v.key(ctx, kid)
	if err != nil {
		return VerifiedWikiEvent{}, err
	}
	if key.Kty != "OKP" || key.Crv != "Ed25519" || key.Alg != "EdDSA" || key.Use != "sig" {
		return VerifiedWikiEvent{}, &ProtocolError{"wiki key is not an Ed25519 signing key"}
	}
	public, err := canonicalBase64(key.X, ed25519.PublicKeySize)
	if err != nil {
		return VerifiedWikiEvent{}, err
	}
	if !ed25519.Verify(public, WikiSigningInput(kid, id, ts, body), signature) {
		return VerifiedWikiEvent{}, &ProtocolError{"wiki signature verification failed"}
	}
	event, err := DecodeWikiWebhook(body)
	if err != nil {
		return VerifiedWikiEvent{}, err
	}
	if event.EventID != id || event.WorkspaceID != v.WorkspaceID {
		return VerifiedWikiEvent{}, &ProtocolError{"wiki identity or workspace mismatch"}
	}
	return event, nil
}
func (v *WikiWebhookVerifier) key(ctx context.Context, kid string) (JWK, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.keys != nil && v.now().Before(v.expires) {
		if key, ok := v.keys[kid]; ok {
			return key, nil
		}
	}
	document, err := v.Client.GetWikiWebhookJWKS(ctx, GetWikiWebhookJWKSParams{})
	if err != nil {
		return JWK{}, err
	}
	keys := map[string]JWK{}
	for _, key := range document.Keys {
		if _, exists := keys[key.Kid]; exists {
			return JWK{}, &ProtocolError{"duplicate wiki signing key ID"}
		}
		keys[key.Kid] = key
	}
	v.keys = keys
	v.expires = v.now().Add(300 * time.Second)
	if key, ok := keys[kid]; ok {
		return key, nil
	}
	return JWK{}, &ProtocolError{"unknown wiki signing key ID"}
}
