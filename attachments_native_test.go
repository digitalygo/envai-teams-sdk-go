package envai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func originalAttachmentUploaded(t *testing.T) json.RawMessage {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(attachmentFixture(t, "original-ready"), &value); err != nil {
		t.Fatal(err)
	}
	value["lifecycle"] = "uploaded"
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestWaitForOriginalAttachmentRequiresReady(t *testing.T) {
	ready := attachmentFixture(t, "original-ready")
	// Compile the contract outside the polling clock: these assertions test
	// lifecycle transitions and a short polling deadline, not cold compilation.
	var validated ChatAttachment
	if err := json.Unmarshal(ready, &validated); err != nil {
		t.Fatal(err)
	}
	uploaded := originalAttachmentUploaded(t)
	var reads atomic.Int32
	c := attachmentClient(t, true, func(w http.ResponseWriter, r *http.Request) {
		if reads.Add(1) == 1 {
			fmt.Fprintf(w, `{"data":%s}`, uploaded)
			return
		}
		fmt.Fprintf(w, `{"data":%s}`, ready)
	})
	value, err := c.WaitForAttachment(context.Background(), "w", "c", "a", time.Millisecond, time.Second)
	if err != nil || value.Lifecycle != "ready" || value.Conversion.Status != "not_required" || reads.Load() != 2 {
		t.Fatal(value, err, reads.Load())
	}
	reads.Store(0)
	c = attachmentClient(t, true, func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		fmt.Fprintf(w, `{"data":%s}`, ready)
	})
	value, err = c.WaitForAttachment(context.Background(), "w", "c", "a", time.Millisecond, time.Second)
	if err != nil || value.Conversion.Status != "not_required" || reads.Load() != 1 {
		t.Fatal(value, err, reads.Load())
	}
	reads.Store(0)
	c = attachmentClient(t, true, func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		fmt.Fprintf(w, `{"data":%s}`, uploaded)
	})
	if _, err = c.WaitForAttachment(context.Background(), "w", "c", "a", time.Millisecond, 50*time.Millisecond); !errors.Is(err, context.DeadlineExceeded) || reads.Load() < 2 {
		t.Fatal(err, reads.Load())
	}
}

func TestAttachmentManifestNativeFieldsAndConversionSchema(t *testing.T) {
	manifest := []byte(`{"attachment_id":"catt","original_name":"notes.md","mime_type":"text/markdown","mode":"native","revision":1,"sha256":"a8257d3af0961ab4c5977d60edbecd1dbbced4bc8623af7b140892abbf884264","bytes":17,"conversion":{"status":"not_required"}}`)
	var value AttachmentManifest
	if err := json.Unmarshal(manifest, &value); err != nil {
		t.Fatal(err)
	}
	if !value.MimeType.Set || value.MimeType.Value != "text/markdown" || !value.Mode.Set || value.Mode.Value != "native" {
		t.Fatal(value)
	}
	out, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var round map[string]json.RawMessage
	if err = json.Unmarshal(out, &round); err != nil {
		t.Fatal(err)
	}
	if string(round["mime_type"]) != `"text/markdown"` || string(round["mode"]) != `"native"` {
		t.Fatal(string(out))
	}
	legacy := []byte(`{"attachment_id":"catt","original_name":"notes.md","revision":1,"sha256":"a8257d3af0961ab4c5977d60edbecd1dbbced4bc8623af7b140892abbf884264","bytes":17,"conversion":{"status":"completed"}}`)
	value = AttachmentManifest{}
	if err = json.Unmarshal(legacy, &value); err != nil || value.MimeType.Set || value.Mode.Set {
		t.Fatal(value, err)
	}
	invalidMode := []byte(`{"attachment_id":"catt","original_name":"notes.md","mode":"converted","revision":1,"sha256":"a8257d3af0961ab4c5977d60edbecd1dbbced4bc8623af7b140892abbf884264","bytes":17,"conversion":{"status":"not_required"}}`)
	if json.Unmarshal(invalidMode, &value) == nil {
		t.Fatal("invalid native mode accepted")
	}
	var conversion AttachmentConversion
	if err = json.Unmarshal([]byte(`{"status":"not_required","error_code":null,"recoverable":false,"next_retry_at":null,"retryable":false}`), &conversion); err != nil || conversion.Status != "not_required" {
		t.Fatal(conversion, err)
	}
	if json.Unmarshal([]byte(`{"status":"skipped","error_code":null,"recoverable":false,"next_retry_at":null,"retryable":false}`), &conversion) == nil {
		t.Fatal("invalid conversion status accepted")
	}
	var attachment ChatAttachment
	if err = json.Unmarshal(attachmentFixture(t, "original-ready"), &attachment); err != nil || attachment.Conversion.Status != "not_required" || attachment.Derived != nil {
		t.Fatal(attachment, err)
	}
}

func TestNativeAttachmentRetryRequiresOnlyChatAttachmentFeature(t *testing.T) {
	posts := 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, capabilitiesWithFeatures(t, `["chat_attachments"]`))
			return
		}
		posts++
		w.WriteHeader(409)
		fmt.Fprint(w, `{"error":{"type":"conflict","code":"attachment_retry_not_permitted","message":"no conversion job"}}`)
	})
	_, err := c.RetryChatAttachment(context.Background(), RetryChatAttachmentParams{WorkspaceID: "w", ChatID: "c", AttachmentID: "a", IdempotencyKey: "retry"})
	if err == nil || posts != 1 {
		t.Fatal(err, posts)
	}
	var api *APIError
	if !errors.As(err, &api) || api.Status != 409 || api.Code != "attachment_retry_not_permitted" {
		t.Fatal(err)
	}
}

func TestAttachmentMessageRequiresOnlyChatAttachmentFeature(t *testing.T) {
	posts, preflights := 0, 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			preflights++
			fmt.Fprint(w, capabilitiesWithFeatures(t, `["chat_attachments"]`))
			return
		}
		posts++
		w.WriteHeader(202)
		fmt.Fprint(w, `{"status":"accepted","turn_id":"turn","attachment_ids":["a"]}`)
	})
	_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "c", IdempotencyKey: Present("stable"), Body: ChatMessageRequest{Message: "read", AttachmentIDs: Present([]string{"a"})}})
	if err != nil || posts != 1 || preflights != 1 {
		t.Fatal(err, posts, preflights)
	}
}
