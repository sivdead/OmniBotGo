// Package app configures and runs application.
package app

//go:generate wire

import (
	"fmt"
	"log"

	"github.com/sivdead/OmniBotGo/internal/config"
)

// Run creates objects via Wire dependency injection and starts the application.
// Returns a non-nil error when listen/runtime failures occur (F03); signal
// shutdown returns nil so main can exit 0.
func Run(cfg *config.Config) error {
	// 使用Wire初始化应用程序
	application, cleanup, err := InitializeApp(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer cleanup()

	return application.Run()
}

// MustRun is retained for callers that historically ignored errors; prefer Run.
func MustRun(cfg *config.Config) {
	if err := Run(cfg); err != nil {
		log.Fatalf("%v", err)
	}
}
