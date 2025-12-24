package types

import (
	"regexp"
	"testing"
)

// TestBaseRecord_UUID tests that UUIDs are generated and are valid UUID v4 format.
func TestBaseRecord_UUID(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	record := NewBaseRecord(line)

	uuid := record.UUID()
	if uuid == "" {
		t.Fatal("UUID should not be empty")
	}

	// Validate UUID v4 format: 8-4-4-4-12 hex digits
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidPattern.MatchString(uuid) {
		t.Errorf("UUID %q does not match UUID v4 format", uuid)
	}
}

// TestBaseRecord_UUID_Uniqueness tests that each record gets a unique UUID.
func TestBaseRecord_UUID_Uniqueness(t *testing.T) {
	line1 := NewGedcomLine(0, "INDI", "", "@I1@")
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	line3 := NewGedcomLine(0, "FAM", "", "@F1@")

	record1 := NewBaseRecord(line1)
	record2 := NewBaseRecord(line2)
	record3 := NewBaseRecord(line3)

	uuid1 := record1.UUID()
	uuid2 := record2.UUID()
	uuid3 := record3.UUID()

	// All UUIDs should be different
	if uuid1 == uuid2 {
		t.Error("Record1 and Record2 should have different UUIDs")
	}
	if uuid1 == uuid3 {
		t.Error("Record1 and Record3 should have different UUIDs")
	}
	if uuid2 == uuid3 {
		t.Error("Record2 and Record3 should have different UUIDs")
	}
}

// TestGedcomTree_UUIDIndex tests that records are indexed by UUID in the tree.
func TestGedcomTree_UUIDIndex(t *testing.T) {
	tree := NewGedcomTree()

	line1 := NewGedcomLine(0, "INDI", "", "@I1@")
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record1 := NewIndividualRecord(line1)
	record2 := NewIndividualRecord(line2)

	tree.AddRecord(record1)
	tree.AddRecord(record2)

	// Test lookup by UUID
	uuid1 := record1.UUID()
	uuid2 := record2.UUID()

	found1 := tree.GetRecordByUUID(uuid1)
	if found1 == nil {
		t.Fatal("Should find record1 by UUID")
	}
	if found1.UUID() != uuid1 {
		t.Errorf("Found record UUID %q, expected %q", found1.UUID(), uuid1)
	}

	found2 := tree.GetRecordByUUID(uuid2)
	if found2 == nil {
		t.Fatal("Should find record2 by UUID")
	}
	if found2.UUID() != uuid2 {
		t.Errorf("Found record UUID %q, expected %q", found2.UUID(), uuid2)
	}

	// Test non-existent UUID
	notFound := tree.GetRecordByUUID("00000000-0000-0000-0000-000000000000")
	if notFound != nil {
		t.Error("Should not find record for non-existent UUID")
	}
}

// TestRecord_UUID_Interface tests that UUID() is accessible via Record interface.
func TestRecord_UUID_Interface(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	record := NewBaseRecord(line)

	// Test via interface
	var r Record = record
	uuid := r.UUID()
	if uuid == "" {
		t.Fatal("UUID should not be empty when accessed via Record interface")
	}
	if uuid != record.UUID() {
		t.Error("UUID from interface should match UUID from concrete type")
	}
}

