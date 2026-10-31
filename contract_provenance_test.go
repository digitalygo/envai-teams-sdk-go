package envai

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

func TestContractSnapshotsCaptureNativeAndNotRequired(t *testing.T) {
	var openapi struct {
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	raw, err := os.ReadFile("contract/openapi.json")
	if err != nil || json.Unmarshal(raw, &openapi) != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Properties map[string]struct {
			Type string   `json:"type"`
			Enum []string `json:"enum"`
		} `json:"properties"`
	}
	if err = json.Unmarshal(openapi.Components.Schemas["AttachmentManifest"], &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Properties["mime_type"].Type != "string" || len(manifest.Properties["mode"].Enum) != 1 || manifest.Properties["mode"].Enum[0] != "native" {
		t.Fatal(manifest)
	}
	var conversion struct {
		Properties map[string]struct {
			Enum []string `json:"enum"`
		} `json:"properties"`
	}
	if err = json.Unmarshal(openapi.Components.Schemas["AttachmentConversion"], &conversion); err != nil || !enumContains(conversion.Properties["status"].Enum, "not_required") {
		t.Fatal(err, conversion)
	}
	var wiki struct {
		Defs map[string]json.RawMessage `json:"$defs"`
	}
	raw, err = os.ReadFile("contract/wiki.json")
	if err != nil || json.Unmarshal(raw, &wiki) != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(wiki.Defs["ConversionState"], &conversion); err != nil || !enumContains(conversion.Properties["status"].Enum, "not_required") {
		t.Fatal(err, conversion)
	}
}

func enumContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestContractProvenanceAndFrozenV2(t *testing.T) {
	raw, err := os.ReadFile("contract/provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	var provenance struct {
		Repository string `json:"repository"`
		Commit     string `json:"commit"`
		Operations int    `json:"operations"`
		Sources    map[string]struct {
			Bytes        int    `json:"bytes"`
			SHA256       string `json:"sha256"`
			GitBlobSHA   string `json:"git_blob_sha"`
			SnapshotPath string `json:"snapshot_path"`
		} `json:"sources"`
		Generated map[string]struct {
			SHA256       string `json:"sha256"`
			SnapshotPath string `json:"snapshot_path"`
		} `json:"generated_fixtures"`
	}
	if err = json.Unmarshal(raw, &provenance); err != nil {
		t.Fatal(err)
	}
	if provenance.Repository != "digitalygo/envai-teams" || provenance.Commit != "20dfd5a6c8ec9e75d3754875d4f94520feb3bbf6" || provenance.Operations != 76 || len(provenance.Sources) != 6 {
		t.Fatal(provenance.Repository, provenance.Commit, provenance.Operations, len(provenance.Sources))
	}
	for name, entry := range provenance.Sources {
		data, err := os.ReadFile(entry.SnapshotPath)
		if err != nil {
			t.Fatal(name, err)
		}
		sum := sha256.Sum256(data)
		blob := sha1.Sum(append([]byte("blob "+strconv.Itoa(len(data))+"\x00"), data...))
		if len(data) != entry.Bytes || hex.EncodeToString(sum[:]) != entry.SHA256 || hex.EncodeToString(blob[:]) != entry.GitBlobSHA {
			t.Fatal(name)
		}
	}
	if provenance.Sources["events-v2.json"].GitBlobSHA != "02b13b23cd06eacd81059d1ed99e0027769d75e4" {
		t.Fatal("frozen v2 source changed")
	}
	for name, entry := range provenance.Generated {
		data, err := os.ReadFile(entry.SnapshotPath)
		if err != nil {
			t.Fatal(name, err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != entry.SHA256 {
			t.Fatal(name)
		}
	}
}
