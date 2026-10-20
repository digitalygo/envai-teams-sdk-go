package envai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func canonicalCapabilities(t *testing.T) json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile("testdata/capabilities/canonical-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Body json.RawMessage `json:"body"`
	}
	if err = json.Unmarshal(raw, &fixture); err != nil || len(fixture.Body) == 0 {
		t.Fatal(err, string(raw))
	}
	return fixture.Body
}

func capabilitiesWithFeatures(t *testing.T, features string) string {
	t.Helper()
	return strings.Replace(testCapabilities, `"features":[]`, `"features":`+features, 1)
}

func sourceOriginalResponse(t *testing.T) []byte {
	t.Helper()
	var data map[string]any
	if err := json.Unmarshal(wikiFixtureData(t, "source-accepted"), &data); err != nil {
		t.Fatal(err)
	}
	data["lifecycle"] = "inbox"
	data["conversion"] = map[string]any{"status": "not_required"}
	data["indexing"] = map[string]any{"status": "pending"}
	response, err := json.Marshal(map[string]any{"data": data})
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func TestCanonicalCapabilitiesProvenanceAndNativeMutations(t *testing.T) {
	rawFixture, err := os.ReadFile("testdata/capabilities/canonical-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	var rawProvenance []byte
	rawProvenance, err = os.ReadFile("testdata/capabilities/provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	var provenance struct {
		SourceRepository string `json:"source_repository"`
		SourcePath       string `json:"source_path"`
		SourceCommit     string `json:"source_commit"`
		CapturedFrom     string `json:"captured_from"`
		Parity           struct {
			Repository string `json:"repository"`
			Commit     string `json:"commit"`
			Path       string `json:"path"`
			SHA256     string `json:"sha256"`
		} `json:"parity"`
		Files map[string]string `json:"files"`
	}
	if err = json.Unmarshal(rawProvenance, &provenance); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(rawFixture)
	digest := hex.EncodeToString(sum[:])
	if provenance.SourceRepository != "digitalygo/envai-teams" || provenance.SourcePath != "apps/execution-plane/internal/gateway/capabilities.go" || provenance.SourceCommit != "7192fc05ebec817bddc705186c8f195cc0dc5ec5" || provenance.CapturedFrom != "GET /api/v1/atp/v3/capabilities" {
		t.Fatal(provenance)
	}
	if provenance.Parity.Repository != "digitalygo/envai-teams-sdk-php" || provenance.Parity.Commit != "93eafc02271035065ac975f91cd4f77d148c9d2c" || provenance.Parity.Path != "tests/fixtures/capabilities/canonical-v3.json" || provenance.Parity.SHA256 != digest || provenance.Files["canonical-v3.json"] != digest {
		t.Fatal(provenance)
	}
	var capabilities Capabilities
	body := canonicalCapabilities(t)
	if err = json.Unmarshal(body, &capabilities); err != nil {
		t.Fatal(err)
	}
	if capabilities.ProtocolVersion != 3 || capabilities.ContractVersion != 3 || !featureAdvertised(capabilities.Features, "chat_attachments") || !featureAdvertised(capabilities.Features, "chat_attachment_originals") || !featureAdvertised(capabilities.Features, "wiki_original_sources") {
		t.Fatal(capabilities)
	}

	content := []byte("canonical native\000bytes\n")
	path := filepath.Join(t.TempDir(), "native.bin")
	if err = os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	var methods []string
	var preflight string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			preflight = r.URL.Path
			w.Write(body)
			return
		}
		if r.Header.Get("Idempotency-Key") != "canonical-native-key" {
			t.Error("idempotency key")
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
		uploaded, err := io.ReadAll(part)
		if err != nil || part.FormName() != "file" || part.FileName() != "native.bin" || !bytes.Equal(uploaded, content) {
			t.Error(err, part.FormName(), part.FileName(), len(uploaded))
		}
		if _, err = reader.NextPart(); err != io.EOF {
			t.Error("unexpected attachment part", err)
		}
		w.WriteHeader(202)
		fmt.Fprintf(w, `{"data":%s}`, attachmentFixture(t, "original-ready"))
	})
	attachment, err := c.UploadAttachmentFile(context.Background(), "wsp_canonical", "cht_canonical", path, "canonical-native-key", "application/octet-stream")
	if err != nil {
		t.Fatal(err)
	}
	if attachment.Lifecycle != "ready" || attachment.Conversion.Status != "not_required" || attachment.Derived != nil {
		t.Fatal(attachment)
	}
	if strings.Join(methods, ",") != "GET,POST" || preflight != "/api/v1/atp/v3/capabilities" {
		t.Fatal(methods, preflight)
	}

	content = []byte("canonical wiki original\000bytes\n")
	path = filepath.Join(t.TempDir(), "source.txt")
	if err = os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	methods = nil
	preflight = ""
	response := sourceOriginalResponse(t)
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			preflight = r.URL.Path
			w.Write(body)
			return
		}
		if r.Header.Get("Idempotency-Key") != "canonical-wiki-key" {
			t.Error("idempotency key")
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
		if err != nil || processing.FormName() != "processing" || processing.FileName() != "" || string(value) != "original" {
			t.Error(err, processing.FormName(), processing.FileName(), string(value))
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			return
		}
		uploaded, err := io.ReadAll(part)
		if err != nil || part.FormName() != "file" || part.FileName() != "source.txt" || !bytes.Equal(uploaded, content) {
			t.Error(err, part.FormName(), part.FileName(), len(uploaded))
		}
		if _, err = reader.NextPart(); err != io.EOF {
			t.Error("unexpected wiki part", err)
		}
		w.WriteHeader(202)
		w.Write(response)
	})
	source, err := c.UploadSourceFile(context.Background(), "wsp_canonical", path, "canonical-wiki-key", "text/plain", "original")
	if err != nil {
		t.Fatal(err)
	}
	if source.Conversion.Status != "not_required" || source.Indexing.Status != "pending" || source.Lifecycle != "inbox" {
		t.Fatal(source)
	}
	if strings.Join(methods, ",") != "GET,POST" || preflight != "/api/v1/atp/v3/capabilities" {
		t.Fatal(methods, preflight)
	}
}

func TestNativeAttachmentFeatureMatrixFailsClosed(t *testing.T) {
	cases := []struct {
		name     string
		features string
		ok       bool
	}{
		{"canonical list", `["chat_attachments","chat_attachment_originals"]`, true},
		{"boolean map", `{"chat_attachments":true,"chat_attachment_originals":true}`, true},
		{"unrelated list name", `["chat_attachments","chat_attachment_originals","future"]`, true},
		{"unrelated false map entry", `{"chat_attachments":true,"chat_attachment_originals":true,"future":false}`, true},
		{"missing originals list", `["chat_attachments"]`, false},
		{"missing originals map", `{"chat_attachments":true}`, false},
		{"false originals map", `{"chat_attachments":true,"chat_attachment_originals":false}`, false},
		{"false chat map", `{"chat_attachments":false,"chat_attachment_originals":true}`, false},
		{"duplicate originals enabled then disabled", `{"chat_attachments":true,"chat_attachment_originals":true,"chat_attachment_originals":false}`, false},
		{"duplicate originals disabled then enabled", `{"chat_attachments":true,"chat_attachment_originals":false,"chat_attachment_originals":true}`, false},
		{"duplicate originals disabled", `{"chat_attachments":true,"chat_attachment_originals":false,"chat_attachment_originals":false}`, false},
		{"duplicate chat enabled then disabled", `{"chat_attachments":true,"chat_attachments":false,"chat_attachment_originals":true}`, false},
		{"unknown originals map", `{"chat_attachments":true,"chat_attachment_originals_beta":true}`, false},
		{"unknown originals list", `["chat_attachments","chat_attachment_originals_beta"]`, false},
		{"nonboolean originals", `{"chat_attachments":true,"chat_attachment_originals":"true"}`, false},
		{"mixed list", `["chat_attachments","chat_attachment_originals",42]`, false},
		{"nonstring list", `["chat_attachments",42]`, false},
		{"null features", `null`, false},
		{"string features", `"chat_attachments"`, false},
		{"null unrelated map value", `{"chat_attachments":true,"chat_attachment_originals":true,"future":null}`, false},
	}
	path := filepath.Join(t.TempDir(), "native.bin")
	if err := os.WriteFile(path, []byte("bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gets, posts := 0, 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					gets++
					fmt.Fprint(w, capabilitiesWithFeatures(t, tc.features))
					return
				}
				posts++
				w.WriteHeader(202)
				fmt.Fprintf(w, `{"data":%s}`, attachmentFixture(t, "original-ready"))
			})
			_, err := c.UploadAttachmentFile(context.Background(), "w", "c", path, "key", "text/plain")
			if (err == nil) != tc.ok {
				t.Fatal(err, tc.ok)
			}
			if gets != 1 || (tc.ok && posts != 1) || (!tc.ok && posts != 0) {
				t.Fatal(gets, posts)
			}
		})
	}
}

func TestWikiOriginalFeatureMatrixFailsClosed(t *testing.T) {
	cases := []struct {
		name     string
		features string
		ok       bool
	}{
		{"list original", `["wiki_original_sources"]`, true},
		{"map original", `{"wiki_original_sources":true}`, true},
		{"missing original list", `["chat_attachments","chat_attachment_originals"]`, false},
		{"missing original map", `{"chat_attachments":true,"chat_attachment_originals":true}`, false},
		{"false original map", `{"wiki_original_sources":false}`, false},
		{"duplicate original enabled then disabled", `{"wiki_original_sources":true,"wiki_original_sources":false}`, false},
		{"duplicate original disabled then enabled", `{"wiki_original_sources":false,"wiki_original_sources":true}`, false},
		{"duplicate original disabled", `{"wiki_original_sources":false,"wiki_original_sources":false}`, false},
		{"nonboolean original", `{"wiki_original_sources":"true"}`, false},
		{"unknown original map", `{"wiki_original_sources_beta":true}`, false},
		{"unknown original list", `["wiki_original_sources_beta"]`, false},
		{"mixed list", `["wiki_original_sources",42]`, false},
		{"null unrelated map value", `{"wiki_original_sources":true,"future":null}`, false},
	}
	path := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(path, []byte("bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gets, posts := 0, 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					gets++
					fmt.Fprint(w, capabilitiesWithFeatures(t, tc.features))
					return
				}
				posts++
				w.WriteHeader(202)
				w.Write(sourceOriginalResponse(t))
			})
			_, err := c.UploadSourceFile(context.Background(), "w", path, "key", "text/plain", "original")
			if (err == nil) != tc.ok {
				t.Fatal(err, tc.ok)
			}
			if gets != 1 || (tc.ok && posts != 1) || (!tc.ok && posts != 0) {
				t.Fatal(gets, posts)
			}
		})
	}
}

func TestCapabilitiesFeaturesNormalizeStrictly(t *testing.T) {
	cases := []struct {
		name   string
		value  string
		expect string
		ok     bool
	}{
		{"list", `["chat_attachments","chat_attachment_originals"]`, "chat_attachments,chat_attachment_originals", true},
		{"map", `{"chat_attachments":true,"chat_attachment_originals":true,"future":false}`, "chat_attachments,chat_attachment_originals", true},
		{"empty list", `[]`, "", true},
		{"empty map", `{}`, "", true},
		{"null", `null`, "", false},
		{"string", `"chat_attachments"`, "", false},
		{"mixed list", `["chat_attachments",42]`, "", false},
		{"duplicate list", `["chat_attachments","chat_attachment_originals","chat_attachment_originals"]`, "chat_attachments,chat_attachment_originals,chat_attachment_originals", true},
		{"nonboolean map", `{"chat_attachments":1}`, "", false},
		{"null map value", `{"chat_attachments":null}`, "", false},
		{"duplicate map enabled then disabled", `{"chat_attachments":true,"chat_attachment_originals":true,"chat_attachment_originals":false}`, "", false},
		{"duplicate map disabled then enabled", `{"chat_attachments":true,"chat_attachment_originals":false,"chat_attachment_originals":true}`, "", false},
		{"duplicate map disabled", `{"chat_attachments":true,"chat_attachment_originals":false,"chat_attachment_originals":false}`, "", false},
		{"duplicate map enabled", `{"chat_attachments":true,"chat_attachment_originals":true,"chat_attachment_originals":true}`, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			names, err := normalizeFeatureNames(json.RawMessage(tc.value))
			if (err == nil) != tc.ok || strings.Join(names, ",") != tc.expect {
				t.Fatal(names, err)
			}
		})
	}
	var legacy CapabilitiesLegacy
	if err := json.Unmarshal([]byte(`{"protocol_version":2,"features":{"background_delegation":true,"ping":true}}`), &legacy); err != nil || strings.Join(legacy.Features, ",") != "background_delegation,ping" {
		t.Fatal(legacy, err)
	}
	if err := json.Unmarshal([]byte(`{"protocol_version":2,"features":[1]}`), &legacy); err == nil {
		t.Fatal("legacy mixed features accepted")
	}
}
