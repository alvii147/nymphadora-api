package timekeeper_test

import (
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

func TestNewSystemProvider(t *testing.T) {
	t.Parallel()

	provider := timekeeper.NewSystemProvider()
	require.WithinDuration(t, time.Now().UTC(), provider.Now(), testkit.TimeToleranceTentative)
}
