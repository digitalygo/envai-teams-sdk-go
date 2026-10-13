package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func wikiFixtureData(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/wiki/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var value struct{ Data json.RawMessage }
	if err = json.Unmarshal(b, &value); err != nil {
		t.Fatal(err)
	}
	return value.Data
}
func TestSourceUploadDurableStatusDownload(t *testing.T) {
	content := []byte("source data\nexact bytes")
	path := filepath.Join(t.TempDir(), "a report.txt")
	os.WriteFile(path, content, 0600)
	var posts atomic.Int32
	c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			writeCapabilities(w)
			return
		}
		if r.Header.Get(ContractHeader) != "3" {
			t.Error("selector")
		}
		w.Header().Set(ContractHeader, "3")
		if r.Method == "POST" {
			posts.Add(1)
			if r.Header.Get("Idempotency-Key") != "stable-key" {
				t.Error("idempotency missing")
			}
			if r.ContentLength >= int64(len(content)) {
				t.Error("streamed source became buffered")
			}
			reader, err := r.MultipartReader()
			if err != nil {
				t.Error(err)
				return
			}
			part, err := reader.NextPart()
			if err != nil {
				t.Error(err)
				return
			}
			data, err := io.ReadAll(part)
			if err != nil || !bytes.Equal(data, content) || part.FormName() != "file" || part.FileName() != "a report.txt" || part.Header.Get("Content-Type") != "text/plain" {
				t.Error(err, string(data), part)
			}
			if _, err = reader.NextPart(); !errors.Is(err, io.EOF) {
				t.Error("unexpected part", err)
			}
			w.WriteHeader(202)
		}
		if strings.HasSuffix(r.URL.Path, "/download") {
			w.Write(content)
			return
		}
		w.Write(append(append([]byte(`{"data":`), wikiFixtureData(t, "source-accepted")...), '}'))
	})
	source, err := c.UploadSourceFile(context.Background(), "workspace", path, "stable-key", "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	if source.Conversion.Status != "queued" || source.Indexing.Status != "pending_conversion" || source.UsageCost.Totals != nil {
		t.Fatal(source)
	}
	if _, err = c.GetWikiSource(context.Background(), GetWikiSourceParams{WorkspaceID: "workspace", FileID: source.FileID}); err != nil {
		t.Fatal(err)
	}
	download, err := c.DownloadWikiSource(context.Background(), DownloadWikiSourceParams{WorkspaceID: "workspace", FileID: source.FileID})
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(download.Body)
	download.Close()
	if err != nil || !bytes.Equal(data, content) {
		t.Fatal(err, string(data))
	}
	if posts.Load() != 1 {
		t.Fatal(posts.Load())
	}
}
func TestWikiRequestValidationAndNoMutationRetry(t *testing.T) {
	var mutations atomic.Int32
	c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			writeCapabilities(w)
			return
		}
		mutations.Add(1)
		w.WriteHeader(503)
	})
	c.config.RetryAttempts = 5
	for _, key := range []string{"", "space key", "bad\nkey", strings.Repeat("x", 256)} {
		if _, err := c.CreateWikiInboxRun(context.Background(), CreateWikiInboxRunParams{WorkspaceID: "w", IdempotencyKey: key, Body: CreateWikiInboxRunBody{WebhookURL: "https://receiver.test/wiki"}}); err == nil {
			t.Fatal("invalid key", key)
		}
	}
	for _, url := range []string{"", "http://receiver.test", "https://user:pass@receiver.test", "https://receiver.test/#fragment", ":bad"} {
		if _, err := c.CreateWikiSyncRun(context.Background(), CreateWikiSyncRunParams{WorkspaceID: "w", IdempotencyKey: "key", Body: CreateWikiSyncRunBody{WebhookURL: url}}); err == nil {
			t.Fatal("invalid webhook", url)
		}
	}
	if _, err := c.WithdrawWikiSource(context.Background(), WithdrawWikiSourceParams{WorkspaceID: "w"}); err == nil {
		t.Fatal("missing source")
	}
	if mutations.Load() != 0 {
		t.Fatal("invalid request sent")
	}
	for _, parts := range [][]MultipartPart{nil, {{Name: "wrong", Filename: "x"}}, {{Name: "file"}}, {{Name: "file", Filename: "x", Reader: strings.NewReader("x"), Size: -1}}, {{Name: "file", Filename: "x", Reader: strings.NewReader("x"), Size: MaxWikiSourceBytes + 1}}, {{Name: "file", Filename: "x", Reader: strings.NewReader("x"), Data: []byte("x"), Size: 1}}} {
		if _, err := c.UploadWikiSource(context.Background(), UploadWikiSourceParams{WorkspaceID: "w", IdempotencyKey: "key", Body: parts}); err == nil {
			t.Fatal("invalid upload", parts)
		}
	}
	if _, err := c.CreateWikiOutboxRun(context.Background(), CreateWikiOutboxRunParams{WorkspaceID: "w", IdempotencyKey: "key", Body: CreateWikiOutboxRunBody{WebhookURL: "https://receiver.test/wiki"}}); err == nil {
		t.Fatal("503 lost")
	}
	if mutations.Load() != 1 {
		t.Fatal("mutation replayed", mutations.Load())
	}
	if err := validateWikiBody(operation{ID: "uploadWikiSource"}, []MultipartPart{{Name: "file", Filename: "x", Reader: strings.NewReader(""), Size: MaxWikiSourceBytes}}); err != nil {
		t.Fatal("inclusive limit rejected", err)
	}
	if _, err := c.UploadSourceFile(context.Background(), "w", "/nonexistent/source", "key", ""); err == nil {
		t.Fatal("missing file")
	}
	if _, err := c.UploadSourceFile(context.Background(), "w", t.TempDir(), "key", ""); err == nil {
		t.Fatal("directory")
	}
	path := filepath.Join(t.TempDir(), "large")
	f, _ := os.Create(path)
	f.Truncate(MaxWikiSourceBytes + 1)
	f.Close()
	if _, err := c.UploadSourceFile(context.Background(), "w", path, "key", ""); err == nil {
		t.Fatal("large original")
	}
}
func TestWikiExactCatalogQuery(t *testing.T) {
	var queries []string
	c := rawClient(t, func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		if !strings.Contains(r.URL.EscapedPath(), "category%20x%25") {
			t.Error(r.URL.EscapedPath())
		}
		w.Header().Set(ContractHeader, "3")
		w.Write(append(append([]byte(`{"data":`), wikiFixtureData(t, "category-index")...), '}'))
	})
	req := GetWikiCategoryIndexParams{WorkspaceID: "w", CategoryID: "category x%", From: Present(int64(1)), To: Present(int64(25)), Type: Present("raw"), CatalogRevision: Present(int64(9))}
	if _, err := c.GetWikiCategoryIndex(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if queries[0] != "catalog_revision=9&from=1&to=25&type=raw" {
		t.Fatal(queries)
	}
	req.From = Optional[int64]{}
	req.To = Optional[int64]{}
	req.All = Present(true)
	req.Type = Present("artifact")
	if _, err := c.GetWikiCategoryIndex(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []GetWikiCategoryIndexParams{{}, {All: Present(true), From: Present(int64(1))}, {From: Present(int64(0)), To: Present(int64(2))}, {From: Present(int64(3)), To: Present(int64(2))}, {All: Present(true), Type: Present("other")}, {All: Present(true), CatalogRevision: Present(int64(-1))}} {
		bad.WorkspaceID = "w"
		bad.CategoryID = "c"
		if _, err := c.GetWikiCategoryIndex(context.Background(), bad); err == nil {
			t.Fatal("bad catalog query")
		}
	}
	if len(queries) != 2 {
		t.Fatal(queries)
	}
}
func TestStreamingMultipartBoundariesAndFailures(t *testing.T) {
	parts := []MultipartPart{{Name: "a", Data: []byte("one")}, {Name: "file", Filename: `a"b.txt`, Reader: strings.NewReader("two"), Size: 3}}
	reader, contentType, err := multipartReader(parts)
	if err != nil {
		t.Fatal(err)
	}
	boundary := strings.TrimPrefix(contentType, "multipart/form-data; boundary=")
	mr := multipart.NewReader(reader, boundary)
	for _, want := range []string{"one", "two"} {
		part, err := mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(part)
		if err != nil || string(b) != want {
			t.Fatal(string(b), err)
		}
	}
	if _, err = mr.NextPart(); !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	reader, _, err = multipartReader([]MultipartPart{{Name: "file", Reader: strings.NewReader("x"), Size: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.ReadAll(reader); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal(err)
	}
	for _, p := range []MultipartPart{{Name: ""}, {Name: "bad\n"}, {Name: "x", Filename: "bad\x00"}, {Name: "x", Reader: strings.NewReader("x"), Size: -1}} {
		if _, _, err = multipartReader([]MultipartPart{p}); err == nil {
			t.Fatal("bad multipart")
		}
	}
}

type wikiRoundTripper func(*http.Request) (*http.Response, error)

func (f wikiRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestMutationHasNoTransportReplayBody(t *testing.T) {
	calls := 0
	client, err := New(Config{BaseURL: "https://api.example", APIKey: "test", HTTPClient: &http.Client{Transport: wikiRoundTripper(func(r *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Set(ContractHeader, "3")
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			return &http.Response{StatusCode: 200, Header: h, Body: io.NopCloser(strings.NewReader(testCapabilities))}, nil
		}
		calls++
		if r.GetBody != nil {
			t.Error("automatic net/http mutation replay remains possible")
		}
		if r.Header.Get("Idempotency-Key") != "stable" {
			t.Error("missing explicit replay identity")
		}
		return nil, io.ErrUnexpectedEOF
	})}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.CreateWikiInboxRun(context.Background(), CreateWikiInboxRunParams{WorkspaceID: "workspace", IdempotencyKey: "stable", Body: CreateWikiInboxRunBody{WebhookURL: "https://receiver.example/wiki"}}); err == nil {
		t.Fatal("ambiguous outcome lost")
	}
	if calls != 1 {
		t.Fatal(calls)
	}
}
