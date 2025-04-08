package reporting

import (
	"fmt"

	"github.com/squarehole/package-scanner/pkg/models"
	"github.com/squarehole/package-scanner/pkg/osv"
)

// ConsoleReporter handles reporting vulnerability scan results to the console
type ConsoleReporter struct{}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

// DisplayResults displays the vulnerability results to the console
func (r *ConsoleReporter) DisplayResults(results models.ScanResults, packageName string) {
	if len(results.Vulnerabilities) == 0 {
		fmt.Println("No vulnerabilities found for the specified package and version.")
	} else {
		fmt.Printf("Found %d vulnerabilities:\n", len(results.Vulnerabilities))
		for i, vuln := range results.Vulnerabilities {
			fmt.Printf("%d. ID: %s\n", i+1, vuln.ID)
			fmt.Printf("   Summary: %s\n", vuln.Summary)
			fmt.Printf("   Published: %s\n", vuln.Published)

			// Add severity rating
			severityRating := osv.GetSeverityRating(vuln)
			fmt.Printf("   Severity Rating: %s\n", severityRating)

			// Find fix version for the queried package
			fixVersion := osv.FindFixVersion(vuln, packageName)
			fmt.Printf("   Fix Version: %s\n", fixVersion)

			fmt.Println("   ---")
		}
	}
}

// DisplayScanSummary displays a summary of the scan operation
func (r *ConsoleReporter) DisplayScanSummary(packageCount int) {
	fmt.Printf("Scan completed. Processed %d packages.\n", packageCount)
}

// DisplayPackageScanStart displays information about scanning a package
func (r *ConsoleReporter) DisplayPackageScanStart(name, version, ecosystem string) {
	fmt.Printf("Checking %s@%s (%s)...\n", name, version, ecosystem)
}

// DisplayDirectoryScanStart displays information about starting a directory scan
func (r *ConsoleReporter) DisplayDirectoryScanStart(dirPath, fileExt string) {
	fmt.Printf("Scanning directory %s for %s files...\n", dirPath, fileExt)
}

// DisplayPackagesFound displays information about found packages
func (r *ConsoleReporter) DisplayPackagesFound(count int) {
	fmt.Printf("Found %d package files to scan.\n", count)
}

// DisplayError displays an error message
func (r *ConsoleReporter) DisplayError(format string, args ...interface{}) {
	fmt.Printf("Error: "+format+"\n", args...)
}

// DisplayWarning displays a warning message
func (r *ConsoleReporter) DisplayWarning(format string, args ...interface{}) {
	fmt.Printf("Warning: "+format+"\n", args...)
}

// DisplayInfo displays an information message
func (r *ConsoleReporter) DisplayInfo(format string, args ...interface{}) {
	fmt.Printf("Info: "+format+"\n", args...)
}
