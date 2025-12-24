package types

// DateNode represents a structured date node.
// Wraps GedcomDate and DateRange to provide a node-like interface.
type DateNode struct {
	// Date is the parsed GedcomDate
	Date *GedcomDate

	// DateRange is the parsed DateRange (for compatibility with elliotchance API)
	DateRange DateRange

	// Original is the original date string
	Original string

	// AlreadyParsed tracks if we've already parsed the date (for caching)
	alreadyParsed bool
}

// NewDateNode creates a new DateNode from a date string.
func NewDateNode(dateStr string) *DateNode {
	if dateStr == "" {
		return &DateNode{}
	}

	dn := &DateNode{
		Original: dateStr,
	}

	// Parse the date
	dn.parse()

	return dn
}

// NewDateNodeFromLine creates a DateNode from a GedcomLine (DATE tag).
func NewDateNodeFromLine(line *GedcomLine) *DateNode {
	if line == nil || line.Tag != "DATE" {
		return nil
	}

	return NewDateNode(line.Value)
}

// parse parses the date string into GedcomDate and DateRange.
func (dn *DateNode) parse() {
	if dn.alreadyParsed || dn.Original == "" {
		return
	}

	// Parse as GedcomDate
	date, err := ParseDate(dn.Original)
	if err == nil {
		dn.Date = date
	}

	// Also create DateRange for compatibility
	dn.DateRange = NewDateRangeWithString(dn.Original)

	dn.alreadyParsed = true
}

// IsValid returns true if the date was successfully parsed.
func (dn *DateNode) IsValid() bool {
	if dn == nil {
		return false
	}
	dn.parse()
	return dn.Date != nil && dn.Date.IsValid()
}

// StartDate returns the start date of the date range.
func (dn *DateNode) StartDate() *GedcomDate {
	if dn == nil || !dn.IsValid() {
		return nil
	}
	return dn.DateRange.StartDate()
}

// EndDate returns the end date of the date range.
func (dn *DateNode) EndDate() *GedcomDate {
	if dn == nil || !dn.IsValid() {
		return nil
	}
	return dn.DateRange.EndDate()
}

// StartAndEndDates returns both start and end dates.
func (dn *DateNode) StartAndEndDates() (*GedcomDate, *GedcomDate) {
	if dn == nil || !dn.IsValid() {
		return nil, nil
	}
	return dn.DateRange.StartAndEndDates()
}

// Years returns the years value of the date.
func (dn *DateNode) Years() float64 {
	if dn == nil || !dn.IsValid() {
		return 0
	}
	return dn.DateRange.Years()
}

// IsExact returns true if the date is exact (specific day).
func (dn *DateNode) IsExact() bool {
	if dn == nil || !dn.IsValid() {
		return false
	}
	return dn.DateRange.IsExact()
}

// Equals compares two date nodes for equality.
func (dn *DateNode) Equals(other *DateNode) bool {
	if dn == nil || other == nil {
		return dn == other
	}

	if !dn.IsValid() || !other.IsValid() {
		return false
	}

	return dn.DateRange.Equals(other.DateRange)
}

// Similarity returns the similarity between two date nodes.
func (dn *DateNode) Similarity(other *DateNode, maxYears float64) float64 {
	if dn == nil || other == nil {
		return 0.5
	}

	if !dn.IsValid() || !other.IsValid() {
		return 0.5
	}

	return dn.DateRange.Similarity(other.DateRange, maxYears)
}

// String returns the string representation of the date.
func (dn *DateNode) String() string {
	if dn == nil {
		return ""
	}
	dn.parse()
	if dn.Date != nil {
		return dn.Date.String()
	}
	return dn.Original
}

