package envai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func wikiSourceFeatures(t *testing.T, features string) *Client {
	t.Helper()
	return testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, capabilitiesWithFeatures(t, features))
			return
		}
		w.WriteHeader(202)
		w.Write(sourceOriginalResponse(t))
	})
}

func TestUploadSourceFileOriginalAndDefaultWire(t *testing.T) {
	content := []byte("source\000bytes\n")
	path := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		options []string
		parts   int
	}{
		{"default", nil, 1},
		{"explicit convert", []string{"convert"}, 1},
		{"original", []string{"original"}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parts := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					fmt.Fprint(w, capabilitiesWithFeatures(t, `["wiki_original_sources"]`))
					return
				}
				if r.Header.Get("Idempotency-Key") != "stable-key" {
					t.Error("idempotency key")
				}
				reader, err := r.MultipartReader()
				if err != nil {
					t.Error(err)
					return
				}
				if tc.parts == 2 {
					processing, err := reader.NextPart()
					if err != nil {
						t.Error(err)
						return
					}
					value, err := io.ReadAll(processing)
					if err != nil || processing.FormName() != "processing" || processing.FileName() != "" || string(value) != "original" {
						t.Error(err, processing.FormName(), processing.FileName(), string(value))
					}
				}
				part, err := reader.NextPart()
				if err != nil {
					t.Error(err)
					return
				}
				uploaded, err := io.ReadAll(part)
				if err != nil || part.FormName() != "file" || part.FileName() != "source.txt" || part.Header.Get("Content-Type") != "text/plain" || !bytes.Equal(uploaded, content) {
					t.Error(err, part.FormName(), part.FileName(), len(uploaded))
				}
				if _, err = reader.NextPart(); err != io.EOF {
					t.Error("unexpected part", err)
				}
				parts++
				w.WriteHeader(202)
				if tc.parts == 2 {
					w.Write(sourceOriginalResponse(t))
				} else {
					fmt.Fprintf(w, `{"data":%s}`, wikiFixtureData(t, "source-accepted"))
				}
			})
			source, err := c.UploadSourceFile(context.Background(), "w", path, "stable-key", "text/plain", tc.options...)
			if err != nil || parts != 1 {
				t.Fatal(source, err, parts)
			}
			if tc.parts == 2 && source.Conversion.Status != "not_required" {
				t.Fatal(source)
			}
		})
	}
	if _, err := wikiSourceFeatures(t, `["wiki_original_sources"]`).UploadSourceFile(context.Background(), "w", path, "key", "text/plain", "ocr"); err == nil {
		t.Fatal("invalid processing accepted")
	}
	if _, err := wikiSourceFeatures(t, `["wiki_original_sources"]`).UploadSourceFile(context.Background(), "w", path, "key", "text/plain", "original", "original"); err == nil {
		t.Fatal("duplicate options accepted")
	}
	if _, err := wikiSourceFeatures(t, `["wiki_original_sources"]`).UploadSourceFile(context.Background(), "w", path, "key", "text/plain", "ORIGINAL"); err == nil {
		t.Fatal("uppercase processing accepted")
	}
}

func TestUploadSourceFileInvalidProcessingDoesNotReachTransport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(path, []byte("bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, options := range [][]string{{"ocr"}, {"original", "original"}, {"ORIGINAL"}, {""}} {
		requests := 0
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { requests++ })
		if _, err := c.UploadSourceFile(context.Background(), "w", path, "key", "text/plain", options...); err == nil {
			t.Fatal("invalid processing accepted", options)
		}
		if requests != 0 {
			t.Fatal("invalid processing reached transport", requests)
		}
	}
}

func TestUploadWikiSourceProcessingValidation(t *testing.T) {
	file := MultipartPart{Name: "file", Filename: "source.txt", Data: []byte("bytes")}
	cases := [][]MultipartPart{
		nil,
		{{Name: "processing", Data: []byte("original")}},
		{{Name: "processing", Data: []byte("ocr")}, file},
		{{Name: "processing", Data: []byte("original")}, {Name: "processing", Data: []byte("convert")}, file},
		{{Name: "processing", Data: []byte("original"), Filename: "field.txt"}, file},
		{{Name: "processing", Reader: strings.NewReader("original")}, file},
		{{Name: "unexpected", Data: []byte("x")}, file},
		{file, file},
		{{Name: "file", Filename: "source.txt", Data: []byte("x"), Reader: strings.NewReader("x"), Size: 1}},
	}
	for _, parts := range cases {
		requests := 0
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { requests++ })
		if _, err := c.UploadWikiSource(context.Background(), UploadWikiSourceParams{WorkspaceID: "w", IdempotencyKey: "key", Body: parts}); err == nil {
			t.Fatal("invalid source parts accepted", parts)
		}
		if requests != 0 {
			t.Fatal("invalid source parts reached transport", requests)
		}
	}
	requests := 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { requests++ })
	if _, err := c.UploadChatAttachment(context.Background(), UploadChatAttachmentParams{WorkspaceID: "w", ChatID: "c", IdempotencyKey: "key", Body: []MultipartPart{{Name: "processing", Data: []byte("original")}, file}}); err == nil {
		t.Fatal("processing accepted for chat attachment")
	}
	if _, err := c.UploadImage(context.Background(), UploadImageParams{Body: []MultipartPart{{Name: "processing", Data: []byte("original")}, file}}); err == nil {
		t.Fatal("processing accepted for image upload")
	}
	if requests != 0 {
		t.Fatal("source-only processing reached transport", requests)
	}
}

func TestLowLevelOriginalSourceFeatureGateAndExactBytes(t *testing.T) {
	content := []byte("low-level\000bytes\n")
	parts := []MultipartPart{{Name: "processing", Data: []byte("original")}, {Name: "file", Filename: "source.bin", Data: content}}
	gets, posts := 0, 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			gets++
			fmt.Fprint(w, capabilitiesWithFeatures(t, `[]`))
			return
		}
		posts++
	})
	if _, err := c.UploadWikiSource(context.Background(), UploadWikiSourceParams{WorkspaceID: "w", IdempotencyKey: "key", Body: parts}); err == nil {
		t.Fatal("original upload bypassed capability")
	}
	if gets != 1 || posts != 0 {
		t.Fatal(gets, posts)
	}
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			gets++
			fmt.Fprint(w, capabilitiesWithFeatures(t, `{"wiki_original_sources":true}`))
			return
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Error(err)
			return
		}
		processing, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			return
		}
		value, err := io.ReadAll(processing)
		if err != nil || processing.FormName() != "processing" || string(value) != "original" {
			t.Error(err, processing.FormName(), string(value))
		}
		file, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			return
		}
		uploaded, err := io.ReadAll(file)
		if err != nil || file.FormName() != "file" || file.FileName() != "source.bin" || !bytes.Equal(uploaded, content) {
			t.Error(err, file.FormName(), file.FileName(), len(uploaded))
		}
		posts++
		w.WriteHeader(202)
		w.Write(sourceOriginalResponse(t))
	})
	source, err := c.UploadWikiSource(context.Background(), UploadWikiSourceParams{WorkspaceID: "w", IdempotencyKey: "key", Body: parts})
	if err != nil || source.Conversion.Status != "not_required" || posts != 1 {
		t.Fatal(source, err, posts)
	}
	gets, posts = 0, 0
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			gets++
			fmt.Fprint(w, capabilitiesWithFeatures(t, `[]`))
			return
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Error(err)
			return
		}
		processing, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			return
		}
		value, err := io.ReadAll(processing)
		if err != nil || string(value) != "convert" {
			t.Error(err, string(value))
		}
		file, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			return
		}
		if _, err = io.ReadAll(file); err != nil {
			t.Error(err)
		}
		posts++
		w.WriteHeader(202)
		fmt.Fprintf(w, `{"data":%s}`, wikiFixtureData(t, "source-accepted"))
	})
	if _, err = c.UploadWikiSource(context.Background(), UploadWikiSourceParams{WorkspaceID: "w", IdempotencyKey: "key", Body: []MultipartPart{{Name: "processing", Data: []byte("convert")}, {Name: "file", Filename: "source.bin", Data: content}}}); err != nil || gets != 1 || posts != 1 {
		t.Fatal(err, gets, posts)
	}
}

func TestOriginalSourceNoAutomaticRetryAndReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(path, []byte("bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	posts := 0
	var keys []string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, capabilitiesWithFeatures(t, `["wiki_original_sources"]`))
			return
		}
		posts++
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		if posts == 1 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(202)
		w.Write(sourceOriginalResponse(t))
	})
	c.config.RetryAttempts = 5
	if _, err := c.UploadSourceFile(context.Background(), "w", path, "same-key", "text/plain", "original"); err == nil {
		t.Fatal("503 lost")
	}
	if posts != 1 {
		t.Fatal("mutation retried", posts)
	}
	source, err := c.UploadSourceFile(context.Background(), "w", path, "same-key", "text/plain", "original")
	if err != nil || source.Conversion.Status != "not_required" || posts != 2 {
		t.Fatal(source, err, posts)
	}
	if strings.Join(keys, ",") != "same-key,same-key" {
		t.Fatal(keys)
	}
}

func TestWikiSourceNotRequiredHydrates(t *testing.T) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal(wikiFixtureData(t, "source-accepted"), &data); err != nil {
		t.Fatal(err)
	}
	data["conversion"] = json.RawMessage(`{"status":"not_required"}`)
	data["indexing"] = json.RawMessage(`{"status":"pending"}`)
	data["lifecycle"] = json.RawMessage(`"inbox"`)
	raw, _ := json.Marshal(data)
	var source WikiSource
	if err := json.Unmarshal(raw, &source); err != nil || source.Conversion.Status != "not_required" || source.Indexing.Status != "pending" || source.Lifecycle != "inbox" {
		t.Fatal(source, err)
	}
	data["conversion"] = json.RawMessage(`{"status":"skipped"}`)
	raw, _ = json.Marshal(data)
	if json.Unmarshal(raw, &source) == nil {
		t.Fatal("invalid conversion status accepted")
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.Write(sourceOriginalResponse(t)) })
	out, err := c.GetWikiSource(context.Background(), GetWikiSourceParams{WorkspaceID: "w", FileID: "f"})
	if err != nil || out.Conversion.Status != "not_required" || out.Indexing.Status != "pending" {
		t.Fatal(out, err)
	}
}

func TestChatAttachmentFileOnlyValidation(t *testing.T) {
	file := MultipartPart{Name: "file", Filename: "source.txt", Data: []byte("bytes")}
	for _, parts := range [][]MultipartPart{
		nil,
		{{Name: "other", Filename: "source.txt", Data: []byte("bytes")}},
		{{Name: "file", Data: []byte("bytes")}},
		{{Name: "file", Filename: "source.txt", Reader: strings.NewReader("x"), Size: -1}},
		{{Name: "file", Filename: "source.txt", Reader: strings.NewReader("x"), Data: []byte("x"), Size: 1}},
		{{Name: "file", Filename: "source.txt", Reader: strings.NewReader("x"), Size: MaxWikiSourceBytes + 1}},
		{file, file},
	} {
		if err := validateWikiBody(operation{ID: "uploadChatAttachment"}, parts); err == nil {
			t.Fatal("invalid attachment parts accepted", parts)
		}
	}
	if err := validateWikiBody(operation{ID: "uploadChatAttachment"}, []MultipartPart{file}); err != nil {
		t.Fatal(err)
	}
}
