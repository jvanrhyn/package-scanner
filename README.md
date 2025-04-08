# Package Scanner

A vulnerability scanning tool for software packages that uses the OSV (Open Source Vulnerability) database.

## Overview

Package Scanner is a Go-based command-line utility that scans software packages for known security vulnerabilities. It can query the OSV API for specific package versions or scan directories of package files (such as `.nupkg`, `.tgz`, etc.), automatically extracting package name and version information from filenames.

Results can be displayed to the console and optionally stored in a PostgreSQL database for historical tracking and analysis.

## Features

- Query vulnerabilities for specific package versions
- Scan directories for package files with automatic package data extraction
- Parallel processing of multiple packages
- Support for multiple package ecosystems:
  - NuGet (`.nupkg`)
  - npm (`.tgz`, `.tar.gz`) 
  - Python (`.whl`, `.egg`)
  - Java/Maven (`.jar`)
  - And more with generic fallback parsing
- Detailed vulnerability information including:
  - Vulnerability ID and summary
  - Severity rating (out of 10)
  - Minimum version to fix vulnerability
- PostgreSQL database integration for storing results
- Environment variable configuration via `.env` files

## Requirements

- Go 1.16 or higher
- PostgreSQL (if using the database storage feature)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/squarehole/package-scanner.git
cd package-scanner

# Build the project
go build -o package-scanner
```

### Dependencies

Package Scanner uses the following external dependencies:

- `github.com/lib/pq` - PostgreSQL driver
- `github.com/joho/godotenv` - Environment variable loading from .env files

To install dependencies:

```bash
go mod download
```

## Configuration

### Database Configuration

Database connection parameters can be configured through:

1. Environment variables
2. A `.env` file in the project root
3. Command-line flags

Create a `.env` file in the project root with the following structure:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=package_scanner
DB_SSL_MODE=disable
```

### API Configuration

You can optionally override the OSV API URL:

```
OSV_API_URL=https://api.osv.dev/v1/query
```

### Database Setup

Package Scanner will automatically create the necessary tables on first run. Ensure your PostgreSQL user has sufficient privileges to create tables.

## Usage

### Single Package Scan

To check a specific package for vulnerabilities:

```bash
# Basic usage
./package-scanner --package="Microsoft.AspNetCore.Identity" --version="2.3.0" --ecosystem="NuGet"

# Save results to database
./package-scanner --package="lodash" --version="4.17.15" --ecosystem="npm" --save-db
```

### Directory Scan

To scan a directory for package files:

```bash
# Scan a directory of NuGet packages
./package-scanner --dir="./packages" --ext="nupkg" --save-db

# Scan npm packages with a custom concurrency level
./package-scanner --dir="./node_packages" --ext="tgz" --ecosystem="npm" --concurrency=10
```

### Command Line Options

#### Package Query Parameters

| Flag | Description | Default |
|------|-------------|---------|
| `--package` | Package name to query | "Microsoft.AspNetCore.Identity" |
| `--version` | Package version to query | "2.3.0" |
| `--ecosystem` | Package ecosystem (npm, NuGet, PyPI, etc.) | "NuGet" |

#### Directory Scanning Parameters

| Flag | Description | Default |
|------|-------------|---------|
| `--dir` | Directory path to scan for package files | "" |
| `--ext` | File extension to scan for (e.g., nupkg, tgz) | "" |
| `--concurrency` | Number of concurrent API requests when scanning | 5 |

#### Database Parameters

| Flag | Description | Default/Source |
|------|-------------|---------------|
| `--save-db` | Save results to PostgreSQL | false |
| `--db-host` | PostgreSQL host | From `.env` or "localhost" |
| `--db-port` | PostgreSQL port | From `.env` or 5432 |
| `--db-user` | PostgreSQL user | From `.env` or "postgres" |
| `--db-password` | PostgreSQL password | From `.env` or "" |
| `--db-name` | PostgreSQL database name | From `.env` or "package_scanner" |
| `--db-sslmode` | PostgreSQL SSL mode | From `.env` or "disable" |

#### API Parameters

| Flag | Description | Default/Source |
|------|-------------|---------------|
| `--osv-api` | OSV API URL | From `.env` or "https://api.osv.dev/v1/query" |

## Example Outputs

### Vulnerability Found

```
Info: Querying OSV API for:
Info:   Package: Microsoft.AspNetCore.Identity
Info:   Version: 2.3.0
Info:   Ecosystem: NuGet
Found 1 vulnerabilities:
1. ID: GHSA-2865-hh9g-w894
   Summary: Microsoft Security Advisory CVE-2025-24070: .NET Elevation of Privilege Vulnerability
   Published: 2025-03-11 19:24:11 +0000 UTC
   Severity Rating: 7.0-8.9/10
   Fix Version: 2.3.1
   ---
Info: Raw API response written to api_response.json
```

### No Vulnerabilities Found

```
Info: Querying OSV API for:
Info:   Package: secure-package
Info:   Version: 1.0.0
Info:   Ecosystem: npm
No vulnerabilities found for the specified package and version.
Info: Raw API response written to api_response.json
```

## Project Structure

```
.
├── main.go                       # Main application entry point 
├── .env                          # Configuration environment variables
├── pkg/                          # Package directory
│   ├── cli/                      # Command line interface
│   │   └── config.go             # Configuration management
│   ├── db/                       # Database integration
│   │   └── postgres.go           # PostgreSQL operations
│   ├── models/                   # Data models
│   │   └── vulnerability.go      # Vulnerability data structures
│   ├── osv/                      # OSV API integration
│   │   └── client.go             # OSV API client
│   ├── reporting/                # Output formatting
│   │   └── console.go            # Console reporting
│   └── scanner/                  # Package scanning utilities
│       ├── controller.go         # Scanning orchestration
│       └── scanner.go            # Package file scanning logic
```

## Architecture

Package Scanner follows a modular design with clear separation of concerns:

1. **CLI** (`pkg/cli`) - Handles command-line arguments and environment configuration
2. **Scanner** (`pkg/scanner`) - Core scanning functionality and orchestration
3. **OSV** (`pkg/osv`) - Interacts with the Open Source Vulnerability API
4. **DB** (`pkg/db`) - Manages database operations for storing scan results
5. **Models** (`pkg/models`) - Data structures shared across the application
6. **Reporting** (`pkg/reporting`) - Formats and displays scan results

This modular architecture makes the application easier to extend and maintain.

## Database Schema

The application creates the following database table:

**vulnerability_scans**

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| package_name | VARCHAR(255) | Package name |
| ecosystem | VARCHAR(100) | Package ecosystem |
| version | VARCHAR(100) | Package version |
| vuln_id | VARCHAR(100) | Vulnerability ID |
| summary | TEXT | Vulnerability summary |
| published | TIMESTAMP | Vulnerability publish date |
| severity_rating | VARCHAR(50) | Severity rating (e.g., "7.5/10") |
| fix_version | VARCHAR(100) | Version that fixes the vulnerability |
| raw_response | JSONB | Complete API response as JSON |
| created_at | TIMESTAMP | Record creation time |

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.