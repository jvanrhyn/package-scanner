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

## Example Outputs

### Vulnerability Found

```
Querying OSV API for:
  Package: Microsoft.AspNetCore.Identity
  Version: 2.3.0
  Ecosystem: NuGet
Found 1 vulnerabilities:
1. ID: GHSA-2865-hh9g-w894
   Summary: Microsoft Security Advisory CVE-2025-24070: .NET Elevation of Privilege Vulnerability
   Published: 2025-03-11 19:24:11 +0000 UTC
   Severity Rating: 7.0-8.9/10
   Fix Version: 2.3.1
   ---
Raw API response written to api_response.json
```

### No Vulnerabilities Found

```
Querying OSV API for:
  Package: secure-package
  Version: 1.0.0
  Ecosystem: npm
No vulnerabilities found for the specified package and version.
Raw API response written to api_response.json
```

## Project Structure

```
.
├── main.go                       # Main application entry point
├── .env                          # Configuration environment variables
├── pkg/                          # Package directory
│   ├── db/                       # Database integration
│   │   └── postgres.go           # PostgreSQL operations
│   ├── models/                   # Data models
│   │   └── vulnerability.go      # Vulnerability data structures
│   └── scanner/                  # Package scanning utilities
│       └── scanner.go            # Package file scanning logic
```

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