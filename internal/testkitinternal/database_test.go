package testkitinternal_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/stretchr/testify/require"
)

func TestMustNewDatabasePoolSuccess(t *testing.T) {
	t.Parallel()

	dbPool := testkitinternal.MustNewDatabasePool()
	defer dbPool.Close()
}

func TestMustNewDatabasePoolInvalidConfig(t *testing.T) {
	t.Setenv("NYMPHADORAAPI_PORT", "P0RT")

	require.Panics(t, func() {
		testkitinternal.MustNewDatabasePool()
	})
}
