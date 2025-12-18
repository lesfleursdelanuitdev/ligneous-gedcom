// Package duplicate provides duplicate detection functionality for GEDCOM individuals.
//
// The duplicate detection system identifies potential duplicate individuals
// within a single GEDCOM file or across multiple files, using weighted similarity
// scoring based on names, dates, places, sex, and relationships.
//
// Basic Usage:
//
//	// Create a detector with default configuration
//	detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())
//
//	// Find duplicates in a single file
//	result, err := detector.FindDuplicates(tree)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Process matches
//	for _, match := range result.Matches {
//		fmt.Printf("Match: %s and %s (similarity: %.2f, confidence: %s)\n",
//			match.Individual1.XrefID(),
//			match.Individual2.XrefID(),
//			match.SimilarityScore,
//			match.Confidence)
//	}
//
// Cross-File Detection:
//
//	// Find duplicates between two files
//	result, err := detector.FindDuplicatesBetween(tree1, tree2)
//
// Single Individual Matching:
//
//	// Find matches for a specific individual
//	matches, err := detector.FindMatches(individual, tree)
//
// Configuration:
//
// The duplicate detection can be customized using DuplicateConfig:
//
//	config := &duplicate.DuplicateConfig{
//		MinThreshold:          0.70,  // Minimum similarity to report
//		HighConfidenceThreshold: 0.85,  // High confidence threshold
//		ExactMatchThreshold:   0.95,  // Exact match threshold
//		NameWeight:            0.40,  // Weight for name similarity
//		DateWeight:            0.30,  // Weight for date similarity
//		PlaceWeight:           0.15,  // Weight for place similarity
//		SexWeight:             0.05,  // Weight for sex match
//		RelationshipWeight:    0.10,  // Weight for relationship similarity
//		DateTolerance:        2,     // Years tolerance for dates
//	}
//
// Similarity Metrics:
//
// The system calculates similarity using multiple metrics:
//
//   - Name Similarity (40% weight): Exact, normalized, component, and fuzzy matching
//   - Date Similarity (30% weight): Year comparison with tolerance for imprecise dates
//   - Place Similarity (15% weight): Exact and component matching
//   - Sex Match (5% weight): Match/mismatch indicator
//   - Relationship Similarity (10% weight): Common parents, spouses, children (Phase 2)
//
// Confidence Levels:
//
//   - exact: 0.95-1.0 - Almost certainly the same person
//   - high: 0.85-0.94 - Very likely the same person
//   - medium: 0.70-0.84 - Possibly the same person
//   - low: 0.60-0.69 - Unlikely but possible
//
// Performance:
//
// The system uses pre-filtering and indexing to optimize performance:
//   - Indexing by surname, birth year, and place
//   - Early termination for low-probability matches
//   - Configurable comparison limits
//
// For large files (10,000+ individuals), expect processing times of
// approximately 1 minute or less with default settings.
package duplicate
