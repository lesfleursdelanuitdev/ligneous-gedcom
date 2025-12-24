package types

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Duration represents a duration that only considers whole-day resolution.
type Duration struct {
	Duration time.Duration

	// IsEstimate indicates if the duration is an estimate
	IsEstimate bool

	// IsKnown indicates if the duration is known (not estimated)
	IsKnown bool
}

// NewExactDuration creates a new exact duration.
func NewExactDuration(duration time.Duration) Duration {
	return NewDuration(duration, true, false)
}

// NewDuration creates a new duration with the specified properties.
// Durations must always be positive.
func NewDuration(duration time.Duration, isKnown, isEstimate bool) Duration {
	// Durations must always be positive
	if duration < 0 {
		duration = -duration
	}

	return Duration{
		Duration:   duration,
		IsEstimate: isEstimate,
		IsKnown:    isKnown,
	}
}

// pluralize returns a pluralized string for the given value and word.
func pluralize(value int, word string) string {
	switch value {
	case 0:
		return ""
	case 1:
		return "one " + word
	default:
		return fmt.Sprintf("%d %ss", value, word)
	}
}

// String returns a human-readable string representation of the duration.
func (d Duration) String() string {
	oneDay := time.Duration(24 * time.Hour)
	oneMonth := time.Duration(30.4166 * float64(oneDay))
	oneYear := time.Duration(365 * float64(oneDay))

	if d.Duration < oneDay {
		return "one day"
	}

	var parts []string

	if years := int(d.Duration / oneYear); years != 0 {
		d.Duration -= time.Duration(years) * oneYear
		parts = append(parts, pluralize(years, "year"))
	}

	if months := int(d.Duration / oneMonth); months != 0 {
		d.Duration -= time.Duration(months) * oneMonth
		parts = append(parts, pluralize(months, "month"))
	}

	if days := int(math.Ceil(float64(d.Duration) / float64(oneDay))); days != 0 {
		parts = append(parts, pluralize(days, "day"))
	}

	return strings.Join(parts, " and ")
}

