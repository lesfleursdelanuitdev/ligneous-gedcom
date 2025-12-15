# Long-Term Feature Implementation Plan

**Document Version:** 1.0  
**Last Updated:** 2025-01-27

This document outlines the implementation plan for three major long-term features identified in the codebase analysis.

---

## 1. Streaming Parser for Very Large Files

### Current State
- **Status**: Entire GEDCOM tree loaded into memory
- **Limitation**: Files >100MB can cause memory issues
- **Use Case**: Large genealogy databases with 100K+ individuals

### Proposed Solution

#### Architecture
```go
// Streaming parser interface
type StreamingParser interface {
    ParseStream(filePath string, handler RecordHandler) error
}

// Record handler callback
type RecordHandler func(record Record) error
```

#### Implementation Approach

**Phase 1: Iterator-Based API**
```go
type RecordIterator struct {
    parser *HierarchicalParser
    recordChan chan Record
    errChan chan error
}

func (hp *HierarchicalParser) ParseStream(filePath string) (*RecordIterator, error) {
    // Open file, start goroutine to parse incrementally
    // Return iterator that yields records as they're parsed
}

func (ri *RecordIterator) Next() (Record, error) {
    // Return next record from channel
}
```

**Phase 2: Callback-Based Processing**
```go
type RecordHandler func(record Record) error

func (hp *HierarchicalParser) ParseWithHandler(filePath string, handler RecordHandler) error {
    // Parse file, call handler for each level-0 record
    // Allows processing records without storing entire tree
}
```

**Phase 3: Windowed Processing**
```go
type WindowedParser struct {
    windowSize int  // Number of records to keep in memory
}

func (wp *WindowedParser) ParseWithWindow(filePath string, windowSize int) error {
    // Process records in batches
    // Only keep N records in memory at a time
}
```

#### Benefits
- **Memory Efficiency**: Process files of any size
- **Faster Initial Response**: Start processing immediately
- **Scalability**: Handle very large genealogy databases

#### Implementation Complexity
- **Effort**: Medium (2-3 weeks)
- **Risk**: Low (can coexist with current parser)
- **Breaking Changes**: None (new API, existing API unchanged)

---

## 2. Advanced Validation Rules

### Current State
- **Status**: Basic GEDCOM 5.5.1 compliance validation
- **Limitation**: Fixed validation rules, no custom rules
- **Use Case**: Data quality checks, custom business rules

### Proposed Solution

#### Architecture
```go
// Validation rule interface
type ValidationRule interface {
    Name() string
    Validate(tree *GedcomTree) []*GedcomError
    Severity() ErrorSeverity
}

// Rule registry
type RuleRegistry struct {
    rules []ValidationRule
}

// Custom validator with pluggable rules
type AdvancedValidator struct {
    registry *RuleRegistry
    errorManager *ErrorManager
}
```

#### Built-in Advanced Rules

**1. Date Consistency Rules**
```go
type DateConsistencyRule struct {
    // Check birth date before death date
    // Check marriage date after birth date
    // Check age at marriage (reasonable range)
    // Check age at death (reasonable range)
}

func (dcr *DateConsistencyRule) Validate(tree *GedcomTree) []*GedcomError {
    // Validate date relationships
}
```

**2. Relationship Validation**
```go
type RelationshipRule struct {
    // Check parent-child age differences
    // Check sibling age differences
    // Check marriage age (minimum age)
    // Check duplicate relationships
}
```

**3. Data Quality Rules**
```go
type DataQualityRule struct {
    // Check for duplicate individuals
    // Check for missing required fields
    // Check for inconsistent data
    // Check for orphaned records
}
```

**4. Custom Business Rules**
```go
// User-defined validation rules
type CustomRule struct {
    name string
    validateFunc func(*GedcomTree) []*GedcomError
}
```

#### Usage Example
```go
// Create advanced validator
validator := NewAdvancedValidator(errorManager)

// Add built-in rules
validator.AddRule(NewDateConsistencyRule())
validator.AddRule(NewRelationshipRule())
validator.AddRule(NewDataQualityRule())

// Add custom rule
validator.AddRule(&CustomRule{
    name: "NoDuplicateNames",
    validateFunc: func(tree *GedcomTree) []*GedcomError {
        // Custom validation logic
    },
})

// Validate
err := validator.Validate(tree)
```

#### Benefits
- **Flexibility**: Pluggable rule system
- **Extensibility**: Easy to add new rules
- **Data Quality**: Comprehensive validation beyond GEDCOM spec
- **Customization**: Support for domain-specific rules

#### Implementation Complexity
- **Effort**: Medium (2-3 weeks)
- **Risk**: Low (additive feature)
- **Breaking Changes**: None (new validator, existing unchanged)

---

## 3. Date/Place Parsing Utilities

### Current State
- **Status**: Dates/places stored as raw strings
- **Example**: `"Dec 1859"`, `"Rapid City"`, `"15 JAN 1800"`
- **Limitation**: No structured access, no normalization, no validation

### Proposed Solution

#### Date Parsing Package (`pkg/gedcom/date`)

**Date Structure**
```go
type GedcomDate struct {
    Original    string      // Original GEDCOM date string
    Type        DateType    // EXACT, ABOUT, BEFORE, AFTER, BETWEEN, etc.
    Calendar    Calendar    // GREGORIAN, JULIAN, HEBREW, etc.
    
    // Exact date
    Year        int
    Month       int
    Day         int
    
    // Range dates
    StartYear   int
    StartMonth  int
    StartDay    int
    EndYear     int
    EndMonth    int
    EndDay      int
    
    // Parsed components
    IsParsed    bool
    ParseError  error
}
```

**Date Parsing**
```go
func ParseDate(dateStr string) (*GedcomDate, error) {
    // Parse GEDCOM date formats:
    // - "15 JAN 1800" (exact)
    // - "ABT 1850" (about)
    // - "BEF 1900" (before)
    // - "AFT 1900" (after)
    // - "BET 1800 AND 1850" (between)
    // - "FROM 1800 TO 1850" (range)
    // - "1800" (year only)
    // - "JAN 1800" (month-year)
}
```

**Date Methods**
```go
func (gd *GedcomDate) ToTime() (time.Time, error)
func (gd *GedcomDate) ToISO8601() string
func (gd *GedcomDate) ToFormatted(format string) string
func (gd *GedcomDate) IsValid() bool
func (gd *GedcomDate) IsRange() bool
func (gd *GedcomDate) Earliest() time.Time
func (gd *GedcomDate) Latest() time.Time
func (gd *GedcomDate) Compare(other *GedcomDate) int
```

**Usage Example**
```go
// Parse date
date, err := date.ParseDate("15 JAN 1800")
if err != nil {
    // Handle error
}

// Access components
year := date.Year  // 1800
month := date.Month  // 1
day := date.Day  // 15

// Convert to time.Time
t, err := date.ToTime()

// Format
iso := date.ToISO8601()  // "1800-01-15"
formatted := date.ToFormatted("January 15, 1800")
```

#### Place Parsing Package (`pkg/gedcom/place`)

**Place Structure**
```go
type GedcomPlace struct {
    Original    string   // Original GEDCOM place string
    Components  []string // Parsed components
    
    // Hierarchical components
    City        string
    County      string
    State       string
    Country     string
    PostalCode  string
    
    // Geographic data (optional)
    Latitude    float64
    Longitude   float64
    
    // Parsed status
    IsParsed    bool
    ParseError  error
}
```

**Place Parsing**
```go
func ParsePlace(placeStr string) (*GedcomPlace, error) {
    // Parse GEDCOM place formats:
    // - "Rapid City" (simple)
    // - "Rapid City, South Dakota" (city, state)
    // - "Rapid City, Pennington, South Dakota, USA" (full hierarchy)
    // - "New York, NY, USA" (with abbreviations)
}
```

**Place Methods**
```go
func (gp *GedcomPlace) ToFormatted(separator string) string
func (gp *GedcomPlace) GetComponent(level int) string
func (gp *GedcomPlace) IsValid() bool
func (gp *GedcomPlace) Normalize() *GedcomPlace
func (gp *GedcomPlace) Geocode() error  // Optional: lookup coordinates
```

**Usage Example**
```go
// Parse place
place, err := place.ParsePlace("Rapid City, South Dakota")
if err != nil {
    // Handle error
}

// Access components
city := place.City  // "Rapid City"
state := place.State  // "South Dakota"

// Format
formatted := place.ToFormatted(", ")  // "Rapid City, South Dakota"
```

#### Integration with Records

**Enhanced Record Methods**
```go
// IndividualRecord enhancements
func (ir *IndividualRecord) GetBirthDateParsed() (*date.GedcomDate, error) {
    dateStr := ir.GetBirthDate()
    return date.ParseDate(dateStr)
}

func (ir *IndividualRecord) GetBirthPlaceParsed() (*place.GedcomPlace, error) {
    placeStr := ir.GetBirthPlace()
    return place.ParsePlace(placeStr)
}

// Backward compatibility: existing string methods still work
func (ir *IndividualRecord) GetBirthDate() string  // Still available
```

#### Date/Place Validation

**Date Validation Rules**
```go
type DateValidationRule struct {
    // Validate date format
    // Validate date ranges
    // Validate calendar consistency
    // Validate date relationships
}
```

**Place Validation Rules**
```go
type PlaceValidationRule struct {
    // Validate place format
    // Validate place hierarchy
    // Validate place consistency
}
```

#### Benefits
- **Structured Access**: Type-safe date/place objects
- **Normalization**: Consistent date/place formats
- **Validation**: Built-in format checking
- **Flexibility**: Support for various GEDCOM date/place formats
- **Backward Compatibility**: Existing string methods still work

#### Implementation Complexity
- **Effort**: High (4-6 weeks)
- **Risk**: Medium (new dependencies, parsing complexity)
- **Breaking Changes**: None (additive feature, backward compatible)

---

## Implementation Priority

### Recommended Order

1. **Date/Place Parsing Utilities** (Highest Value)
   - Most requested feature
   - High user value
   - Foundation for advanced validation

2. **Advanced Validation Rules** (Medium Priority)
   - Builds on date/place parsing
   - Improves data quality
   - Moderate complexity

3. **Streaming Parser** (Lower Priority)
   - Only needed for very large files
   - Lower user impact
   - Can be added later

### Estimated Timeline

- **Date/Place Parsing**: 4-6 weeks
- **Advanced Validation**: 2-3 weeks (after date/place)
- **Streaming Parser**: 2-3 weeks

**Total**: 8-12 weeks for all three features

---

## Design Considerations

### Backward Compatibility
- All new features are **additive**
- Existing APIs remain unchanged
- New methods added alongside old ones
- Gradual migration path for users

### Testing Strategy
- Comprehensive unit tests for parsing
- Integration tests with real GEDCOM files
- Edge case testing (malformed dates/places)
- Performance benchmarks

### Documentation
- API documentation for new packages
- Usage examples
- Migration guide (if needed)
- Best practices guide

---

## Related Documentation

- `CODEBASE_ANALYSIS_COMPREHENSIVE.md`: Overall codebase analysis
- `ANALYSIS_SUMMARY_2024.md`: Summary with recommendations
- `EXAMPLES.md`: Current usage examples (to be updated)

---

**Document Version:** 1.0  
**Last Updated:** 2025-01-27
