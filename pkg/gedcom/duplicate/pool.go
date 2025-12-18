package duplicate

import (
	"sync"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// Memory pools for temporary data structures to reduce allocations.

var (
	// Pool for match slices
	matchSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]DuplicateMatch, 0, 32)
		},
	}

	// Pool for individual slices
	individualSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]*gedcom.IndividualRecord, 0, 64)
		},
	}

	// Pool for string slices (for matching fields, differences)
	stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 16)
		},
	}

	// Pool for comparison job slices
	jobSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]comparisonJob, 0, 128)
		},
	}
)

// getMatchSlice gets a match slice from the pool.
func getMatchSlice() []DuplicateMatch {
	return matchSlicePool.Get().([]DuplicateMatch)
}

// putMatchSlice returns a match slice to the pool.
func putMatchSlice(s []DuplicateMatch) {
	if s == nil {
		return
	}
	// Clear the slice but keep capacity
	s = s[:0]
	matchSlicePool.Put(s)
}

// getIndividualSlice gets an individual slice from the pool.
func getIndividualSlice() []*gedcom.IndividualRecord {
	return individualSlicePool.Get().([]*gedcom.IndividualRecord)
}

// putIndividualSlice returns an individual slice to the pool.
func putIndividualSlice(s []*gedcom.IndividualRecord) {
	if s == nil {
		return
	}
	// Clear the slice but keep capacity
	s = s[:0]
	individualSlicePool.Put(s)
}

// getStringSlice gets a string slice from the pool.
func getStringSlice() []string {
	return stringSlicePool.Get().([]string)
}

// putStringSlice returns a string slice to the pool.
func putStringSlice(s []string) {
	if s == nil {
		return
	}
	// Clear the slice but keep capacity
	s = s[:0]
	stringSlicePool.Put(s)
}

// getJobSlice gets a job slice from the pool.
func getJobSlice() []comparisonJob {
	return jobSlicePool.Get().([]comparisonJob)
}

// putJobSlice returns a job slice to the pool.
func putJobSlice(s []comparisonJob) {
	if s == nil {
		return
	}
	// Clear the slice but keep capacity
	s = s[:0]
	jobSlicePool.Put(s)
}
