package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bytechainx/baselib-template/pkg/templatex"
)

func main() {
	client, err := templatex.New(context.Background(), templatex.Config{Name: "templatex"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create client: %v\n", err)
		return
	}

	status := client.HealthCheck(context.Background())
	fmt.Println(status.Status)
}
