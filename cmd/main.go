// Package main starts the gRPC student-service server.
package main

import (
	"log"

	"example.com/student-service/internal/app"
	"example.com/student-service/internal/config"
)

// ConfigPath defines the path to the YAML configuration file.
const ConfigPath = "config.yaml"

func main() {
	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.NewApp(cfg)

	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("application exited with error: %v", err)
	}
}
