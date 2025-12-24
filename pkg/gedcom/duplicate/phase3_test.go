package duplicate

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom"
)

func TestParallelProcessing(t *testing.T) {
	config := DefaultConfig()
	config.UseParallelProcessing = true
	detector := NewDuplicateDetector(config)

	tree := gedcom.NewGedcomTree()

	// Create multiple individuals to test parallel processing
	for i := 0; i < 20; i++ {
		indi := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
		indi.FirstLine().XrefID = gedcom.NewGedcomLine(0, "INDI", "", "").XrefID
		if indi.FirstLine().XrefID == "" {
			indi.FirstLine().XrefID = fmt.Sprintf("@I%d@", i+1)
		}
		tree.AddRecord(indi)
	}

	result, err := detector.FindDuplicates(tree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find many duplicates
	if len(result.Matches) == 0 {
		t.Error("expected to find duplicate matches")
	}

	// Check that metrics are populated
	if result.Metrics == nil {
		t.Error("expected performance metrics to be populated")
	} else {
		if result.Metrics.TotalComparisons == 0 {
			t.Error("expected total comparisons to be > 0")
		}
		if result.Metrics.ParallelWorkers == 0 {
			t.Error("expected parallel workers to be > 0")
		}
		if result.Metrics.Throughput <= 0 {
			t.Error("expected throughput to be > 0")
		}
	}
}

func TestSequentialVsParallel(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	// Create test individuals
	for i := 0; i < 15; i++ {
		indi := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
		indi.FirstLine().XrefID = fmt.Sprintf("@I%d@", i+1)
		tree.AddRecord(indi)
	}

	// Test sequential
	configSeq := DefaultConfig()
	configSeq.UseParallelProcessing = false
	detectorSeq := NewDuplicateDetector(configSeq)

	resultSeq, err := detectorSeq.FindDuplicates(tree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test parallel
	configPar := DefaultConfig()
	configPar.UseParallelProcessing = true
	detectorPar := NewDuplicateDetector(configPar)

	resultPar, err := detectorPar.FindDuplicates(tree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should find the same matches (order may differ)
	if len(resultSeq.Matches) != len(resultPar.Matches) {
		t.Errorf("expected same number of matches: sequential=%d, parallel=%d",
			len(resultSeq.Matches), len(resultPar.Matches))
	}

	// Parallel should be faster (or at least not slower)
	if resultPar.Metrics != nil && resultSeq.Metrics != nil {
		t.Logf("Sequential time: %v, Parallel time: %v",
			resultSeq.Metrics.ProcessingTime, resultPar.Metrics.ProcessingTime)
		t.Logf("Parallel workers: %d, Throughput: %.2f comparisons/sec",
			resultPar.Metrics.ParallelWorkers, resultPar.Metrics.Throughput)
	}
}

func TestNumWorkersConfiguration(t *testing.T) {
	config := DefaultConfig()
	config.UseParallelProcessing = true
	config.NumWorkers = 4
	detector := NewDuplicateDetector(config)

	// Check that getNumWorkers returns configured value
	numWorkers := detector.getNumWorkers()
	if numWorkers != 4 {
		t.Errorf("expected 4 workers, got %d", numWorkers)
	}
}

func TestPerformanceMetrics(t *testing.T) {
	config := DefaultConfig()
	config.UseParallelProcessing = true
	detector := NewDuplicateDetector(config)

	tree := gedcom.NewGedcomTree()

	// Create test data
	for i := 0; i < 10; i++ {
		indi := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
		indi.FirstLine().XrefID = fmt.Sprintf("@I%d@", i+1)
		tree.AddRecord(indi)
	}

	result, err := detector.FindDuplicates(tree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metrics == nil {
		t.Fatal("expected metrics to be populated")
	}

	// Verify metrics fields
	if result.Metrics.ProcessingTime <= 0 {
		t.Error("expected processing time to be > 0")
	}
	if result.Metrics.TotalComparisons <= 0 {
		t.Error("expected total comparisons to be > 0")
	}
	if result.Metrics.ParallelWorkers <= 0 {
		t.Error("expected parallel workers to be > 0")
	}
	if result.Metrics.Throughput <= 0 {
		t.Error("expected throughput to be > 0")
	}

	t.Logf("Metrics: %+v", result.Metrics)
}
