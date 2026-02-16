package main

import (
	"log"

	"github.com/leal-hospital/server/app"
	"github.com/leal-hospital/server/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize modules
	modules := []app.Module{
		// Modules will be added here as we implement them
		// Example: &user.UserModule{},
		// Example: &patient.PatientModule{},
	}

	// Create and bootstrap application
	application := app.NewApp(cfg, modules)
	if err := application.Bootstrap(); err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}

	// Run application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
