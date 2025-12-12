package gedcom

// RecordFactory creates specialized Record instances based on the record type.
type RecordFactory struct{}

// NewRecordFactory creates a new RecordFactory.
func NewRecordFactory() *RecordFactory {
	return &RecordFactory{}
}

// CreateRecord creates a specialized Record from a GedcomLine.
// Returns the appropriate record type based on the line's tag.
func (rf *RecordFactory) CreateRecord(line *GedcomLine) Record {
	if line == nil {
		return nil
	}

	switch RecordType(line.Tag) {
	case RecordTypeINDI:
		return NewIndividualRecord(line)
	case RecordTypeFAM:
		return NewFamilyRecord(line)
	case RecordTypeHEAD:
		return NewHeaderRecord(line)
	case RecordTypeNOTE:
		return NewNoteRecord(line)
	case RecordTypeSOUR:
		return NewSourceRecord(line)
	case RecordTypeREPO:
		return NewRepositoryRecord(line)
	case RecordTypeSUBM:
		return NewSubmitterRecord(line)
	case RecordTypeOBJE:
		return NewMultimediaRecord(line)
	case RecordTypeTRLR:
		// TRLR doesn't need special handling, use BaseRecord
		return NewBaseRecord(line)
	default:
		// Unknown record type, use BaseRecord
		return NewBaseRecord(line)
	}
}

