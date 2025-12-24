package duplicate

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom"
)

// findDuplicatesSequential finds duplicates using sequential processing.
func (dd *DuplicateDetector) findDuplicatesSequential(individuals []*gedcom.IndividualRecord) ([]DuplicateMatch, int, *BlockingMetrics, error) {
	if len(individuals) < 2 {
		return []DuplicateMatch{}, 0, nil, nil
	}

	// Use blocking-based candidate generation (much faster than O(n²))
	blockIndex := dd.buildBlockIndex(individuals)
	maxCandidatesPerPerson := dd.config.MaxCandidatesPerPerson
	if maxCandidatesPerPerson <= 0 {
		maxCandidatesPerPerson = 200 // Default
	}

	// Find duplicates using blocking
	matches := make([]DuplicateMatch, 0)
	comparisonCount := 0
	candidatesPerPerson := make([]int, len(individuals))

	for i := uint32(0); i < uint32(len(individuals)); i++ {
		candidates := blockIndex.findCandidates(i, maxCandidatesPerPerson)
		candidatesPerPerson[i] = len(candidates)
		for candidateID := range candidates {
			comparisonCount++
			if dd.config.MaxComparisons > 0 && comparisonCount > dd.config.MaxComparisons {
				break
			}

			indi1 := individuals[i]
			indi2 := individuals[candidateID]

			// Calculate similarity
			match, err := dd.compare(indi1, indi2)
			if err != nil {
				continue
			}

			// Filter by threshold
			if match.SimilarityScore >= dd.config.MinThreshold {
				matches = append(matches, *match)
			}
		}
		if dd.config.MaxComparisons > 0 && comparisonCount >= dd.config.MaxComparisons {
			break
		}
	}

	// Compute blocking metrics
	blockingMetrics := blockIndex.computeBlockingMetrics(len(individuals), candidatesPerPerson)

	return matches, comparisonCount, blockingMetrics, nil
}

// findDuplicatesBetweenSequential finds duplicates between two trees using sequential processing.
func (dd *DuplicateDetector) findDuplicatesBetweenSequential(
	individuals1, individuals2 []*gedcom.IndividualRecord) ([]DuplicateMatch, int, error) {
	if len(individuals1) == 0 || len(individuals2) == 0 {
		return []DuplicateMatch{}, 0, nil
	}

	// Build indexes for pre-filtering
	indexes1 := dd.buildIndexes(individuals1)
	indexes2 := dd.buildIndexes(individuals2)

	// Find duplicates
	matches := make([]DuplicateMatch, 0)
	comparisonCount := 0

	for _, indi1 := range individuals1 {
		for _, indi2 := range individuals2 {
			// Pre-filter: skip if not in same index buckets
			if !dd.shouldCompareCrossFile(indi1, indi2, indexes1, indexes2) {
				continue
			}

			comparisonCount++
			if dd.config.MaxComparisons > 0 && comparisonCount > dd.config.MaxComparisons {
				break
			}

			// Calculate similarity
			match, err := dd.compare(indi1, indi2)
			if err != nil {
				continue
			}

			// Filter by threshold
			if match.SimilarityScore >= dd.config.MinThreshold {
				matches = append(matches, *match)
			}
		}
		if dd.config.MaxComparisons > 0 && comparisonCount >= dd.config.MaxComparisons {
			break
		}
	}

	return matches, comparisonCount, nil
}
