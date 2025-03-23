package timekeeper

import "time"

// systemProvider implements Provider and returns the current system timestamp.
type systemProvider struct{}

// NewSystemProvider returns a new systemProvider.
func NewSystemProvider() *systemProvider {
	return &systemProvider{}
}

// Now provides the current time.
func (provider *systemProvider) Now() time.Time {
	return time.Now().UTC()
}
