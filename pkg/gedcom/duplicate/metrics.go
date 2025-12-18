package duplicate

import (
	"time"
)

// PerformanceMetrics holds performance statistics for duplicate detection.
type PerformanceMetrics struct {
	TotalComparisons    int
	FilteredComparisons int // Comparisons skipped by pre-filtering
	ProcessingTime      time.Duration
	IndexBuildTime      time.Duration
	ComparisonTime      time.Duration
	SortTime            time.Duration
	ParallelWorkers     int
	Throughput          float64 // Comparisons per second
}

// calculateMetrics calculates performance metrics from timing data.
func (dd *DuplicateDetector) calculateMetrics(
	startTime time.Time,
	indexBuildTime time.Duration,
	comparisonTime time.Duration,
	sortTime time.Duration,
	totalComparisons int,
	filteredComparisons int,
	numWorkers int) *PerformanceMetrics {

	totalTime := time.Since(startTime)
	throughput := 0.0
	if totalTime > 0 {
		throughput = float64(totalComparisons) / totalTime.Seconds()
	}

	return &PerformanceMetrics{
		TotalComparisons:    totalComparisons,
		FilteredComparisons: filteredComparisons,
		ProcessingTime:      totalTime,
		IndexBuildTime:      indexBuildTime,
		ComparisonTime:      comparisonTime,
		SortTime:            sortTime,
		ParallelWorkers:     numWorkers,
		Throughput:          throughput,
	}
}

// DuplicateResultWithMetrics extends DuplicateResult with performance metrics.
type DuplicateResultWithMetrics struct {
	*DuplicateResult
	Metrics *PerformanceMetrics
}
