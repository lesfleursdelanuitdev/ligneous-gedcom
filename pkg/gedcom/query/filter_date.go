package query

import (
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// ByBirthDate filters by birth date range.
// Uses index for fast lookup.
func (fq *FilterQuery) ByBirthDate(start, end time.Time) *FilterQuery {
	fq.birthDateStart = &start
	fq.birthDateEnd = &end
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		birthDate, err := indi.GetBirthDateParsed()
		if err != nil || birthDate == nil || !birthDate.IsValid() {
			return false
		}

		birthTime := birthDate.Earliest()
		return !birthTime.Before(start) && !birthTime.After(end)
	})
}

// ByBirthDateBefore filters individuals born before the specified year.
func (fq *FilterQuery) ByBirthDateBefore(year int) *FilterQuery {
	start := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)
	return fq.ByBirthDate(start, end)
}

// ByBirthDateAfter filters individuals born after the specified year.
func (fq *FilterQuery) ByBirthDateAfter(year int) *FilterQuery {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	return fq.ByBirthDate(start, end)
}

// ByBirthYear filters individuals born in the specified year.
func (fq *FilterQuery) ByBirthYear(year int) *FilterQuery {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)
	return fq.ByBirthDate(start, end)
}

