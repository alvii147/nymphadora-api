package testkitinternal_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/stretchr/testify/require"
)

func TestMustCreateConfigSuccess(t *testing.T) {
	t.Parallel()

	testkitinternal.MustCreateConfig()
}

func TestMustCreateConfigInvalidPort(t *testing.T) {
	t.Setenv("NYMPHADORAAPI_PORT", "P0RT")

	require.Panics(t, func() {
		testkitinternal.MustCreateConfig()
	})
}
