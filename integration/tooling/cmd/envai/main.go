// Optional CLI with MySQL and pure-Go SQLite drivers linked explicitly.
package main

import (
	"context"
	envai "github.com/digitalygo/envai-teams-sdk-go"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
	"os"
)

func main() {
	os.Exit(envai.RunDefinitionsCLI(context.Background(), os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}
