package scanner

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PackageInfo represents extracted package information
type PackageInfo struct {
	Name      string
	Version   string
	Ecosystem string
}

// PackageScanner handles scanning for package files
type PackageScanner struct {
	FileExtension string
	Ecosystem     string
	logger        *slog.Logger
}

// NewPackageScanner creates a new package scanner
func NewPackageScanner(extension string, ecosystem string, logger *slog.Logger) *PackageScanner {
	// Remove leading dot if present in the extension
	extension = strings.TrimPrefix(extension, ".")

	// Determine ecosystem if not provided
	if ecosystem == "" {
		ecosystem = determineEcosystem(extension)
	}

	// Use default logger if none provided
	if logger == nil {
		logger = slog.Default()
	}

	return &PackageScanner{
		FileExtension: extension,
		Ecosystem:     ecosystem,
		logger:        logger,
	}
}

// ScanDirectory scans a directory for packages with the specified extension
func (ps *PackageScanner) ScanDirectory(dirPath string) ([]PackageInfo, error) {
	var packages []PackageInfo

	// Ensure path exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error accessing directory %s: %w", dirPath, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dirPath)
	}

	// Walk the directory recursively
	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check file extension - case insensitive matching
		if !strings.HasSuffix(strings.ToLower(d.Name()), "."+strings.ToLower(ps.FileExtension)) {
			return nil
		}

		// Extract package info from filename - preserve original case
		pkg, err := ps.ExtractPackageInfo(d.Name())
		if err != nil {
			ps.logger.Warn("Could not parse package information",
				"filename", d.Name(),
				"error", err)
			return nil
		}

		// Add additional case sensitivity warning if applicable
		ps.logCaseSensitivityWarning(pkg.Name, pkg.Ecosystem)

		packages = append(packages, pkg)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %w", err)
	}

	return packages, nil
}

// ExtractPackageInfo extracts package name and version from filename
func (ps *PackageScanner) ExtractPackageInfo(filename string) (PackageInfo, error) {
	// Different parsing strategies based on extension/ecosystem
	var name, version string
	var err error

	switch strings.ToLower(ps.FileExtension) {
	case "nupkg":
		name, version, err = parseNuGetPackage(filename)
	case "tgz", "tar.gz":
		name, version, err = parseNpmPackage(filename)
	case "whl", "egg":
		name, version, err = parsePythonPackage(filename)
	case "jar":
		name, version, err = parseJavaPackage(filename)
	default:
		// Generic parsing strategy
		name, version, err = parseGenericPackage(filename)
	}

	if err != nil {
		return PackageInfo{}, err
	}

	return PackageInfo{
		Name:      name,
		Version:   version,
		Ecosystem: ps.Ecosystem,
	}, nil
}

// parseNuGetPackage extracts name and version from a NuGet package filename
// Format: PackageName.Version.nupkg (where PackageName may contain periods)
func parseNuGetPackage(filename string) (string, string, error) {
	logger := slog.Default()

	// Remove extension - case insensitive matching but preserve original case
	base := strings.TrimSuffix(strings.TrimSuffix(filename, ".nupkg"), ".NUPKG")

	// NuGet packages use the format PackageName.Version.nupkg or PackageName.Version.Suffix.nupkg
	// PackageName may contain periods (e.g., Microsoft.AspNetCore.Identity)

	// Start from the end and look for a valid version pattern
	parts := strings.Split(base, ".")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid NuGet package filename format: %s", filename)
	}

	// Define regex for version pattern that matches common NuGet version formats
	// This pattern looks for sequences like: 1.0, 1.0.0, 1.0.0.0, 1.0-beta, etc.
	versionRegex := regexp.MustCompile(`^\d+(\.\d+)+(-[a-zA-Z0-9.-]+)?$`)

	// Find the starting index of the version
	// We need to identify where the package name ends and the version begins
	versionStartIndex := -1

	// Start from the end and work backwards
	// We're looking for the longest consecutive sequence of version-like segments
	for i := len(parts) - 1; i >= 0; i-- {
		// Check if this part could be the start of a version
		if versionRegex.MatchString(strings.Join(parts[i:], ".")) {
			versionStartIndex = i
		} else if versionStartIndex != -1 {
			// If we've already found a version start and this part doesn't match,
			// we've gone too far back, so break
			break
		}
	}

	// If no version pattern was found, fall back to simpler detection
	if versionStartIndex == -1 {
		// Try a more lenient approach - look for any segment that starts with a number
		for i := len(parts) - 1; i >= 0; i-- {
			if regexp.MustCompile(`^\d+`).MatchString(parts[i]) {
				versionStartIndex = i
				break
			}
		}

		// If still no match, assume the last component is the version
		if versionStartIndex == -1 {
			versionStartIndex = len(parts) - 1
		}
	}

	// Extract the name and version
	name := strings.Join(parts[:versionStartIndex], ".")
	version := strings.Join(parts[versionStartIndex:], ".")

	// Log package extraction details at debug level
	logger.Debug("Extracted package details",
		"name", name,
		"version", version,
		"filename", filename)

	return name, version, nil
}

// parseNpmPackage extracts name and version from an NPM package filename
// Format: package-name-version.tgz
func parseNpmPackage(filename string) (string, string, error) {
	// Remove extension - case insensitive matching but preserve original case
	base := filename
	base = strings.TrimSuffix(strings.TrimSuffix(base, ".tgz"), ".TGZ")
	base = strings.TrimSuffix(strings.TrimSuffix(base, ".tar.gz"), ".TAR.GZ")

	// NPM packages often use format package-name-version.tgz
	versionRegex := regexp.MustCompile(`-(\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?)$`)
	matches := versionRegex.FindStringSubmatch(base)

	if len(matches) < 2 {
		return "", "", fmt.Errorf("could not extract version from npm package: %s", filename)
	}

	version := matches[1]
	name := strings.TrimSuffix(base, "-"+version)

	return name, version, nil
}

// parsePythonPackage extracts name and version from a Python package filename
// Format: package_name-version-info.whl or package_name-version.egg
func parsePythonPackage(filename string) (string, string, error) {
	// Remove extension - case insensitive matching but preserve original case
	base := filename
	base = strings.TrimSuffix(strings.TrimSuffix(base, ".whl"), ".WHL")
	base = strings.TrimSuffix(strings.TrimSuffix(base, ".egg"), ".EGG")

	// Split by first hyphen to separate name from version
	parts := strings.SplitN(base, "-", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid Python package filename: %s", filename)
	}

	// For wheels, the version might be followed by additional info
	// Try to extract just the version part
	versionParts := strings.SplitN(parts[1], "-", 2)
	version := versionParts[0]

	return parts[0], version, nil
}

// parseJavaPackage extracts name and version from a Java package filename
// Format: name-version.jar or name-version-classifier.jar
func parseJavaPackage(filename string) (string, string, error) {
	// Remove extension - case insensitive matching but preserve original case
	base := strings.TrimSuffix(strings.TrimSuffix(filename, ".jar"), ".JAR")

	// Try to find version pattern in the name
	versionRegex := regexp.MustCompile(`-(\d+\.\d+(\.\d+)?(-[a-zA-Z0-9.-]+)?)`)
	matches := versionRegex.FindStringSubmatch(base)

	if len(matches) < 2 {
		return "", "", fmt.Errorf("could not extract version from Java package: %s", filename)
	}

	version := matches[1]
	name := strings.TrimSuffix(base, "-"+version)

	return name, version, nil
}

// parseGenericPackage attempts to extract name and version using common patterns
func parseGenericPackage(filename string) (string, string, error) {
	// Case insensitive extension matching but preserve original filename case
	lastDotLower := strings.LastIndex(strings.ToLower(filename), ".")
	if lastDotLower < 0 {
		return "", "", fmt.Errorf("filename has no extension: %s", filename)
	}

	// Preserve the original case of the name portion
	base := filename[:lastDotLower]

	// Try common separators and patterns
	separators := []string{"-", "_", "."}
	versionRegex := regexp.MustCompile(`(\d+\.\d+(\.\d+)?(-[a-zA-Z0-9.-]+)?)$`)

	for _, sep := range separators {
		parts := strings.Split(base, sep)
		if len(parts) >= 2 {
			// Try to find a version-like string in the last part
			lastPart := parts[len(parts)-1]
			if versionRegex.MatchString(lastPart) {
				return strings.Join(parts[:len(parts)-1], sep), lastPart, nil
			}
		}
	}

	return "", "", fmt.Errorf("could not extract package name and version from: %s", filename)
}

// determineEcosystem determines the package ecosystem based on file extension
func determineEcosystem(extension string) string {
	switch strings.ToLower(extension) {
	case "nupkg":
		return "NuGet"
	case "tgz", "tar.gz":
		return "npm"
	case "whl", "egg":
		return "PyPI"
	case "jar":
		return "Maven"
	case "gem":
		return "RubyGems"
	case "deb":
		return "Debian"
	case "rpm":
		return "RPM"
	default:
		return "Unknown"
	}
}

// logCaseSensitivityWarning logs warnings about case-sensitive package ecosystems
func (ps *PackageScanner) logCaseSensitivityWarning(packageName string, ecosystem string) {
	// Certain ecosystems are known to be case-sensitive
	caseSensitiveEcosystems := map[string]bool{
		"npm":   true,
		"PyPI":  true,
		"Maven": true,
		"Cargo": true,
		"NuGet": true, // Adding NuGet to the list of case-sensitive ecosystems
	}

	if caseSensitiveEcosystems[ecosystem] &&
		(strings.ToLower(packageName) != packageName && strings.ToUpper(packageName) != packageName) {
		ps.logger.Warn("Case-sensitive package ecosystem",
			"ecosystem", ecosystem,
			"packageName", packageName)
	}
}
