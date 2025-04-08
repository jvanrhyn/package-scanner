package main

import (
	"log"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/squarehole/package-scanner/pkg/cli"
	"github.com/squarehole/package-scanner/pkg/logging"
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

	// Initialize logging system
	logConfig := logging.LogConfig{
		LogToFile:   config.LogToFile,
		LogFilePath: config.LogFilePath,
		MaxSize:     config.LogMaxSize,
		MaxBackups:  config.LogMaxBackups,
		MaxAge:      config.LogMaxAge,
		Compress:    config.LogCompress,
		Level:       parseLogLevel(config.LogLevel),
	}

	logger, err := logging.SetupLogger(logConfig)
	if err != nil {
		log.Fatalf("Error setting up logger: %v", err)
	}

	logger.Info("Package Scanner starting", "version", "1.0.0")

	// Create and run the scanner controller
	controller := scanner.NewController(config)
	defer controller.Close()

	controller.Run()
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
