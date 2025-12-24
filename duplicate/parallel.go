package duplicate

import (
	"runtime"
	"sync"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// comparisonJob represents a single comparison task.
type comparisonJob struct {
	indi1 *types.IndividualRecord
	indi2 *types.IndividualRecord
	index int // For ordering results
}

// comparisonResult represents the result of a comparison.
type comparisonResult struct {
	match *DuplicateMatch
	index int
	err   error
}

// findDuplicatesParallel finds duplicates using parallel processing.
func (dd *DuplicateDetector) findDuplicatesParallel(individuals []*types.IndividualRecord) ([]DuplicateMatch, int, *BlockingMetrics, error) {
	if len(individuals) < 2 {
		return []DuplicateMatch{}, 0, nil, nil
	}

	// Determine number of workers
	numWorkers := dd.getNumWorkers()
	if numWorkers > len(individuals) {
		numWorkers = len(individuals)
	}

	// Use blocking-based candidate generation (much faster than O(n²))
	// This replaces the old index-based approach with proper blocking
	jobs, blockingMetrics := dd.generateBlockedComparisonJobs(individuals)
	if len(jobs) == 0 {
		return []DuplicateMatch{}, 0, blockingMetrics, nil
	}

	// Limit comparisons if configured
	if dd.config.MaxComparisons > 0 && len(jobs) > dd.config.MaxComparisons {
		jobs = jobs[:dd.config.MaxComparisons]
	}

	// Create channels
	jobChan := make(chan comparisonJob, len(jobs))
	resultChan := make(chan comparisonResult, len(jobs))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go dd.worker(jobChan, resultChan, &wg)
	}

	// Send jobs
	go func() {
		for _, job := range jobs {
			jobChan <- job
		}
		close(jobChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	matches := make([]DuplicateMatch, 0)
	comparisonCount := 0
	results := make([]comparisonResult, 0, len(jobs))

	for result := range resultChan {
		comparisonCount++
		if result.err == nil && result.match != nil {
			if result.match.SimilarityScore >= dd.config.MinThreshold {
				results = append(results, result)
			}
		}
	}

	// Convert results to matches (maintain order if needed)
	for _, result := range results {
		matches = append(matches, *result.match)
	}

	return matches, comparisonCount, blockingMetrics, nil
}

// worker processes comparison jobs.
func (dd *DuplicateDetector) worker(jobChan <-chan comparisonJob, resultChan chan<- comparisonResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobChan {
		match, err := dd.compare(job.indi1, job.indi2)
		resultChan <- comparisonResult{
			match: match,
			index: job.index,
			err:   err,
		}
	}
}

// generateComparisonJobs generates all comparison jobs with pre-filtering.
func (dd *DuplicateDetector) generateComparisonJobs(individuals []*types.IndividualRecord, idx *indexes) []comparisonJob {
	jobs := make([]comparisonJob, 0)
	jobIndex := 0

	for i := 0; i < len(individuals); i++ {
		for j := i + 1; j < len(individuals); j++ {
			indi1 := individuals[i]
			indi2 := individuals[j]

			// Pre-filter: skip if not in same index buckets
			if !dd.shouldCompare(indi1, indi2, idx) {
				continue
			}

			jobs = append(jobs, comparisonJob{
				indi1: indi1,
				indi2: indi2,
				index: jobIndex,
			})
			jobIndex++
		}
	}

	return jobs
}

// getNumWorkers determines the optimal number of worker goroutines.
func (dd *DuplicateDetector) getNumWorkers() int {
	// Use configured number if set
	if dd.config.NumWorkers > 0 {
		return dd.config.NumWorkers
	}

	// Default to number of CPU cores
	numCPU := runtime.NumCPU()

	// For small datasets, use fewer workers
	// For large datasets, use more workers (up to 2x CPU cores)
	if numCPU < 4 {
		return numCPU
	}

	// Use 1.5x CPU cores for better throughput
	return int(float64(numCPU) * 1.5)
}

// findDuplicatesBetweenParallel finds duplicates between two trees using parallel processing.
func (dd *DuplicateDetector) findDuplicatesBetweenParallel(
	individuals1, individuals2 []*types.IndividualRecord) ([]DuplicateMatch, int, error) {
	if len(individuals1) == 0 || len(individuals2) == 0 {
		return []DuplicateMatch{}, 0, nil
	}

	// Determine number of workers
	numWorkers := dd.getNumWorkers()

	// Build indexes for pre-filtering
	indexes1 := dd.buildIndexes(individuals1)
	indexes2 := dd.buildIndexes(individuals2)

	// Generate comparison jobs
	jobs := make([]comparisonJob, 0)
	jobIndex := 0

	for _, indi1 := range individuals1 {
		for _, indi2 := range individuals2 {
			// Pre-filter: skip if not in same index buckets
			if !dd.shouldCompareCrossFile(indi1, indi2, indexes1, indexes2) {
				continue
			}

			jobs = append(jobs, comparisonJob{
				indi1: indi1,
				indi2: indi2,
				index: jobIndex,
			})
			jobIndex++

			// Limit comparisons if configured
			if dd.config.MaxComparisons > 0 && len(jobs) >= dd.config.MaxComparisons {
				break
			}
		}
		if dd.config.MaxComparisons > 0 && len(jobs) >= dd.config.MaxComparisons {
			break
		}
	}

	if len(jobs) == 0 {
		return []DuplicateMatch{}, 0, nil
	}

	// Create channels
	jobChan := make(chan comparisonJob, len(jobs))
	resultChan := make(chan comparisonResult, len(jobs))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go dd.worker(jobChan, resultChan, &wg)
	}

	// Send jobs
	go func() {
		for _, job := range jobs {
			jobChan <- job
		}
		close(jobChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	matches := make([]DuplicateMatch, 0)
	comparisonCount := 0
	results := make([]comparisonResult, 0, len(jobs))

	for result := range resultChan {
		comparisonCount++
		if result.err == nil && result.match != nil {
			if result.match.SimilarityScore >= dd.config.MinThreshold {
				results = append(results, result)
			}
		}
	}

	// Convert results to matches
	for _, result := range results {
		matches = append(matches, *result.match)
	}

	return matches, comparisonCount, nil
}
