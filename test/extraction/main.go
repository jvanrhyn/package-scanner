package main

import (
	"fmt"

	"github.com/squarehole/package-scanner/pkg/scanner"
)

func main() {
	// Create a scanner for NuGet packages
	packageScanner := scanner.NewPackageScanner("nupkg", "NuGet")

	// Test with the example filename
	testFilenames := []string{
		"System.Text.RegularExpressions.4.3.0.nupkg",
		"System.Threading.4.3.0.nupkg",
		"Microsoft.AspNetCore.Identity.3.1.10.nupkg",
		"Newtonsoft.Json.13.0.1.nupkg",
	}

	for _, filename := range testFilenames {
		// Extract package info
		pkgInfo, err := packageScanner.ExtractPackageInfo(filename)
		if err != nil {
			fmt.Printf("Error extracting package info from %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("Extracted package: name='%s', version='%s' from '%s'\n",
			pkgInfo.Name, pkgInfo.Version, filename)
	}
}
