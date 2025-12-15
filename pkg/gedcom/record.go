package gedcom

// RecordType represents the type of a GEDCOM record.
// Valid types are defined as constants below.
type RecordType string

const (
	RecordTypeHEAD RecordType = "HEAD"
	RecordTypeINDI RecordType = "INDI"
	RecordTypeFAM  RecordType = "FAM"
	RecordTypeNOTE RecordType = "NOTE"
	RecordTypeSOUR RecordType = "SOUR"
	RecordTypeREPO RecordType = "REPO"
	RecordTypeSUBM RecordType = "SUBM"
	RecordTypeOBJE RecordType = "OBJE"
	RecordTypeTRLR RecordType = "TRLR"
)

// Record represents a GEDCOM record (INDI, FAM, NOTE, etc.).
type Record interface {
	Type() RecordType
	XrefID() string
	FirstLine() *GedcomLine
	GetValue(selector string) string
	GetValues(selector string) []string
	GetLines(selector string) []*GedcomLine
}

// BaseRecord provides a basic implementation of the Record interface.
type BaseRecord struct {
	firstLine  *GedcomLine
	recordType RecordType
}

// NewBaseRecord creates a new BaseRecord from a GedcomLine.
func NewBaseRecord(line *GedcomLine) *BaseRecord {
	return &BaseRecord{
		firstLine:  line,
		recordType: RecordType(line.Tag),
	}
}

// Type returns the record type.
func (br *BaseRecord) Type() RecordType {
	return br.recordType
}

// XrefID returns the cross-reference ID of the record.
func (br *BaseRecord) XrefID() string {
	return br.firstLine.XrefID
}

// FirstLine returns the first line (level 0) of the record.
func (br *BaseRecord) FirstLine() *GedcomLine {
	return br.firstLine
}

// GetValue retrieves a value using dot notation selector.
func (br *BaseRecord) GetValue(selector string) string {
	return br.firstLine.GetValue(selector)
}

// GetValues retrieves all values matching the selector.
func (br *BaseRecord) GetValues(selector string) []string {
	lines := br.firstLine.GetLines(selector)
	values := make([]string, 0, len(lines))
	for _, line := range lines {
		if line.Value != "" {
			values = append(values, line.Value)
		}
	}
	return values
}

// GetLines retrieves all lines matching the selector.
func (br *BaseRecord) GetLines(selector string) []*GedcomLine {
	return br.firstLine.GetLines(selector)
}

