package server_test

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/stretchr/testify/require"
)

var TestServerURL = ""

func TestMain(m *testing.M) {
	ctrl, srv := testkitinternal.MustCreateTestServer()
	TestServerURL = srv.URL

	code := m.Run()

	testkitinternal.MustCloseTestServer(ctrl, srv)
	os.Exit(code)
}

func TestHandlePing(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	req, err := http.NewRequest(http.MethodGet, TestServerURL+"/ping", http.NoBody)
	require.NoError(t, err)

	res, err := httpClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := res.Body.Close()
		require.NoError(t, err)
	})

	require.Equal(t, http.StatusOK, res.StatusCode)

	resp, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "pong", string(resp))
}
