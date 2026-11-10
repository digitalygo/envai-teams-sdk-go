package envai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestProjectKeyResourceHydration(t *testing.T) {
	fixtures := fixture[map[string]map[string]map[string]any](t, "project-tenancy.json")
	for _, strict := range []bool{false, true} {
		for operation, response := range fixtures {
			for _, creator := range []any{nil, "usr_01ARZ3NDEKTSV4RRFFQ69G5FAV"} {
				t.Run(operation, func(t *testing.T) {
					row := response["data"]
					owner := "created_by_user_id"
					if operation == "getCompletion" {
						owner = "user_id"
					}
					row[owner] = creator
					calls := 0
					c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
						calls++
						if r.Header.Get("Authorization") != "Bearer fixture-key" {
							t.Fatal("missing project credential")
						}
						_ = json.NewEncoder(w).Encode(response)
					})
					c.config.StrictValidation = strict
					method := reflect.ValueOf(c).MethodByName(strings.ToUpper(operation[:1]) + operation[1:])
					req := reflect.New(method.Type().In(1)).Elem()
					for i := 0; i < req.NumField(); i++ {
						f := req.Type().Field(i).Name
						var value any
						switch f {
						case "WorkspaceID":
							value = row["workspace_id"]
						case "ChatID":
							value = row["chat_id"]
						}
						if value == nil {
							value = row["id"]
						}
						req.Field(i).SetString(value.(string))
					}
					result := method.Call([]reflect.Value{reflect.ValueOf(context.Background()), req})
					if err := result[1]; !err.IsNil() {
						t.Fatal(err.Interface())
					}
					raw, err := json.Marshal(result[0].Interface())
					if err != nil {
						t.Fatal(err)
					}
					var actual map[string]any
					if json.Unmarshal(raw, &actual) != nil || !reflect.DeepEqual(actual, row) || calls != 1 {
						t.Fatal(string(raw), row, calls)
					}
					if strict {
						project := row["project_id"]
						delete(row, "project_id")
						wire, _ := json.Marshal(response)
						if c.validateStrictResponse(c.operations[operation], 200, wire) == nil {
							t.Fatal("missing project ownership accepted")
						}
						row["project_id"] = project
					}
				})
			}
		}
	}
}

func TestAPIKeyManagementUnsupportedBeforeHTTP(t *testing.T) {
	for _, strict := range []bool{false, true} {
		c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("session-only endpoint reached HTTP") })
		c.config.StrictValidation = strict
		_, e1 := c.ListAPIKeys(context.Background(), ListAPIKeysParams{})
		_, e2 := c.CreateAPIKey(context.Background(), CreateAPIKeyParams{Body: APIKeyCreate{Name: "test"}})
		e3 := c.DeleteAPIKey(context.Background(), DeleteAPIKeyParams{KeyID: "key_test"})
		for _, err := range []error{e1, e2, e3} {
			if !errors.Is(err, ErrAPIKeyManagementUnsupported) {
				t.Fatal(err)
			}
		}
	}
}

func TestDefinitionsIgnoreProjectOwnershipMetadata(t *testing.T) {
	desired := DefinitionPayload{"name": "researcher", "mode": "primary"}
	observed := DefinitionPayload{"name": "researcher", "mode": "primary", "id": "agt_x", "project_id": "prj_x", "created_by_user_id": nil}
	if !reflect.DeepEqual(definitionNormalize("agent", desired, nil), definitionNormalize("agent", observed, nil)) {
		t.Fatal("ownership metadata caused drift")
	}
}

func TestDeprecatedToolAliasesUseProjectOperations(t *testing.T) {
	cases := fixture[[]struct {
		ID       string          `json:"id"`
		GoName   string          `json:"go_name"`
		Response json.RawMessage `json:"response"`
	}](t, "api-cases.json")
	for _, tc := range cases {
		if !strings.Contains(tc.GoName, "ProjectTool") {
			continue
		}
		t.Run(tc.ID, func(t *testing.T) {
			calls := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if !strings.HasPrefix(r.URL.Path, "/api/v1/tools") {
					t.Fatal(r.URL.Path)
				}
				if tc.ID == "deleteProjectTool" {
					w.WriteHeader(204)
				} else {
					_, _ = w.Write(tc.Response)
				}
			})
			canonical := reflect.ValueOf(c).MethodByName(tc.GoName)
			alias := reflect.ValueOf(c).MethodByName(strings.Replace(tc.GoName, "ProjectTool", "UserTool", 1))
			if !alias.IsValid() || alias.Type() != canonical.Type() {
				t.Fatal("alias signature drift")
			}
			req := reflect.New(alias.Type().In(1)).Elem()
			if name := req.FieldByName("Name"); name.IsValid() {
				name.SetString("lookup")
			}
			results := alias.Call([]reflect.Value{reflect.ValueOf(context.Background()), req})
			if err := results[len(results)-1]; !err.IsNil() {
				t.Fatal(err.Interface())
			}
			if calls != 1 {
				t.Fatal(calls)
			}
		})
	}
}

func TestOnlyReviewedOperationsHaveClientMethods(t *testing.T) {
	raw, err := os.ReadFile("scripts/sdk-operations.json")
	if err != nil {
		t.Fatal(err)
	}
	var allowed []string
	if json.Unmarshal(raw, &allowed) != nil || len(allowed) != 78 {
		t.Fatal("invalid reviewed inventory")
	}
	c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected network") })
	if len(c.operations) != len(allowed) {
		t.Fatal("unreviewed operation generated")
	}
	for _, id := range allowed {
		if _, ok := c.operations[id]; !ok {
			t.Fatal(id)
		}
	}
	for _, method := range []string{"CreateProject", "ListProjects", "GetInternalBillingAccount", "CreateBillingCheckoutSession"} {
		if reflect.ValueOf(c).MethodByName(method).IsValid() {
			t.Fatal("unexpected administration method", method)
		}
	}
}

func TestWorkspaceMetadataObjectOrListRoundTrip(t *testing.T) {
	responses := fixture[map[string]map[string]map[string]any](t, "project-tenancy.json")
	response := responses["getWorkspace"]
	for _, metadata := range []any{map[string]any{"environment": "staging"}, []any{}, []any{"one", float64(2)}} {
		response["data"]["metadata"] = metadata
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(response) })
		c.config.StrictValidation = true
		workspace, err := c.GetWorkspace(context.Background(), GetWorkspaceParams{WorkspaceID: response["data"]["id"].(string)})
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(workspace)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if json.Unmarshal(raw, &decoded) != nil || !reflect.DeepEqual(decoded["metadata"], metadata) {
			t.Fatal(string(raw), metadata)
		}
	}
	response["data"]["metadata"] = "invalid scalar"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(response) })
	c.config.StrictValidation = true
	if _, err := c.GetWorkspace(context.Background(), GetWorkspaceParams{WorkspaceID: response["data"]["id"].(string)}); err == nil {
		t.Fatal("scalar metadata accepted")
	}
}

func TestStrictLegacyCompletionPreservesFrozenOwnership(t *testing.T) {
	response := fixture[map[string]map[string]any](t, "v3-http/legacy-completion.json")
	for _, operation := range []string{"getCompletion", "listCompletions"} {
		row := response["data"]
		var wire any = response
		if operation == "listCompletions" {
			wire = map[string]any{"data": []any{row}, "links": map[string]any{}, "meta": map[string]any{}}
		}
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(wire) })
		c.config.StrictValidation = true
		c.config.LegacyReadOnly = true
		var value Completion
		var err error
		if operation == "getCompletion" {
			value, err = c.GetCompletion(context.Background(), GetCompletionParams{WorkspaceID: row["workspace_id"].(string), ChatID: row["chat_id"].(string), CompletionID: row["id"].(string)})
		} else {
			var rows CompletionCollection
			rows, err = c.ListCompletions(context.Background(), ListCompletionsParams{WorkspaceID: row["workspace_id"].(string), ChatID: row["chat_id"].(string)})
			if len(rows.Data) == 1 {
				value = rows.Data[0]
			}
		}
		if err != nil || value.UserID == nil || *value.UserID != row["user_id"] {
			t.Fatal(value, err)
		}
		raw, _ := json.Marshal(wire)
		c.config.LegacyReadOnly = false
		if c.validateStrictResponse(c.operations[operation], 200, raw) == nil {
			t.Fatal("legacy response accepted under current contract")
		}
	}
}
