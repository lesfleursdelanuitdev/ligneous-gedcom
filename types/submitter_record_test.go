package types

import "testing"

func TestSubmitterRecord_AllMethods(t *testing.T) {
	submLine := NewGedcomLine(0, "SUBM", "", "@U1@")
	
	nameLine := NewGedcomLine(1, "NAME", "John Doe", "")
	addrLine := NewGedcomLine(1, "ADDR", "123 Main St", "")
	phoneLine := NewGedcomLine(1, "PHON", "555-1234", "")
	
	submLine.AddChild(nameLine)
	submLine.AddChild(addrLine)
	submLine.AddChild(phoneLine)

	subm := NewSubmitterRecord(submLine)

	if subm.GetName() != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %q", subm.GetName())
	}

	addresses := subm.GetAddress()
	if len(addresses) != 1 {
		t.Errorf("Expected 1 address, got %d", len(addresses))
	}
	if addresses[0] != "123 Main St" {
		t.Errorf("Expected address '123 Main St', got %q", addresses[0])
	}

	if subm.GetPhone() != "555-1234" {
		t.Errorf("Expected phone '555-1234', got %q", subm.GetPhone())
	}
}


