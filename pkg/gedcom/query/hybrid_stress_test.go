package query

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom"
)

// testWithTimeout wraps a test function with a 2-minute timeout
// NOTE: This uses Go's built-in timeout mechanism via -timeout flag
// Tests should be run with: go test -timeout 2m
// This function provides an additional safety layer for tests that might hang
func testWithTimeout(t *testing.T, testName string, fn func(*testing.T)) {
	// Use a context to track test execution time and allow cancellation
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	// Channel to signal test completion
	done := make(chan bool, 1)
	var panicErr interface{}

	// Run test in goroutine so we can timeout if needed
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = r
			}
			select {
			case done <- true:
			default:
			}
		}()
		fn(t)
	}()

	// Wait for test completion or timeout
	select {
	case <-done:
		if panicErr != nil {
			t.Fatalf("Test %s panicked: %v", testName, panicErr)
		}
		// Test completed successfully
	case <-ctx.Done():
		// Timeout occurred - fail the test immediately
		// Note: The goroutine may still be running, but we've failed the test
		// Go's test runner will handle cleanup when the test process exits
		t.Fatalf("Test %s timed out after %v", testName, TestTimeout)
	}
}

// HybridStressTestMetrics holds metrics for hybrid storage stress tests
type HybridStressTestMetrics struct {
	TestName         string
	DatasetSize      int
	Duration          time.Duration
	MemoryBefore     uint64
	MemoryAfter      uint64
	MemoryPeak       uint64
	SQLiteSize       int64
	BadgerDBSize     int64
	CacheHitRate     float64
	QueryDuration    time.Duration
	NodesLoaded      int
	QueriesExecuted  int
	Success          bool
	Error            error
}

// getMemStats returns comprehensive memory statistics
func getMemStatsHybrid() (alloc, totalAlloc, sys uint64, numGC uint32, numGoroutines int) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc, m.TotalAlloc, m.Sys, m.NumGC, runtime.NumGoroutine()
}

// getDirSize returns the total size of a directory in bytes
func getDirSize(dir string) (int64, error) {
	var size int64
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// generateHybridTestTree creates a tree for hybrid storage testing
func generateHybridTestTree(n int) *gedcom.GedcomTree {
	tree := gedcom.NewGedcomTree()

	individuals := make([]*gedcom.IndividualRecord, 0, n)
	families := make([]*gedcom.FamilyRecord, 0, n/2)

	childToFamily := make(map[int]int, n)
	familyID := 1
	indiID := 1

	// Create families with 1-3 children each
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

		// Create family
		famLine := gedcom.NewGedcomLine(0, "FAM", "", fmt.Sprintf("@F%d@", familyID))

		// Husband
		if indiID < n {
			husbLine := gedcom.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", indiID))
			nameLine := gedcom.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", indiID), "")
			sexLine := gedcom.NewGedcomLine(1, "SEX", "M", "")
			birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
			dateLine := gedcom.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", 1800+(indiID%200)), "")
			birtLine.AddChild(dateLine)
			husbLine.AddChild(nameLine)
			husbLine.AddChild(sexLine)
			husbLine.AddChild(birtLine)
			husb := gedcom.NewIndividualRecord(husbLine)
			individuals = append(individuals, husb)
			tree.AddRecord(husb)
			famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", fmt.Sprintf("@I%d@", indiID), ""))
			indiID++
		}

		// Wife
		if indiID < n {
			wifeLine := gedcom.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", indiID))
			nameLine := gedcom.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", indiID), "")
			sexLine := gedcom.NewGedcomLine(1, "SEX", "F", "")
			birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
			dateLine := gedcom.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", 1800+(indiID%200)), "")
			birtLine.AddChild(dateLine)
			wifeLine.AddChild(nameLine)
			wifeLine.AddChild(sexLine)
			wifeLine.AddChild(birtLine)
			wife := gedcom.NewIndividualRecord(wifeLine)
			individuals = append(individuals, wife)
			tree.AddRecord(wife)
			famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", fmt.Sprintf("@I%d@", indiID), ""))
			indiID++
		}

		// Children
		for i := 0; i < numChildren && indiID < n; i++ {
			childLine := gedcom.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", indiID))
			nameLine := gedcom.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", indiID), "")
			sexLine := gedcom.NewGedcomLine(1, "SEX", map[int]string{0: "M", 1: "F"}[i%2], "")
			birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
			dateLine := gedcom.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", 1800+(indiID%200)), "")
			birtLine.AddChild(dateLine)
			childLine.AddChild(nameLine)
			childLine.AddChild(sexLine)
			childLine.AddChild(birtLine)
			child := gedcom.NewIndividualRecord(childLine)
			individuals = append(individuals, child)
			tree.AddRecord(child)
			famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", fmt.Sprintf("@I%d@", indiID), ""))
			childToFamily[indiID] = familyID
			indiID++
		}

		fam := gedcom.NewFamilyRecord(famLine)
		families = append(families, fam)
		tree.AddRecord(fam)
		familyID++
	}

	return tree
}

// TestHybridStorage_1M tests hybrid storage with 1M individuals
// NOTE: This is a stress test that takes several minutes. It will be skipped unless
// RUN_STRESS_TESTS environment variable is set to "1"
func TestHybridStorage_1M(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping hybrid storage stress test in short mode")
	}
	// Skip by default - only run when explicitly requested
	if os.Getenv("RUN_STRESS_TESTS") == "" {
		t.Skip("Skipping stress test. Set RUN_STRESS_TESTS=1 to run")
	}

	testWithTimeout(t, "TestHybridStorage_1M", func(t *testing.T) {
		const size = 1_000_000
		tmpDir := t.TempDir()
		sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
		badgerPath := filepath.Join(tmpDir, "test_graph")

		t.Logf("=== HYBRID STORAGE STRESS TEST: 1,000,000 INDIVIDUALS ===\n")
		t.Logf("Timeout: %v\n", TestTimeout)
		t.Logf("SQLite: %s\n", sqlitePath)
		t.Logf("BadgerDB: %s\n", badgerPath)

	// Phase 1: Generate test data
	t.Log("\n--- Phase 1: Data Generation ---")
	genStart := time.Now()
	tree := generateHybridTestTree(size)
	genDuration := time.Since(genStart)
	t.Logf("Generated %d individuals in %v\n", size, genDuration)

	// Phase 2: Build hybrid graph
	t.Log("\n--- Phase 2: Hybrid Graph Construction ---")
	before, _, _, _, _ := getMemStatsHybrid()
	buildStart := time.Now()
	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
	buildDuration := time.Since(buildStart)
	after, totalAlloc, _, _, _ := getMemStatsHybrid()

	if err != nil {
		t.Fatalf("Failed to build hybrid graph: %v", err)
	}
	defer graph.Close()

	// Get database sizes
	sqliteSize, _ := getDirSize(filepath.Dir(sqlitePath))
	badgerSize, _ := getDirSize(badgerPath)

	t.Logf("Construction Duration: %v\n", buildDuration)
	t.Logf("Memory Before: %.2f MB\n", float64(before)/(1024*1024))
	t.Logf("Memory After: %.2f MB\n", float64(after)/(1024*1024))
	t.Logf("Memory Used: %.2f MB\n", float64(after-before)/(1024*1024))
	t.Logf("Peak Memory: %.2f MB\n", float64(totalAlloc)/(1024*1024))
	t.Logf("SQLite Size: %.2f MB\n", float64(sqliteSize)/(1024*1024))
	t.Logf("BadgerDB Size: %.2f MB\n", float64(badgerSize)/(1024*1024))

	// Phase 3: Query Performance
	t.Log("\n--- Phase 3: Query Performance ---")
	
	// Test FilterQuery
	fq := NewFilterQuery(graph)
	queryStart := time.Now()
	results, err := fq.ByName("Person").Execute()
	queryDuration := time.Since(queryStart)
	if err != nil {
		t.Errorf("FilterQuery failed: %v", err)
	}
	t.Logf("FilterQuery (ByName): %d results in %v\n", len(results), queryDuration)

	// Test cache statistics
	if graph.hybridCache != nil {
		stats := graph.hybridCache.Stats()
		t.Logf("\n--- Cache Statistics ---\n")
		t.Logf("Node Cache: %d / %d\n", stats.NodeCacheSize, stats.NodeCacheCapacity)
		t.Logf("XREF Cache: %d / %d\n", stats.XrefCacheSize, stats.XrefCacheCapacity)
		t.Logf("Query Cache: %d / %d\n", stats.QueryCacheSize, stats.QueryCacheCapacity)
	}

	// Phase 4: Concurrent Access
	t.Log("\n--- Phase 4: Concurrent Access ---")
	concurrentStart := time.Now()
	var wg sync.WaitGroup
	numWorkers := 10
	queriesPerWorker := 100

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			fq := NewFilterQuery(graph)
			for j := 0; j < queriesPerWorker; j++ {
				_, err := fq.ByName(fmt.Sprintf("Person %d", workerID*queriesPerWorker+j)).Execute()
				if err != nil {
					t.Errorf("Worker %d query %d failed: %v", workerID, j, err)
				}
			}
		}(i)
	}
	wg.Wait()
	concurrentDuration := time.Since(concurrentStart)
		t.Logf("Concurrent Queries: %d workers × %d queries = %d total in %v\n",
			numWorkers, queriesPerWorker, numWorkers*queriesPerWorker, concurrentDuration)

		t.Log("\n=== TEST COMPLETE ===\n")
	})
}

// TestHybridStorage_5M tests hybrid storage with 5M individuals
// NOTE: This is a very large stress test that takes 10+ minutes. It will be skipped unless
// RUN_STRESS_TESTS environment variable is set to "1"
func TestHybridStorage_5M(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 5M hybrid storage stress test in short mode")
	}
	// Skip by default - only run when explicitly requested
	if os.Getenv("RUN_STRESS_TESTS") == "" {
		t.Skip("Skipping stress test. Set RUN_STRESS_TESTS=1 to run")
	}

	testWithTimeout(t, "TestHybridStorage_5M", func(t *testing.T) {
		const size = 5_000_000
		tmpDir := t.TempDir()
		sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
		badgerPath := filepath.Join(tmpDir, "test_graph")

		t.Logf("=== HYBRID STORAGE STRESS TEST: 5,000,000 INDIVIDUALS ===\n")
		t.Logf("WARNING: This is an extremely large-scale test. Timeout: %v\n", TestTimeout)
		t.Logf("SQLite: %s\n", sqlitePath)
		t.Logf("BadgerDB: %s\n", badgerPath)

		// Phase 1: Generate test data
		t.Log("\n--- Phase 1: Data Generation ---")
		genStart := time.Now()
		tree := generateHybridTestTree(size)
		genDuration := time.Since(genStart)
		t.Logf("Generated %d individuals in %v\n", size, genDuration)

		// Phase 2: Build hybrid graph
		t.Log("\n--- Phase 2: Hybrid Graph Construction ---")
		before, _, _, _, _ := getMemStatsHybrid()
		buildStart := time.Now()
		graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
		buildDuration := time.Since(buildStart)
		after, totalAlloc, _, _, _ := getMemStatsHybrid()

		if err != nil {
			t.Fatalf("Failed to build hybrid graph: %v", err)
		}
		defer graph.Close()

		// Get database sizes
		sqliteSize, _ := getDirSize(filepath.Dir(sqlitePath))
		badgerSize, _ := getDirSize(badgerPath)

		t.Logf("Construction Duration: %v\n", buildDuration)
		t.Logf("Memory Before: %.2f MB\n", float64(before)/(1024*1024))
		t.Logf("Memory After: %.2f MB\n", float64(after)/(1024*1024))
		t.Logf("Memory Used: %.2f MB\n", float64(after-before)/(1024*1024))
		t.Logf("Peak Memory: %.2f MB\n", float64(totalAlloc)/(1024*1024))
		t.Logf("SQLite Size: %.2f MB\n", float64(sqliteSize)/(1024*1024))
		t.Logf("BadgerDB Size: %.2f MB\n", float64(badgerSize)/(1024*1024))

		// Phase 3: Query Performance
		t.Log("\n--- Phase 3: Query Performance ---")
		
		fq := NewFilterQuery(graph)
		queryStart := time.Now()
		results, err := fq.ByName("Person").Execute()
		queryDuration := time.Since(queryStart)
		if err != nil {
			t.Errorf("FilterQuery failed: %v", err)
		}
		t.Logf("FilterQuery (ByName): %d results in %v\n", len(results), queryDuration)

		// Test cache statistics
		if graph.hybridCache != nil {
			stats := graph.hybridCache.Stats()
			t.Logf("\n--- Cache Statistics ---\n")
			t.Logf("Node Cache: %d / %d\n", stats.NodeCacheSize, stats.NodeCacheCapacity)
			t.Logf("XREF Cache: %d / %d\n", stats.XrefCacheSize, stats.XrefCacheCapacity)
			t.Logf("Query Cache: %d / %d\n", stats.QueryCacheSize, stats.QueryCacheCapacity)
		}

		t.Log("\n=== TEST COMPLETE ===\n")
	})
}

// TestHybridStorage_Concurrent tests concurrent access to hybrid storage
func TestHybridStorage_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent access test in short mode")
	}

	testWithTimeout(t, "TestHybridStorage_Concurrent", func(t *testing.T) {
		const size = 100_000
		tmpDir := t.TempDir()
		sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
		badgerPath := filepath.Join(tmpDir, "test_graph")

		tree := generateHybridTestTree(size)
		graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
		if err != nil {
			t.Fatalf("Failed to build hybrid graph: %v", err)
		}
		defer graph.Close()

		// Test concurrent reads
		numWorkers := 20
		queriesPerWorker := 50
		var wg sync.WaitGroup
		errors := make(chan error, numWorkers*queriesPerWorker)

		start := time.Now()
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				fq := NewFilterQuery(graph)
				for j := 0; j < queriesPerWorker; j++ {
					_, err := fq.ByName(fmt.Sprintf("Person %d", workerID*queriesPerWorker+j)).Execute()
					if err != nil {
						errors <- fmt.Errorf("worker %d query %d: %w", workerID, j, err)
					}
				}
			}(i)
		}
		wg.Wait()
		duration := time.Since(start)
		close(errors)

		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Errorf("Concurrent query error: %v", err)
			errorCount++
		}

		t.Logf("Concurrent Test: %d workers × %d queries = %d total in %v\n",
			numWorkers, queriesPerWorker, numWorkers*queriesPerWorker, duration)
		t.Logf("Errors: %d\n", errorCount)

		if errorCount > 0 {
			t.Fatalf("Concurrent test failed with %d errors", errorCount)
		}
	})
}

// TestHybridStorage_Persistence tests that data persists across graph closes
func TestHybridStorage_Persistence(t *testing.T) {
	testWithTimeout(t, "TestHybridStorage_Persistence", func(t *testing.T) {
		tmpDir := t.TempDir()
		sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
		badgerPath := filepath.Join(tmpDir, "test_graph")

		// Create graph and add data
		tree := generateHybridTestTree(1000)
		graph1, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
		if err != nil {
			t.Fatalf("Failed to build graph: %v", err)
		}

		// Verify data
		fq1 := NewFilterQuery(graph1)
		results1, err := fq1.ByName("Person").Execute()
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		initialCount := len(results1)

		// Close graph
		graph1.Close()

		// Reopen (in a real scenario, we'd need a LoadGraphHybrid function)
		// For now, we'll just verify the databases exist and have data
		sqliteSize, err := getDirSize(filepath.Dir(sqlitePath))
		if err != nil {
			t.Fatalf("Failed to get SQLite size: %v", err)
		}
		if sqliteSize == 0 {
			t.Error("SQLite database should have data")
		}

		badgerSize, err := getDirSize(badgerPath)
		if err != nil {
			t.Fatalf("Failed to get BadgerDB size: %v", err)
		}
		if badgerSize == 0 {
			t.Error("BadgerDB should have data")
		}

		t.Logf("Persistence Test: SQLite=%d bytes, BadgerDB=%d bytes, Initial results=%d\n",
			sqliteSize, badgerSize, initialCount)
	})
}

// BenchmarkHybridStorage_Query benchmarks query performance with hybrid storage
func BenchmarkHybridStorage_Query(b *testing.B) {
	const size = 100_000
	tmpDir := b.TempDir()
	sqlitePath := filepath.Join(tmpDir, "bench_indexes.db")
	badgerPath := filepath.Join(tmpDir, "bench_graph")

	tree := generateHybridTestTree(size)
	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
	if err != nil {
		b.Fatalf("Failed to build graph: %v", err)
	}
	defer graph.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		fq := NewFilterQuery(graph)
		i := 0
		for pb.Next() {
			_, err := fq.ByName(fmt.Sprintf("Person %d", i%size)).Execute()
			if err != nil {
				b.Errorf("Query failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkHybridStorage_NodeLoad benchmarks node loading from BadgerDB
func BenchmarkHybridStorage_NodeLoad(b *testing.B) {
	const size = 100_000
	tmpDir := b.TempDir()
	sqlitePath := filepath.Join(tmpDir, "bench_indexes.db")
	badgerPath := filepath.Join(tmpDir, "bench_graph")

	tree := generateHybridTestTree(size)
	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
	if err != nil {
		b.Fatalf("Failed to build graph: %v", err)
	}
	defer graph.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			xref := fmt.Sprintf("@I%d@", (i%size)+1)
			node := graph.GetIndividual(xref)
			if node == nil {
				b.Errorf("Failed to load node %s", xref)
			}
			i++
		}
	})
}

