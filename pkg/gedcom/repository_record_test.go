package gedcom

import "testing"

func TestRepositoryRecord_AllMethods(t *testing.T) {
	repoLine := NewGedcomLine(0, "REPO", "", "@R1@")
	
	nameLine := NewGedcomLine(1, "NAME", "Library Name", "")
	addr1 := NewGedcomLine(1, "ADDR", "123 Main St", "")
	addr2 := NewGedcomLine(2, "CONT", "City, State", "")
	addr1.AddChild(addr2)
	
	repoLine.AddChild(nameLine)
	repoLine.AddChild(addr1)

	repo := NewRepositoryRecord(repoLine)

	if repo.GetName() != "Library Name" {
		t.Errorf("Expected name 'Library Name', got %q", repo.GetName())
	}

	addresses := repo.GetAddress()
	if len(addresses) != 1 {
		t.Errorf("Expected 1 address, got %d", len(addresses))
	}
	if addresses[0] != "123 Main St" {
		t.Errorf("Expected address '123 Main St', got %q", addresses[0])
	}
}


