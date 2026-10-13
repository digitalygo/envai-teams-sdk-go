// The example requires explicit credentials, stable replay keys, a local source,
// and an HTTPS application webhook. It never fabricates idempotency keys.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	envai "github.com/digitalygo/envai-teams-sdk-go"
)

func run(ctx context.Context, client *envai.Client, workspace, path, uploadKey, runKey, webhook string, out io.Writer) error {
	source, err := client.UploadSourceFile(ctx, workspace, path, uploadKey, "application/octet-stream")
	if err != nil {
		return err
	}
	current, err := client.GetWikiSource(ctx, envai.GetWikiSourceParams{WorkspaceID: workspace, FileID: source.FileID})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "source=%s conversion=%s lifecycle=%s\n", current.FileID, current.Conversion.Status, current.Lifecycle)
	if current.Conversion.Status != "completed" || current.Lifecycle != "inbox" {
		fmt.Fprintln(out, "Poll source status later; conversion does not automatically index the original.")
		return nil
	}
	operation, err := client.CreateWikiInboxRun(ctx, envai.CreateWikiInboxRunParams{WorkspaceID: workspace, IdempotencyKey: runKey, Body: envai.CreateWikiInboxRunBody{WebhookURL: webhook}})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "operation=%s status=%s provider_cost=%s\n", operation.OperationID, operation.Status, operation.UsageCost.Status)
	return nil
}
func start(env func(string) string, out io.Writer) error {
	client, err := envai.New(envai.Config{BaseURL: env("ENVAI_BASE_URL"), APIKey: env("ENVAI_API_KEY")})
	if err != nil {
		return err
	}
	return run(context.Background(), client, env("ENVAI_WORKSPACE_ID"), env("ENVAI_SOURCE_PATH"), env("ENVAI_UPLOAD_KEY"), env("ENVAI_RUN_KEY"), env("ENVAI_WEBHOOK_URL"), out)
}
func main() {
	if err := start(os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
