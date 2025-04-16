package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	// Package scanning options
	PackageName      string
	PackageVersion   string
	PackageEcosystem string

	// Directory scanning options
	DirectoryPath string
	FileExtension string
	Concurrency   int

	// Database options
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	UseDB      bool

	// API options
	OSVAPI string

	// Logging options
	LogToFile     bool
	LogFilePath   string
	LogMaxSize    int
	LogMaxBackups int
	LogMaxAge     int
	LogCompress   bool
	LogLevel      string
	LogFormat     string
}

// NewConfig creates a new configuration by parsing command-line flags
// and environment variables
func NewConfig() *Config {
	config := &Config{}

	// Define command line flags for OSV query
	packageVersion := flag.String("version", "2.3.0", "The package version to query")
	packageName := flag.String("package", "Microsoft.AspNetCore.Identity", "The package name to query")
	packageEcosystem := flag.String("ecosystem", "NuGet", "The package ecosystem (npm, NuGet, PyPI, etc.)")

	// Define flags for directory scanning mode
	dirPath := flag.String("dir", "", "Directory path to scan for package files")
	fileExt := flag.String("ext", "", "File extension to scan for (e.g., nupkg, tgz)")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent API requests when scanning a directory")

	// Define database connection flags - use environment variables as defaults
	dbHost := flag.String("db-host", getEnvWithDefault("DB_HOST", "localhost"), "PostgreSQL database host")
	dbPort := flag.Int("db-port", getEnvIntWithDefault("DB_PORT", 5432), "PostgreSQL database port")
	dbUser := flag.String("db-user", getEnvWithDefault("DB_USER", "postgres"), "PostgreSQL database user")
	dbPassword := flag.String("db-password", getEnvWithDefault("DB_PASSWORD", ""), "PostgreSQL database password")
	dbName := flag.String("db-name", getEnvWithDefault("DB_NAME", "package_scanner"), "PostgreSQL database name")
	dbSSLMode := flag.String("db-sslmode", getEnvWithDefault("DB_SSL_MODE", "disable"), "PostgreSQL SSL mode (disable, require, verify-ca, verify-full)")

	// Debug the USE_DB value from environment
	useDbFromEnv := os.Getenv("USE_DB")
	fmt.Println("USE_DB environment value:", useDbFromEnv)

	useDb := flag.Bool("save-db", getEnvBoolWithDefault("USE_DB", false), "Save results to PostgreSQL database")

	// API options
	osvAPI := flag.String("osv-api", getEnvWithDefault("OSV_API_URL", "https://api.osv.dev/v1/query"), "OSV API URL")

	// Logging options
	logToFile := flag.Bool("log-to-file", getEnvBoolWithDefault("LOG_TO_FILE", true), "Whether to log to file (in addition to stdout)")
	logFilePath := flag.String("log-file", getEnvWithDefault("LOG_FILE_PATH", "logs/package-scanner.log"), "Log file path")
	logMaxSize := flag.Int("log-max-size", getEnvIntWithDefault("LOG_MAX_SIZE", 10), "Maximum size of log file in MB before rotation")
	logMaxBackups := flag.Int("log-max-backups", getEnvIntWithDefault("LOG_MAX_BACKUPS", 5), "Maximum number of log files to keep")
	logMaxAge := flag.Int("log-max-age", getEnvIntWithDefault("LOG_MAX_AGE", 30), "Maximum age of log files in days")
	logCompress := flag.Bool("log-compress", getEnvBoolWithDefault("LOG_COMPRESS", true), "Whether to compress rotated logs")
	logLevel := flag.String("log-level", getEnvWithDefault("LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
	logFormat := flag.String("log-format", getEnvWithDefault("LOG_FORMAT", "json"), "Log format (json, text)")

	// Parse command line flags
	flag.Parse()

	// Set config from parsed flags
	config.PackageName = *packageName
	config.PackageVersion = *packageVersion
	config.PackageEcosystem = *packageEcosystem
	config.DirectoryPath = *dirPath
	config.FileExtension = *fileExt
	config.Concurrency = *concurrency
	config.DBHost = *dbHost
	config.DBPort = *dbPort
	config.DBUser = *dbUser
	config.DBPassword = *dbPassword
	config.DBName = *dbName
	config.DBSSLMode = *dbSSLMode
	config.UseDB = *useDb
	config.OSVAPI = *osvAPI
	config.LogToFile = *logToFile
	config.LogFilePath = *logFilePath
	config.LogMaxSize = *logMaxSize
	config.LogMaxBackups = *logMaxBackups
	config.LogMaxAge = *logMaxAge
	config.LogCompress = *logCompress
	config.LogLevel = *logLevel
	config.LogFormat = *logFormat

	return config
}

// getEnvWithDefault gets an environment variable or returns a default value if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvIntWithDefault gets an environment variable as an int or returns a default value if not set
func getEnvIntWithDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvBoolWithDefault gets an environment variable as a bool or returns a default value if not set
func getEnvBoolWithDefault(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	fmt.Printf("DEBUG: Reading environment variable %s: '%s'\n", key, valueStr)

	if valueStr == "" {
		fmt.Printf("DEBUG: Using default value for %s: %v\n", key, defaultValue)
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		fmt.Printf("DEBUG: Error parsing %s as bool: %v, using default: %v\n", key, err, defaultValue)
		return defaultValue
	}

	fmt.Printf("DEBUG: Successfully parsed %s as bool: %v\n", key, value)
	return value
}
