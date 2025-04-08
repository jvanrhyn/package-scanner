package osv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/squarehole/package-scanner/pkg/models"
)

const defaultOSVAPIURL = "https://api.osv.dev/v1/query"

// Client represents an OSV API client
type Client struct {
	apiURL string
}

// PackageQuery represents the request structure for the OSV API
type PackageQuery struct {
	Version string       `json:"version"`
	Package *PackageInfo `json:"package"`
}

// PackageInfo represents the package part of an OSV query
type PackageInfo struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

// NewClient creates a new OSV API client
func NewClient(apiURL string) *Client {
	// Use default URL if not provided
	if apiURL == "" {
		apiURL = defaultOSVAPIURL
	}

	return &Client{
		apiURL: apiURL,
	}
}

// QueryPackage queries the OSV API for vulnerabilities in a package
func (c *Client) QueryPackage(packageName, packageVersion, packageEcosystem string) (models.ScanResults, []byte, error) {
	// Create the query payload
	query := PackageQuery{
		Version: packageVersion,
		Package: &PackageInfo{
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
	req, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(queryJSON))
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

// Helper functions for vulnerability analysis

// ExtractCVSSScore extracts a numeric CVSS score from a CVSS string
func ExtractCVSSScore(cvssString string) string {
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
func FindFixVersion(vuln models.Vulnerability, packageName string) string {
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
func GetSeverityRating(vuln models.Vulnerability) string {
	// Try to get CVSS score first
	if len(vuln.Severity) > 0 {
		for _, sev := range vuln.Severity {
			if sev.Type == "CVSS_V3" {
				return ExtractCVSSScore(sev.Score)
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
