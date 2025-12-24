package validator

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// ValidationRule is the interface that all advanced validation rules must implement.
type ValidationRule interface {
	// Name returns the name of the validation rule.
	Name() string

	// Validate validates the GEDCOM tree and returns all errors found.
	// Errors should be categorized by severity (Severe, Warning, Info, Hint).
	Validate(tree *types.GedcomTree, config *ValidationConfig) []*types.GedcomError

	// Description returns a human-readable description of what this rule checks.
	Description() string
}

// ValidationConfig holds configuration for advanced validation rules.
// All thresholds are configurable to support different cultures and historical periods.
type ValidationConfig struct {
	// Age thresholds
	MinParentAge     int // Minimum age for parent at child's birth (default: 10)
	MaxParentAge     int // Maximum age for parent at child's birth (default: 80)
	MinMarriageAge   int // Minimum age for marriage (default: 12)
	MaxMarriageAge   int // Maximum age for marriage (default: 100)
	MaxDeathAge      int // Maximum reasonable age at death (default: 120)
	SpouseAgeGapWarn int // Age gap between spouses to trigger warning (default: 30)
	SpouseAgeGapHint int // Age gap between spouses to trigger hint (default: 40)

	// Date thresholds
	MinHistoricalDate int // Minimum reasonable historical date in CE (default: 500)
	MaxFutureDate     int // Maximum future date allowed (default: current year + 1)
	DateRangeWarn     int // Date range width to trigger warning in years (default: 50)

	// Duplicate detection thresholds
	NameSimilarity float64 // Name similarity threshold for duplicates (default: 0.85)
	DateSimilarity int     // Date similarity threshold in years (default: ±1)

	// Severity filtering
	MinSeverity types.ErrorSeverity // Minimum severity to report (default: Hint)
}

// NewValidationConfig creates a new ValidationConfig with sensible defaults.
func NewValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MinParentAge:      10,
		MaxParentAge:      80,
		MinMarriageAge:    12,
		MaxMarriageAge:    100,
		MaxDeathAge:       120,
		SpouseAgeGapWarn:  30,
		SpouseAgeGapHint:  40,
		MinHistoricalDate: 500,
		MaxFutureDate:     2026, // Current year + 1, should be dynamic
		DateRangeWarn:     50,
		NameSimilarity:    0.85,
		DateSimilarity:    1,
		MinSeverity:       types.SeverityHint, // Show all by default
	}
}

// AdvancedValidator orchestrates advanced validation rules.
// It provides a pluggable rule system for data quality and consistency checks.
type AdvancedValidator struct {
	errorManager *types.ErrorManager
	config       *ValidationConfig
	rules        []ValidationRule
}

// NewAdvancedValidator creates a new AdvancedValidator with default configuration.
func NewAdvancedValidator(errorManager *types.ErrorManager) *AdvancedValidator {
	return &AdvancedValidator{
		errorManager: errorManager,
		config:       NewValidationConfig(),
		rules:        make([]ValidationRule, 0),
	}
}

// NewAdvancedValidatorWithConfig creates a new AdvancedValidator with custom configuration.
func NewAdvancedValidatorWithConfig(errorManager *types.ErrorManager, config *ValidationConfig) *AdvancedValidator {
	return &AdvancedValidator{
		errorManager: errorManager,
		config:       config,
		rules:        make([]ValidationRule, 0),
	}
}

// AddRule adds a validation rule to the validator.
func (av *AdvancedValidator) AddRule(rule ValidationRule) {
	av.rules = append(av.rules, rule)
}

// GetRules returns all registered rules.
func (av *AdvancedValidator) GetRules() []ValidationRule {
	return av.rules
}

// Validate runs all registered validation rules on the tree.
func (av *AdvancedValidator) Validate(tree *types.GedcomTree) error {
	for _, rule := range av.rules {
		errors := rule.Validate(tree, av.config)
		for _, err := range errors {
			// Only add errors that meet the minimum severity threshold
			if av.shouldReportError(err) {
				av.errorManager.AddError(err.Severity, err.Message, err.LineNumber, err.Context)
			}
		}
	}
	return nil
}

// shouldReportError checks if an error should be reported based on severity threshold.
func (av *AdvancedValidator) shouldReportError(err *types.GedcomError) bool {
	severityOrder := map[types.ErrorSeverity]int{
		types.SeverityHint:    0,
		types.SeverityInfo:    1,
		types.SeverityWarning: 2,
		types.SeveritySevere:  3,
	}

	minLevel := severityOrder[av.config.MinSeverity]
	errLevel := severityOrder[err.Severity]

	return errLevel >= minLevel
}

// GetErrorManager returns the error manager.
func (av *AdvancedValidator) GetErrorManager() *types.ErrorManager {
	return av.errorManager
}

// GetConfig returns the validation configuration.
func (av *AdvancedValidator) GetConfig() *ValidationConfig {
	return av.config
}

// SetConfig updates the validation configuration.
func (av *AdvancedValidator) SetConfig(config *ValidationConfig) {
	av.config = config
}
