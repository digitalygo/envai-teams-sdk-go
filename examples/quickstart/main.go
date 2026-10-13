// The quickstart creates real remote resources and a potentially billable turn.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	envai "github.com/digitalygo/envai-teams-sdk-go"
)

func run(ctx context.Context, client *envai.Client, teamID string, out io.Writer) error {
	if teamID == "" {
		return fmt.Errorf("ENVAI_TEAM_ID is required")
	}
	workspace, err := client.CreateWorkspace(ctx, envai.CreateWorkspaceParams{
		Body: envai.WorkspaceCreateRequest{Name: "SDK quickstart"},
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "workspace:", workspace.ID)
	chat, err := client.CreateChat(ctx, envai.CreateChatParams{
		WorkspaceID: workspace.ID,
		Body:        envai.ChatCreateRequest{TeamID: teamID},
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "chat:", chat.ID)
	result, err := client.RunTurn(ctx, chat.ID, "Say hello in one sentence.", envai.TurnOptions{
		Stream: envai.StreamOptions{MaxReconnects: 2},
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(out, result.OrchestratorText)
	fmt.Fprintln(out, "terminal:", result.Terminal.Type)
	if result.Terminal.Type != "turn.completed" {
		return fmt.Errorf("turn ended with %s", result.Terminal.Type)
	}
	return nil
}

func main() {
	client, err := envai.New(envai.Config{BaseURL: os.Getenv("ENVAI_BASE_URL"), APIKey: os.Getenv("ENVAI_API_KEY")})
	if err == nil {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		err = run(ctx, client, os.Getenv("ENVAI_TEAM_ID"), os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
