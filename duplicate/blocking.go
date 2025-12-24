package duplicate

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"sync"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// BlockKey represents a composite blocking key for candidate generation.
// Using uint64 for efficient hashing and map lookups.
type BlockKey uint64

// PersonBlock represents precomputed blocking data for a person.
type PersonBlock struct {
	ID            uint32 // Index in the individuals slice
	Individual    *types.IndividualRecord
	SurnameSoundex string
	GivenInitial   byte
	BirthYear      int
	BirthYearBucket int // Year / 5 (5-year buckets)
	BirthPlaceToken string // First significant token from place
	SurnamePrefix  string // First 4 chars of surname
	GivenPrefix    string // First 2 chars of given name
}

// BlockIndex is the inverted index for blocking.
type BlockIndex struct {
	// Primary block: surname_soundex + birthYear
	primaryBlocks map[BlockKey][]uint32

	// Fallback blocks
	surnameYearBlocks    map[BlockKey][]uint32 // surname_soundex + birthYear (expanded)
	surnameYearBucketBlocks map[BlockKey][]uint32 // surname_soundex + birthYearBucket (5-year buckets)
	surnameInitialBlocks map[BlockKey][]uint32 // surname_soundex + given_initial
	surnamePrefixBlocks  map[BlockKey][]uint32 // surname_prefix(4) + birth_place_token
	surnameGivenPrefixBlocks map[BlockKey][]uint32 // surname_soundex + given_prefix(2)
	rescueBlocks         map[BlockKey][]uint32 // given_prefix(3) + surname_prefix(3) + place_token

	// Person blocks (for lookup)
	personBlocks []PersonBlock

	// Metrics
	metrics *BlockingMetrics

	// Configuration
	maxBlockSize int // Skip blocks larger than this (adaptive blocking)

	mu sync.RWMutex
}

// NewBlockIndex creates a new block index.
func NewBlockIndex() *BlockIndex {
	return &BlockIndex{
		primaryBlocks:           make(map[BlockKey][]uint32),
		surnameYearBlocks:       make(map[BlockKey][]uint32),
		surnameYearBucketBlocks: make(map[BlockKey][]uint32),
		surnameInitialBlocks:    make(map[BlockKey][]uint32),
		surnamePrefixBlocks:     make(map[BlockKey][]uint32),
		surnameGivenPrefixBlocks: make(map[BlockKey][]uint32),
		rescueBlocks:            make(map[BlockKey][]uint32),
		personBlocks:            make([]PersonBlock, 0),
		maxBlockSize:            5000, // Skip blocks larger than 5000 (adaptive)
	}
}

// hashKey creates a hash from multiple string/int components.
func hashKey(components ...interface{}) BlockKey {
	h := fnv.New64a()
	for _, c := range components {
		switch v := c.(type) {
		case string:
			h.Write([]byte(v))
		case int:
			h.Write([]byte(fmt.Sprintf("%d", v)))
		case byte:
			h.Write([]byte{v})
		}
		h.Write([]byte("|")) // Separator
	}
	return BlockKey(h.Sum64())
}

// buildBlockIndex builds the blocking index from individuals.
func (dd *DuplicateDetector) buildBlockIndex(individuals []*types.IndividualRecord) *BlockIndex {
	index := NewBlockIndex()
	index.personBlocks = make([]PersonBlock, len(individuals))

	// Precompute all person blocks
	for i, indi := range individuals {
		block := dd.computePersonBlock(uint32(i), indi)
		index.personBlocks[i] = block

		// PRIMARY BLOCK: surname_soundex + birthYear (only if both present)
		// Don't create blocks with "unknown" values - that creates giant junk blocks
		if block.SurnameSoundex != "" && block.BirthYear > 0 {
			key := hashKey(block.SurnameSoundex, block.BirthYear)
			index.primaryBlocks[key] = append(index.primaryBlocks[key], uint32(i))

			// Expanded year buckets (±1 year for fuzzy matching)
			for yearOffset := -1; yearOffset <= 1; yearOffset++ {
				expandedYear := block.BirthYear + yearOffset
				if expandedYear > 0 {
					key := hashKey(block.SurnameSoundex, expandedYear)
					index.surnameYearBlocks[key] = append(index.surnameYearBlocks[key], uint32(i))
				}
			}

			// 5-year bucket block (for missing/uncertain dates)
			if block.BirthYearBucket > 0 {
				key := hashKey(block.SurnameSoundex, block.BirthYearBucket)
				index.surnameYearBucketBlocks[key] = append(index.surnameYearBucketBlocks[key], uint32(i))
			}
		}

		// FALLBACK 1: surname_soundex + given_initial (when birth year missing)
		if block.SurnameSoundex != "" && block.GivenInitial != 0 {
			key := hashKey(block.SurnameSoundex, block.GivenInitial)
			index.surnameInitialBlocks[key] = append(index.surnameInitialBlocks[key], uint32(i))
		}

		// FALLBACK 2: surname_soundex + given_prefix(2) (looser than initial)
		if block.SurnameSoundex != "" && block.GivenPrefix != "" && len(block.GivenPrefix) >= 2 {
			key := hashKey(block.SurnameSoundex, block.GivenPrefix)
			index.surnameGivenPrefixBlocks[key] = append(index.surnameGivenPrefixBlocks[key], uint32(i))
		}

		// FALLBACK 3: surname_prefix(4) + birth_place_token (when year missing)
		if block.SurnamePrefix != "" && block.BirthPlaceToken != "" {
			key := hashKey(block.SurnamePrefix, block.BirthPlaceToken)
			index.surnamePrefixBlocks[key] = append(index.surnamePrefixBlocks[key], uint32(i))
		}

		// RESCUE BLOCK: for people with no other blocks
		// Only create if we have enough signal (given + surname + place)
		if block.GivenPrefix != "" && len(block.GivenPrefix) >= 3 &&
			block.SurnamePrefix != "" && block.BirthPlaceToken != "" {
			key := hashKey(block.GivenPrefix[:minInt(3, len(block.GivenPrefix))],
				block.SurnamePrefix[:minInt(3, len(block.SurnamePrefix))],
				block.BirthPlaceToken)
			index.rescueBlocks[key] = append(index.rescueBlocks[key], uint32(i))
		}
	}

	return index
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// computePersonBlock computes blocking data for a person.
func (dd *DuplicateDetector) computePersonBlock(id uint32, indi *types.IndividualRecord) PersonBlock {
	block := PersonBlock{
		ID:         id,
		Individual: indi,
	}

	// Compute surname soundex
	// Handle multi-part surnames: "van der Berg" -> use "Berg" for Soundex
	surname := normalizeString(indi.GetSurname())
	if surname == "" {
		// Fallback: Extract surname from NAME value if SURN sub-tag is missing
		// GEDCOM format: "Given Name /Surname/" or "Given Name /Surname1/Surname2/"
		nameValue := indi.GetName()
		if nameValue != "" {
			// Try to extract surname between slashes
			// Pattern: "text /surname/" or "text /surname1/surname2/"
			slashIdx := strings.Index(nameValue, "/")
			if slashIdx >= 0 {
				// Find closing slash
				closeIdx := strings.Index(nameValue[slashIdx+1:], "/")
				if closeIdx >= 0 {
					surname = strings.TrimSpace(nameValue[slashIdx+1 : slashIdx+1+closeIdx])
					surname = normalizeString(surname)
				}
			}
		}
	}
	if surname != "" {
		// Extract last significant word (handles "van der Berg", "de la Cruz", etc.)
		surnameParts := strings.Fields(surname)
		surnameForSoundex := surname
		if len(surnameParts) > 1 {
			// Common prefixes to skip
			prefixes := []string{"van", "von", "de", "del", "de la", "der", "den", "du", "le", "la", "les"}
			// Try to find last non-prefix word
			for i := len(surnameParts) - 1; i >= 0; i-- {
				part := strings.ToLower(surnameParts[i])
				isPrefix := false
				for _, prefix := range prefixes {
					if part == prefix {
						isPrefix = true
						break
					}
				}
				if !isPrefix {
					surnameForSoundex = surnameParts[i]
					break
				}
			}
		}
		
		block.SurnameSoundex = Soundex(surnameForSoundex)
		// Surname prefix (first 4 chars of last significant word, padded)
		if len(surnameForSoundex) >= 4 {
			block.SurnamePrefix = strings.ToUpper(surnameForSoundex[:4])
		} else {
			block.SurnamePrefix = strings.ToUpper(surnameForSoundex + strings.Repeat(" ", 4-len(surnameForSoundex)))
		}
	}

	// Compute given name initial
	given := normalizeString(indi.GetGivenName())
	if given != "" {
		block.GivenInitial = strings.ToUpper(given)[0]
		if len(given) >= 2 {
			block.GivenPrefix = strings.ToUpper(given[:2])
		} else {
			block.GivenPrefix = strings.ToUpper(given)
		}
	}

	// Compute birth year
	birthDate := indi.GetBirthDate()
	if birthDate != "" {
		block.BirthYear = extractYear(birthDate)
		if block.BirthYear > 0 {
			// 5-year buckets
			block.BirthYearBucket = block.BirthYear / 5
		}
	}

	// Compute birth place token (first significant token)
	birthPlace := normalizeString(indi.GetBirthPlace())
	if birthPlace != "" {
		tokens := strings.Fields(birthPlace)
		if len(tokens) > 0 {
			// Take first non-trivial token (skip common words)
			for _, token := range tokens {
				token = strings.ToUpper(token)
				if len(token) > 2 && !isCommonPlaceWord(token) {
					block.BirthPlaceToken = token
					break
				}
			}
			// If no good token found, use first token
			if block.BirthPlaceToken == "" && len(tokens) > 0 {
				block.BirthPlaceToken = strings.ToUpper(tokens[0])
			}
		}
	}

	return block
}

// isCommonPlaceWord checks if a word is a common place word (to skip).
func isCommonPlaceWord(word string) bool {
	common := []string{"THE", "OF", "IN", "ON", "AT", "TO", "FOR", "AND", "OR", "COUNTY", "CITY", "TOWN", "STATE", "PROVINCE"}
	for _, c := range common {
		if word == c {
			return true
		}
	}
	return false
}

// candidateInfo holds information about a candidate for prioritization.
type candidateInfo struct {
	id           uint32
	block        PersonBlock
	priority     int // Higher = better match
	yearDiff     int
	surnameMatch bool
	placeMatch   bool
}

// findCandidates finds candidate pairs using blocking with adaptive strategies.
// Returns a map of candidate IDs, prioritized by match quality.
func (index *BlockIndex) findCandidates(personID uint32, maxCandidates int) map[uint32]bool {
	block := index.personBlocks[personID]
	candidates := make(map[uint32]bool)
	candidateList := make([]candidateInfo, 0)

	// Strategy: try primary block first, then fallbacks if needed
	// Use adaptive blocking: skip giant blocks

	// PRIMARY: surname_soundex + birthYear (and ±1 year, ±2 year for better recall)
	if block.SurnameSoundex != "" && block.BirthYear > 0 {
		for yearOffset := -2; yearOffset <= 2; yearOffset++ {
			year := block.BirthYear + yearOffset
			if year > 0 {
				key := hashKey(block.SurnameSoundex, year)
				
				// Check primary blocks (adaptive: skip if too large)
				if ids, ok := index.primaryBlocks[key]; ok {
					if len(ids) <= index.maxBlockSize {
						for _, id := range ids {
							if id != personID && id > personID {
								candBlock := index.personBlocks[id]
								priority := index.computePriority(block, candBlock, yearOffset)
								candidateList = append(candidateList, candidateInfo{
									id:           id,
									block:        candBlock,
									priority:     priority,
									yearDiff:     absInt(block.BirthYear - candBlock.BirthYear),
									surnameMatch: block.SurnameSoundex == candBlock.SurnameSoundex,
									placeMatch:   block.BirthPlaceToken != "" && block.BirthPlaceToken == candBlock.BirthPlaceToken,
								})
							}
						}
					}
				}

				// Check expanded year blocks
				if ids, ok := index.surnameYearBlocks[key]; ok {
					if len(ids) <= index.maxBlockSize {
						for _, id := range ids {
							if id != personID && id > personID {
								// Check if already added
								alreadyAdded := false
								for _, cand := range candidateList {
									if cand.id == id {
										alreadyAdded = true
										break
									}
								}
								if !alreadyAdded {
									candBlock := index.personBlocks[id]
									priority := index.computePriority(block, candBlock, yearOffset)
									candidateList = append(candidateList, candidateInfo{
										id:           id,
										block:        candBlock,
										priority:     priority,
										yearDiff:     absInt(block.BirthYear - candBlock.BirthYear),
										surnameMatch: block.SurnameSoundex == candBlock.SurnameSoundex,
										placeMatch:   block.BirthPlaceToken != "" && block.BirthPlaceToken == candBlock.BirthPlaceToken,
									})
								}
							}
						}
					}
				}
			}
		}

		// 5-year bucket block (for missing/uncertain dates)
		if block.BirthYearBucket > 0 {
			key := hashKey(block.SurnameSoundex, block.BirthYearBucket)
			if ids, ok := index.surnameYearBucketBlocks[key]; ok {
				if len(ids) <= index.maxBlockSize {
					for _, id := range ids {
						if id != personID && id > personID {
							alreadyAdded := false
							for _, cand := range candidateList {
								if cand.id == id {
									alreadyAdded = true
									break
								}
							}
							if !alreadyAdded {
								candBlock := index.personBlocks[id]
								candidateList = append(candidateList, candidateInfo{
									id:       id,
									block:    candBlock,
									priority: 5, // Lower priority for bucket matches
								})
							}
						}
					}
				}
			}
		}
	}

	// FALLBACKS: only if we don't have enough candidates
	if maxCandidates == 0 || len(candidateList) < maxCandidates {
		// Fallback 1: surname_soundex + given_initial
		if block.SurnameSoundex != "" && block.GivenInitial != 0 {
			key := hashKey(block.SurnameSoundex, block.GivenInitial)
			if ids, ok := index.surnameInitialBlocks[key]; ok {
				if len(ids) <= index.maxBlockSize {
					for _, id := range ids {
						if id != personID && id > personID {
							alreadyAdded := false
							for _, cand := range candidateList {
								if cand.id == id {
									alreadyAdded = true
									break
								}
							}
							if !alreadyAdded {
								candBlock := index.personBlocks[id]
								candidateList = append(candidateList, candidateInfo{
									id:       id,
									block:    candBlock,
									priority: 3, // Lower priority for fallback
								})
							}
						}
					}
				}
			}
		}

		// Fallback 2: surname_soundex + given_prefix(2)
		if block.SurnameSoundex != "" && block.GivenPrefix != "" {
			key := hashKey(block.SurnameSoundex, block.GivenPrefix)
			if ids, ok := index.surnameGivenPrefixBlocks[key]; ok {
				if len(ids) <= index.maxBlockSize {
					for _, id := range ids {
						if id != personID && id > personID {
							alreadyAdded := false
							for _, cand := range candidateList {
								if cand.id == id {
									alreadyAdded = true
									break
								}
							}
							if !alreadyAdded {
								candBlock := index.personBlocks[id]
								candidateList = append(candidateList, candidateInfo{
									id:       id,
									block:    candBlock,
									priority: 2,
								})
							}
						}
					}
				}
			}
		}

		// Fallback 3: surname_prefix + birth_place_token
		if block.SurnamePrefix != "" && block.BirthPlaceToken != "" {
			key := hashKey(block.SurnamePrefix, block.BirthPlaceToken)
			if ids, ok := index.surnamePrefixBlocks[key]; ok {
				if len(ids) <= index.maxBlockSize {
					for _, id := range ids {
						if id != personID && id > personID {
							alreadyAdded := false
							for _, cand := range candidateList {
								if cand.id == id {
									alreadyAdded = true
									break
								}
							}
							if !alreadyAdded {
								candBlock := index.personBlocks[id]
								candidateList = append(candidateList, candidateInfo{
									id:       id,
									block:    candBlock,
									priority: 1,
								})
							}
						}
					}
				}
			}
		}

		// RESCUE: only if still no candidates
		if len(candidateList) == 0 {
			if block.GivenPrefix != "" && len(block.GivenPrefix) >= 3 &&
				block.SurnamePrefix != "" && block.BirthPlaceToken != "" {
				key := hashKey(block.GivenPrefix[:minInt(3, len(block.GivenPrefix))],
					block.SurnamePrefix[:minInt(3, len(block.SurnamePrefix))],
					block.BirthPlaceToken)
				if ids, ok := index.rescueBlocks[key]; ok {
					if len(ids) <= index.maxBlockSize {
						for _, id := range ids {
							if id != personID && id > personID {
								candBlock := index.personBlocks[id]
								candidateList = append(candidateList, candidateInfo{
									id:       id,
									block:    candBlock,
									priority: 0, // Lowest priority
								})
							}
						}
					}
				}
			}
		}
	}

	// Sort by priority (highest first) and take top candidates
	sort.Slice(candidateList, func(i, j int) bool {
		if candidateList[i].priority != candidateList[j].priority {
			return candidateList[i].priority > candidateList[j].priority
		}
		// Tie-breaker: prefer smaller year difference, then place match
		if candidateList[i].yearDiff != candidateList[j].yearDiff {
			return candidateList[i].yearDiff < candidateList[j].yearDiff
		}
		return candidateList[i].placeMatch && !candidateList[j].placeMatch
	})

	// Take top candidates up to limit
	takeCount := len(candidateList)
	if maxCandidates > 0 && takeCount > maxCandidates {
		takeCount = maxCandidates
	}

	for i := 0; i < takeCount; i++ {
		candidates[candidateList[i].id] = true
	}

	return candidates
}

// computePriority computes a priority score for a candidate match.
// Higher priority = better match quality.
func (index *BlockIndex) computePriority(block1, block2 PersonBlock, yearOffset int) int {
	priority := 10 // Base priority

	// Exact year match gets highest priority
	if yearOffset == 0 {
		priority += 5
	} else if absInt(yearOffset) == 1 {
		priority += 3
	} else if absInt(yearOffset) == 2 {
		priority += 1
	}

	// Surname exact match (not just Soundex)
	if block1.SurnamePrefix == block2.SurnamePrefix {
		priority += 3
	}

	// Place match
	if block1.BirthPlaceToken != "" && block1.BirthPlaceToken == block2.BirthPlaceToken {
		priority += 2
	}

	// Given name prefix match
	if block1.GivenPrefix != "" && block2.GivenPrefix != "" &&
		block1.GivenPrefix == block2.GivenPrefix {
		priority += 2
	}

	return priority
}

// absInt returns the absolute value of an integer.
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// generateBlockedComparisonJobs generates comparison jobs using blocking.
// This replaces the O(n²) approach with O(n * avg_block_size).
func (dd *DuplicateDetector) generateBlockedComparisonJobs(individuals []*types.IndividualRecord) ([]comparisonJob, *BlockingMetrics) {
	// Build block index
	blockIndex := dd.buildBlockIndex(individuals)

	// Generate jobs from blocks
	jobs := make([]comparisonJob, 0)
	maxCandidatesPerPerson := dd.config.MaxCandidatesPerPerson
	if maxCandidatesPerPerson <= 0 {
		maxCandidatesPerPerson = 200 // Default
	}

	candidatesPerPerson := make([]int, len(individuals))

	for i := uint32(0); i < uint32(len(individuals)); i++ {
		candidates := blockIndex.findCandidates(i, maxCandidatesPerPerson)
		candidatesPerPerson[i] = len(candidates)
		for candidateID := range candidates {
			jobs = append(jobs, comparisonJob{
				indi1: individuals[i],
				indi2: individuals[candidateID],
				index: len(jobs),
			})
		}
	}

	// Compute metrics
	metrics := blockIndex.computeBlockingMetrics(len(individuals), candidatesPerPerson)
	blockIndex.metrics = metrics

	return jobs, metrics
}

