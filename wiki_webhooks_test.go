package envai

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type wikiVector struct {
	RawBody      string            `json:"raw_body"`
	SigningInput string            `json:"signing_input"`
	Headers      map[string]string `json:"headers"`
	JWKS         JWKSDocument      `json:"jwks"`
	Now          int64             `json:"now"`
}

func wikiHarness(t *testing.T) (*WikiWebhookVerifier, http.Header, []byte, *atomic.Int32) {
	t.Helper()
	vector := fixture[wikiVector](t, "wiki/webhook-signature.json")
	headers := http.Header{}
	for k, v := range vector.Headers {
		headers.Set(k, v)
	}
	calls := &atomic.Int32{}
	c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/.well-known/wiki-webhook-jwks.json" || r.Header.Get("Authorization") != "" || r.Header.Get(ContractHeader) != "" {
			t.Errorf("wrong JWKS route or credentials: %s %+v", r.URL.Path, r.Header)
		}
		json.NewEncoder(w).Encode(vector.JWKS)
	})
	return &WikiWebhookVerifier{Client: c, WorkspaceID: "wsp_01j5abcdefghjkmnpqrstvwxyz", Clock: func() time.Time { return time.Unix(vector.Now, 0) }}, headers, []byte(vector.RawBody), calls
}
func signWiki(t *testing.T, headers http.Header, body []byte) {
	t.Helper()
	seed, err := hex.DecodeString("9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60")
	if err != nil {
		t.Fatal(err)
	}
	private := ed25519.NewKeyFromSeed(seed)
	signature := ed25519.Sign(private, WikiSigningInput(headers.Get("Envai-Webhook-Key-Id"), headers.Get("Envai-Webhook-Id"), headers.Get("Envai-Webhook-Timestamp"), body))
	headers.Set("Envai-Webhook-Signature", base64.RawURLEncoding.EncodeToString(signature))
}
func TestWikiVerificationSharedVectorAndCache(t *testing.T) {
	v, h, b, calls := wikiHarness(t)
	vector := fixture[wikiVector](t, "wiki/webhook-signature.json")
	if string(WikiSigningInput(h.Get("Envai-Webhook-Key-Id"), h.Get("Envai-Webhook-Id"), h.Get("Envai-Webhook-Timestamp"), b)) != vector.SigningInput {
		t.Fatal("signature preimage changed")
	}
	for n := 0; n < 3; n++ {
		event, err := v.Verify(context.Background(), "POST", h, b)
		if err != nil {
			t.Fatal(err)
		}
		if event.FileProcessed == nil || event.Operation != nil || event.FileProcessed.SourceRevision != 3 || string(event.Raw) != string(b) {
			t.Fatal(event)
		}
	}
	if calls.Load() != 1 {
		t.Fatal(calls.Load())
	}
	// Fresh delivery timestamps are valid even when the logical occurrence is old.
	v.Clock = func() time.Time { return time.Unix(vector.Now+86400, 0) }
	h.Set("Envai-Webhook-Timestamp", fmt.Sprint(vector.Now+86400))
	signWiki(t, h, b)
	if _, err := v.Verify(context.Background(), "POST", h, b); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatal("expired keys not refreshed")
	}
}
func TestWikiVerificationRejectsMalformedInputs(t *testing.T) {
	for _, name := range []string{"method", "missing", "duplicate", "caseDuplicate", "comma", "kid", "id", "zeroTimestamp", "leadingZeroTimestamp", "overflowTimestamp", "old", "future", "signaturePadding", "signatureBadBits", "signatureLength", "tamper", "workspace", "bodyLimit", "negativeLimit", "negativeTolerance", "nilClient", "emptyWorkspace", "invalidUTF8", "eventIDMismatch", "invalidPayload"} {
		t.Run(name, func(t *testing.T) {
			v, h, b, _ := wikiHarness(t)
			method := "POST"
			switch name {
			case "method":
				method = "GET"
			case "missing":
				h.Del("Envai-Webhook-Id")
			case "duplicate":
				h.Add("Envai-Webhook-Id", h.Get("Envai-Webhook-Id"))
			case "caseDuplicate":
				h["envai-webhook-id"] = []string{h.Get("Envai-Webhook-Id")}
			case "comma":
				h.Set("Envai-Webhook-Id", "one,two")
			case "kid":
				h.Set("Envai-Webhook-Key-Id", "../bad")
			case "id":
				h.Set("Envai-Webhook-Id", "bad\nvalue")
			case "zeroTimestamp":
				h.Set("Envai-Webhook-Timestamp", "0")
			case "leadingZeroTimestamp":
				h.Set("Envai-Webhook-Timestamp", "01791907200")
			case "overflowTimestamp":
				h.Set("Envai-Webhook-Timestamp", "999999999999999999999")
			case "old":
				v.Clock = func() time.Time { return time.Unix(1791907501, 0) }
			case "future":
				v.Clock = func() time.Time { return time.Unix(1791906899, 0) }
			case "signaturePadding":
				h.Set("Envai-Webhook-Signature", h.Get("Envai-Webhook-Signature")+"=")
			case "signatureBadBits":
				h.Set("Envai-Webhook-Signature", strings.Repeat("A", 85)+"B")
			case "signatureLength":
				h.Set("Envai-Webhook-Signature", "YQ")
			case "tamper":
				b = append(b, ' ')
			case "workspace":
				v.WorkspaceID = "other"
			case "bodyLimit":
				v.MaxBodyBytes = 1
			case "negativeLimit":
				v.MaxBodyBytes = -1
			case "negativeTolerance":
				v.Tolerance = -time.Second
			case "nilClient":
				v.Client = nil
			case "emptyWorkspace":
				v.WorkspaceID = ""
			case "invalidUTF8":
				b = []byte{0xff}
				signWiki(t, h, b)
			case "eventIDMismatch":
				b = []byte(strings.Replace(string(b), "evt_01j5abcdefghjkmnpqrstvwxyz", "evt_other", 1))
				signWiki(t, h, b)
			case "invalidPayload":
				b = []byte(`{}`)
				signWiki(t, h, b)
			}
			if _, err := v.Verify(context.Background(), method, h, b); err == nil {
				t.Fatal("accepted invalid", name)
			}
		})
	}
}
func TestWikiKeyRotationMetadataAndErrors(t *testing.T) {
	for _, kind := range []string{"rotation", "unknown", "duplicate", "badType", "badCurve", "badAlg", "missingAlg", "badUse", "badEncoding", "badSize", "httpError"} {
		t.Run(kind, func(t *testing.T) {
			v, h, b, _ := wikiHarness(t)
			vector := fixture[wikiVector](t, "wiki/webhook-signature.json")
			key := vector.JWKS.Keys[0]
			switch kind {
			case "badType":
				key.Kty = "RSA"
			case "badCurve":
				key.Crv = "X25519"
			case "badAlg":
				key.Alg = "HS256"
			case "missingAlg":
				key.Alg = ""
			case "badUse":
				key.Use = "enc"
			case "badEncoding":
				key.X += "="
			case "badSize":
				key.X = "YQ"
			case "unknown":
				key.Kid = "other"
			}
			v.keys = map[string]JWK{"old": key}
			v.expires = v.now().Add(time.Minute)
			var calls atomic.Int32
			c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if kind == "httpError" {
					w.WriteHeader(500)
					return
				}
				keys := []JWK{key}
				if kind == "duplicate" {
					keys = append(keys, key)
				}
				json.NewEncoder(w).Encode(JWKSDocument{Keys: keys})
			})
			v.Client = c
			_, err := v.Verify(context.Background(), "POST", h, b)
			if kind == "rotation" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("invalid key accepted")
			}
			if calls.Load() != 1 {
				t.Fatal("unbounded refresh", calls.Load())
			}
		})
	}
}
func TestWikiVerifierConcurrency(t *testing.T) {
	v, h, b, calls := wikiHarness(t)
	var wg sync.WaitGroup
	for n := 0; n < 20; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := v.Verify(context.Background(), "POST", h, b); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatal(calls.Load())
	}
}
func TestAllWikiFixturesTypedRoundTripAndInvalid(t *testing.T) {
	models := map[string]reflect.Type{"BillingState": reflect.TypeOf(BillingState{}), "CurrencyAmount": reflect.TypeOf(CurrencyAmount{}), "UsageCost": reflect.TypeOf(UsageCost{}), "ConversionState": reflect.TypeOf(ConversionState{}), "IndexingState": reflect.TypeOf(IndexingState{}), "WikiSource": reflect.TypeOf(WikiSource{}), "BatchSizes": reflect.TypeOf(BatchSizes{}), "FileOutcome": reflect.TypeOf(FileOutcome{}), "PhaseResult": reflect.TypeOf(PhaseResult{}), "WikiOperation": reflect.TypeOf(WikiOperation{}), "Category": reflect.TypeOf(Category{}), "Categories": reflect.TypeOf(Categories{}), "CatalogEntry": reflect.TypeOf(CatalogEntry{}), "IndexRange": reflect.TypeOf(IndexRange{}), "CategoryIndex": reflect.TypeOf(CategoryIndex{}), "WikiFileProcessedEvent": reflect.TypeOf(WikiFileProcessedEvent{}), "WikiFileDeletedEvent": reflect.TypeOf(WikiFileDeletedEvent{}), "WikiOperationEvent": reflect.TypeOf(WikiOperationEvent{}), "ATPBillingState": reflect.TypeOf(ATPBillingState{}), "ATPCurrencyAmount": reflect.TypeOf(ATPCurrencyAmount{}), "ATPUsageCost": reflect.TypeOf(ATPUsageCost{})}
	files, _ := os.ReadDir("testdata/wiki")
	seen := map[string]bool{}
	for _, file := range files {
		var fixture struct {
			Schema string          `json:"schema"`
			Data   json.RawMessage `json:"data"`
		}
		b, _ := os.ReadFile("testdata/wiki/" + file.Name())
		json.Unmarshal(b, &fixture)
		typ, ok := models[fixture.Schema]
		if !ok {
			continue
		}
		value := reflect.New(typ)
		if err := json.Unmarshal(fixture.Data, value.Interface()); err != nil {
			t.Fatal(file.Name(), err)
		}
		out, err := json.Marshal(value.Interface())
		if err != nil {
			t.Fatal(err)
		}
		if err = validateDefinition("wiki.json", fixture.Schema, out); err != nil {
			t.Fatal(file.Name(), err)
		}
		seen[fixture.Schema] = true
		if strings.HasSuffix(fixture.Schema, "Event") {
			event, err := DecodeWikiWebhook(fixture.Data)
			if err != nil || event.EventID == "" {
				t.Fatal(event, err)
			}
		}
	}
	for name, typ := range models {
		t.Run(name, func(t *testing.T) {
			value := reflect.New(typ)
			for _, invalid := range []string{"null", "[]", "{", `{}`} {
				if json.Unmarshal([]byte(invalid), value.Interface()) == nil {
					t.Fatal("invalid object accepted", name, invalid)
				}
			}
			// Marshal every model with an unknown value and ensure all generated hooks run.
			field := value.Elem().FieldByName("Extra")
			field.Set(reflect.ValueOf(map[string]json.RawMessage{"extension": json.RawMessage(`{"precise":9007199254740993}`)}))
			if _, err := json.Marshal(value.Interface()); err != nil {
				t.Fatal(err)
			}
		})
	}
	if !seen["WikiSource"] || !seen["UsageCost"] || !seen["WikiOperation"] {
		t.Fatal("missing fixture coverage")
	}
}
