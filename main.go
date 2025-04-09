package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/squarehole/package-scanner/pkg/cli"
	"github.com/squarehole/package-scanner/pkg/logging"
	"github.com/squarehole/package-scanner/pkg/scanner"
	"github.com/squarehole/package-scanner/pkg/tui"
)

func main() {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or cannot be read. Using defaults and command line flags.")
	}

	// Check if any command-line arguments were provided
	tuiMode := len(os.Args) == 1 // If no args, run in TUI mode

	var config *cli.Config
	if tuiMode {
		// Launch TUI to collect configuration
		tuiConfig, err := tui.RunTUI()
		if err != nil {
			log.Fatalf("Error running TUI: %v", err)
		}

		// If user exited without submitting the form
		if tuiConfig == nil {
			log.Println("TUI closed without submitting. Exiting.")
			return
		}

		// Convert the TUI config to the CLI config
		config = convertTUIConfigToCLIConfig(tuiConfig)
	} else {
		// Create config from command-line flags and environment variables
		config = cli.NewConfig()
	}

	// Initialize logging system
	logConfig := logging.LogConfig{
		LogToFile:   config.LogToFile,
		LogFilePath: config.LogFilePath,
		MaxSize:     config.LogMaxSize,
		MaxBackups:  config.LogMaxBackups,
		MaxAge:      config.LogMaxAge,
		Compress:    config.LogCompress,
		Level:       parseLogLevel(config.LogLevel),
		Format:      logging.ParseLogFormat(config.LogFormat),
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

// convertTUIConfigToCLIConfig converts a TUI configuration to CLI configuration
func convertTUIConfigToCLIConfig(tuiConfig *tui.AppConfig) *cli.Config {
	config := &cli.Config{
		// Common fields for both modes
		DBHost:        tuiConfig.DBHost,
		DBPort:        tuiConfig.DBPort,
		DBUser:        tuiConfig.DBUser,
		DBPassword:    tuiConfig.DBPassword,
		DBName:        tuiConfig.DBName,
		DBSSLMode:     tuiConfig.DBSSLMode,
		UseDB:         tuiConfig.UseDB,
		LogToFile:     tuiConfig.LogToFile,
		LogFilePath:   tuiConfig.LogFilePath,
		LogMaxSize:    tuiConfig.LogMaxSize,
		LogMaxBackups: tuiConfig.LogMaxBackups,
		LogMaxAge:     tuiConfig.LogMaxAge,
		LogCompress:   tuiConfig.LogCompress,
		LogLevel:      tuiConfig.LogLevel,
		LogFormat:     tuiConfig.LogFormat,
	}

	// Mode-specific fields
	if tuiConfig.Mode == tui.SinglePackageMode {
		config.PackageName = tuiConfig.PackageName
		config.PackageVersion = tuiConfig.PackageVersion
		config.PackageEcosystem = tuiConfig.PackageEcosystem
		// Clear directory scan fields
		config.DirectoryPath = ""
		config.FileExtension = ""
	} else {
		// Directory scan mode
		config.DirectoryPath = tuiConfig.DirectoryPath
		config.FileExtension = tuiConfig.FileExtension
		config.Concurrency = tuiConfig.Concurrency
		// Clear single package fields
		config.PackageName = ""
		config.PackageVersion = ""
		config.PackageEcosystem = ""
	}

	return config
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
