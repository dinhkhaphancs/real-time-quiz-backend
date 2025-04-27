package main

import (
	"log"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/bootstrap"
)

func main() {
	// Initialize the application
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Stop()

	// Start the application
	app.Start()
}
