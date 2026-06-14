package main

import (
	"log"
	"todoapp/config"
	"todoapp/internal/app"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app.RunTasks(cfg)
}
