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
	client, err := taosx.New(context.Background(), cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "create client: %v\n", err)
		return
	}
	defer func() {
		_ = client.Close(context.Background())
	}()

	_, _ = fmt.Fprintln(stdout, taosx.ModuleName)
}
