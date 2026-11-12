package envai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	projectOpProjectID   = "prj_01j5abcdefghjkmnpqrstvwxyz"
	projectOpWorkspaceID = "wsp_01j5abcdefghjkmnpqrstvwxyz"
	projectOpChatID      = "cht_01j5abcdefghjkmnpqrstvwxyz"
	projectOpCronID      = "cron_01j5abcdefghjkmnpqrstvwxyz"
	projectOpSkillID     = "psk_01j5abcdefghjkmnpqrstvwxyz"
	projectOpKeyID       = "key_01j5abcdefghjkmnpqrstvwxyz"
	projectOpTeamID      = "team_01j5abcdefghjkmnpqrstvwxyz"
	projectOpTime        = "2026-11-12T10:00:00Z"
)

func projectSkillJSON() string {
	return fmt.Sprintf(`{"id":%q,"project_id":%q,"created_by_user_id":null,"name":"incident_response","description":null,"content":"# Runbook\n","created_at":%q,"updated_at":%q}`, projectOpSkillID, projectOpProjectID, projectOpTime, projectOpTime)
}

func projectSkillResource() string {
	return `{"data":` + projectSkillJSON() + `}`
}

func projectSkillCollection() string {
	return `{"data":[` + projectSkillJSON() + `]}`
}

func projectChatJSON() string {
	return fmt.Sprintf(`{"data":{"id":%q,"workspace_id":%q,"project_id":%q,"created_by_user_id":null,"team_id":%q,"title":"new title","created_at":%q,"updated_at":%q}}`, projectOpChatID, projectOpWorkspaceID, projectOpProjectID, projectOpTeamID, projectOpTime, projectOpTime)
}

func projectChatCronJSON() string {
	return fmt.Sprintf(`{"id":%q,"workspace_id":%q,"origin_chat_id":%q,"team_id":%q,"target_mode":"same_chat","prompt":"wake up","interval_minutes":15,"daily_at":null,"timezone":"UTC","start_at":null,"repeat_count":null,"until":null,"status":"active","scheduled_count":0,"delivery_count":0,"next_due_at":null,"max_until":%q,"created_at":%q,"updated_at":%q}`, projectOpCronID, projectOpWorkspaceID, projectOpChatID, projectOpTeamID, projectOpTime, projectOpTime, projectOpTime)
}

func projectChatCronResource() string {
	return `{"data":` + projectChatCronJSON() + `}`
}

func projectChatCronCollection() string {
	return `{"data":[` + projectChatCronJSON() + `]}`
}

func projectCleanupResponse(resourceType, resourceID string) string {
	return fmt.Sprintf(`{"data":{"resource_type":%q,"resource_id":%q,"cleanup":[{"id":"cln_01j5abcdefghjkmnpqrstvwxyz","kind":%q,"status":"pending","attempts":0,"deadline_at":null,"started_at":null,"finished_at":null,"last_error_code":null}]}}`, resourceType, resourceID, resourceType)
}

func projectUsagePageResponse(current, last int64) string {
	return fmt.Sprintf(`{"data":[],"links":{"next":"https://foreign.invalid/next"},"meta":{"current_page":%d,"last_page":%d}}`, current, last)
}

func projectLogPageResponse(current, last int64, agent, data string) string {
	entry := fmt.Sprintf(`{"id":"evt_01j5abcdefghjkmnpqrstvwxyz","project_id":%q,"api_key_id":null,"actor_api_key_id":null,"chat_id":%q,"workspace_id":%q,"seq":1,"turn_id":null,"type":"turn.started","ts":%q,"agent":%s,"data":%s,"extension":{"amount":"123456789.00000000000000000001"}}`, projectOpProjectID, projectOpChatID, projectOpWorkspaceID, projectOpTime, agent, data)
	return fmt.Sprintf(`{"data":[%s],"links":{"next":"https://foreign.invalid/next"},"meta":{"current_page":%d,"last_page":%d}}`, entry, current, last)
}

func projectAnalyticsResponse() string {
	bounds := []int{0, 100, 500, 1000, 2000, 5000, 10000, 30000}
	buckets := make([]map[string]any, 0, len(bounds))
	for index, lower := range bounds {
		var upper any
		if index+1 < len(bounds) {
			upper = bounds[index+1]
		}
		buckets = append(buckets, map[string]any{"lower_ms": lower, "upper_ms": upper, "count": 0})
	}
	duration := map[string]any{
		"candidate_count": 0,
		"sample_count":    0,
		"missing_count":   0,
		"invalid_count":   0,
		"zero_count":      0,
		"min_ms":          nil,
		"max_ms":          nil,
		"avg_ms":          nil,
		"p50_ms":          nil,
		"p95_ms":          nil,
		"buckets":         buckets,
	}
	tools := []string{"read", "wiki_read", "skill_read", "wiki_read_source", "wiki_read_artifact"}
	fileReads := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		fileReads = append(fileReads, map[string]any{"tool_name": tool, "count": 0})
	}
	data := map[string]any{
		"window": map[string]any{"from": "2026-11-01", "to": "2026-11-02", "timezone": "UTC", "api_key_id": nil, "chat_id": nil},
		"coverage": map[string]any{
			"source":                     "retained_chat_events",
			"aggregation":                "all_retained_matching_rows",
			"execution_history_complete": false,
			"configured_retention_days":  30,
			"observed_event_count":       0,
			"oldest_observed_at":         nil,
			"newest_observed_at":         nil,
		},
		"daily":       []map[string]any{{"date": "2026-11-01", "event_count": 0, "tool_start_count": 0, "file_read_start_count": 0, "turn_failure_count": 0, "agent_failure_count": 0, "tool_failure_count": 0}},
		"event_types": map[string]any{"items": []any{}, "other_count": 0, "omitted_group_count": 0},
		"tools":       map[string]any{"items": []any{}, "other_count": 0, "omitted_group_count": 0, "missing_name_count": 0},
		"failures":    map[string]any{"turn_event_count": 0, "agent_event_count": 0, "tool_event_count": 0, "tool_result_unknown_status_count": 0},
		"durations":   map[string]any{"tool_execution": duration, "turn_execution": duration},
		"file_reads":  fileReads,
	}
	raw, _ := json.Marshal(map[string]any{"data": data})
	return string(raw)
}

func projectFiltersResponse() string {
	return `{"data":{"types":["turn.started","agent.tool.start"]}}`
}

func projectUsageDecimalResponse() string {
	day := map[string]any{
		"date":                      "2026-11-12",
		"totals":                    []map[string]any{{"currency": "EUR", "amount": "123456789.00000000000000000001", "future": json.RawMessage(`{"n":9007199254740993}`)}},
		"invocation_count":          0,
		"unknown_invocations":       0,
		"not_submitted_invocations": 0,
		"completion_count":          0,
		"by_api_key":                []any{},
		"media":                     map[string]any{"totals": []any{}, "priced_observations": 0, "unknown_observations": 0, "unknown_currency_observations": 0, "by_api_key": []any{}},
		"by_project":                nil,
		"unknown":                   json.RawMessage("null"),
	}
	raw, _ := json.Marshal(map[string]any{
		"data":  []map[string]any{day},
		"links": map[string]any{"next": "https://foreign.invalid/next"},
		"meta":  map[string]any{"current_page": 1, "last_page": 1, "extension": json.RawMessage(`{"safe":true}`)},
	})
	return string(raw)
}

func projectTestRequest() GetProjectLogsParams {
	return GetProjectLogsParams{ProjectID: projectOpProjectID, From: Present("2026-11-01"), To: Present("2026-11-02")}
}

func TestProjectOperationMethods(t *testing.T) {
	logQuery := url.Values{
		"view":            {"events"},
		"from":            {"2026-11-01"},
		"to":              {"2026-11-02"},
		"types[]":         {"turn.started", "turn.failed"},
		"workspace_ids[]": {projectOpWorkspaceID},
		"api_key_ids[]":   {projectOpKeyID},
		"page":            {"1"},
	}
	cases := []struct {
		id, method, path, response string
		status                     int
		query                      url.Values
		body                       map[string]any
		call                       func(context.Context, *Client) error
	}{
		{
			id: "deleteWorkspace", method: http.MethodDelete, path: "/workspaces/" + projectOpWorkspaceID, status: 200,
			response: projectCleanupResponse("workspace", projectOpWorkspaceID),
			call: func(ctx context.Context, c *Client) error {
				result, err := c.DeleteWorkspace(ctx, DeleteWorkspaceParams{WorkspaceID: projectOpWorkspaceID})
				if err != nil {
					return err
				}
				if result.ResourceType != "workspace" || len(result.Cleanup) != 1 || result.Cleanup[0].Status != "pending" {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "updateChat", method: http.MethodPut, path: "/workspaces/" + projectOpWorkspaceID + "/chats/" + projectOpChatID, status: 200,
			response: projectChatJSON(), body: map[string]any{"title": "new title"},
			call: func(ctx context.Context, c *Client) error {
				result, err := c.UpdateChat(ctx, UpdateChatParams{WorkspaceID: projectOpWorkspaceID, ChatID: projectOpChatID, Body: ChatUpdateRequest{Title: Present("new title")}})
				if err != nil {
					return err
				}
				if result.Title == nil || *result.Title != "new title" {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "deleteChat", method: http.MethodDelete, path: "/workspaces/" + projectOpWorkspaceID + "/chats/" + projectOpChatID, status: 200,
			response: projectCleanupResponse("chat", projectOpChatID),
			call: func(ctx context.Context, c *Client) error {
				result, err := c.DeleteChat(ctx, DeleteChatParams{WorkspaceID: projectOpWorkspaceID, ChatID: projectOpChatID})
				if err != nil {
					return err
				}
				if result.ResourceType != "chat" || len(result.Cleanup) != 1 || result.Cleanup[0].Status != "pending" {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "listChatCrons", method: http.MethodGet, path: "/workspaces/" + projectOpWorkspaceID + "/crons", status: 200,
			response: projectChatCronCollection(),
			call: func(ctx context.Context, c *Client) error {
				result, err := c.ListChatCrons(ctx, ListChatCronsParams{WorkspaceID: projectOpWorkspaceID, Status: Present("active")})
				if err != nil {
					return err
				}
				if len(result) != 1 || result[0].ID != projectOpCronID {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "getChatCron", method: http.MethodGet, path: "/workspaces/" + projectOpWorkspaceID + "/crons/" + projectOpCronID, status: 200,
			response: projectChatCronResource(),
			call: func(ctx context.Context, c *Client) error {
				result, err := c.GetChatCron(ctx, GetChatCronParams{WorkspaceID: projectOpWorkspaceID, CronID: projectOpCronID})
				if err != nil {
					return err
				}
				if result.ID != projectOpCronID || result.TeamID != projectOpTeamID {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "updateChatCron", method: http.MethodPatch, path: "/workspaces/" + projectOpWorkspaceID + "/crons/" + projectOpCronID, status: 200,
			response: projectChatCronResource(), body: map[string]any{"target_mode": "same_chat"},
			call: func(ctx context.Context, c *Client) error {
				result, err := c.UpdateChatCron(ctx, UpdateChatCronParams{WorkspaceID: projectOpWorkspaceID, CronID: projectOpCronID, Body: ChatCronUpdateRequest{TargetMode: Present("same_chat")}})
				if err != nil {
					return err
				}
				if result.TargetMode != "same_chat" {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "deleteChatCron", method: http.MethodDelete, path: "/workspaces/" + projectOpWorkspaceID + "/crons/" + projectOpCronID, status: 204,
			call: func(ctx context.Context, c *Client) error {
				return c.DeleteChatCron(ctx, DeleteChatCronParams{WorkspaceID: projectOpWorkspaceID, CronID: projectOpCronID})
			},
		},
		{
			id: "listProjectSkills", method: http.MethodGet, path: "/projects/" + projectOpProjectID + "/skills", status: 200,
			response: projectSkillCollection(),
			call: func(ctx context.Context, c *Client) error {
				result, err := c.ListProjectSkills(ctx, ListProjectSkillsParams{ProjectID: projectOpProjectID})
				if err != nil {
					return err
				}
				if len(result) != 1 || result[0].Name != "incident_response" {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "createProjectSkill", method: http.MethodPost, path: "/projects/" + projectOpProjectID + "/skills", status: 201,
			response: projectSkillResource(), body: map[string]any{"name": "incident_response", "content": "# Runbook\n"},
			call: func(ctx context.Context, c *Client) error {
				result, err := c.CreateProjectSkill(ctx, CreateProjectSkillParams{ProjectID: projectOpProjectID, Body: ProjectSkillCreateRequest{"name": json.RawMessage(`"incident_response"`), "content": json.RawMessage(`"# Runbook\n"`)}})
				if err != nil {
					return err
				}
				if result.ID != projectOpSkillID {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "getProjectSkill", method: http.MethodGet, path: "/projects/" + projectOpProjectID + "/skills/" + projectOpSkillID, status: 200,
			response: projectSkillResource(),
			call: func(ctx context.Context, c *Client) error {
				result, err := c.GetProjectSkill(ctx, GetProjectSkillParams{ProjectID: projectOpProjectID, SkillID: projectOpSkillID})
				if err != nil {
					return err
				}
				if result.ID != projectOpSkillID {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "updateProjectSkill", method: http.MethodPut, path: "/projects/" + projectOpProjectID + "/skills/" + projectOpSkillID, status: 200,
			response: projectSkillResource(), body: map[string]any{"content": "# Runbook v2\n"},
			call: func(ctx context.Context, c *Client) error {
				result, err := c.UpdateProjectSkill(ctx, UpdateProjectSkillParams{ProjectID: projectOpProjectID, SkillID: projectOpSkillID, Body: ProjectSkillUpdateRequest{"content": json.RawMessage(`"# Runbook v2\n"`)}})
				if err != nil {
					return err
				}
				if result.Name != "incident_response" {
					t.Fatal(result)
				}
				return nil
			},
		},
		{
			id: "deleteProjectSkill", method: http.MethodDelete, path: "/projects/" + projectOpProjectID + "/skills/" + projectOpSkillID, status: 204,
			call: func(ctx context.Context, c *Client) error {
				return c.DeleteProjectSkill(ctx, DeleteProjectSkillParams{ProjectID: projectOpProjectID, SkillID: projectOpSkillID})
			},
		},
		{
			id: "getProjectUsage", method: http.MethodGet, path: "/projects/" + projectOpProjectID + "/usage", status: 200,
			query:    url.Values{"from": {"2026-11-01"}, "to": {"2026-11-02"}, "page": {"1"}},
			response: projectUsagePageResponse(1, 1),
			call: func(ctx context.Context, c *Client) error {
				result, err := c.GetProjectUsage(ctx, GetProjectUsageParams{ProjectID: projectOpProjectID, From: Present("2026-11-01"), To: Present("2026-11-02"), Page: Present(int64(1))})
				if err != nil {
					return err
				}
				if !result.Meta.CurrentPage.Set || result.Meta.CurrentPage.Value != 1 || !result.Meta.LastPage.Set || result.Meta.LastPage.Value != 1 {
					t.Fatal(result.Meta)
				}
				return nil
			},
		},
		{
			id: "getProjectLogs", method: http.MethodGet, path: "/projects/" + projectOpProjectID + "/logs", status: 200,
			query: logQuery, response: projectLogPageResponse(1, 1, "{}", "{}"),
			call: func(ctx context.Context, c *Client) error {
				raw, err := c.GetProjectLogs(ctx, GetProjectLogsParams{
					ProjectID:    projectOpProjectID,
					View:         Present("events"),
					From:         Present("2026-11-01"),
					To:           Present("2026-11-02"),
					Types:        Present([]string{"turn.started", "turn.failed"}),
					WorkspaceIDs: Present([]string{projectOpWorkspaceID}),
					APIKeyIDs:    Present([]string{projectOpKeyID}),
					Page:         Present(int64(1)),
				})
				if err != nil {
					return err
				}
				if !strings.Contains(string(raw), `"extension":{"amount":"123456789.00000000000000000001"}`) {
					t.Fatal(string(raw))
				}
				return nil
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			calls := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tc.method || r.URL.Path != "/api/v1"+tc.path {
					t.Fatal(r.Method, r.URL.Path)
				}
				if tc.query != nil && !reflect.DeepEqual(r.URL.Query(), tc.query) {
					t.Fatal(r.URL.Query(), tc.query)
				}
				if tc.body != nil {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !reflect.DeepEqual(body, tc.body) {
						t.Fatal(body, err)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					_, _ = io.WriteString(w, tc.response)
				}
			})
			if err := tc.call(context.Background(), c); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatal(calls)
			}
		})
	}
}

func TestProjectLogViewHelpers(t *testing.T) {
	for _, strict := range []bool{false, true} {
		t.Run(fmt.Sprintf("strict=%t", strict), func(t *testing.T) {
			cases := []struct {
				view, response string
				call           func(context.Context, *Client) error
			}{
				{
					view: "events", response: projectLogPageResponse(1, 1, "[]", "[]"),
					call: func(ctx context.Context, c *Client) error {
						result, err := c.GetProjectLogsEvents(ctx, projectTestRequest())
						if err != nil {
							return err
						}
						if len(result.Data) != 1 || len(result.Data[0].Agent) != 0 || len(result.Data[0].Data) != 0 || !bytes.Contains(result.Data[0].Extra["extension"], []byte(`123456789.00000000000000000001`)) {
							t.Fatal(result.Data)
						}
						return nil
					},
				},
				{
					view: "analytics", response: projectAnalyticsResponse(),
					call: func(ctx context.Context, c *Client) error {
						result, err := c.GetProjectLogsAnalytics(ctx, projectTestRequest())
						if err != nil {
							return err
						}
						if result.Window.From != "2026-11-01" || result.Window.To != "2026-11-02" || len(result.Daily) != 1 || len(result.Durations.TurnExecution.Buckets) != 8 || len(result.FileReads) != 5 {
							t.Fatal(result)
						}
						return nil
					},
				},
				{
					view: "filters", response: projectFiltersResponse(),
					call: func(ctx context.Context, c *Client) error {
						result, err := c.GetProjectLogsFilterCatalog(ctx, projectTestRequest())
						if err != nil {
							return err
						}
						if !reflect.DeepEqual(result.Types, []string{"turn.started", "agent.tool.start"}) {
							t.Fatal(result.Types)
						}
						return nil
					},
				},
			}
			for _, tc := range cases {
				t.Run(tc.view, func(t *testing.T) {
					calls := 0
					c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
						calls++
						if r.URL.Query().Get("view") != tc.view {
							t.Fatal(r.URL.Query())
						}
						_, _ = io.WriteString(w, tc.response)
					})
					c.config.StrictValidation = strict
					if err := tc.call(context.Background(), c); err != nil {
						t.Fatal(err)
					}
					if calls != 1 {
						t.Fatal(calls)
					}
				})
			}
		})
	}
}

func TestProjectLogViewCorrelationAndNormalization(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		call           func(context.Context, *Client) error
	}{
		{"analytics_gets_events", projectLogPageResponse(1, 1, "{}", "{}"), func(ctx context.Context, c *Client) error {
			_, err := c.GetProjectLogsAnalytics(ctx, projectTestRequest())
			return err
		}},
		{"events_gets_filters", projectFiltersResponse(), func(ctx context.Context, c *Client) error {
			_, err := c.GetProjectLogsEvents(ctx, projectTestRequest())
			return err
		}},
		{"filters_gets_analytics", projectAnalyticsResponse(), func(ctx context.Context, c *Client) error {
			_, err := c.GetProjectLogsFilterCatalog(ctx, projectTestRequest())
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, tc.response) })
			if err := tc.call(context.Background(), c); err == nil {
				t.Fatal("mismatched view accepted")
			}
		})
	}
	for _, view := range []string{"analytics", "unknown"} {
		request := projectTestRequest()
		if view == "unknown" {
			request.View = Null[string]()
		} else {
			request.View = Present(view)
		}
		c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("view mismatch reached HTTP") })
		if _, err := c.GetProjectLogsEvents(context.Background(), request); err == nil {
			t.Fatal("conflicting helper view accepted", view)
		}
	}
	if err := validateProjectLogsViewResponse("unknown", []byte("{}")); err == nil {
		t.Fatal("unknown view accepted")
	}
	if _, err := decodeProjectLogData[struct{ Types string }]("filters", []byte(projectFiltersResponse())); err == nil {
		t.Fatal("typed filter decoder accepted wrong scalar type")
	}
	for _, strict := range []bool{false, true} {
		for _, field := range []string{"agent", "data"} {
			t.Run(fmt.Sprintf("%s/strict=%t", field, strict), func(t *testing.T) {
				agent, data := "{}", "{}"
				if field == "agent" {
					agent = `["x"]`
				} else {
					data = `["x"]`
				}
				response := projectLogPageResponse(1, 1, agent, data)
				c := testClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, response) })
				c.config.StrictValidation = strict
				if _, err := c.GetProjectLogs(context.Background(), projectTestRequest()); err == nil {
					t.Fatal("nonempty metadata array accepted")
				}
			})
		}
	}
	raw := []byte(projectLogPageResponse(1, 1, "[]", "[]"))
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(raw) })
	out, err := c.GetProjectLogs(context.Background(), projectTestRequest())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"agent":{}`) || !strings.Contains(string(out), `"data":{}`) || !bytes.Contains(out, []byte(`123456789.00000000000000000001`)) {
		t.Fatal(string(out))
	}
	for _, invalid := range []string{`{"data":[`, `[]`, `{"data":[[]]}`, `{"data":{"types":[]}}`} {
		normalized, err := normalizeProjectLogEntryMetadata([]byte(invalid))
		if invalid == `{"data":{"types":[]}}` {
			if err != nil || !bytes.Equal(normalized, []byte(invalid)) {
				t.Fatal(invalid, string(normalized), err)
			}
			continue
		}
		if err == nil {
			t.Fatal("invalid log metadata accepted", invalid)
		}
	}
}

func TestProjectLogHelperTransportErrorsAndViewMismatch(t *testing.T) {
	calls := []struct {
		name string
		call func(context.Context, *Client, GetProjectLogsParams) error
	}{
		{"events", func(ctx context.Context, c *Client, request GetProjectLogsParams) error {
			_, err := c.GetProjectLogsEvents(ctx, request)
			return err
		}},
		{"analytics", func(ctx context.Context, c *Client, request GetProjectLogsParams) error {
			_, err := c.GetProjectLogsAnalytics(ctx, request)
			return err
		}},
		{"filters", func(ctx context.Context, c *Client, request GetProjectLogsParams) error {
			_, err := c.GetProjectLogsFilterCatalog(ctx, request)
			return err
		}},
	}
	for _, tc := range calls {
		t.Run(tc.name+"/transport", func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = io.WriteString(w, `{"error":{"type":"server_error","code":"failure","message":"failed"}}`)
			})
			if err := tc.call(context.Background(), c, projectTestRequest()); err == nil {
				t.Fatal("transport failure accepted")
			}
		})
		t.Run(tc.name+"/view_mismatch", func(t *testing.T) {
			c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("view mismatch reached HTTP") })
			request := projectTestRequest()
			request.View = Present("other")
			if err := tc.call(context.Background(), c, request); err == nil {
				t.Fatal("view mismatch accepted")
			}
		})
	}
}

func TestProjectLogQueryValidation(t *testing.T) {
	c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("invalid project log query reached HTTP") })
	op := c.operations["getProjectLogs"]
	valid := map[string]any{
		"project_id":         projectOpProjectID,
		"X-Envai-Project-Id": projectOpProjectID,
		"view":               "events",
		"from":               "2026-11-01",
		"to":                 "2026-11-02",
		"types[]":            []string{"turn.started"},
		"workspace_ids[]":    []string{projectOpWorkspaceID},
		"api_key_ids[]":      []string{projectOpKeyID},
		"page":               int64(1),
		"per_page":           int64(25),
	}
	if err := validateProjectLogQuery(op, valid); err != nil {
		t.Fatal(err)
	}
	for name, params := range map[string]map[string]any{
		"nil_view":    {"view": nil},
		"typed_view":  {"view": []string{"events"}},
		"unknown":     {"unknown": "x"},
		"bad_scalar":  {"from": "not-a-date"},
		"typed_list":  {"types[]": "turn.started"},
		"bad_pointer": {"view": "events"},
	} {
		t.Run(name, func(t *testing.T) {
			targetOp := op
			if name == "bad_pointer" {
				targetOp = operation{ID: "getProjectLogs", Method: http.MethodGet, Path: "/missing"}
			}
			if err := validateProjectLogQuery(targetOp, params); err == nil {
				t.Fatal("invalid query accepted", name)
			}
		})
	}
	if _, err := projectLogParameterSchema(op, "missing"); err == nil {
		t.Fatal("missing parameter schema accepted")
	}
	tooMany := make([]string, 51)
	for index := range tooMany {
		tooMany[index] = "type_" + strconv.Itoa(index)
	}
	invalid := []GetProjectLogsParams{
		{View: Present("unknown")},
		{View: Null[string]()},
		{Types: Present([]string{})},
		{Types: Null[[]string]()},
		{Types: Present(tooMany)},
		{Types: Present([]string{"turn.started", "turn.started"})},
		{Types: Present([]string{strings.Repeat("x", 81)})},
		{Type: Present("turn.started"), Types: Present([]string{"turn.failed"})},
		{APIKeyID: Present(projectOpKeyID), APIKeyIDs: Present([]string{projectOpKeyID})},
		{View: Present("analytics"), Type: Present("turn.started")},
		{View: Present("analytics"), Types: Present([]string{"turn.started"})},
		{View: Present("analytics"), Page: Present(int64(1))},
		{View: Present("filters"), APIKeyID: Present(projectOpKeyID)},
		{View: Present("filters"), ChatID: Present(projectOpChatID)},
		{View: Present("filters"), PerPage: Present(int64(25))},
		{View: Present("filters"), WorkspaceIDs: Present([]string{projectOpWorkspaceID})},
	}
	for index, request := range invalid {
		request.ProjectID = projectOpProjectID
		request.From = Present("2026-11-01")
		request.To = Present("2026-11-02")
		for _, strict := range []bool{false, true} {
			t.Run(fmt.Sprintf("case_%d/strict=%t", index, strict), func(t *testing.T) {
				client := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("invalid request reached HTTP") })
				client.config.StrictValidation = strict
				if _, err := client.GetProjectLogs(context.Background(), request); err == nil {
					t.Fatal("invalid request accepted")
				}
			})
		}
	}
}

func TestProjectDeletionEnvelopesAndStatuses(t *testing.T) {
	for _, strict := range []bool{false, true} {
		t.Run(fmt.Sprintf("strict=%t", strict), func(t *testing.T) {
			calls := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				switch {
				case r.URL.Path == "/api/v1/workspaces/"+projectOpWorkspaceID:
					_, _ = io.WriteString(w, projectCleanupResponse("workspace", projectOpWorkspaceID))
				case r.URL.Path == "/api/v1/workspaces/"+projectOpWorkspaceID+"/chats/"+projectOpChatID:
					_, _ = io.WriteString(w, projectCleanupResponse("chat", projectOpChatID))
				case strings.Contains(r.URL.Path, "/crons/") || strings.Contains(r.URL.Path, "/skills/"):
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Fatal(r.URL.Path)
				}
			})
			c.config.StrictValidation = strict
			workspace, err := c.DeleteWorkspace(context.Background(), DeleteWorkspaceParams{WorkspaceID: projectOpWorkspaceID})
			if err != nil || workspace.Cleanup[0].Status != "pending" {
				t.Fatal(workspace, err)
			}
			chat, err := c.DeleteChat(context.Background(), DeleteChatParams{WorkspaceID: projectOpWorkspaceID, ChatID: projectOpChatID})
			if err != nil || chat.Cleanup[0].Status != "pending" {
				t.Fatal(chat, err)
			}
			if err = c.DeleteChatCron(context.Background(), DeleteChatCronParams{WorkspaceID: projectOpWorkspaceID, CronID: projectOpCronID}); err != nil {
				t.Fatal(err)
			}
			if err = c.DeleteProjectSkill(context.Background(), DeleteProjectSkillParams{ProjectID: projectOpProjectID, SkillID: projectOpSkillID}); err != nil {
				t.Fatal(err)
			}
			if calls != 4 {
				t.Fatal(calls)
			}
		})
	}
}

func TestProjectSkillRequestMapsAndMutableWhitelist(t *testing.T) {
	var create ProjectSkillCreateRequest = map[string]json.RawMessage{"name": json.RawMessage(`"incident_response"`), "content": json.RawMessage(`"# Runbook\n"`)}
	var update ProjectSkillUpdateRequest = map[string]json.RawMessage{"description": json.RawMessage(`null`)}
	if reflect.TypeOf(create).Kind() != reflect.Map || reflect.TypeOf(update).Kind() != reflect.Map {
		t.Fatal("skill request maps changed to structs")
	}
	if raw, err := json.Marshal(create); err != nil || !strings.Contains(string(raw), `"content":"# Runbook\n"`) {
		t.Fatal(string(raw), err)
	}
	c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("invalid request reached HTTP") })
	c.config.StrictValidation = true
	for _, call := range []func() error{
		func() error {
			_, err := c.UpdateChat(context.Background(), UpdateChatParams{WorkspaceID: projectOpWorkspaceID, ChatID: projectOpChatID, Body: ChatUpdateRequest{Title: Present("new title"), Extra: map[string]json.RawMessage{"team_id": json.RawMessage(`"other"`)}}})
			return err
		},
		func() error {
			_, err := c.UpdateChatCron(context.Background(), UpdateChatCronParams{WorkspaceID: projectOpWorkspaceID, CronID: projectOpCronID, Body: ChatCronUpdateRequest{TargetMode: Present("same_chat"), Extra: map[string]json.RawMessage{"status": json.RawMessage(`"completed"`)}}})
			return err
		},
		func() error {
			_, err := c.CreateProjectSkill(context.Background(), CreateProjectSkillParams{ProjectID: projectOpProjectID, Body: ProjectSkillCreateRequest{"name": json.RawMessage(`"incident_response"`), "content": json.RawMessage(`"# Runbook\n"`), "project_id": json.RawMessage("null")}})
			return err
		},
		func() error {
			_, err := c.UpdateProjectSkill(context.Background(), UpdateProjectSkillParams{ProjectID: projectOpProjectID, SkillID: projectOpSkillID, Body: ProjectSkillUpdateRequest{"content": json.RawMessage(`"# Runbook\n"`), "id": json.RawMessage("null")}})
			return err
		},
	} {
		if err := call(); err == nil {
			t.Fatal("non-whitelisted mutable field accepted")
		}
	}
	if err := validateMutableBody(c.operations["getProjectUsage"], nil); err != nil {
		t.Fatal(err)
	}
	if err := validateMutableBody(c.operations["updateChat"], "not an object"); err == nil {
		t.Fatal("non-object mutable body accepted")
	}
	if err := validateMutableBody(c.operations["updateChat"], func() {}); err == nil {
		t.Fatal("unmarshalable mutable body accepted")
	}
	if err := validateMutableBody(operation{ID: "updateChat", Method: http.MethodPut, Path: "/missing"}, ChatUpdateRequest{}); err == nil {
		t.Fatal("missing mutable body schema accepted")
	}
	if err := validateMutableBody(operation{ID: "updateChat", Method: http.MethodPost, Path: "/billing/webhooks/stripe"}, map[string]any{}); err == nil {
		t.Fatal("mutable body schema without canonical properties accepted")
	}
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/skills"):
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, projectSkillResource())
		case strings.HasSuffix(r.URL.Path, "/skills"):
			_, _ = io.WriteString(w, projectSkillResource())
		case strings.Contains(r.URL.Path, "/skills/"):
			_, _ = io.WriteString(w, projectSkillResource())
		case strings.Contains(r.URL.Path, "/crons/"):
			_, _ = io.WriteString(w, projectChatCronResource())
		default:
			_, _ = io.WriteString(w, projectChatJSON())
		}
	})
	c.config.StrictValidation = true
	if _, err := c.UpdateChat(context.Background(), UpdateChatParams{WorkspaceID: projectOpWorkspaceID, ChatID: projectOpChatID, Body: ChatUpdateRequest{Title: Present("new title")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UpdateChatCron(context.Background(), UpdateChatCronParams{WorkspaceID: projectOpWorkspaceID, CronID: projectOpCronID, Body: ChatCronUpdateRequest{TargetMode: Present("same_chat")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateProjectSkill(context.Background(), CreateProjectSkillParams{ProjectID: projectOpProjectID, Body: create}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UpdateProjectSkill(context.Background(), UpdateProjectSkillParams{ProjectID: projectOpProjectID, SkillID: projectOpSkillID, Body: ProjectSkillUpdateRequest{"content": json.RawMessage(`"# Runbook v2\n"`)}}); err != nil {
		t.Fatal(err)
	}
}

func TestProjectNumericPageIterators(t *testing.T) {
	for _, strict := range []bool{false, true} {
		for _, stop := range []bool{false, true} {
			for _, kind := range []string{"usage", "logs"} {
				t.Run(fmt.Sprintf("%s/strict=%t/stop=%t", kind, strict, stop), func(t *testing.T) {
					var pages []string
					var client *Client
					client = testClient(t, func(w http.ResponseWriter, r *http.Request) {
						page := r.URL.Query().Get("page")
						pages = append(pages, page)
						if r.Host != client.base.Host || strings.Contains(r.URL.Path, "foreign.invalid") {
							t.Fatal(r.Host, r.URL.Path)
						}
						current, err := strconv.ParseInt(page, 10, 64)
						if err != nil {
							t.Fatal(err)
						}
						if kind == "usage" {
							if r.URL.Path != "/api/v1/projects/"+projectOpProjectID+"/usage" || r.URL.Query().Get("from") != "2026-11-01" || r.URL.Query().Get("per_page") != "25" {
								t.Fatal(r.URL.String())
							}
							_, _ = io.WriteString(w, projectUsagePageResponse(current, 4))
						} else {
							if r.URL.Path != "/api/v1/projects/"+projectOpProjectID+"/logs" || !reflect.DeepEqual(r.URL.Query()["types[]"], []string{"turn.started"}) || r.URL.Query().Get("view") != "events" {
								t.Fatal(r.URL.String())
							}
							_, _ = io.WriteString(w, projectLogPageResponse(current, 4, "{}", "{}"))
						}
					})
					client.config.StrictValidation = strict
					yielded := 0
					if kind == "usage" {
						request := GetProjectUsageParams{ProjectID: projectOpProjectID, From: Present("2026-11-01"), To: Present("2026-11-02"), Page: Present(int64(3)), PerPage: Present(int64(25))}
						for _, err := range client.ProjectUsagePages(context.Background(), request) {
							if err != nil {
								t.Fatal(err)
							}
							yielded++
							if stop {
								break
							}
						}
					} else {
						request := GetProjectLogsParams{ProjectID: projectOpProjectID, From: Present("2026-11-01"), To: Present("2026-11-02"), Types: Present([]string{"turn.started"}), WorkspaceIDs: Present([]string{projectOpWorkspaceID}), APIKeyIDs: Present([]string{projectOpKeyID}), Page: Present(int64(3)), PerPage: Present(int64(25))}
						for _, err := range client.ProjectLogPages(context.Background(), request) {
							if err != nil {
								t.Fatal(err)
							}
							yielded++
							if stop {
								break
							}
						}
					}
					want := []string{"3", "4"}
					wantYielded := 2
					if stop {
						want = []string{"3"}
						wantYielded = 1
					}
					if !reflect.DeepEqual(pages, want) || yielded != wantYielded {
						t.Fatal(pages, yielded)
					}
				})
			}
		}
	}
}

func TestNumericPagesEdgeCases(t *testing.T) {
	t.Run("invalid_start", func(t *testing.T) {
		c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("invalid page reached HTTP") })
		yielded := 0
		for _, err := range c.ProjectUsagePages(context.Background(), GetProjectUsageParams{ProjectID: projectOpProjectID, Page: Present(int64(0))}) {
			yielded++
			var protocol *ProtocolError
			if !errors.As(err, &protocol) {
				t.Fatal(err)
			}
		}
		if yielded != 1 {
			t.Fatal(yielded)
		}
	})
	t.Run("canceled", func(t *testing.T) {
		c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("canceled page reached HTTP") })
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		yielded := 0
		for _, err := range c.ProjectUsagePages(ctx, GetProjectUsageParams{ProjectID: projectOpProjectID}) {
			yielded++
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		}
		if yielded != 1 {
			t.Fatal(yielded)
		}
	})
	t.Run("fetch_error", func(t *testing.T) {
		calls := 0
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"error":{"type":"server_error","code":"failure","message":"failed","details":{"safe":true}}}`)
		})
		yielded := 0
		for _, err := range c.ProjectUsagePages(context.Background(), GetProjectUsageParams{ProjectID: projectOpProjectID}) {
			yielded++
			var api *APIError
			if !errors.As(err, &api) || api.Status != 500 {
				t.Fatal(err)
			}
		}
		if yielded != 1 || calls != 1 {
			t.Fatal(yielded, calls)
		}
	})
	for name, response := range map[string]string{"wrong_current": projectUsagePageResponse(9, 4), "last_before_current": projectUsagePageResponse(3, 2)} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				_, _ = io.WriteString(w, response)
			})
			yielded := 0
			for _, err := range c.ProjectUsagePages(context.Background(), GetProjectUsageParams{ProjectID: projectOpProjectID, Page: Present(int64(3))}) {
				yielded++
				var protocol *ProtocolError
				if !errors.As(err, &protocol) {
					t.Fatal(err)
				}
			}
			if yielded != 1 || calls != 1 {
				t.Fatal(yielded, calls)
			}
		})
	}
}

func TestProjectUsageDecimalStringsAndUnknownFields(t *testing.T) {
	for _, strict := range []bool{false, true} {
		t.Run(fmt.Sprintf("strict=%t", strict), func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, projectUsageDecimalResponse()) })
			c.config.StrictValidation = strict
			result, err := c.GetProjectUsage(context.Background(), GetProjectUsageParams{ProjectID: projectOpProjectID, From: Present("2026-11-12"), To: Present("2026-11-12")})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Data) != 1 || len(result.Data[0].Totals) != 1 || result.Data[0].Totals[0].Amount != "123456789.00000000000000000001" || result.Data[0].Totals[0].Currency != "EUR" {
				t.Fatal(result.Data)
			}
			if string(result.Data[0].Totals[0].Extra["future"]) != `{"n":9007199254740993}` || string(result.Data[0].Extra["unknown"]) != "null" || result.Data[0].ByProject != nil {
				t.Fatal(result.Data[0])
			}
			raw, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			for _, part := range []string{`"amount":"123456789.00000000000000000001"`, `"future":{"n":9007199254740993}`, `"by_project":null`, `"unknown":null`} {
				if !strings.Contains(string(raw), part) {
					t.Fatal(part, string(raw))
				}
			}
		})
	}
}

func TestProjectMutationErrorsRemainTyped(t *testing.T) {
	cases := []struct {
		name, code, errorType, payload string
		status                         int
		call                           func(*Client) error
	}{
		{
			name: "402_create_project_skill", code: "credit_exhausted", errorType: "billing_error", status: 402,
			payload: `{"error":{"type":"billing_error","code":"credit_exhausted","message":"no credit","details":{"scope":"company"}}}`,
			call: func(c *Client) error {
				_, err := c.CreateProjectSkill(context.Background(), CreateProjectSkillParams{ProjectID: projectOpProjectID, Body: ProjectSkillCreateRequest{"name": json.RawMessage(`"incident_response"`), "content": json.RawMessage(`"# Runbook\n"`)}})
				return err
			},
		},
		{
			name: "503_create_image_generation", code: "provider_configuration_error", errorType: "provider_configuration_error", status: 503,
			payload: `{"error":{"type":"provider_configuration_error","code":"provider_configuration_error","message":"provider unavailable","details":{"provider":"x"}}}`,
			call: func(c *Client) error {
				_, err := c.CreateImageGeneration(context.Background(), CreateImageGenerationParams{Body: ImageGenerationRequest{Model: "model", Prompt: "prompt", WebhookURL: "https://receiver.example/hook"}})
				return err
			},
		},
		{
			name: "409_delete_workspace", code: "active_cron", errorType: "conflict", status: 409,
			payload: `{"error":{"type":"conflict","code":"active_cron","message":"active cron","details":{"cron_id":"cron_x"}}}`,
			call: func(c *Client) error {
				_, err := c.DeleteWorkspace(context.Background(), DeleteWorkspaceParams{WorkspaceID: projectOpWorkspaceID})
				return err
			},
		},
		{
			name: "409_delete_chat", code: "active_cron", errorType: "conflict", status: 409,
			payload: `{"error":{"type":"conflict","code":"active_cron","message":"active cron","details":{"cron_id":"cron_x"}}}`,
			call: func(c *Client) error {
				_, err := c.DeleteChat(context.Background(), DeleteChatParams{WorkspaceID: projectOpWorkspaceID, ChatID: projectOpChatID})
				return err
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			var paths []string
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				paths = append(paths, r.Method+" "+r.URL.Path)
				w.Header().Set("X-Request-ID", "request1")
				w.Header().Set("Retry-After", "3")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.payload)
			})
			c.config.RetryAttempts = 2
			err := tc.call(c)
			var api *APIError
			if !errors.As(err, &api) || api.Status != tc.status || api.Type != tc.errorType || api.Code != tc.code || api.RequestID != "request1" || api.RetryAfter != 3*time.Second || !strings.Contains(string(api.Details), `"safe"`) && !strings.Contains(string(api.Details), `scope`) && !strings.Contains(string(api.Details), `provider`) && !strings.Contains(string(api.Details), `cron_id`) {
				t.Fatal(err, api)
			}
			if calls != 1 || len(paths) != 1 || strings.Contains(paths[0], "/crons") {
				t.Fatal(calls, paths)
			}
		})
	}
}

func TestNoPublicChatCronCreateOperation(t *testing.T) {
	c := testClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected network") })
	if _, ok := c.operations["createChatCron"]; ok {
		t.Fatal("public cron create generated")
	}
	if reflect.ValueOf(c).MethodByName("CreateChatCron").IsValid() {
		t.Fatal("public cron create method exists")
	}
}

func TestProviderConfigurationErrorContract(t *testing.T) {
	raw, err := os.ReadFile("contract/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var root any
	if json.Unmarshal(raw, &root) != nil {
		t.Fatal("invalid derived OpenAPI JSON")
	}
	path := jsonPath(t, root, "paths", "/images/generations", "post", "responses", "503")
	if ref, _ := path.(map[string]any)["$ref"].(string); ref != "#/components/responses/ProviderConfigurationError" {
		t.Fatal(path)
	}
	example := jsonPath(t, root, "components", "responses", "ProviderConfigurationError", "content", "application/json", "example", "error")
	if value := jsonPath(t, example, "type"); value != "provider_configuration_error" {
		t.Fatal(value)
	}
	if value := jsonPath(t, example, "code"); value != "provider_configuration_error" {
		t.Fatal(value)
	}
	if value := jsonPath(t, example, "details"); value != nil {
		t.Fatal(value)
	}
	enum, ok := jsonPath(t, root, "components", "schemas", "Error", "properties", "error", "properties", "type", "enum").([]any)
	if !ok {
		t.Fatal("missing error type enum")
	}
	found := false
	for _, value := range enum {
		if value == "provider_configuration_error" {
			found = true
		}
	}
	if !found {
		t.Fatal(enum)
	}
}

func jsonPath(t *testing.T, value any, keys ...string) any {
	t.Helper()
	for _, key := range keys {
		object, ok := value.(map[string]any)
		if !ok {
			t.Fatal(key, value)
		}
		value, ok = object[key]
		if !ok {
			t.Fatal(key, object)
		}
	}
	return value
}

func TestGeneratedFixtureDoesNotClaimThreeLogViews(t *testing.T) {
	raw, err := os.ReadFile("testdata/api-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		ID           string `json:"id"`
		ResponseView string `json:"response_view"`
	}
	if json.Unmarshal(raw, &cases) != nil || len(cases) != 92 {
		t.Fatal("invalid generated fixture inventory")
	}
	found := false
	for _, entry := range cases {
		if entry.ID == "getProjectLogs" {
			found = true
			if entry.ResponseView != "events" {
				t.Fatal(entry)
			}
		} else if entry.ResponseView != "" {
			t.Fatal(entry)
		}
	}
	if !found {
		t.Fatal("missing getProjectLogs generated fixture")
	}
	provenance, err := os.ReadFile("contract/provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(provenance), "events view only") {
		t.Fatal("generated fixture origin overclaims view coverage")
	}
}

func TestBillingEventsRemainNonTerminalAndV2Frozen(t *testing.T) {
	for _, name := range []string{"usage-pending", "usage-empty"} {
		var usage ATPUsageCost
		if json.Unmarshal(wikiFixtureData(t, name), &usage) != nil {
			t.Fatal(name)
		}
		if _, err := json.Marshal(usage); err != nil {
			t.Fatal(err)
		}
	}
	var billing ATPBillingState
	if json.Unmarshal([]byte(`{"status":"inactive","extension":true}`), &billing) != nil {
		t.Fatal("billing state")
	}
	var amount ATPCurrencyAmount
	if json.Unmarshal([]byte(`{"currency":"EUR","amount":"1.0000000000001","source":"reported"}`), &amount) != nil || amount.Amount != "1.0000000000001" {
		t.Fatal(amount)
	}
	var title ATPChatTitleUpdated
	if json.Unmarshal([]byte(`{"title":"New chat title"}`), &title) != nil || title.Title != "New chat title" {
		t.Fatal(title)
	}
	for _, eventType := range []string{"billing.limit_reached", "billing.credit_exhausted"} {
		payload := `{"reason":"credit_exhausted","scope":"company","currency":"EUR"}`
		if eventType == "billing.limit_reached" {
			payload = `{"reason":"limit_reached","scope":"project","currency":"EUR","reset_at":null}`
		}
		raw := fmt.Sprintf(`{"v":3,"id":"evt_01j5abcdefghjkmnpqrstvwxyz","seq":1,"ts":%q,"chat_id":%q,"turn_id":null,"agent":null,"type":%q,"data":%s}`, projectOpTime, projectOpChatID, eventType, payload)
		event, err := DecodeEvent([]byte(raw))
		if err != nil {
			t.Fatal(eventType, err)
		}
		if event.Terminal() || event.TurnID != nil {
			t.Fatal(eventType, event)
		}
		typed, err := event.TypedPayload()
		if err != nil {
			t.Fatal(eventType, err)
		}
		switch eventType {
		case "billing.limit_reached":
			if value, ok := typed.(ATPBillingLimitReached); !ok || value.Scope != "project" || value.Currency != "EUR" {
				t.Fatal(typed)
			}
		case "billing.credit_exhausted":
			if value, ok := typed.(ATPBillingCreditExhausted); !ok || value.Reason != "credit_exhausted" {
				t.Fatal(typed)
			}
		}
	}
	raw, err := os.ReadFile("contract/events-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "88d41835f88c0cc4ad071b06ed4623882c6b64144baf312c4a095504c98cd7d6" {
		t.Fatal("frozen v2 source changed")
	}
}
