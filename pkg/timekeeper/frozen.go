package timekeeper

import "time"

// frozenProvider implements Provider and returns a frozen timestamp.
type frozenProvider struct {
	time time.Time
}

// NewFrozenProvider returns a new frozenProvider.
func NewFrozenProvider() *frozenProvider {
	return &frozenProvider{
		time: time.Now().UTC(),
	}
}

// Now provides the current time.
func (provider *frozenProvider) Now() time.Time {
	return provider.time
}

// SetTime sets the current time.
func (provider *frozenProvider) SetTime(t time.Time) {
	provider.time = t
}

// Add advances the current time by the given duration.
func (provider *frozenProvider) Add(d time.Duration) {
	provider.time = provider.time.Add(d)
}

// Add advances the current time by the given number of days, month, and years.
func (provider *frozenProvider) AddDate(years int, months int, days int) {
	provider.time = provider.time.AddDate(years, months, days)
}
