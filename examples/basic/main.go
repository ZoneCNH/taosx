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
	defer func() {
		_ = client.Close(context.Background())
	}()

	fmt.Println(templatex.ModuleName)
}
