package envai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func attachmentFixture(t *testing.T, name string) json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile("testdata/chat-attachments/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var value struct {
		Data json.RawMessage `json:"data"`
	}
	if err = json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value.Data
}
func attachmentClient(t *testing.T, feature bool, fn http.HandlerFunc) *Client {
	return testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			value := testCapabilities
			if feature {
				value = strings.Replace(value, `"features":[]`, `"features":["chat_attachments","chat_attachment_originals"]`, 1)
			}
			fmt.Fprint(w, value)
			return
		}
		fn(w, r)
	})
}
func TestAttachmentFilesUploadPollAndPrivateDownloads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.txt")
	if err := os.WriteFile(path, []byte("synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	reads, posts := 0, 0
	c := attachmentClient(t, true, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.RequestURI, "/workspaces/w/chats/c/attachments") {
			t.Error(r.RequestURI)
		}
		if r.Header.Get(ContractHeader) != "3" {
			t.Error("missing contract selector")
		}
		if r.Method == "POST" {
			posts++
			if r.Header.Get("Idempotency-Key") != "upload" {
				t.Error("key")
			}
			if err := r.ParseMultipartForm(1024); err != nil {
				t.Error(err)
			}
			file, header, err := r.FormFile("file")
			if err != nil {
				t.Fatal(err)
			}
			file.Close()
			if header.Filename != "note.txt" {
				t.Error(header.Filename)
			}
		}
		if strings.HasSuffix(r.URL.Path, "/download") {
			fmt.Fprint(w, "original")
			return
		}
		name := "ready"
		if r.Method == "GET" {
			reads++
			if reads == 1 {
				name = "pending"
			}
		}
		fmt.Fprintf(w, `{"data":%s}`, attachmentFixture(t, name))
	})
	value, err := c.UploadAttachmentFile(context.Background(), "w", "c", path, "upload", "text/plain")
	if err != nil || value.Lifecycle != "ready" {
		t.Fatal(value, err)
	}
	value, err = c.WaitForAttachment(context.Background(), "w", "c", "a", time.Millisecond, time.Second)
	if err != nil || value.Lifecycle != "ready" || reads != 2 || posts != 1 {
		t.Fatal(value, err, reads, posts)
	}
	download, err := c.DownloadChatAttachment(context.Background(), DownloadChatAttachmentParams{WorkspaceID: "w", ChatID: "c", AttachmentID: "a", Kind: "original"})
	if err != nil {
		t.Fatal(err)
	}
	download.Close()
	if _, err = c.UploadAttachmentFile(context.Background(), "w", "c", path+"missing", "upload", ""); err == nil {
		t.Fatal("missing file accepted")
	}
	if _, err = c.UploadAttachmentFile(context.Background(), "w", "c", filepath.Dir(path), "upload", ""); err == nil {
		t.Fatal("directory accepted")
	}
	if _, err = c.WaitForAttachment(context.Background(), "w", "c", "a", 0, time.Second); err == nil {
		t.Fatal("invalid bounds")
	}
}
func TestAttachmentMessagesRequireFeatureKeyAndExactAcknowledgment(t *testing.T) {
	for _, mode := range []string{"feature", "key", "missing", "changed", "success"} {
		t.Run(mode, func(t *testing.T) {
			posts := 0
			c := attachmentClient(t, mode != "feature", func(w http.ResponseWriter, r *http.Request) {
				posts++
				if r.Header.Get("Idempotency-Key") != "stable" {
					t.Error("key dropped")
				}
				var body ChatMessageRequest
				if json.NewDecoder(r.Body).Decode(&body) != nil || len(body.AttachmentIDs.Value) != 1 {
					t.Error("files dropped")
				}
				ids := `["a"]`
				if mode == "missing" {
					ids = `null`
				}
				if mode == "changed" {
					ids = `["b"]`
				}
				fmt.Fprintf(w, `{"status":"accepted","turn_id":"turn","attachment_ids":%s}`, ids)
			})
			request := SendChatMessageParams{ChatID: "c", Body: ChatMessageRequest{Message: "read", AttachmentIDs: Present([]string{"a"})}}
			if mode != "key" {
				request.IdempotencyKey = Present("stable")
			}
			_, err := c.SendChatMessage(context.Background(), request)
			if mode == "success" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("unsafe acceptance")
			}
			if (mode == "feature" || mode == "key") && posts != 0 || posts > 1 {
				t.Fatal("mutation resubmitted", posts)
			}
		})
	}
}
func TestAttachmentPollingFailuresDeadlineAndCancellation(t *testing.T) {
	for _, name := range []string{"failed", "recoverable", "pending"} {
		t.Run(name, func(t *testing.T) {
			c := attachmentClient(t, true, func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{"data":%s}`, attachmentFixture(t, name))
			})
			timeout := 10 * time.Millisecond
			if name == "failed" {
				timeout = time.Second
			}
			result, err := c.WaitForAttachment(context.Background(), "w", "c", "a", time.Millisecond, timeout)
			if name == "failed" {
				if err != nil || !result.Conversion.Retryable {
					t.Fatal(result, err)
				}
			} else if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err = c.WaitForAttachment(ctx, "w", "c", "a", time.Millisecond, time.Second); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		})
	}
}

func TestRunTurnPreservesAttachmentsAndKeyAfterLiveSubscription(t *testing.T) {
	events := fixture[[]Event](t, "events-valid.json")
	sent := make(chan struct{})
	c := attachmentClient(t, true, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/messages") {
			if r.Header.Get("Idempotency-Key") != "stable-turn" {
				t.Error("message key dropped")
			}
			var request ChatMessageRequest
			if json.NewDecoder(r.Body).Decode(&request) != nil || !sameAttachmentIDs(request.AttachmentIDs.Value, []string{"a"}) {
				t.Error("attachment IDs dropped")
			}
			w.WriteHeader(202)
			fmt.Fprintf(w, `{"status":"accepted","turn_id":%q,"attachment_ids":["a"]}`, *events[1].TurnID)
			close(sent)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		writeEvent(w, events[0])
		w.(http.Flusher).Flush()
		select {
		case <-sent:
		case <-r.Context().Done():
			return
		}
		terminal := events[7]
		terminal.Seq = 1
		writeEvent(w, terminal)
	})
	result, err := c.RunTurn(context.Background(), events[0].ChatID, "Read", TurnOptions{AttachmentIDs: []string{"a"}, IdempotencyKey: "stable-turn"})
	if err != nil || !result.Terminal.Terminal() {
		t.Fatal(result, err)
	}
}

func TestRunTurnReplaysCompletedIdempotentAcceptance(t *testing.T) {
	events := fixture[[]Event](t, "events-valid.json")
	posts := 0
	c := attachmentClient(t, true, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/messages") {
			posts++
			fmt.Fprintf(w, `{"status":"accepted","turn_id":%q,"attachment_ids":["a"]}`, *events[1].TurnID)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		terminal := events[7]
		terminal.Seq = 1
		writeEvent(w, terminal)
		writeEvent(w, events[0])
	})
	result, err := c.RunTurn(context.Background(), events[0].ChatID, "Read", TurnOptions{AttachmentIDs: []string{"a"}, IdempotencyKey: "existing"})
	if err != nil || !result.Terminal.Terminal() || posts != 1 {
		t.Fatal(result, err, posts)
	}
}

func TestAttachmentMessageRejectsExplicitNullBeforeAnyRequest(t *testing.T) {
	calls := 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { calls++ })
	_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "c", Body: ChatMessageRequest{Message: "read", AttachmentIDs: Null[[]string]()}})
	if err == nil || calls != 0 {
		t.Fatal(err, calls)
	}
}
func TestAttachmentMessageAllowsEmptyArrayWithoutIdempotencyKey(t *testing.T) {
	posts := 0
	c := negotiatedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		posts++
		w.Header().Set(ContractHeader, "3")
		fmt.Fprintf(w, `{"status":"accepted","turn_id":%q}`, *fixture[[]Event](t, "events-valid.json")[1].TurnID)
	})
	_, err := c.SendChatMessage(context.Background(), SendChatMessageParams{ChatID: "c", Body: ChatMessageRequest{Message: "read", AttachmentIDs: Present([]string{})}})
	if err != nil || posts != 1 {
		t.Fatal(err, posts)
	}
}
