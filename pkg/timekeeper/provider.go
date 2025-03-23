package timekeeper

import "time"

// Provider represents a component that provides the current timestamp.
type Provider interface {
	Now() time.Time
}
