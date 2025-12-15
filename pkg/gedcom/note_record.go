package gedcom

// NoteRecord represents a Note (NOTE) record.
type NoteRecord struct {
	*BaseRecord
}

// NewNoteRecord creates a new NoteRecord from a GedcomLine.
func NewNoteRecord(line *GedcomLine) *NoteRecord {
	return &NoteRecord{
		BaseRecord: NewBaseRecord(line),
	}
}

// GetText returns the note text (value of the NOTE line or CONT/CONC lines).
func (nr *NoteRecord) GetText() string {
	return nr.GetValue("")
}



