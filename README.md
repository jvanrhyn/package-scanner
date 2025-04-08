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
- PostgreSQL database integration for storing vulnerability results
- Environment variable configuration via `.env` files
- Smart package name and version extraction from filenames
- Structured logging with log rotation

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
- `gopkg.in/natefinch/lumberjack.v2` - Log rotation and management

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

### Logging Configuration

The application can log to both the console and a rotating log file. Configure logging with:

```
# Control whether to write logs to file (in addition to stdout)
LOG_TO_FILE=true
# Path to log file
LOG_FILE_PATH=logs/package-scanner.log
# Maximum size of log file in MB before rotation
LOG_MAX_SIZE=10
# Maximum number of old log files to retain
LOG_MAX_BACKUPS=5
# Maximum age of log files in days
LOG_MAX_AGE=30
# Whether to compress old log files
LOG_COMPRESS=true
# Minimum log level (debug, info, warn, error)
LOG_LEVEL=info
```

### Database Setup

Package Scanner will automatically create the necessary tables on first run. Ensure your PostgreSQL user has sufficient privileges to create tables.

**Note:** The application only saves results to the database when vulnerabilities are found. This keeps your database clean and focused on actual security issues.

## Usage

### Single Package Scan

To check a specific package for vulnerabilities:

```bash
# Basic usage
./package-scanner --package="Microsoft.AspNetCore.Identity" --version="2.3.0" --ecosystem="NuGet"

# Save results to database (only if vulnerabilities are found)
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

#### Logging Parameters

| Flag | Description | Default/Source |
|------|-------------|---------------|
| `--log-to-file` | Whether to log to file (in addition to stdout) | From `.env` or "true" |
| `--log-file` | Log file path | From `.env` or "logs/package-scanner.log" |
| `--log-max-size` | Maximum size of log file in MB before rotation | From `.env` or 10 |
| `--log-max-backups` | Maximum number of old log files to retain | From `.env` or 5 |
| `--log-max-age` | Maximum age of log files in days | From `.env` or 30 |
| `--log-compress` | Whether to compress old log files | From `.env` or "true" |
| `--log-level` | Minimum log level (debug, info, warn, error) | From `.env` or "info" |

## Example Outputs

### Console Output (JSON structured logging)
```json
{"time":"2025-04-08T10:45:22.123Z","level":"INFO","msg":"Package Scanner starting","version":"1.0.0"}
{"time":"2025-04-08T10:45:22.234Z","level":"INFO","msg":"Scanning package","name":"Microsoft.AspNetCore.Identity","version":"2.3.0","ecosystem":"NuGet"}
{"time":"2025-04-08T10:45:23.345Z","level":"INFO","msg":"Vulnerabilities found","count":1}
{"time":"2025-04-08T10:45:23.456Z","level":"INFO","msg":"Vulnerability details","index":1,"id":"GHSA-2865-hh9g-w894","summary":"Microsoft Security Advisory CVE-2025-24070","published":"2025-03-11T19:24:11Z","severity":"7.0-8.9/10","fixVersion":"2.3.1"}
{"time":"2025-04-08T10:45:23.567Z","level":"INFO","msg":"Results successfully saved to PostgreSQL database."}
{"time":"2025-04-08T10:45:23.678Z","level":"INFO","msg":"Raw API response written to api_response.json"}
```

### Log File Output

The same structured logs are written to the configured log file with automatic rotation when it reaches the configured maximum size.

## Package Name Extraction

The scanner automatically detects package names and versions from filenames. It has specific handling for:

- **NuGet** packages: Correctly extracts names with multiple segments (like `Microsoft.AspNetCore.Identity.2.3.0.nupkg`)
- **npm** packages: Handles packages with hyphens in names and versions (like `lodash-4.17.15.tgz`)
- **Python** packages: Processes wheel and egg formats correctly
- **Java** packages: Extracts Maven artifact information from JAR files

If the package format is not recognized, a fallback parser attempts to extract names and versions using common conventions.

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
│   ├── logging/                  # Logging subsystem
│   │   └── logger.go             # Structured logging with rotation
│   ├── models/                   # Data models
│   │   └── vulnerability.go      # Vulnerability data structures
│   ├── osv/                      # OSV API integration
│   │   └── client.go             # OSV API client
│   ├── reporting/                # Output formatting
│   │   └── console.go            # Console reporting
│   └── scanner/                  # Package scanning utilities
│       ├── controller.go         # Scanning orchestration
│       └── scanner.go            # Package file scanning logic
├── logs/                         # Directory for log files
│   └── package-scanner.log       # Application logs with rotation
├── test/                         # Test directory
│   └── extraction/               # Package extraction tests
│       └── main.go               # Test code for extraction logic
```

## Architecture

Package Scanner follows a modular design with clear separation of concerns:

1. **CLI** (`pkg/cli`) - Handles command-line arguments and environment configuration
2. **Scanner** (`pkg/scanner`) - Core scanning functionality and orchestration
3. **OSV** (`pkg/osv`) - Interacts with the Open Source Vulnerability API
4. **DB** (`pkg/db`) - Manages database operations for storing scan results
5. **Models** (`pkg/models`) - Data structures shared across the application
6. **Reporting** (`pkg/reporting`) - Formats and displays scan results
7. **Logging** (`pkg/logging`) - Structured logging with file rotation

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