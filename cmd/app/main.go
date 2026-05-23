package main

import (
	"fmt"
	"os"

	"todoapp/internal/app"
	"todoapp/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	app.Run(cfg)
}
