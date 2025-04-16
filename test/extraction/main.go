package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/squarehole/package-scanner/pkg/scanner"
)

func main() {
	// Create a logger with text output for easier reading in the test
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Initialize the scanner for NuGet packages
	packageScanner := scanner.NewPackageScanner("nupkg", "NuGet", logger)

	fmt.Println("Testing package name and version extraction from filenames...")

	// Test with sample filenames (not real files)
	testFilenames := []string{
		"System.Text.RegularExpressions.4.3.0.nupkg",
		"System.Threading.4.3.0.nupkg",
		"Microsoft.AspNetCore.Identity.3.1.10.nupkg",
		"Newtonsoft.Json.13.0.1.nupkg",
	}

	for _, filename := range testFilenames {
		// Extract package info from the filename string
		// Note: This is only testing the parsing logic, not actual file operations
		pkgInfo, err := packageScanner.ExtractPackageInfo(filename)
		if err != nil {
			fmt.Printf("Error extracting package info from %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("Extracted package: name='%s', version='%s' from '%s'\n",
			pkgInfo.Name, pkgInfo.Version, filename)
	}
}
