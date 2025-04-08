package reporting

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/squarehole/package-scanner/pkg/models"
	"github.com/squarehole/package-scanner/pkg/osv"
)

// Reporter handles reporting vulnerability scan results
type Reporter struct {
	logger *slog.Logger
}

// NewReporter creates a new reporter with structured logging
func NewReporter(logger *slog.Logger) *Reporter {
	if logger == nil {
		logger = slog.Default()
	}
	return &Reporter{
		logger: logger,
	}
}

// DisplayResults displays the vulnerability results
func (r *Reporter) DisplayResults(results models.ScanResults, packageName string) {
	if len(results.Vulnerabilities) == 0 {
		r.logger.Info("No vulnerabilities found for the specified package and version.")
	} else {
		r.logger.Info("Vulnerabilities found", "count", len(results.Vulnerabilities))

		for i, vuln := range results.Vulnerabilities {
			// Extract severity rating and fix version
			severityRating := osv.GetSeverityRating(vuln)
			fixVersion := osv.FindFixVersion(vuln, packageName)

			// Log each vulnerability as a structured log entry
			r.logger.Info("Vulnerability details",
				"index", i+1,
				"id", vuln.ID,
				"summary", vuln.Summary,
				"published", vuln.Published,
				"severity", severityRating,
				"fixVersion", fixVersion,
			)
		}
	}
}

// DisplayScanSummary displays a summary of the scan operation
func (r *Reporter) DisplayScanSummary(packageCount int) {
	r.logger.Info("Scan completed", "packagesProcessed", packageCount)
}

// DisplayPackageScanStart displays information about scanning a package
func (r *Reporter) DisplayPackageScanStart(name, version, ecosystem string) {
	r.logger.Info("Scanning package",
		"name", name,
		"version", version,
		"ecosystem", ecosystem,
	)
}

// DisplayDirectoryScanStart displays information about starting a directory scan
func (r *Reporter) DisplayDirectoryScanStart(dirPath, fileExt string) {
	r.logger.Info("Scanning directory",
		"path", dirPath,
		"fileExtension", fileExt,
	)
}

// DisplayPackagesFound displays information about found packages
func (r *Reporter) DisplayPackagesFound(count int) {
	r.logger.Info("Package files found", "count", count)
}

// DisplayError displays an error message
func (r *Reporter) DisplayError(format string, args ...interface{}) {
	// Check if this is a printf-style message that needs formatting
	if len(args) > 0 && (strings.Contains(format, "%s") ||
		strings.Contains(format, "%d") ||
		strings.Contains(format, "%v")) {
		// This is a printf-style format string, format it properly
		formattedMsg := fmt.Sprintf(format, args...)
		r.logger.Error(formattedMsg)
	} else if len(args) > 0 {
		// Non-format string with attributes for structured logging
		// Build key-value pairs for structured logging
		keyVals := make([]any, 0, len(args)*2)
		for i := 0; i < len(args); i++ {
			// Create a default key name if needed
			key := fmt.Sprintf("param%d", i)
			keyVals = append(keyVals, key, args[i])
		}
		r.logger.Error(format, keyVals...)
	} else {
		// Simple message with no args
		r.logger.Error(format)
	}
}

// DisplayWarning displays a warning message
func (r *Reporter) DisplayWarning(format string, args ...interface{}) {
	// Check if this is a printf-style message that needs formatting
	if len(args) > 0 && (strings.Contains(format, "%s") ||
		strings.Contains(format, "%d") ||
		strings.Contains(format, "%v")) {
		// This is a printf-style format string, format it properly
		formattedMsg := fmt.Sprintf(format, args...)
		r.logger.Warn(formattedMsg)
	} else if len(args) > 0 {
		// Non-format string with attributes for structured logging
		// Build key-value pairs for structured logging
		keyVals := make([]any, 0, len(args)*2)
		for i := 0; i < len(args); i++ {
			// Create a default key name if needed
			key := fmt.Sprintf("param%d", i)
			keyVals = append(keyVals, key, args[i])
		}
		r.logger.Warn(format, keyVals...)
	} else {
		// Simple message with no args
		r.logger.Warn(format)
	}
}

// DisplayInfo displays an information message
func (r *Reporter) DisplayInfo(format string, args ...interface{}) {
	// Check if this is a printf-style message that needs formatting
	if len(args) > 0 && (strings.Contains(format, "%s") ||
		strings.Contains(format, "%d") ||
		strings.Contains(format, "%v")) {
		// This is a printf-style format string, format it properly
		formattedMsg := fmt.Sprintf(format, args...)
		r.logger.Info(formattedMsg)
	} else if len(args) > 0 {
		// Non-format string with attributes for structured logging
		// Build key-value pairs for structured logging
		keyVals := make([]any, 0, len(args)*2)
		for i := 0; i < len(args); i++ {
			// Create a default key name if needed
			key := fmt.Sprintf("param%d", i)
			keyVals = append(keyVals, key, args[i])
		}
		r.logger.Info(format, keyVals...)
	} else {
		// Simple message with no args
		r.logger.Info(format)
	}
}
