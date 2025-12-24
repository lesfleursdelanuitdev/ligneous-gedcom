package duplicate

import (
	"strings"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// DuplicateConfig holds configuration for duplicate detection.
type DuplicateConfig struct {
	// Thresholds
	MinThreshold            float64 // Minimum similarity to report (default: 0.60)
	HighConfidenceThreshold float64 // High confidence threshold (default: 0.85)
	ExactMatchThreshold     float64 // Exact match threshold (default: 0.95)

	// Weights
	NameWeight         float64 // Name similarity weight (default: 0.40)
	DateWeight         float64 // Date similarity weight (default: 0.30)
	PlaceWeight        float64 // Place similarity weight (default: 0.15)
	SexWeight          float64 // Sex match weight (default: 0.05)
	RelationshipWeight float64 // Relationship weight (default: 0.10)

	// Options
	UsePhoneticMatching   bool // Use Soundex/Metaphone (default: true for Phase 2)
	UseRelationshipData   bool // Use family relationships (default: true for Phase 2)
	UseParallelProcessing bool // Use parallel processing (default: true for Phase 3)
	UseBlocking           bool // Use blocking for candidate generation (default: true)
	DateTolerance         int  // Years tolerance for dates (default: 2)
	MaxComparisons        int  // Limit comparisons for performance (0 = unlimited)
	MaxCandidatesPerPerson int // Max candidates per person when blocking (default: 200)
	NumWorkers            int  // Number of worker goroutines (0 = auto-detect)
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *DuplicateConfig {
	return &DuplicateConfig{
		MinThreshold:            0.60,
		HighConfidenceThreshold: 0.85,
		ExactMatchThreshold:     0.95,
		NameWeight:              0.40,
		DateWeight:              0.30,
		PlaceWeight:             0.15,
		SexWeight:               0.05,
		RelationshipWeight:      0.10,
		UsePhoneticMatching:     true, // Phase 2: enabled
		UseRelationshipData:     true, // Phase 2: enabled
		UseParallelProcessing:   true, // Phase 3: enabled
		UseBlocking:             true, // Use blocking for O(n²) -> O(n) reduction
		DateTolerance:           2,
		MaxComparisons:          0, // Unlimited
		MaxCandidatesPerPerson:  200, // Max candidates per person when blocking
		NumWorkers:              0, // Auto-detect
	}
}

// DuplicateMatch represents a potential duplicate match between two individuals.
type DuplicateMatch struct {
	Individual1     *types.IndividualRecord
	Individual2     *types.IndividualRecord
	SimilarityScore float64
	Confidence      string // "exact", "high", "medium", "low"
	MatchingFields  []string
	Differences     []string

	// Breakdown
	NameScore         float64
	DateScore         float64
	PlaceScore        float64
	SexScore          float64
	RelationshipScore float64
}

// DuplicateResult holds the results of duplicate detection.
type DuplicateResult struct {
	Matches          []DuplicateMatch
	TotalComparisons int
	ProcessingTime   time.Duration
	Metrics          *PerformanceMetrics // Optional performance metrics
	BlockingMetrics  *BlockingMetrics   // Optional blocking metrics
}

// DuplicateDetector detects potential duplicate individuals.
type DuplicateDetector struct {
	config *DuplicateConfig
	tree   *types.GedcomTree // Optional: for relationship matching
}

// NewDuplicateDetector creates a new duplicate detector with the given configuration.
func NewDuplicateDetector(config *DuplicateConfig) *DuplicateDetector {
	if config == nil {
		config = DefaultConfig()
	}
	return &DuplicateDetector{
		config: config,
		tree:   nil,
	}
}

// SetTree sets the GEDCOM tree for relationship matching.
// This is required for relationship similarity calculations.
func (dd *DuplicateDetector) SetTree(tree *types.GedcomTree) {
	dd.tree = tree
}

// FindDuplicates finds potential duplicates within a single GEDCOM tree.
func (dd *DuplicateDetector) FindDuplicates(tree *types.GedcomTree) (*DuplicateResult, error) {
	// Set tree for relationship matching
	dd.SetTree(tree)
	startTime := time.Now()

	// Get all individuals
	allIndividuals := tree.GetAllIndividuals()
	individuals := make([]*types.IndividualRecord, 0, len(allIndividuals))
	for _, record := range allIndividuals {
		if indi, ok := record.(*types.IndividualRecord); ok {
			individuals = append(individuals, indi)
		}
	}

	if len(individuals) < 2 {
		return &DuplicateResult{
			Matches:          []DuplicateMatch{},
			TotalComparisons: 0,
			ProcessingTime:   time.Since(startTime),
		}, nil
	}

	// Build indexes (measure time) - indexes are built inside parallel/sequential functions
	indexStartTime := time.Now()
	indexBuildTime := time.Since(indexStartTime)

	// Use parallel processing if enabled
	var matches []DuplicateMatch
	var comparisonCount int
	var err error
	var numWorkers int
	comparisonStartTime := time.Now()

	var blockingMetrics *BlockingMetrics
	if dd.config.UseParallelProcessing && len(individuals) > 10 {
		// Use parallel processing for larger datasets
		numWorkers = dd.getNumWorkers()
		matches, comparisonCount, blockingMetrics, err = dd.findDuplicatesParallel(individuals)
		if err != nil {
			return nil, err
		}
	} else {
		// Use sequential processing for small datasets
		numWorkers = 1
		matches, comparisonCount, blockingMetrics, err = dd.findDuplicatesSequential(individuals)
		if err != nil {
			return nil, err
		}
	}
	
	// blockingMetrics is now available for attachment to result
	comparisonTime := time.Since(comparisonStartTime)

	// Sort by similarity score (descending)
	sortStartTime := time.Now()
	dd.sortMatches(matches)
	sortTime := time.Since(sortStartTime)

	// Calculate metrics
	metrics := dd.calculateMetrics(
		startTime,
		indexBuildTime,
		comparisonTime,
		sortTime,
		comparisonCount,
		0, // Filtered comparisons (could be calculated if needed)
		numWorkers,
	)

	return &DuplicateResult{
		Matches:          matches,
		TotalComparisons: comparisonCount,
		ProcessingTime:   time.Since(startTime),
		Metrics:          metrics,
		BlockingMetrics:  blockingMetrics,
	}, nil
}

// FindDuplicatesBetween finds potential duplicates between two GEDCOM trees.
func (dd *DuplicateDetector) FindDuplicatesBetween(tree1, tree2 *types.GedcomTree) (*DuplicateResult, error) {
	startTime := time.Now()

	// Get all individuals from both trees
	allIndi1 := tree1.GetAllIndividuals()
	allIndi2 := tree2.GetAllIndividuals()

	individuals1 := make([]*types.IndividualRecord, 0, len(allIndi1))
	for _, record := range allIndi1 {
		if indi, ok := record.(*types.IndividualRecord); ok {
			individuals1 = append(individuals1, indi)
		}
	}

	individuals2 := make([]*types.IndividualRecord, 0, len(allIndi2))
	for _, record := range allIndi2 {
		if indi, ok := record.(*types.IndividualRecord); ok {
			individuals2 = append(individuals2, indi)
		}
	}

	if len(individuals1) == 0 || len(individuals2) == 0 {
		return &DuplicateResult{
			Matches:          []DuplicateMatch{},
			TotalComparisons: 0,
			ProcessingTime:   time.Since(startTime),
		}, nil
	}

	// Build indexes (measure time) - indexes are built inside parallel/sequential functions
	indexStartTime := time.Now()
	indexBuildTime := time.Since(indexStartTime)

	// Use parallel processing if enabled
	var matches []DuplicateMatch
	var comparisonCount int
	var err error
	var numWorkers int
	comparisonStartTime := time.Now()

	totalComparisons := len(individuals1) * len(individuals2)
	if dd.config.UseParallelProcessing && totalComparisons > 100 {
		// Use parallel processing for larger datasets
		numWorkers = dd.getNumWorkers()
		matches, comparisonCount, err = dd.findDuplicatesBetweenParallel(individuals1, individuals2)
		if err != nil {
			return nil, err
		}
	} else {
		// Use sequential processing for small datasets
		numWorkers = 1
		matches, comparisonCount, err = dd.findDuplicatesBetweenSequential(individuals1, individuals2)
		if err != nil {
			return nil, err
		}
	}
	comparisonTime := time.Since(comparisonStartTime)

	// Sort by similarity score (descending)
	sortStartTime := time.Now()
	dd.sortMatches(matches)
	sortTime := time.Since(sortStartTime)

	// Calculate metrics
	metrics := dd.calculateMetrics(
		startTime,
		indexBuildTime,
		comparisonTime,
		sortTime,
		comparisonCount,
		0, // Filtered comparisons
		numWorkers,
	)

	return &DuplicateResult{
		Matches:          matches,
		TotalComparisons: comparisonCount,
		ProcessingTime:   time.Since(startTime),
		Metrics:          metrics,
	}, nil
}

// FindMatches finds potential matches for a single individual within a tree.
func (dd *DuplicateDetector) FindMatches(individual *types.IndividualRecord, tree *types.GedcomTree) ([]DuplicateMatch, error) {
	// Set tree for relationship matching
	dd.SetTree(tree)
	startTime := time.Now()
	defer func() { _ = time.Since(startTime) }()

	// Get all individuals
	allIndividuals := tree.GetAllIndividuals()
	matches := make([]DuplicateMatch, 0)

	for _, record := range allIndividuals {
		indi, ok := record.(*types.IndividualRecord)
		if !ok {
			continue
		}

		// Skip self
		if indi.XrefID() == individual.XrefID() {
			continue
		}

		// Calculate similarity
		match, err := dd.compare(individual, indi)
		if err != nil {
			continue
		}

		// Filter by threshold
		if match.SimilarityScore >= dd.config.MinThreshold {
			matches = append(matches, *match)
		}
	}

	// Sort by similarity score (descending)
	dd.sortMatches(matches)

	return matches, nil
}

// Compare compares two individuals and returns their similarity score.
func (dd *DuplicateDetector) Compare(indi1, indi2 *types.IndividualRecord) (float64, error) {
	match, err := dd.compare(indi1, indi2)
	if err != nil {
		return 0.0, err
	}
	return match.SimilarityScore, nil
}

// compare is the internal comparison method that returns a full match.
func (dd *DuplicateDetector) compare(indi1, indi2 *types.IndividualRecord) (*DuplicateMatch, error) {
	// Calculate individual similarity scores
	nameScore := dd.calculateNameSimilarity(indi1, indi2)
	dateScore := dd.calculateDateSimilarity(indi1, indi2)
	placeScore := dd.calculatePlaceSimilarity(indi1, indi2)
	sexScore := dd.calculateSexSimilarity(indi1, indi2)

	// Calculate relationship similarity if enabled and tree is available
	relationshipScore := 0.0
	if dd.config.UseRelationshipData && dd.tree != nil {
		relationshipScore = dd.calculateRelationshipSimilarity(indi1, indi2)
	}

	// Calculate weighted sum
	totalScore := (nameScore * dd.config.NameWeight) +
		(dateScore * dd.config.DateWeight) +
		(placeScore * dd.config.PlaceWeight) +
		(sexScore * dd.config.SexWeight) +
		(relationshipScore * dd.config.RelationshipWeight)

	// Determine confidence level
	confidence := dd.determineConfidence(totalScore)

	// Determine matching fields and differences
	matchingFields, differences := dd.analyzeFields(indi1, indi2, nameScore, dateScore, placeScore, sexScore)

	return &DuplicateMatch{
		Individual1:       indi1,
		Individual2:       indi2,
		SimilarityScore:   totalScore,
		Confidence:        confidence,
		MatchingFields:    matchingFields,
		Differences:       differences,
		NameScore:         nameScore,
		DateScore:         dateScore,
		PlaceScore:        placeScore,
		SexScore:          sexScore,
		RelationshipScore: relationshipScore,
	}, nil
}

// determineConfidence determines the confidence level based on the similarity score.
func (dd *DuplicateDetector) determineConfidence(score float64) string {
	if score >= dd.config.ExactMatchThreshold {
		return "exact"
	} else if score >= dd.config.HighConfidenceThreshold {
		return "high"
	} else if score >= 0.70 {
		return "medium"
	}
	return "low"
}

// analyzeFields determines which fields match and which differ.
func (dd *DuplicateDetector) analyzeFields(indi1, indi2 *types.IndividualRecord,
	nameScore, dateScore, placeScore, sexScore float64) ([]string, []string) {
	matchingFields := make([]string, 0)
	differences := make([]string, 0)

	if nameScore >= 0.8 {
		matchingFields = append(matchingFields, "name")
	} else if nameScore > 0.0 {
		differences = append(differences, "name")
	}

	if dateScore >= 0.8 {
		matchingFields = append(matchingFields, "birth_date")
	} else if dateScore > 0.0 {
		differences = append(differences, "birth_date")
	}

	if placeScore >= 0.8 {
		matchingFields = append(matchingFields, "birth_place")
	} else if placeScore > 0.0 {
		differences = append(differences, "birth_place")
	}

	if sexScore >= 0.8 {
		matchingFields = append(matchingFields, "sex")
	} else if sexScore < 0.5 {
		differences = append(differences, "sex")
	}

	return matchingFields, differences
}

// sortMatches sorts matches by similarity score in descending order.
func (dd *DuplicateDetector) sortMatches(matches []DuplicateMatch) {
	// Simple insertion sort for small lists, or use sort.Slice for larger lists
	for i := 1; i < len(matches); i++ {
		key := matches[i]
		j := i - 1
		for j >= 0 && matches[j].SimilarityScore < key.SimilarityScore {
			matches[j+1] = matches[j]
			j--
		}
		matches[j+1] = key
	}
}

// Indexes for pre-filtering
type indexes struct {
	surnameIndex   map[string][]*types.IndividualRecord // surname -> individuals
	birthYearIndex map[int][]*types.IndividualRecord    // birth year -> individuals
	placeIndex     map[string][]*types.IndividualRecord // place (normalized) -> individuals
}

// buildIndexes builds indexes for pre-filtering.
func (dd *DuplicateDetector) buildIndexes(individuals []*types.IndividualRecord) *indexes {
	idx := &indexes{
		surnameIndex:   make(map[string][]*types.IndividualRecord),
		birthYearIndex: make(map[int][]*types.IndividualRecord),
		placeIndex:     make(map[string][]*types.IndividualRecord),
	}

	for _, indi := range individuals {
		// Index by surname
		surname := normalizeString(indi.GetSurname())
		if surname != "" {
			idx.surnameIndex[surname] = append(idx.surnameIndex[surname], indi)
		}

		// Index by birth year
		birthDate := indi.GetBirthDate()
		if birthDate != "" {
			if year := extractYear(birthDate); year > 0 {
				idx.birthYearIndex[year] = append(idx.birthYearIndex[year], indi)
			}
		}

		// Index by birth place
		birthPlace := normalizeString(indi.GetBirthPlace())
		if birthPlace != "" {
			idx.placeIndex[birthPlace] = append(idx.placeIndex[birthPlace], indi)
		}
	}

	return idx
}

// shouldCompare determines if two individuals should be compared (single file).
func (dd *DuplicateDetector) shouldCompare(indi1, indi2 *types.IndividualRecord, idx *indexes) bool {
	// Early termination: if name similarity is too low, skip
	name1 := normalizeString(indi1.GetName())
	name2 := normalizeString(indi2.GetName())
	if name1 != "" && name2 != "" {
		if quickNameSimilarity(name1, name2) < 0.3 {
			return false
		}
	}

	// Check if surnames match (if both have surnames)
	surname1 := normalizeString(indi1.GetSurname())
	surname2 := normalizeString(indi2.GetSurname())
	if surname1 != "" && surname2 != "" && surname1 != surname2 {
		return false
	}

	// Check if birth years are within tolerance
	year1 := extractYear(indi1.GetBirthDate())
	year2 := extractYear(indi2.GetBirthDate())
	if year1 > 0 && year2 > 0 {
		diff := abs(year1 - year2)
		if diff > dd.config.DateTolerance+10 { // Add buffer for pre-filtering
			return false
		}
	}

	return true
}

// shouldCompareCrossFile determines if two individuals should be compared (cross-file).
func (dd *DuplicateDetector) shouldCompareCrossFile(indi1, indi2 *types.IndividualRecord,
	idx1, idx2 *indexes) bool {
	// Similar logic to shouldCompare, but for cross-file
	return dd.shouldCompare(indi1, indi2, idx1)
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// quickNameSimilarity performs a quick name similarity check for pre-filtering.
func quickNameSimilarity(name1, name2 string) float64 {
	norm1 := normalizeName(name1)
	norm2 := normalizeName(name2)

	if norm1 == norm2 {
		return 1.0
	}

	// Quick check: if first word matches, likely similar
	words1 := strings.Fields(norm1)
	words2 := strings.Fields(norm2)
	if len(words1) > 0 && len(words2) > 0 {
		if words1[0] == words2[0] {
			return 0.5 // Partial match
		}
	}

	return 0.0
}
