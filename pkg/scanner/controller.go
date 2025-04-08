package scanner

import (
	"log"
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
	reporter   *reporting.ConsoleReporter
	dbInstance *db.PostgresDB
}

// NewController creates a new scanner controller
func NewController(config *cli.Config) *Controller {
	controller := &Controller{
		config:    config,
		osvClient: osv.NewClient(config.OSVAPI),
		reporter:  reporting.NewConsoleReporter(),
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
			log.Fatalf("Error connecting to PostgreSQL: %v", err)
		}

		// Initialize the database schema
		if err := controller.dbInstance.InitializeSchema(); err != nil {
			log.Fatalf("Error initializing database schema: %v", err)
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
	c.reporter.DisplayInfo("Querying OSV API for:")
	c.reporter.DisplayInfo("  Package: %s", c.config.PackageName)
	c.reporter.DisplayInfo("  Version: %s", c.config.PackageVersion)
	c.reporter.DisplayInfo("  Ecosystem: %s", c.config.PackageEcosystem)

	results, body, err := c.osvClient.QueryPackage(
		c.config.PackageName,
		c.config.PackageVersion,
		c.config.PackageEcosystem,
	)

	if err != nil {
		log.Fatalf("Error checking package vulnerabilities: %v", err)
	}

	// Display results
	c.reporter.DisplayResults(results, c.config.PackageName)

	// Save to database if requested
	if c.config.UseDB && c.dbInstance != nil {
		err = c.dbInstance.SaveVulnerabilityResults(
			c.config.PackageName,
			c.config.PackageEcosystem,
			c.config.PackageVersion,
			results.Vulnerabilities,
			body,
		)

		if err != nil {
			log.Fatalf("Error saving results to database: %v", err)
		}

		c.reporter.DisplayInfo("Results successfully saved to PostgreSQL database.")
	}

	// Write raw response to file
	err = os.WriteFile("api_response.json", body, 0644)
	if err != nil {
		c.reporter.DisplayWarning("Could not write API response to file: %v", err)
	} else {
		c.reporter.DisplayInfo("Raw API response written to api_response.json")
	}
}

// runDirectoryScan scans a directory for packages and checks their vulnerabilities
func (c *Controller) runDirectoryScan() {
	c.reporter.DisplayDirectoryScanStart(c.config.DirectoryPath, c.config.FileExtension)

	// Create scanner
	packageScanner := NewPackageScanner(c.config.FileExtension, c.config.PackageEcosystem)

	// Scan directory
	packages, err := packageScanner.ScanDirectory(c.config.DirectoryPath)
	if err != nil {
		log.Fatalf("Error scanning directory: %v", err)
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

			// Save to database if requested
			if c.config.UseDB && c.dbInstance != nil {
				err = c.dbInstance.SaveVulnerabilityResults(
					pkg.Name,
					pkg.Ecosystem,
					pkg.Version,
					results.Vulnerabilities,
					body,
				)

				if err != nil {
					c.reporter.DisplayError("Error saving results to database for %s: %v", pkg.Name, err)
				}
			}
		}(pkg)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	c.reporter.DisplayScanSummary(len(packages))
}
