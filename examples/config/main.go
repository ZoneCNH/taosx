package main

import (
	"fmt"
	"time"

	"github.com/ZoneCNH/taosx/pkg/taosx"
)

func main() {
	cfg := taosx.Config{
		Endpoint: "localhost:6041",
		Database: "metrics",
		Username: "root",
		Password: "taosdata",
		Timeout:  time.Second,
	}

	fmt.Println(cfg.Sanitized().Password)
}
