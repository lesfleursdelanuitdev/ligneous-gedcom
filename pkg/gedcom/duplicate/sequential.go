package duplicate

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// findDuplicatesSequential finds duplicates using sequential processing.
func (dd *DuplicateDetector) findDuplicatesSequential(individuals []*gedcom.IndividualRecord) ([]DuplicateMatch, int, error) {
	if len(individuals) < 2 {
		return []DuplicateMatch{}, 0, nil
	}

	// Build indexes for pre-filtering
	indexes := dd.buildIndexes(individuals)

	// Find duplicates
	matches := make([]DuplicateMatch, 0)
	comparisonCount := 0

	for i := 0; i < len(individuals); i++ {
		for j := i + 1; j < len(individuals); j++ {
			indi1 := individuals[i]
			indi2 := individuals[j]

			// Pre-filter: skip if not in same index buckets
			if !dd.shouldCompare(indi1, indi2, indexes) {
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
	}

	return matches, comparisonCount, nil
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
