package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/squarehole/package-scanner/pkg/cli"
	"github.com/squarehole/package-scanner/pkg/scanner"
)

func main() {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or cannot be read. Using defaults and command line flags.")
	}

	// Create config from command-line flags and environment variables
	config := cli.NewConfig()

	// Create and run the scanner controller
	controller := scanner.NewController(config)
	defer controller.Close()

	controller.Run()
}
