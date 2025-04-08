package cli

import (
	"flag"
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
	useDb := flag.Bool("save-db", false, "Save results to PostgreSQL database")

	// API options
	osvAPI := flag.String("osv-api", getEnvWithDefault("OSV_API_URL", "https://api.osv.dev/v1/query"), "OSV API URL")

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
