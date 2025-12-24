package duplicate

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// calculateRelationshipSimilarity calculates similarity based on family relationships.
// Returns a score from 0.0 to 1.0 based on common parents, spouses, and children.
func (dd *DuplicateDetector) calculateRelationshipSimilarity(indi1, indi2 *types.IndividualRecord) float64 {
	if dd.tree == nil {
		return 0.0
	}

	score := 0.0
	matches := 0

	// Check for common parents (FAMC)
	parents1 := getParents(indi1, dd.tree)
	parents2 := getParents(indi2, dd.tree)
	commonParents := countCommonXrefs(parents1, parents2)
	if commonParents > 0 {
		// Same parents is a strong indicator
		if commonParents >= 2 {
			score += 0.2 // Both parents match
			matches++
		} else if commonParents == 1 {
			score += 0.1 // One parent matches
			matches++
		}
	}

	// Check for common spouses (FAMS)
	spouses1 := getSpouses(indi1, dd.tree)
	spouses2 := getSpouses(indi2, dd.tree)
	commonSpouses := countCommonXrefs(spouses1, spouses2)
	if commonSpouses > 0 {
		score += 0.2 // Same spouse is a strong indicator
		matches++
	}

	// Check for common children
	children1 := getChildren(indi1, dd.tree)
	children2 := getChildren(indi2, dd.tree)
	commonChildren := countCommonXrefs(children1, children2)
	if commonChildren > 0 {
		// Bonus for each common child (max 0.3)
		childBonus := float64(commonChildren) * 0.1
		if childBonus > 0.3 {
			childBonus = 0.3
		}
		score += childBonus
		matches++
	}

	// Normalize score to 0.0-1.0 range
	// Maximum possible score is 0.7 (0.2 parents + 0.2 spouse + 0.3 children)
	// Normalize to 0.0-1.0 range
	if score > 0.0 {
		// Scale to 0.0-1.0, but cap at 1.0
		normalizedScore := score / 0.7
		if normalizedScore > 1.0 {
			normalizedScore = 1.0
		}
		return normalizedScore
	}

	return 0.0
}

// getParents returns the xref IDs of an individual's parents.
func getParents(indi *types.IndividualRecord, tree *types.GedcomTree) []string {
	parents := make([]string, 0)
	famcXrefs := indi.GetFamiliesAsChild()

	for _, famcXref := range famcXrefs {
		famRecord := tree.GetFamily(famcXref)
		if famRecord == nil {
			continue
		}

		fam, ok := famRecord.(*types.FamilyRecord)
		if !ok {
			continue
		}

		husband := fam.GetHusband()
		if husband != "" {
			parents = append(parents, husband)
		}

		wife := fam.GetWife()
		if wife != "" {
			parents = append(parents, wife)
		}
	}

	return parents
}

// getSpouses returns the xref IDs of an individual's spouses.
func getSpouses(indi *types.IndividualRecord, tree *types.GedcomTree) []string {
	spouses := make([]string, 0)
	famsXrefs := indi.GetFamiliesAsSpouse()

	for _, famsXref := range famsXrefs {
		famRecord := tree.GetFamily(famsXref)
		if famRecord == nil {
			continue
		}

		fam, ok := famRecord.(*types.FamilyRecord)
		if !ok {
			continue
		}

		husband := fam.GetHusband()
		wife := fam.GetWife()

		// Get the other spouse (if this individual is in the family)
		indiXref := indi.XrefID()
		if husband == indiXref && wife != "" && wife != indiXref {
			spouses = append(spouses, wife)
		} else if wife == indiXref && husband != "" && husband != indiXref {
			spouses = append(spouses, husband)
		}
	}

	return spouses
}

// getChildren returns the xref IDs of an individual's children.
func getChildren(indi *types.IndividualRecord, tree *types.GedcomTree) []string {
	children := make([]string, 0)
	famsXrefs := indi.GetFamiliesAsSpouse()

	for _, famsXref := range famsXrefs {
		famRecord := tree.GetFamily(famsXref)
		if famRecord == nil {
			continue
		}

		fam, ok := famRecord.(*types.FamilyRecord)
		if !ok {
			continue
		}

		children = append(children, fam.GetChildren()...)
	}

	return children
}

// countCommonXrefs counts the number of common xref IDs in two slices.
func countCommonXrefs(list1, list2 []string) int {
	if len(list1) == 0 || len(list2) == 0 {
		return 0
	}

	// Create a map for faster lookup
	set1 := make(map[string]bool)
	for _, xref := range list1 {
		set1[xref] = true
	}

	count := 0
	for _, xref := range list2 {
		if set1[xref] {
			count++
		}
	}

	return count
}
