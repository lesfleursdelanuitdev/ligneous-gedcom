package duplicate

import (
	"fmt"
	"sort"
)

// BlockingMetrics holds metrics about the blocking process.
type BlockingMetrics struct {
	TotalPeople           int
	PeopleWithPrimaryBlock int
	PeopleWithAnyBlock     int
	PeopleWithNoBlocks     int

	TotalBlocks           int
	PrimaryBlocks         int
	SurnameYearBlocks     int
	SurnameInitialBlocks  int
	SurnamePrefixBlocks   int

	TotalCandidatesGenerated int64
	TotalCandidatesScored    int64
	AverageCandidatesPerPerson float64
	MaxCandidatesPerPerson     int

	BlockSizeDistribution map[int]int // block size -> count
	TopBlockSizes         []BlockSizeInfo

	BlockTypeUsage map[string]int // block type -> usage count

	PeopleWithZeroCandidates int
	PeopleWithOneCandidate    int
	PeopleWithManyCandidates int // > 10

	// Warnings
	HasGiantBlocks        bool   // True if any blocks exceed maxBlockSize
	LargestBlockSize     int    // Size of largest block
	PeopleInGiantBlocks  int    // Number of people in blocks that were skipped
	MostCommonSurname    string // Most common surname (if available)
	RepetitionWarning     string // Human-readable warning message
}

// BlockSizeInfo represents information about a block size.
type BlockSizeInfo struct {
	Size      int
	Count     int
	BlockType string
}

// String returns a string representation of blocking metrics.
func (bm *BlockingMetrics) String() string {
	s := fmt.Sprintf("Blocking Metrics:\n")
	s += fmt.Sprintf("  Total People: %d\n", bm.TotalPeople)
	s += fmt.Sprintf("  People with Primary Block: %d (%.1f%%)\n",
		bm.PeopleWithPrimaryBlock,
		float64(bm.PeopleWithPrimaryBlock)/float64(bm.TotalPeople)*100)
	s += fmt.Sprintf("  People with Any Block: %d (%.1f%%)\n",
		bm.PeopleWithAnyBlock,
		float64(bm.PeopleWithAnyBlock)/float64(bm.TotalPeople)*100)
	s += fmt.Sprintf("  People with No Blocks: %d (%.1f%%)\n",
		bm.PeopleWithNoBlocks,
		float64(bm.PeopleWithNoBlocks)/float64(bm.TotalPeople)*100)
	s += fmt.Sprintf("\n  Total Blocks: %d\n", bm.TotalBlocks)
	s += fmt.Sprintf("  Primary Blocks: %d\n", bm.PrimaryBlocks)
	s += fmt.Sprintf("  Surname+Year Blocks: %d\n", bm.SurnameYearBlocks)
	s += fmt.Sprintf("  Surname+Initial Blocks: %d\n", bm.SurnameInitialBlocks)
	s += fmt.Sprintf("  Surname+Prefix Blocks: %d\n", bm.SurnamePrefixBlocks)
	s += fmt.Sprintf("\n  Total Candidates Generated: %d\n", bm.TotalCandidatesGenerated)
	s += fmt.Sprintf("  Total Candidates Scored: %d\n", bm.TotalCandidatesScored)
	s += fmt.Sprintf("  Avg Candidates/Person: %.2f\n", bm.AverageCandidatesPerPerson)
	s += fmt.Sprintf("  Max Candidates/Person: %d\n", bm.MaxCandidatesPerPerson)
	s += fmt.Sprintf("\n  People with 0 candidates: %d\n", bm.PeopleWithZeroCandidates)
	s += fmt.Sprintf("  People with 1 candidate: %d\n", bm.PeopleWithOneCandidate)
	s += fmt.Sprintf("  People with >10 candidates: %d\n", bm.PeopleWithManyCandidates)

	if len(bm.TopBlockSizes) > 0 {
		s += fmt.Sprintf("\n  Top Block Sizes:\n")
		for i, info := range bm.TopBlockSizes {
			if i >= 20 {
				break
			}
			s += fmt.Sprintf("    %s: size %d (count: %d)\n",
				info.BlockType, info.Size, info.Count)
		}
	}

	// Add warnings if present
	if bm.RepetitionWarning != "" {
		s += fmt.Sprintf("\n  ⚠️  WARNING: %s\n", bm.RepetitionWarning)
	}

	return s
}

// GetWarnings returns human-readable warnings about blocking issues.
func (bm *BlockingMetrics) GetWarnings() []string {
	warnings := make([]string, 0)

	if bm.HasGiantBlocks {
		warning := fmt.Sprintf(
			"Duplicate detection could not evaluate %d records (%.1f%%) because the dataset has extremely common surnames/years (largest block: %d people). "+
				"Try adding a place filter, widening given-name prefix matching, or running per-region.",
			bm.PeopleInGiantBlocks,
			float64(bm.PeopleInGiantBlocks)/float64(bm.TotalPeople)*100,
			bm.LargestBlockSize)
		warnings = append(warnings, warning)
	}

	if bm.PeopleWithZeroCandidates > bm.TotalPeople/2 {
		warning := fmt.Sprintf(
			"Over half of records (%d, %.1f%%) produced no candidate matches. "+
				"This may indicate missing data (birth dates, surnames) or extremely repetitive names. "+
				"Consider using fallback blocking strategies or filtering by place/time period.",
			bm.PeopleWithZeroCandidates,
			float64(bm.PeopleWithZeroCandidates)/float64(bm.TotalPeople)*100)
		warnings = append(warnings, warning)
	}

	if bm.MostCommonSurname != "" && bm.PeopleWithPrimaryBlock > 0 {
		// Estimate if surname is too common (heuristic: if >30% of people share it)
		// This is approximate since we don't track exact surname distribution
		if bm.LargestBlockSize > bm.TotalPeople/3 {
			warning := fmt.Sprintf(
				"Very common surname detected (largest block: %d people). "+
					"For better results, try filtering by place, time period, or given name prefix.",
				bm.LargestBlockSize)
			warnings = append(warnings, warning)
		}
	}

	return warnings
}

// computeBlockingMetrics computes metrics from a block index and candidate generation.
func (index *BlockIndex) computeBlockingMetrics(
	totalPeople int,
	candidatesPerPerson []int,
) *BlockingMetrics {
	metrics := &BlockingMetrics{
		TotalPeople:                totalPeople,
		BlockSizeDistribution:      make(map[int]int),
		BlockTypeUsage:            make(map[string]int),
		TopBlockSizes:             make([]BlockSizeInfo, 0),
	}

	// Count people with blocks
	peopleWithPrimary := 0
	peopleWithAny := 0
	peopleWithNoBlocks := 0

	for _, block := range index.personBlocks {
		hasPrimary := block.SurnameSoundex != "" && block.BirthYear > 0
		hasAny := hasPrimary ||
			(block.SurnameSoundex != "" && block.GivenInitial != 0) ||
			(block.SurnamePrefix != "" && block.BirthPlaceToken != "")

		if hasPrimary {
			peopleWithPrimary++
		}
		if hasAny {
			peopleWithAny++
		} else {
			peopleWithNoBlocks++
		}
	}

	metrics.PeopleWithPrimaryBlock = peopleWithPrimary
	metrics.PeopleWithAnyBlock = peopleWithAny
	metrics.PeopleWithNoBlocks = peopleWithNoBlocks

	// Count blocks and sizes
	metrics.PrimaryBlocks = len(index.primaryBlocks)
	for _, ids := range index.primaryBlocks {
		size := len(ids)
		metrics.BlockSizeDistribution[size]++
		metrics.TopBlockSizes = append(metrics.TopBlockSizes, BlockSizeInfo{
			Size:      size,
			Count:     1,
			BlockType: "primary",
		})
		metrics.BlockTypeUsage["primary"] += size
	}

	metrics.SurnameYearBlocks = len(index.surnameYearBlocks)
	for _, ids := range index.surnameYearBlocks {
		size := len(ids)
		metrics.BlockSizeDistribution[size]++
		metrics.TopBlockSizes = append(metrics.TopBlockSizes, BlockSizeInfo{
			Size:      size,
			Count:     1,
			BlockType: "surname+year",
		})
		metrics.BlockTypeUsage["surname+year"] += size
	}

	metrics.SurnameInitialBlocks = len(index.surnameInitialBlocks)
	for _, ids := range index.surnameInitialBlocks {
		size := len(ids)
		metrics.BlockSizeDistribution[size]++
		metrics.TopBlockSizes = append(metrics.TopBlockSizes, BlockSizeInfo{
			Size:      size,
			Count:     1,
			BlockType: "surname+initial",
		})
		metrics.BlockTypeUsage["surname+initial"] += size
	}

	metrics.SurnamePrefixBlocks = len(index.surnamePrefixBlocks)
	for _, ids := range index.surnamePrefixBlocks {
		size := len(ids)
		metrics.BlockSizeDistribution[size]++
		metrics.TopBlockSizes = append(metrics.TopBlockSizes, BlockSizeInfo{
			Size:      size,
			Count:     1,
			BlockType: "surname+prefix",
		})
		metrics.BlockTypeUsage["surname+prefix"] += size
	}

	metrics.TotalBlocks = metrics.PrimaryBlocks + metrics.SurnameYearBlocks +
		metrics.SurnameInitialBlocks + metrics.SurnamePrefixBlocks

	// Sort top block sizes
	sort.Slice(metrics.TopBlockSizes, func(i, j int) bool {
		return metrics.TopBlockSizes[i].Size > metrics.TopBlockSizes[j].Size
	})

	// Aggregate block sizes
	aggregated := make(map[int]int)
	for _, info := range metrics.TopBlockSizes {
		aggregated[info.Size] += info.Count
	}

	// Rebuild top block sizes with aggregated counts
	metrics.TopBlockSizes = make([]BlockSizeInfo, 0, len(aggregated))
	for size, count := range aggregated {
		metrics.TopBlockSizes = append(metrics.TopBlockSizes, BlockSizeInfo{
			Size:  size,
			Count: count,
		})
	}
	sort.Slice(metrics.TopBlockSizes, func(i, j int) bool {
		return metrics.TopBlockSizes[i].Size > metrics.TopBlockSizes[j].Size
	})

	// Candidate statistics
	totalCandidates := int64(0)
	maxCandidates := 0
	zeroCandidates := 0
	oneCandidate := 0
	manyCandidates := 0

	for _, count := range candidatesPerPerson {
		totalCandidates += int64(count)
		if count > maxCandidates {
			maxCandidates = count
		}
		if count == 0 {
			zeroCandidates++
		} else if count == 1 {
			oneCandidate++
		} else if count > 10 {
			manyCandidates++
		}
	}

	metrics.TotalCandidatesGenerated = totalCandidates
	metrics.TotalCandidatesScored = totalCandidates // Same for now
	if totalPeople > 0 {
		metrics.AverageCandidatesPerPerson = float64(totalCandidates) / float64(totalPeople)
	}
	metrics.MaxCandidatesPerPerson = maxCandidates
	metrics.PeopleWithZeroCandidates = zeroCandidates
	metrics.PeopleWithOneCandidate = oneCandidate
	metrics.PeopleWithManyCandidates = manyCandidates

	// Detect giant blocks and generate warnings
	maxBlockSize := 5000 // Should match BlockIndex.maxBlockSize
	if len(metrics.TopBlockSizes) > 0 {
		metrics.LargestBlockSize = metrics.TopBlockSizes[0].Size
		if metrics.LargestBlockSize > maxBlockSize {
			metrics.HasGiantBlocks = true
			// Estimate people in giant blocks (sum of people in blocks > maxBlockSize)
			// Note: TopBlockSizes may have duplicate sizes, so we need to aggregate properly
			peopleInGiantBlocks := 0
			seenSizes := make(map[int]bool)
			for _, info := range metrics.TopBlockSizes {
				if info.Size > maxBlockSize {
					// Only count each unique block size once (avoid double counting)
					if !seenSizes[info.Size] {
						// Estimate: assume each block of this size has this many people
						// This is approximate since we don't track exact block membership
						peopleInGiantBlocks += info.Size
						seenSizes[info.Size] = true
					}
				} else {
					break
				}
			}
			// Cap at total people (can't exceed dataset size)
			if peopleInGiantBlocks > metrics.TotalPeople {
				peopleInGiantBlocks = metrics.TotalPeople
			}
			metrics.PeopleInGiantBlocks = peopleInGiantBlocks

			// Generate human-readable warning
			warnings := metrics.GetWarnings()
			if len(warnings) > 0 {
				metrics.RepetitionWarning = warnings[0]
			}
		}
	}

	return metrics
}

