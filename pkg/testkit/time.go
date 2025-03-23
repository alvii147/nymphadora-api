package testkit

import "time"

const (
	// TimeToleranceExact is the tolerance allowed in time comparison in tests
	// where the sources of the times are the same.
	TimeToleranceExact = time.Second
	// TimeToleranceTentative is the tolerance allowed in time comparison in tests
	// where the sources of the times are different.
	TimeToleranceTentative = time.Minute
)
