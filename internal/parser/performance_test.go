package parser

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom/query"
)

// generateLargeGEDCOMFile generates a GEDCOM file with n individuals
func generateLargeGEDCOMFile(filename string, n int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	file.WriteString("0 HEAD\n")
	file.WriteString("1 SOUR Test Generator\n")
	file.WriteString("1 GEDC\n")
	file.WriteString("2 VERS 5.5.1\n")
	file.WriteString("0 @SUBM@ SUBM\n")
	file.WriteString("1 NAME Test\n")

	// Track relationships
	childToFamily := make(map[int]int)
	familyID := 1
	indiID := 1

	// Create families
	for indiID < n {
		numChildren := 2
		if indiID%10 == 0 {
			numChildren = 1
		} else if indiID%5 == 0 {
			numChildren = 3
		}

		if indiID+numChildren+1 >= n {
			numChildren = n - indiID - 1
			if numChildren <= 0 {
				break
			}
		}

		// Family record
		file.WriteString(fmt.Sprintf("0 @F%d@ FAM\n", familyID))

		// Husband
		if indiID < n {
			file.WriteString(fmt.Sprintf("1 HUSB @I%d@\n", indiID))
			indiID++
		}

		// Wife
		if indiID < n {
			file.WriteString(fmt.Sprintf("1 WIFE @I%d@\n", indiID))
			indiID++
		}

		// Children
		for i := 0; i < numChildren && indiID < n; i++ {
			file.WriteString(fmt.Sprintf("1 CHIL @I%d@\n", indiID))
			childToFamily[indiID] = familyID
			indiID++
		}

		familyID++
	}

	// Create individuals
	for i := 1; i <= n; i++ {
		file.WriteString(fmt.Sprintf("0 @I%d@ INDI\n", i))
		file.WriteString(fmt.Sprintf("1 NAME Person %d /Test/\n", i))

		birthYear := 1800 + (i % 200)
		file.WriteString("1 BIRT\n")
		file.WriteString(fmt.Sprintf("2 DATE %d\n", birthYear))

		sex := "M"
		if i%2 == 0 {
			sex = "F"
		}
		file.WriteString(fmt.Sprintf("1 SEX %s\n", sex))

		if famID, ok := childToFamily[i]; ok {
			file.WriteString(fmt.Sprintf("1 FAMC @F%d@\n", famID))
		}
	}

	file.WriteString("0 TRLR\n")
	return nil
}

// measureMemory returns current memory usage
func measureMemory() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// TestPerformance_Parse_100K tests parsing performance with 100K individuals
func TestPerformance_Parse_100K(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large performance test in short mode")
	}

	const size = 100000
	filename := fmt.Sprintf("/tmp/gedcom_100k_%d.ged", time.Now().Unix())

	t.Logf("Generating GEDCOM file with %d individuals...", size)
	if err := generateLargeGEDCOMFile(filename, size); err != nil {
		t.Fatalf("Failed to generate test file: %v", err)
	}
	defer os.Remove(filename)

	before := measureMemory()
	start := time.Now()

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(filename)

	duration := time.Since(start)
	after := measureMemory()

	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	individuals := tree.GetAllIndividuals()
	if len(individuals) != size {
		t.Errorf("Expected %d individuals, got %d", size, len(individuals))
	}

	t.Logf("\n=== Parse Performance (100K) ===")
	t.Logf("Duration: %v", duration)
	t.Logf("Throughput: %.2f individuals/sec", float64(size)/duration.Seconds())
	t.Logf("Memory Used: %.2f MB", float64(after-before)/1024/1024)
}

// TestPerformance_Parse_500K tests parsing performance with 500K individuals
func TestPerformance_Parse_500K(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large performance test in short mode")
	}

	const size = 500000
	filename := fmt.Sprintf("/tmp/gedcom_500k_%d.ged", time.Now().Unix())

	t.Logf("Generating GEDCOM file with %d individuals...", size)
	if err := generateLargeGEDCOMFile(filename, size); err != nil {
		t.Fatalf("Failed to generate test file: %v", err)
	}
	defer os.Remove(filename)

	before := measureMemory()
	start := time.Now()

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(filename)

	duration := time.Since(start)
	after := measureMemory()

	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	individuals := tree.GetAllIndividuals()
	if len(individuals) != size {
		t.Errorf("Expected %d individuals, got %d", size, len(individuals))
	}

	t.Logf("\n=== Parse Performance (500K) ===")
	t.Logf("Duration: %v", duration)
	t.Logf("Throughput: %.2f individuals/sec", float64(size)/duration.Seconds())
	t.Logf("Memory Used: %.2f MB", float64(after-before)/1024/1024)
}

// BenchmarkParse_100K benchmarks parsing with 100K individuals
func BenchmarkParse_100K(b *testing.B) {
	filename := "/tmp/gedcom_bench_100k.ged"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := generateLargeGEDCOMFile(filename, 100000); err != nil {
			b.Fatalf("Failed to generate test file: %v", err)
		}
	}

	parser := NewHierarchicalParser()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(filename)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParse_500K benchmarks parsing with 500K individuals
func BenchmarkParse_500K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	filename := "/tmp/gedcom_bench_500k.ged"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := generateLargeGEDCOMFile(filename, 500000); err != nil {
			b.Fatalf("Failed to generate test file: %v", err)
		}
	}

	parser := NewHierarchicalParser()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(filename)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestPerformance_ParseAndQuery_100K tests full pipeline: parse + graph + query
func TestPerformance_ParseAndQuery_100K(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large performance test in short mode")
	}

	const size = 100000
	filename := fmt.Sprintf("/tmp/gedcom_full_100k_%d.ged", time.Now().Unix())

	t.Logf("Generating GEDCOM file with %d individuals...", size)
	if err := generateLargeGEDCOMFile(filename, size); err != nil {
		t.Fatalf("Failed to generate test file: %v", err)
	}
	defer os.Remove(filename)

	// Parse
	t.Log("Parsing...")
	parseStart := time.Now()
	parser := NewHierarchicalParser()
	tree, err := parser.Parse(filename)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	parseDuration := time.Since(parseStart)

	// Build graph
	t.Log("Building graph...")
	graphStart := time.Now()
	graph, err := query.BuildGraph(tree)
	if err != nil {
		t.Fatalf("Graph construction failed: %v", err)
	}
	graphDuration := time.Since(graphStart)

	// Create query
	t.Log("Creating query...")
	queryStart := time.Now()
	q := query.NewQueryFromGraph(graph)
	queryDuration := time.Since(queryStart)

	// Run filter query
	t.Log("Running filter query...")
	filterStart := time.Now()
	results, err := q.Filter().ByName("Person").Execute()
	if err != nil {
		t.Fatalf("Filter query failed: %v", err)
	}
	filterDuration := time.Since(filterStart)

	t.Logf("\n=== Full Pipeline Performance (100K) ===")
	t.Logf("Parse: %v (%.2f ind/sec)", parseDuration, float64(size)/parseDuration.Seconds())
	t.Logf("Graph Build: %v", graphDuration)
	t.Logf("Query Creation: %v", queryDuration)
	t.Logf("Filter Query: %v (%d results)", filterDuration, len(results))
	t.Logf("Total: %v", parseDuration+graphDuration+queryDuration+filterDuration)
}
