package scanner

import (
	"log/slog"
	"os"
	"sync"

	"github.com/squarehole/package-scanner/pkg/cli"
	"github.com/squarehole/package-scanner/pkg/db"
	"github.com/squarehole/package-scanner/pkg/osv"
	"github.com/squarehole/package-scanner/pkg/reporting"
)

// Controller handles the package scanning operations
type Controller struct {
	config     *cli.Config
	osvClient  *osv.Client
	reporter   *reporting.Reporter
	dbInstance *db.PostgresDB
	logger     *slog.Logger
}

// NewController creates a new scanner controller
func NewController(config *cli.Config) *Controller {
	// Use the default logger
	logger := slog.Default()

	controller := &Controller{
		config:    config,
		osvClient: osv.NewClient(config.OSVAPI),
		reporter:  reporting.NewReporter(logger),
		logger:    logger,
	}

	// Initialize database if needed
	if config.UseDB {
		dbConfig := db.Config{
			Host:     config.DBHost,
			Port:     config.DBPort,
			User:     config.DBUser,
			Password: config.DBPassword,
			DBName:   config.DBName,
			SSLMode:  config.DBSSLMode,
		}

		var err error
		controller.dbInstance, err = db.NewPostgresDB(dbConfig)
		if err != nil {
			logger.Error("Error connecting to PostgreSQL", "error", err)
			os.Exit(1)
		}

		// Initialize the database schema
		if err := controller.dbInstance.InitializeSchema(); err != nil {
			logger.Error("Error initializing database schema", "error", err)
			os.Exit(1)
		}
	}

	return controller
}

// Close cleans up resources
func (c *Controller) Close() {
	if c.dbInstance != nil {
		c.dbInstance.Close()
	}
}

// Run executes the scanning operation based on the current configuration
func (c *Controller) Run() {
	// Check if we're in directory scanning mode
	if c.config.DirectoryPath != "" && c.config.FileExtension != "" {
		c.runDirectoryScan()
	} else {
		c.runSinglePackageScan()
	}
}

// runSinglePackageScan performs a vulnerability check on a single package
func (c *Controller) runSinglePackageScan() {
	// Log query information with structured fields instead of format strings
	c.logger.Info("Querying OSV API for package",
		"name", c.config.PackageName,
		"version", c.config.PackageVersion,
		"ecosystem", c.config.PackageEcosystem)

	results, body, err := c.osvClient.QueryPackage(
		c.config.PackageName,
		c.config.PackageVersion,
		c.config.PackageEcosystem,
	)

	if err != nil {
		c.logger.Error("Error checking package vulnerabilities", "error", err)
		os.Exit(1)
	}

	// Display results
	c.reporter.DisplayResults(results, c.config.PackageName)

	// Save to database if requested AND vulnerabilities were found
	if c.config.UseDB && c.dbInstance != nil && len(results.Vulnerabilities) > 0 {
		err = c.dbInstance.SaveVulnerabilityResults(
			c.config.PackageName,
			c.config.PackageEcosystem,
			c.config.PackageVersion,
			results.Vulnerabilities,
			body,
		)

		if err != nil {
			c.logger.Error("Error saving results to database", "error", err)
			os.Exit(1)
		}

		c.logger.Info("Results saved to database",
			"packageName", c.config.PackageName,
			"vulnerabilitiesCount", len(results.Vulnerabilities))
	} else if c.config.UseDB && len(results.Vulnerabilities) == 0 {
		c.logger.Info("No vulnerabilities found. Nothing saved to database.")
	}

	// Write raw response to file
	err = os.WriteFile("api_response.json", body, 0644)
	if err != nil {
		c.logger.Warn("Could not write API response to file", "error", err)
	} else {
		c.logger.Info("Raw API response written to api_response.json")
	}
}

// runDirectoryScan scans a directory for packages and checks their vulnerabilities
func (c *Controller) runDirectoryScan() {
	c.reporter.DisplayDirectoryScanStart(c.config.DirectoryPath, c.config.FileExtension)

	// Create scanner with the logger
	packageScanner := NewPackageScanner(c.config.FileExtension, c.config.PackageEcosystem, c.logger)

	// Scan directory
	packages, err := packageScanner.ScanDirectory(c.config.DirectoryPath)
	if err != nil {
		c.logger.Error("Error scanning directory", "error", err)
		os.Exit(1)
	}

	c.reporter.DisplayPackagesFound(len(packages))

	// Create a semaphore to limit concurrency
	sem := make(chan bool, c.config.Concurrency)
	var wg sync.WaitGroup

	// Process each package file
	for _, pkg := range packages {
		wg.Add(1)
		sem <- true // Acquire semaphore

		go func(pkg PackageInfo) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			c.reporter.DisplayPackageScanStart(pkg.Name, pkg.Version, pkg.Ecosystem)

			// Run vulnerability check for this package
			results, body, err := c.osvClient.QueryPackage(pkg.Name, pkg.Version, pkg.Ecosystem)
			if err != nil {
				c.reporter.DisplayError("Error checking %s@%s: %v", pkg.Name, pkg.Version, err)
				return
			}

			// Display results
			c.reporter.DisplayResults(results, pkg.Name)

			// Save to database if requested AND vulnerabilities were found
			if c.config.UseDB && c.dbInstance != nil && len(results.Vulnerabilities) > 0 {
				err = c.dbInstance.SaveVulnerabilityResults(
					pkg.Name,
					pkg.Ecosystem,
					pkg.Version,
					results.Vulnerabilities,
					body,
				)

				if err != nil {
					c.reporter.DisplayError("Error saving results to database for %s: %v", pkg.Name, err)
				} else {
					c.reporter.DisplayInfo("Results for %s@%s saved to database.", pkg.Name, pkg.Version)
				}
			} else if c.config.UseDB && len(results.Vulnerabilities) == 0 {
				c.reporter.DisplayInfo("No vulnerabilities found for %s@%s. Nothing saved to database.", pkg.Name, pkg.Version)
			}
		}(pkg)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	c.reporter.DisplayScanSummary(len(packages))
}
