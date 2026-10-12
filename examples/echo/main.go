package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	envai "github.com/digitalygo/envai-teams-sdk-go"
)

func run() error {
	client, err := envai.New(envai.Config{BaseURL: os.Getenv("ENVAI_BASE_URL"), APIKey: os.Getenv("ENVAI_API_KEY")})
	if err != nil {
		return err
	}
	registry, err := envai.NewRegistry(envai.Tool{Name: "echo", Description: "Returns the supplied text", Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`), Handler: func(ctx context.Context, inv envai.ToolInvocation) (envai.ToolResult, error) {
		if err := ctx.Err(); err != nil {
			return envai.ToolResult{}, err
		}
		var args struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(inv.Arguments, &args); err != nil {
			return envai.ToolResult{}, err
		}
		return envai.JSONResult(map[string]string{"text": args.Text})
	}})
	if err != nil {
		return err
	}
	path := os.Getenv("ENVAI_STATE_DIR")
	if path == "" {
		return fmt.Errorf("ENVAI_STATE_DIR is required for durable local duplicate protection")
	}
	store, err := envai.NewFileCallStore(path)
	if err != nil {
		return err
	}
	runner, err := envai.NewToolRunner(client, registry, store, envai.RunnerOptions{Namespace: os.Getenv("ENVAI_ACCOUNT_SCOPE")})
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	result, err := client.RunTurn(ctx, os.Getenv("ENVAI_CHAT_ID"), "Use echo to return hello.", envai.TurnOptions{Stream: envai.StreamOptions{MaxReconnects: 2}, Tools: runner})
	if err != nil {
		return err
	}
	fmt.Println(result.OrchestratorText)
	fmt.Println(result.Terminal.Type)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
