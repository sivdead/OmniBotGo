package main

import (
	"log"
	"os"

	"github.com/sivdead/OmniBotGo/internal/app"
	"github.com/sivdead/OmniBotGo/internal/config"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run — listen/runtime failures exit non-zero; clean signal shutdown exits 0 (F03).
	if err := app.Run(cfg); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}
}
