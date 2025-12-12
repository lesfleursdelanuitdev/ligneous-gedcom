package gedcom

// MultimediaRecord represents a Multimedia Object (OBJE) record.
type MultimediaRecord struct {
	*BaseRecord
}

// NewMultimediaRecord creates a new MultimediaRecord from a GedcomLine.
func NewMultimediaRecord(line *GedcomLine) *MultimediaRecord {
	return &MultimediaRecord{
		BaseRecord: NewBaseRecord(line),
	}
}

// GetFile returns the file reference (FILE).
func (mr *MultimediaRecord) GetFile() string {
	return mr.GetValue("FILE")
}

// GetForm returns the media format (FORM).
func (mr *MultimediaRecord) GetForm() string {
	return mr.GetValue("FORM")
}

// GetTitle returns the media title (TITL).
func (mr *MultimediaRecord) GetTitle() string {
	return mr.GetValue("TITL")
}

