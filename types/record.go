package types

import (
	"crypto/rand"
	"fmt"
)

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
	UUID() string
	FirstLine() *GedcomLine
	GetValue(selector string) string
	GetValues(selector string) []string
	GetLines(selector string) []*GedcomLine
}

// BaseRecord provides a basic implementation of the Record interface.
type BaseRecord struct {
	firstLine  *GedcomLine
	recordType RecordType
	tree       *GedcomTree // Reference to the tree this record belongs to (set when added to tree)
	uuid       string      // System-generated UUID (v4 format)
}

// NewBaseRecord creates a new BaseRecord from a GedcomLine.
// A system-generated UUID is automatically assigned to the record.
func NewBaseRecord(line *GedcomLine) *BaseRecord {
	return &BaseRecord{
		firstLine:  line,
		recordType: RecordType(line.Tag),
		uuid:       generateUUID(),
	}
}

// generateUUID generates a UUID v4 (random UUID).
// Uses crypto/rand for cryptographically secure random number generation.
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	
	// Set version (4) and variant bits according to RFC 4122
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Type returns the record type.
func (br *BaseRecord) Type() RecordType {
	return br.recordType
}

// XrefID returns the cross-reference ID of the record.
func (br *BaseRecord) XrefID() string {
	return br.firstLine.XrefID
}

// UUID returns the system-generated UUID of the record.
func (br *BaseRecord) UUID() string {
	return br.uuid
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

// GetNotes returns all note xrefs.
// This method is shared by IndividualRecord and FamilyRecord.
func (br *BaseRecord) GetNotes() []string {
	return br.GetValues("NOTE")
}

// GetSources returns all source xrefs.
// This method is shared by IndividualRecord and FamilyRecord.
func (br *BaseRecord) GetSources() []string {
	return br.GetValues("SOUR")
}

// getTree returns the tree this record belongs to.
// Returns nil if the record hasn't been added to a tree yet.
func (br *BaseRecord) getTree() *GedcomTree {
	return br.tree
}

// setTree sets the tree reference for this record.
// This is called automatically when a record is added to a tree.
func (br *BaseRecord) setTree(tree *GedcomTree) {
	br.tree = tree
}

