package testkitinternal_test

import (
	"net/http/httptest"
	"testing"

	"github.com/alvii147/nymphadora-api/internal/server"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/stretchr/testify/require"
)

func TestMustCreateTestServerSuccess(t *testing.T) {
	t.Parallel()

	ctrl, srv := testkitinternal.MustCreateTestServer()
	t.Cleanup(ctrl.Close)
	t.Cleanup(srv.Close)
}

func TestMustCreateTestServerCtrlError(t *testing.T) {
	t.Setenv("NYMPHADORAAPI_PORT", "B33F")

	require.Panics(t, func() {
		testkitinternal.MustCreateTestServer()
	})
}

func TestMustCloseTestServer(t *testing.T) {
	t.Parallel()

	ctrl, err := server.NewController()
	require.NoError(t, err)

	srv := httptest.NewServer(ctrl)

	testkitinternal.MustCloseTestServer(ctrl, srv)
}
