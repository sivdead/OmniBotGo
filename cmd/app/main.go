package main

import (
	"log"

	"github.com/sivdead/OmniBotGo/config"
	"github.com/sivdead/OmniBotGo/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
