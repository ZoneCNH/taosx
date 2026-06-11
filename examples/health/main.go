package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ZoneCNH/taosx/pkg/taosx"
)

func main() {
	run(os.Stdout, os.Stderr, taosx.Config{
		Endpoint: "localhost:6041",
		Database: "metrics",
	})
}

func run(stdout, stderr io.Writer, cfg taosx.Config) {
	ctx := context.Background()
	client, err := taosx.New(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "create client: %v\n", err)
		return
	}
	defer func() {
		_ = client.Close(ctx)
	}()

	status := client.Health(ctx)
	_, _ = fmt.Fprintln(stdout, status.Status)
}
