// Package app configures and runs application.
package app

//go:generate wire

import (
	"log"

	"github.com/sivdead/OmniBotGo/config"
)

// Run creates objects via Wire dependency injection and starts the application.
func Run(cfg *config.Config) {
	// 使用Wire初始化应用程序
	app, err := InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// 启动应用程序
	app.Run()
}
