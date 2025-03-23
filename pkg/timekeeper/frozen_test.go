package timekeeper_test

import (
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

func TestNewFrozenProvider(t *testing.T) {
	t.Parallel()

	provider := timekeeper.NewFrozenProvider()
	require.WithinDuration(t, time.Now().UTC(), provider.Now(), testkit.TimeToleranceTentative)
}

func TestFrozenProviderSetTime(t *testing.T) {
	t.Parallel()

	provider := timekeeper.NewFrozenProvider()
	currentTime := time.Date(2018, 9, 5, 10, 32, 12, 0, time.UTC)
	provider.SetTime(currentTime)
	require.Equal(t, currentTime, provider.Now())
}

func TestFrozenProviderAdd(t *testing.T) {
	t.Parallel()

	provider := timekeeper.NewFrozenProvider()
	currentTime := provider.Now()
	provider.Add(24 * time.Hour)
	require.Equal(t, currentTime.Add(24*time.Hour), provider.Now())
}

func TestFrozenProviderAddDate(t *testing.T) {
	t.Parallel()

	provider := timekeeper.NewFrozenProvider()
	currentTime := provider.Now()
	provider.AddDate(3, 1, 4)
	require.Equal(t, currentTime.AddDate(3, 1, 4), provider.Now())
}
