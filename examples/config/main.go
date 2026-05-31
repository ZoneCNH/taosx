package main

import (
	"fmt"
	"time"

	"github.com/bytechainx/baselib-template/pkg/templatex"
)

func main() {
	cfg := templatex.Config{
		Name:    "templatex",
		Timeout: time.Second,
		Secret:  "example",
	}

	fmt.Println(cfg.Sanitize().Secret)
}
