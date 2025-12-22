package query

import (
	"strings"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// BySex filters by sex.
// Uses index for fast lookup.
func (fq *FilterQuery) BySex(sex string) *FilterQuery {
	fq.sexFilter = sex
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return strings.ToUpper(indi.GetSex()) == strings.ToUpper(sex)
	})
}

// ByBirthPlace filters by birth place (case-insensitive substring match).
// Uses index for fast lookup.
func (fq *FilterQuery) ByBirthPlace(place string) *FilterQuery {
	fq.birthPlaceFilter = place
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		birthPlace := strings.ToLower(indi.GetBirthPlace())
		return strings.Contains(birthPlace, strings.ToLower(place))
	})
}

// HasChildren filters individuals with children.
// Uses index for fast lookup.
func (fq *FilterQuery) HasChildren() *FilterQuery {
	hasChildren := true
	fq.hasChildrenFilter = &hasChildren
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return fq.graph.indexes.hasChildren(indi.XrefID())
	})
}

// NoChildren filters individuals without children.
// Uses index for fast lookup.
func (fq *FilterQuery) NoChildren() *FilterQuery {
	hasChildren := false
	fq.hasChildrenFilter = &hasChildren
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return !fq.graph.indexes.hasChildren(indi.XrefID())
	})
}

// HasSpouse filters individuals with spouses.
// Uses index for fast lookup.
func (fq *FilterQuery) HasSpouse() *FilterQuery {
	hasSpouse := true
	fq.hasSpouseFilter = &hasSpouse
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return fq.graph.indexes.hasSpouse(indi.XrefID())
	})
}

// NoSpouse filters individuals without spouses.
// Uses index for fast lookup.
func (fq *FilterQuery) NoSpouse() *FilterQuery {
	hasSpouse := false
	fq.hasSpouseFilter = &hasSpouse
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return !fq.graph.indexes.hasSpouse(indi.XrefID())
	})
}

// Living filters living individuals (no death date).
// Uses index for fast lookup.
func (fq *FilterQuery) Living() *FilterQuery {
	living := true
	fq.livingFilter = &living
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return fq.graph.indexes.isLiving(indi.XrefID())
	})
}

// Deceased filters deceased individuals (has death date).
func (fq *FilterQuery) Deceased() *FilterQuery {
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		deathDate := indi.GetDeathDate()
		return deathDate != ""
	})
}
