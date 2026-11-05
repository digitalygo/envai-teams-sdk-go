package envai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

var independentCapabilities = strings.Replace(backgroundCapabilities, `"strict_model"`, `"strict_model","independent_jobs","finite_choice_classification","originless_deliveries"`, 1)

func TestIndependentJobAndChoiceContracts(t *testing.T) {
	posts := 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			fmt.Fprint(w, independentCapabilities)
			return
		}
		if r.Method == "POST" {
			posts++
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/jobs"):
			var body IndependentJobRequest
			json.NewDecoder(r.Body).Decode(&body)
			if body.RequestID != "job" {
				t.Errorf("wrong request: %+v", body)
			}
			w.WriteHeader(202)
			fmt.Fprint(w, `{"task_id":"task","turn_id":"initial"}`)
		case strings.HasSuffix(r.URL.Path, "/classify-choice"):
			fmt.Fprint(w, `{"request_id":"choice","choice":"yes","model":"fixture","invocation_id":"i"}`)
		case strings.HasSuffix(r.URL.Path, "/tasks/task"):
			fmt.Fprint(w, `{"task_id":"task","origin_turn_id":"","status":"completed","output":"full","result_turn_id":"final"}`)
		case strings.HasSuffix(r.URL.Path, "/deliveries"):
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if _, ok := body["origin_turn_id"]; ok {
				t.Error("empty origin serialized")
			}
			if body["strict_model"] != true {
				t.Error("strict policy lost")
			}
			w.WriteHeader(202)
			fmt.Fprint(w, `{"event_id":"event","status":"queued","duplicate":false}`)
		default:
			t.Errorf("unexpected %s", r.URL.Path)
		}
	})
	c.config.NativeTasks = true
	ctx := context.Background()
	receipt, err := c.StartIndependentJob(ctx, StartIndependentJobParams{ChatID: "c", Body: IndependentJobRequest{WorkspaceID: "w", RequestID: "job", Prompt: "work"}})
	if err != nil || receipt.TurnID != "initial" {
		t.Fatal(receipt, err)
	}
	choice, err := c.ClassifyChoice(ctx, ClassifyChoiceParams{ChatID: "c", Body: ChoiceClassificationRequest{RequestID: "choice", Context: json.RawMessage(`{}`), Question: "ready?", Choices: []ChoiceClassificationRequestChoicesItem{{Label: "yes", Description: "ready"}, {Label: "no", Description: "wait"}}}})
	if err != nil || choice.Choice != "yes" {
		t.Fatal(choice, err)
	}
	task, err := c.GetTask(ctx, "c", "task")
	if err != nil || task.Output != "full" || task.OriginTurnID != "" {
		t.Fatal(task, err)
	}
	_, err = c.SubmitDelivery(ctx, "c", NativeDeliveryRequest{WorkspaceID: "w", EventID: "event", TaskID: "task", Status: "completed", Output: "full", StrictModel: Present(true)})
	if err != nil {
		t.Fatal(err)
	}
	if posts != 3 {
		t.Fatal(posts)
	}
}

func TestIndependentCapabilityAndChoiceValidation(t *testing.T) {
	for _, scenario := range []string{"missing", "unknown", "mismatch", "invalid", "error"} {
		t.Run(scenario, func(t *testing.T) {
			posts := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					if scenario == "missing" {
						fmt.Fprint(w, backgroundCapabilities)
					} else {
						fmt.Fprint(w, independentCapabilities)
					}
					return
				}
				posts++
				if scenario == "error" {
					w.WriteHeader(503)
					fmt.Fprint(w, `{"error":{"code":"unavailable","message":"fixture"}}`)
					return
				}
				id, label := "choice", "yes"
				if scenario == "unknown" {
					label = "other"
				}
				if scenario == "mismatch" {
					id = "other"
				}
				fmt.Fprintf(w, `{"request_id":%q,"choice":%q,"model":"fixture","invocation_id":"i"}`, id, label)
			})
			c.config.NativeTasks = true
			req := ChoiceClassificationRequest{RequestID: "choice", Context: json.RawMessage(`{}`), Question: "ready?", Choices: []ChoiceClassificationRequestChoicesItem{{Label: "yes", Description: "ready"}, {Label: "no", Description: "wait"}}}
			if scenario == "invalid" {
				req.Context = json.RawMessage(`null`)
			}
			_, err := c.ClassifyChoice(context.Background(), ClassifyChoiceParams{ChatID: "c", Body: req})
			if err == nil {
				t.Fatal("accepted invalid request/response")
			}
			expected := 1
			if scenario == "missing" || scenario == "invalid" {
				expected = 0
			}
			if posts != expected {
				t.Fatal(posts)
			}
		})
	}
}

func TestIndependentResponseIdentities(t *testing.T) {
	for _, response := range []string{`{"task_id":"","turn_id":"initial"}`, `{"task_id":"task","turn_id":""}`} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/capabilities") {
				fmt.Fprint(w, independentCapabilities)
				return
			}
			w.WriteHeader(202)
			fmt.Fprint(w, response)
		})
		c.config.NativeTasks = true
		_, err := c.StartIndependentJob(context.Background(), StartIndependentJobParams{ChatID: "c", Body: IndependentJobRequest{WorkspaceID: "w", RequestID: "job", Prompt: "work"}})
		if err == nil {
			t.Fatal("empty receipt identity accepted")
		}
	}
	for _, field := range []string{"model", "invocation_id"} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/capabilities") {
				fmt.Fprint(w, independentCapabilities)
				return
			}
			data := map[string]string{"request_id": "choice", "choice": "yes", "model": "fixture", "invocation_id": "i"}
			data[field] = ""
			json.NewEncoder(w).Encode(data)
		})
		c.config.NativeTasks = true
		_, err := c.ClassifyChoice(context.Background(), ClassifyChoiceParams{ChatID: "c", Body: ChoiceClassificationRequest{RequestID: "choice", Context: json.RawMessage(`{}`), Question: "ready?", Choices: []ChoiceClassificationRequestChoicesItem{{Label: "yes", Description: "ready"}, {Label: "no", Description: "wait"}}}})
		if err == nil {
			t.Fatal("empty classifier metadata accepted")
		}
	}
}

func TestOriginlessCapabilityRequired(t *testing.T) {
	for _, kind := range []string{"delivery", "snapshot"} {
		t.Run(kind, func(t *testing.T) {
			posts := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					fmt.Fprint(w, backgroundCapabilities)
					return
				}
				if r.Method == "POST" {
					posts++
				}
				fmt.Fprint(w, `{"task_id":"task","origin_turn_id":"","status":"completed"}`)
			})
			c.config.NativeTasks = true
			var err error
			if kind == "delivery" {
				_, err = c.SubmitDelivery(context.Background(), "c", NativeDeliveryRequest{WorkspaceID: "w", EventID: "event", TaskID: "task", Status: "completed", Output: "full"})
			} else {
				_, err = c.GetTask(context.Background(), "c", "task")
			}
			if err == nil || posts != 0 {
				t.Fatal("missing capability allowed", err, posts)
			}
		})
	}
}
