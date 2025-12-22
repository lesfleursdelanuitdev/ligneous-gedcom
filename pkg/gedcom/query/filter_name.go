package query

import (
	"strings"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// ByName filters by name (case-insensitive substring match).
// Uses index for fast lookup.
func (fq *FilterQuery) ByName(pattern string) *FilterQuery {
	fq.nameFilter = pattern
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		name := strings.ToLower(indi.GetName())
		return strings.Contains(name, strings.ToLower(pattern))
	})
}

// ByNameExact filters by exact name match (case-insensitive).
func (fq *FilterQuery) ByNameExact(name string) *FilterQuery {
	fq.nameExactFilter = name
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return strings.EqualFold(indi.GetName(), name)
	})
}

// ByNameStarts filters by name starting with prefix (case-insensitive).
func (fq *FilterQuery) ByNameStarts(prefix string) *FilterQuery {
	fq.nameStartsFilter = prefix
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		name := strings.ToLower(indi.GetName())
		return strings.HasPrefix(name, strings.ToLower(prefix))
	})
}

// ByNameEnds filters by name ending with suffix (case-insensitive).
func (fq *FilterQuery) ByNameEnds(suffix string) *FilterQuery {
	fq.nameEndsFilter = suffix
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		name := strings.ToLower(indi.GetName())
		return strings.HasSuffix(name, strings.ToLower(suffix))
	})
}

