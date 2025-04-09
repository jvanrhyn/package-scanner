package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LogFormat defines the logging format
type LogFormat string

const (
	// JSONFormat represents structured JSON logging format
	JSONFormat LogFormat = "json"
	// TextFormat represents human-readable text logging format
	TextFormat LogFormat = "text"
)

// LogConfig represents configuration options for logging
type LogConfig struct {
	// Whether to log to file (in addition to stdout)
	LogToFile bool
	// Log file path
	LogFilePath string
	// Maximum size of log file in MB before rotation
	MaxSize int
	// Maximum number of log files to keep
	MaxBackups int
	// Maximum age of log files in days
	MaxAge int
	// Whether to compress rotated logs
	Compress bool
	// Minimum log level
	Level slog.Level
	// Logging format (json or text)
	Format LogFormat
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() LogConfig {
	return LogConfig{
		LogToFile:   true,
		LogFilePath: "logs/package-scanner.log",
		MaxSize:     10, // 10 MB
		MaxBackups:  5,
		MaxAge:      30, // 30 days
		Compress:    true,
		Level:       slog.LevelInfo,
		Format:      JSONFormat,
	}
}

// ParseLogFormat converts a string to LogFormat
func ParseLogFormat(format string) LogFormat {
	format = strings.ToLower(format)
	if format == "text" {
		return TextFormat
	}
	// Default to JSON for any other value
	return JSONFormat
}

// SetupLogger configures the global logger with file rotation
func SetupLogger(config LogConfig) (*slog.Logger, error) {
	var writer io.Writer

	// Create multi-writer if logging to file
	if config.LogToFile {
		// Create logs directory if it doesn't exist
		err := os.MkdirAll(filepath.Dir(config.LogFilePath), 0755)
		if err != nil {
			return nil, err
		}

		// Configure lumberjack for log rotation
		fileLogger := &lumberjack.Logger{
			Filename:   config.LogFilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}

		// Log to both stdout and file
		writer = io.MultiWriter(os.Stdout, fileLogger)
	} else {
		// Log to stdout only
		writer = os.Stdout
	}

	// Create slog handler based on format
	var handler slog.Handler

	switch config.Format {
	case TextFormat:
		// Use text handler for human-readable format
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: config.Level,
		})
	default:
		// Use JSON handler (default)
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: config.Level,
		})
	}

	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	// Log the format we're using
	logger.Info("Logger initialized",
		"format", string(config.Format),
		"level", config.Level.String(),
		"toFile", config.LogToFile)

	return logger, nil
}
