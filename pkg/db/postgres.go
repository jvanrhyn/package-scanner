package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/squarehole/package-scanner/pkg/models"
)

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// VulnerabilityRecord represents a database record for vulnerability data
type VulnerabilityRecord struct {
	ID             int64
	PackageName    string
	Ecosystem      string
	Version        string
	VulnID         string
	Summary        string
	Published      time.Time
	SeverityRating string
	FixVersion     string
	RawResponse    []byte // JSON data
	CreatedAt      time.Time
}

// PostgresDB wraps a connection to PostgreSQL
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config Config) (*PostgresDB, error) {
	// Connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	// Open a connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to PostgreSQL: %v", err)
	}

	// Check the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping PostgreSQL: %v", err)
	}

	return &PostgresDB{db: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// InitializeSchema ensures the necessary tables exist
func (p *PostgresDB) InitializeSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS vulnerability_scans (
		id SERIAL PRIMARY KEY,
		package_name VARCHAR(255) NOT NULL,
		ecosystem VARCHAR(100) NOT NULL,
		version VARCHAR(100) NOT NULL,
		vuln_id VARCHAR(100) NOT NULL,
		summary TEXT,
		published TIMESTAMP,
		severity_rating VARCHAR(50),
		fix_version VARCHAR(100),
		raw_response JSONB,
		created_at TIMESTAMP DEFAULT NOW()
	);
	
	CREATE INDEX IF NOT EXISTS idx_vuln_scans_package ON vulnerability_scans(package_name, ecosystem, version);
	`

	_, err := p.db.Exec(schema)
	return err
}

// SaveVulnerabilityResults saves vulnerability scan results to the database
func (p *PostgresDB) SaveVulnerabilityResults(packageName string, ecosystem string, version string,
	vulnerabilities []models.Vulnerability, rawResponse []byte) error {

	// Begin a transaction
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the statement for inserting vulnerability records
	stmt, err := tx.Prepare(`
		INSERT INTO vulnerability_scans (
			package_name, ecosystem, version, vuln_id, summary,
			published, severity_rating, fix_version, raw_response
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// If no vulnerabilities were found, create a single record with empty vulnerability info
	if len(vulnerabilities) == 0 {
		// Raw response might be a JSON structure indicating no vulnerabilities
		_, err = stmt.Exec(
			packageName,
			ecosystem,
			version,
			"", // No vulnerability ID
			"No vulnerabilities found",
			time.Now(),
			"N/A",
			"N/A",
			rawResponse,
		)
		if err != nil {
			return err
		}
	} else {
		// For each vulnerability, create a record
		for _, vuln := range vulnerabilities {
			// Extract fix version
			fixVersion := findFixVersion(vuln, packageName)

			// Extract severity rating
			severityRating := getSeverityRating(vuln)

			// Insert the record
			_, err = stmt.Exec(
				packageName,
				ecosystem,
				version,
				vuln.ID,
				vuln.Summary,
				vuln.Published,
				severityRating,
				fixVersion,
				rawResponse,
			)
			if err != nil {
				return err
			}
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// GetLatestScans gets the most recent vulnerability scans
func (p *PostgresDB) GetLatestScans(limit int) ([]VulnerabilityRecord, error) {
	rows, err := p.db.Query(`
		SELECT id, package_name, ecosystem, version, vuln_id, summary, published, 
		       severity_rating, fix_version, raw_response, created_at
		FROM vulnerability_scans
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []VulnerabilityRecord{}
	for rows.Next() {
		var record VulnerabilityRecord
		err := rows.Scan(
			&record.ID,
			&record.PackageName,
			&record.Ecosystem,
			&record.Version,
			&record.VulnID,
			&record.Summary,
			&record.Published,
			&record.SeverityRating,
			&record.FixVersion,
			&record.RawResponse,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// Helper functions ported from main.go to make this package self-contained

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
		switch vuln.DBSpecific.Severity {
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

// Helper function to extract numeric CVSS score from a CVSS string
func extractCVSSScore(cvssString string) string {
	// Implementation copied from main.go
	// CVSS strings are typically in format "CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:L/I:L/A:H"
	if len(cvssString) == 0 {
		return "N/A"
	}

	// Simple heuristic to estimate score out of 10
	highCount := 0
	mediumCount := 0
	lowCount := 0

	if cvssString[0:4] == "CVSS" {
		// This is a proper CVSS string
		highCount = countOccurrences(cvssString, ":H")
		mediumCount = countOccurrences(cvssString, ":M")
		lowCount = countOccurrences(cvssString, ":L")
	}

	// Simple heuristic for a score out of 10
	score := float64(highCount)*3.0 + float64(mediumCount)*2.0 + float64(lowCount)*1.0
	if score > 10.0 {
		score = 10.0
	}

	return fmt.Sprintf("%.1f/10", score)
}

// countOccurrences counts how many times a substring appears in a string
func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
		}
	}
	return count
}
