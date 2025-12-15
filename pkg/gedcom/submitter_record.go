package gedcom

// SubmitterRecord represents a Submitter (SUBM) record.
type SubmitterRecord struct {
	*BaseRecord
}

// NewSubmitterRecord creates a new SubmitterRecord from a GedcomLine.
func NewSubmitterRecord(line *GedcomLine) *SubmitterRecord {
	return &SubmitterRecord{
		BaseRecord: NewBaseRecord(line),
	}
}

// GetName returns the submitter name (NAME).
func (sr *SubmitterRecord) GetName() string {
	return sr.GetValue("NAME")
}

// GetAddress returns the address lines (ADDR).
func (sr *SubmitterRecord) GetAddress() []string {
	return sr.GetValues("ADDR")
}

// GetPhone returns the phone number (PHON).
func (sr *SubmitterRecord) GetPhone() string {
	return sr.GetValue("PHON")
}



