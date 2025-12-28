package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// ParserResult holds the results from parsing a file
type ParserResult struct {
	ParserName   string
	FileName     string
	Individuals  int
	Families     int
	Errors       int
	Warnings     int
	SevereErrors int
	ParseTime    time.Duration
	Success      bool
	ErrorMsg     string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run parser_comparison.go <testdata_dir>")
		fmt.Println("Example: go run parser_comparison.go ../testdata")
		os.Exit(1)
	}

	testDataDir := os.Args[1]

	// Find all .ged files
	gedFiles, err := findGedFiles(testDataDir)
	if err != nil {
		fmt.Printf("Error finding GED files: %v\n", err)
		os.Exit(1)
	}

	if len(gedFiles) == 0 {
		fmt.Printf("No .ged files found in %s\n", testDataDir)
		os.Exit(1)
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("PARSER COMPARISON RESULTS")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Test gedcom-go parser
	for _, file := range gedFiles {
		fmt.Printf("Testing: %s\n", filepath.Base(file))
		fmt.Println(string(make([]byte, 80)))
		fmt.Println()

		// Test gedcom-go parser
		result1 := testGedcomGoParser(file)
		printResult(result1)
		fmt.Println()
		fmt.Println()
	}
}

func findGedFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".ged" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func testGedcomGoParser(filePath string) ParserResult {
	result := ParserResult{
		ParserName: "gedcom-go (NewParser)",
		FileName:   filepath.Base(filePath),
	}

	start := time.Now()

	// Use our parser
	p := parser.NewParser()
	tree, err := p.Parse(filePath)

	result.ParseTime = time.Since(start)

	if err != nil {
		result.Success = false
		result.ErrorMsg = err.Error()
		return result
	}

	result.Success = true

	// Get individuals and families
	individuals := tree.GetAllIndividuals()
	families := tree.GetAllFamilies()

	result.Individuals = len(individuals)
	result.Families = len(families)

	// Get errors
	errors := p.GetErrors()
	result.Errors = len(errors)

	// Count warnings vs severe errors
	for _, err := range errors {
		if err.Severity == types.SeveritySevere {
			result.SevereErrors++
		} else {
			result.Warnings++
		}
	}

	return result
}


func printResult(result ParserResult) {
	fmt.Printf("Parser: %s\n", result.ParserName)
	if !result.Success {
		fmt.Printf("  ❌ FAILED: %s\n", result.ErrorMsg)
		return
	}

	fmt.Printf("  ✅ Success\n")
	fmt.Printf("  Individuals: %d\n", result.Individuals)
	fmt.Printf("  Families:    %d\n", result.Families)
	fmt.Printf("  Total Issues: %d\n", result.Errors)
	if result.SevereErrors > 0 || result.Warnings > 0 {
		fmt.Printf("    - Severe:   %d\n", result.SevereErrors)
		fmt.Printf("    - Warnings: %d\n", result.Warnings)
	}
	fmt.Printf("  Parse Time:   %v\n", result.ParseTime)
}

