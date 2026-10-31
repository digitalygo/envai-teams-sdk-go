// Command envai provides explicit definitions deployment operations.
package main

import (
	"context"
	envai "github.com/digitalygo/envai-teams-sdk-go"
	_ "github.com/lib/pq"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	os.Exit(envai.RunDefinitionsCLI(ctx, os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}
