package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	envai "github.com/digitalygo/envai-teams-sdk-go"
)

func TestWikiExampleOffline(t *testing.T) {
	for _, scenario := range []string{"pending", "ready", "uploadFailure", "statusFailure", "runFailure"} {
		t.Run(scenario, func(t *testing.T) {
			sourceBytes, _ := os.ReadFile("../../testdata/wiki/source-accepted.json")
			var source struct {
				Data map[string]any `json:"data"`
			}
			json.Unmarshal(sourceBytes, &source)
			operationBytes, _ := os.ReadFile("../../testdata/wiki/operation-completed.json")
			var operation struct {
				Data any `json:"data"`
			}
			json.Unmarshal(operationBytes, &operation)
			runs := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(envai.ContractHeader, "3")
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					fmt.Fprint(w, `{"protocol_version":3,"contract_version":3,"supported_contract_versions":[2,3],"legacy_v2_mode":"historical_reads_replay_only"}`)
					return
				}
				if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/sources") {
					if scenario == "uploadFailure" {
						w.WriteHeader(503)
						return
					}
					w.WriteHeader(202)
					json.NewEncoder(w).Encode(source)
					return
				}
				if r.Method == "GET" {
					if scenario == "statusFailure" {
						w.WriteHeader(503)
						return
					}
					if scenario == "ready" || scenario == "runFailure" {
						source.Data["conversion"].(map[string]any)["status"] = "completed"
						source.Data["lifecycle"] = "inbox"
					}
					json.NewEncoder(w).Encode(source)
					return
				}
				runs++
				if scenario == "runFailure" {
					w.WriteHeader(503)
					return
				}
				w.WriteHeader(202)
				json.NewEncoder(w).Encode(operation)
			}))
			defer server.Close()
			client, err := envai.New(envai.Config{BaseURL: server.URL, APIKey: "offline", AllowHTTP: true})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "input.txt")
			os.WriteFile(path, []byte("offline original"), 0600)
			var output bytes.Buffer
			err = run(context.Background(), client, "workspace", path, "upload-key", "run-key", "https://receiver.example/wiki", &output)
			if strings.HasSuffix(scenario, "Failure") {
				if err == nil {
					t.Fatal("failure ignored")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if scenario == "pending" && (runs != 0 || !strings.Contains(output.String(), "Poll source status")) {
				t.Fatal(runs, output.String())
			}
			if scenario == "ready" && (runs != 1 || !strings.Contains(output.String(), "operation=")) {
				t.Fatal(runs, output.String())
			}
		})
	}
}

func TestWikiExampleConfigurationAndMissingFile(t *testing.T) {
	for _, base := range []string{"", "https://api.example.invalid"} {
		var out bytes.Buffer
		err := start(func(name string) string {
			if name == "ENVAI_BASE_URL" {
				return base
			}
			return ""
		}, &out)
		if err == nil {
			t.Fatal("invalid configuration or missing source ignored")
		}
	}
}
