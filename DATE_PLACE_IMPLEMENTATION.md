# Date/Place Parsing Implementation

**Date:** 2025-01-27  
**Status:** ✅ Complete

## Overview

Implemented comprehensive date and place parsing utilities for GEDCOM files, providing structured access to dates and places with normalization and formatting capabilities.

## Features Implemented

### 1. Date Parsing (`pkg/gedcom/date.go`)

**Supported Formats:**
- ✅ Exact dates: `"15 JAN 1800"`
- ✅ Month-year: `"JAN 1800"`
- ✅ Year only: `"1800"`
- ✅ About: `"ABT 1850"`
- ✅ Before: `"BEF 1900"`
- ✅ After: `"AFT 1900"`
- ✅ Between: `"BET 1800 AND 1850"`
- ✅ From-To: `"FROM 1800 TO 1850"`

**Key Methods:**
- `ParseDate(dateStr string) (*GedcomDate, error)` - Parse GEDCOM date string
- `ToTime() (time.Time, error)` - Convert to Go time.Time
- `ToISO8601() string` - Convert to ISO 8601 format
- `IsValid() bool` - Check if date was successfully parsed
- `IsRange() bool` - Check if date is a range
- `Earliest() time.Time` - Get earliest possible time
- `Latest() time.Time` - Get latest possible time
- `Compare(other *GedcomDate) int` - Compare two dates

**Example:**
```go
date, err := ParseDate("15 JAN 1800")
if err != nil {
    // Handle error
}

year := date.Year  // 1800
month := date.Month  // 1
day := date.Day  // 15

iso := date.ToISO8601()  // "1800-01-15"
t, _ := date.ToTime()  // time.Time
```

### 2. Place Parsing (`pkg/gedcom/place.go`)

**Supported Formats:**
- ✅ Simple: `"Rapid City"`
- ✅ City-State: `"Rapid City, South Dakota"`
- ✅ Full hierarchy: `"Rapid City, Pennington, South Dakota, USA"`
- ✅ With abbreviations: `"New York, NY, USA"`

**Key Methods:**
- `ParsePlace(placeStr string) (*GedcomPlace, error)` - Parse GEDCOM place string
- `ToFormatted(separator string) string` - Format with custom separator
- `GetComponent(level int) string` - Get component at specific level
- `IsValid() bool` - Check if place was successfully parsed
- `Normalize() *GedcomPlace` - Normalize place (trim, standardize)
- `Geocode() error` - Placeholder for future geocoding

**Example:**
```go
place, err := ParsePlace("Rapid City, South Dakota")
if err != nil {
    // Handle error
}

city := place.City  // "Rapid City"
state := place.State  // "South Dakota"

formatted := place.ToFormatted(", ")  // "Rapid City, South Dakota"
```

### 3. Integration with Records

**IndividualRecord Methods:**
- `GetBirthDateParsed() (*GedcomDate, error)`
- `GetDeathDateParsed() (*GedcomDate, error)`
- `GetBirthPlaceParsed() (*GedcomPlace, error)`
- `GetDeathPlaceParsed() (*GedcomPlace, error)`

**FamilyRecord Methods:**
- `GetMarriageDateParsed() (*GedcomDate, error)`
- `GetDivorceDateParsed() (*GedcomDate, error)`
- `GetMarriagePlaceParsed() (*GedcomPlace, error)`
- `GetDivorcePlaceParsed() (*GedcomPlace, error)`

**Example:**
```go
indi := tree.GetIndividual("@I1@")
birthDate, err := indi.GetBirthDateParsed()
if err == nil {
    fmt.Printf("Born: %s (%s)\n", birthDate.ToISO8601(), birthDate.String())
}

birthPlace, err := indi.GetBirthPlaceParsed()
if err == nil {
    fmt.Printf("Place: %s, %s\n", birthPlace.City, birthPlace.State)
}
```

## Backward Compatibility

✅ **Fully Backward Compatible**
- Existing string methods (`GetBirthDate()`, `GetBirthPlace()`, etc.) remain unchanged
- New parsed methods are additive
- No breaking changes

## Test Coverage

**Date Parsing Tests:**
- ✅ Exact date parsing
- ✅ Month-year parsing
- ✅ Year-only parsing
- ✅ Date type prefixes (ABT, BEF, AFT)
- ✅ Range dates (BET, FROM-TO)
- ✅ ISO 8601 conversion
- ✅ Time conversion
- ✅ Date comparison

**Place Parsing Tests:**
- ✅ Simple place parsing
- ✅ City-state parsing
- ✅ Full hierarchy parsing
- ✅ Abbreviation handling
- ✅ Component extraction
- ✅ Formatting
- ✅ Normalization

**Integration Tests:**
- ✅ IndividualRecord date/place parsing
- ✅ FamilyRecord date/place parsing

## Files Created/Modified

**New Files:**
1. `pkg/gedcom/date.go` - Date parsing implementation
2. `pkg/gedcom/date_test.go` - Date parsing tests
3. `pkg/gedcom/place.go` - Place parsing implementation
4. `pkg/gedcom/place_test.go` - Place parsing tests
5. `pkg/gedcom/date_place_integration_test.go` - Integration tests

**Modified Files:**
1. `pkg/gedcom/individual_record.go` - Added parsed date/place methods
2. `pkg/gedcom/family_record.go` - Added parsed date/place methods

## Usage Examples

### Basic Date Parsing
```go
// Parse various date formats
dates := []string{
    "15 JAN 1800",
    "ABT 1850",
    "BET 1800 AND 1850",
    "FROM 1800 TO 1850",
}

for _, dateStr := range dates {
    date, err := ParseDate(dateStr)
    if err != nil {
        continue
    }
    fmt.Printf("%s -> %s\n", dateStr, date.ToISO8601())
}
```

### Basic Place Parsing
```go
// Parse various place formats
places := []string{
    "Rapid City",
    "Rapid City, South Dakota",
    "Rapid City, Pennington, South Dakota, USA",
}

for _, placeStr := range places {
    place, err := ParsePlace(placeStr)
    if err != nil {
        continue
    }
    fmt.Printf("%s -> City: %s, State: %s\n", placeStr, place.City, place.State)
}
```

### Using with Records
```go
// Parse a GEDCOM file
parser := parser.NewHierarchicalParser()
tree, err := parser.Parse("family.ged")

// Get individual and parse dates/places
indi := tree.GetIndividual("@I1@")
if indi != nil {
    // Parse birth date
    birthDate, err := indi.GetBirthDateParsed()
    if err == nil {
        fmt.Printf("Born: %s\n", birthDate.ToISO8601())
        fmt.Printf("Year: %d, Month: %d, Day: %d\n", 
            birthDate.Year, birthDate.Month, birthDate.Day)
    }

    // Parse birth place
    birthPlace, err := indi.GetBirthPlaceParsed()
    if err == nil {
        fmt.Printf("Born in: %s, %s, %s\n", 
            birthPlace.City, birthPlace.State, birthPlace.Country)
    }
}
```

## Future Enhancements

Potential future improvements (not implemented):
- Calendar system conversion (Julian, Hebrew, etc.)
- Date validation (check for valid dates like Feb 30)
- Place geocoding (lookup latitude/longitude)
- Place name standardization
- Date range calculations
- Age calculations from dates

## Status

✅ **Complete and Tested**
- All date formats supported
- All place formats supported
- Comprehensive test coverage
- Integration with existing records
- Backward compatible
- Ready for production use

---

**Implementation Date:** 2025-01-27  
**Test Status:** All tests passing ✅
