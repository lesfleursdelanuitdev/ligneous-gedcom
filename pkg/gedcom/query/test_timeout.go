package query

import (
	"time"
)

// TestTimeout is the maximum duration for a test (2 minutes)
// This constant is used by testWithTimeout in hybrid_stress_test.go
const TestTimeout = 2 * time.Minute

