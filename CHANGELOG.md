# Changelog

All notable changes to the Package Scanner project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-05-05

### Added
- Initial release of Package Scanner
- Support for scanning multiple package ecosystems (NuGet, npm, PyPI, Maven)
- Interactive Terminal UI (TUI) for easy parameter entry
- Command-line interface with multiple options
- PostgreSQL database integration for vulnerability storage
- Structured logging with log rotation
- Environment variable configuration via .env files
- Concurrent package scanning with configurable limits

### Changed
- Only write vulnerability results to the database when vulnerabilities are found
- Improved package name and version extraction from filenames
- Enhanced logging for database operations

### Fixed
- Case sensitivity warnings for package ecosystems
- Proper error handling for database connections
- Improved version extraction from package filenames

## [Unreleased]

### Added
- Environment variable `USE_DB` to control database writes from .env file

### Changed
- Enhanced logging for database write operations to show when entries are being saved
- Improved handling of packages without vulnerabilities

### Fixed
- Issues with .env file loading and environment variable recognition