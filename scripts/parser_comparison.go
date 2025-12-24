package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"

	// For cacack parser
	cacackDecoder "github.com/cacack/gedcom-go/decoder"

	// For elliotchance parser
	elliotchance "github.com/elliotchance/gedcom/v39"
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

	// Test all three parsers
	for _, file := range gedFiles {
		fmt.Printf("Testing: %s\n", filepath.Base(file))
		fmt.Println(string(make([]byte, 80)))
		fmt.Println()

		// Test gedcom-go parser
		result1 := testGedcomGoParser(file)
		printResult(result1)
		fmt.Println()

		// Test cacack parser
		result2 := testCacackParser(file)
		printResult(result2)
		fmt.Println()

		// Test elliotchance parser
		result3 := testElliotchanceParser(file)
		printResult(result3)
		fmt.Println()

		// Comparison summary
		printComparison(result1, result2, result3)
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

func testCacackParser(filePath string) ParserResult {
	result := ParserResult{
		ParserName: "gedcom-go-cacack",
		FileName:   filepath.Base(filePath),
	}

	start := time.Now()

	file, err := os.Open(filePath)
	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("Failed to open file: %v", err)
		return result
	}
	defer file.Close()

	doc, err := cacackDecoder.Decode(file)
	result.ParseTime = time.Since(start)

	if err != nil {
		result.Success = false
		result.ErrorMsg = err.Error()
		return result
	}

	result.Success = true
	result.Individuals = len(doc.Individuals())
	result.Families = len(doc.Families())
	// cacack doesn't provide error counting in the same way
	result.Errors = 0
	result.Warnings = 0
	result.SevereErrors = 0

	return result
}

func testElliotchanceParser(filePath string) ParserResult {
	result := ParserResult{
		ParserName: "elliotchance",
		FileName:   filepath.Base(filePath),
	}

	start := time.Now()

	doc, err := elliotchance.NewDocumentFromGEDCOMFile(filePath)
	result.ParseTime = time.Since(start)

	if err != nil {
		result.Success = false
		result.ErrorMsg = err.Error()
		return result
	}

	result.Success = true
	result.Individuals = len(doc.Individuals())
	result.Families = len(doc.Families())

	// Get warnings
	warnings := doc.Warnings()
	result.Warnings = len(warnings)
	result.Errors = result.Warnings
	result.SevereErrors = 0 // elliotchance uses warnings, not separate errors

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

func printComparison(r1, r2, r3 ParserResult) {
	fmt.Println("COMPARISON:")
	fmt.Println("-" + string(make([]byte, 78)))

	if !r1.Success || !r2.Success || !r3.Success {
		fmt.Println("⚠️  Some parsers failed - cannot compare")
		return
	}

	// Compare individuals
	fmt.Printf("Individuals: ")
	if r1.Individuals == r2.Individuals && r2.Individuals == r3.Individuals {
		fmt.Printf("✅ All agree: %d\n", r1.Individuals)
	} else {
		fmt.Printf("⚠️  Discrepancy: gedcom-go=%d, cacack=%d, elliotchance=%d\n",
			r1.Individuals, r2.Individuals, r3.Individuals)
	}

	// Compare families
	fmt.Printf("Families:    ")
	if r1.Families == r2.Families && r2.Families == r3.Families {
		fmt.Printf("✅ All agree: %d\n", r1.Families)
	} else {
		fmt.Printf("⚠️  Discrepancy: gedcom-go=%d, cacack=%d, elliotchance=%d\n",
			r1.Families, r2.Families, r3.Families)
	}

	// Compare parse times
	fmt.Printf("Parse Time:  ")
	times := []time.Duration{r1.ParseTime, r2.ParseTime, r3.ParseTime}
	fastest := times[0]
	fastestName := r1.ParserName
	for i, t := range times {
		if t < fastest {
			fastest = t
			if i == 1 {
				fastestName = r2.ParserName
			} else if i == 2 {
				fastestName = r3.ParserName
			}
		}
	}
	fmt.Printf("Fastest: %s (%v)\n", fastestName, fastest)
	fmt.Printf("  - gedcom-go:    %v\n", r1.ParseTime)
	fmt.Printf("  - cacack:       %v\n", r2.ParseTime)
	fmt.Printf("  - elliotchance: %v\n", r3.ParseTime)
}
