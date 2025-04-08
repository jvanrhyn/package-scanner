package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/squarehole/package-scanner/pkg/db"
	"github.com/squarehole/package-scanner/pkg/models"
	"github.com/squarehole/package-scanner/pkg/scanner"
)

const osvAPIURL = "https://api.osv.dev/v1/query"

// OSVQuery represents the request structure for the OSV API
type OSVQuery struct {
	Version string           `json:"version"`
	Package *OSVPackageQuery `json:"package"`
}

// OSVPackageQuery represents the package part of an OSV query
type OSVPackageQuery struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

// Helper function to extract numeric CVSS score from a CVSS string
func extractCVSSScore(cvssString string) string {
	// CVSS strings are typically in format "CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:L/I:L/A:H"
	// We'll find the numeric part for display purposes
	if len(cvssString) == 0 {
		return "N/A"
	}

	// Extract base score from CVSS
	parts := strings.Split(cvssString, "/")
	if len(parts) < 1 {
		return "N/A"
	}

	// Try to convert the standard format into a numeric score
	// This is a simplified approximation
	highCount := strings.Count(cvssString, ":H")
	mediumCount := strings.Count(cvssString, ":M")
	lowCount := strings.Count(cvssString, ":L")

	// Simple heuristic to estimate score out of 10
	score := float64(highCount)*3.0 + float64(mediumCount)*2.0 + float64(lowCount)*1.0
	if score > 10.0 {
		score = 10.0
	}

	return fmt.Sprintf("%.1f/10", score)
}

// FindFixVersion returns the fixed version for a given package name in a vulnerability
func findFixVersion(vuln models.Vulnerability, packageName string) string {
	for _, affected := range vuln.Affected {
		if affected.Package.Name == packageName {
			for _, r := range affected.Ranges {
				for _, event := range r.Events {
					if event.Fixed != "" {
						return event.Fixed
					}
				}
			}
		}
	}
	return "No fix version found"
}

// GetSeverityRating gets the severity rating as a string from the vulnerability
func getSeverityRating(vuln models.Vulnerability) string {
	// Try to get CVSS score first
	if len(vuln.Severity) > 0 {
		for _, sev := range vuln.Severity {
			if sev.Type == "CVSS_V3" {
				return extractCVSSScore(sev.Score)
			}
		}
	}

	// Fallback to database_specific severity
	if vuln.DBSpecific.Severity != "" {
		switch strings.ToUpper(vuln.DBSpecific.Severity) {
		case "CRITICAL":
			return "9.0+/10"
		case "HIGH":
			return "7.0-8.9/10"
		case "MEDIUM":
			return "4.0-6.9/10"
		case "LOW":
			return "0.1-3.9/10"
		}
	}

	return "Unknown"
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

func main() {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or cannot be read. Using defaults and command line flags.")
	}

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

	// Parse command line flags
	flag.Parse()

	// Initialize DB connection if requested
	var postgres *db.PostgresDB
	if *useDb {
		dbConfig := db.Config{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Password: *dbPassword,
			DBName:   *dbName,
			SSLMode:  *dbSSLMode,
		}

		postgres, err = db.NewPostgresDB(dbConfig)
		if err != nil {
			log.Fatalf("Error connecting to PostgreSQL: %v", err)
		}
		defer postgres.Close()

		// Initialize the database schema
		if err := postgres.InitializeSchema(); err != nil {
			log.Fatalf("Error initializing database schema: %v", err)
		}
	}

	// Check if we're in directory scanning mode
	if *dirPath != "" && *fileExt != "" {
		// Directory scanning mode
		fmt.Printf("Scanning directory %s for %s files...\n", *dirPath, *fileExt)

		// Create scanner
		packageScanner := scanner.NewPackageScanner(*fileExt, *packageEcosystem)

		// Scan directory
		packages, err := packageScanner.ScanDirectory(*dirPath)
		if err != nil {
			log.Fatalf("Error scanning directory: %v", err)
		}

		fmt.Printf("Found %d package files to scan.\n", len(packages))

		// Create a semaphore to limit concurrency
		sem := make(chan bool, *concurrency)
		var wg sync.WaitGroup

		// Process each package file
		for _, pkg := range packages {
			wg.Add(1)
			sem <- true // Acquire semaphore

			go func(pkg scanner.PackageInfo) {
				defer wg.Done()
				defer func() { <-sem }() // Release semaphore

				fmt.Printf("Checking %s@%s (%s)...\n", pkg.Name, pkg.Version, pkg.Ecosystem)

				// Run vulnerability check for this package
				results, body, err := checkPackageVulnerabilities(pkg.Name, pkg.Version, pkg.Ecosystem)
				if err != nil {
					log.Printf("Error checking %s@%s: %v\n", pkg.Name, pkg.Version, err)
					return
				}

				// Display results
				displayResults(results, pkg.Name)

				// Save to database if requested
				if *useDb && postgres != nil {
					err = postgres.SaveVulnerabilityResults(
						pkg.Name,
						pkg.Ecosystem,
						pkg.Version,
						results.Vulnerabilities,
						body,
					)

					if err != nil {
						log.Printf("Error saving results to database for %s: %v\n", pkg.Name, err)
					}
				}
			}(pkg)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		fmt.Println("Scan completed.")

	} else {
		// Single package mode (original behavior)
		fmt.Printf("Querying OSV API for:\n")
		fmt.Printf("  Package: %s\n", *packageName)
		fmt.Printf("  Version: %s\n", *packageVersion)
		fmt.Printf("  Ecosystem: %s\n", *packageEcosystem)

		results, body, err := checkPackageVulnerabilities(*packageName, *packageVersion, *packageEcosystem)
		if err != nil {
			log.Fatalf("Error checking package vulnerabilities: %v", err)
		}

		// Display results
		displayResults(results, *packageName)

		// Save to database if requested
		if *useDb && postgres != nil {
			err = postgres.SaveVulnerabilityResults(
				*packageName,
				*packageEcosystem,
				*packageVersion,
				results.Vulnerabilities,
				body,
			)

			if err != nil {
				log.Fatalf("Error saving results to database: %v", err)
			}

			fmt.Println("Results successfully saved to PostgreSQL database.")
		}

		// Write raw response to file
		err = os.WriteFile("api_response.json", body, 0644)
		if err != nil {
			log.Printf("Warning: Could not write API response to file: %v", err)
		} else {
			fmt.Println("Raw API response written to api_response.json")
		}
	}
}

// checkPackageVulnerabilities queries the OSV API for vulnerabilities in a package
func checkPackageVulnerabilities(packageName, packageVersion, packageEcosystem string) (models.ScanResults, []byte, error) {
	// Create the query payload
	query := OSVQuery{
		Version: packageVersion,
		Package: &OSVPackageQuery{
			Name:      packageName,
			Ecosystem: packageEcosystem,
		},
	}

	// Convert the query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return models.ScanResults{}, nil, fmt.Errorf("error marshaling query to JSON: %v", err)
	}

	// Create an HTTP request
	req, err := http.NewRequest("POST", osvAPIURL, bytes.NewBuffer(queryJSON))
	if err != nil {
		return models.ScanResults{}, nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.ScanResults{}, nil, fmt.Errorf("error sending request to OSV API: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.ScanResults{}, nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		return models.ScanResults{}, body, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, body)
	}

	// Parse the response into our vulnerability model
	var results models.ScanResults
	if err := json.Unmarshal(body, &results); err != nil {
		// Handle different response formats
		var rawResponse map[string]interface{}
		if jsonErr := json.Unmarshal(body, &rawResponse); jsonErr == nil {
			// If the response has a "vulns" field, try to extract it
			if vulns, ok := rawResponse["vulns"]; ok {
				vulnsJSON, _ := json.Marshal(vulns)
				var vulnerabilities []models.Vulnerability
				if err := json.Unmarshal(vulnsJSON, &vulnerabilities); err == nil {
					results.Vulnerabilities = vulnerabilities
				}
			} else {
				// If "vulns" not present, check if the response is directly an array of vulnerabilities
				var vulnerabilities []models.Vulnerability
				if err := json.Unmarshal(body, &vulnerabilities); err == nil {
					results.Vulnerabilities = vulnerabilities
				}
			}
		}
	}

	return results, body, nil
}

// displayResults displays the vulnerability results to the console
func displayResults(results models.ScanResults, packageName string) {
	if len(results.Vulnerabilities) == 0 {
		fmt.Println("No vulnerabilities found for the specified package and version.")
	} else {
		fmt.Printf("Found %d vulnerabilities:\n", len(results.Vulnerabilities))
		for i, vuln := range results.Vulnerabilities {
			fmt.Printf("%d. ID: %s\n", i+1, vuln.ID)
			fmt.Printf("   Summary: %s\n", vuln.Summary)
			fmt.Printf("   Published: %s\n", vuln.Published)

			// Add severity rating
			severityRating := getSeverityRating(vuln)
			fmt.Printf("   Severity Rating: %s\n", severityRating)

			// Find fix version for the queried package
			fixVersion := findFixVersion(vuln, packageName)
			fmt.Printf("   Fix Version: %s\n", fixVersion)

			fmt.Println("   ---")
		}
	}
}
