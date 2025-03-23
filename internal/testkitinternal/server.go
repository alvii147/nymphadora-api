package testkitinternal

import (
	"net/http/httptest"

	"github.com/alvii147/nymphadora-api/internal/server"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
)

// MustCreateTestServer creates a new Controller, sets up a new test HTTP server and panics on error.
func MustCreateTestServer() (*server.Controller, *httptest.Server) {
	ctrl, err := server.NewController()
	if err != nil {
		panic(errutils.FormatError(err))
	}

	srv := httptest.NewServer(ctrl)

	return ctrl, srv
}

// MustCloseTestServer closes a Controller and an HTTP server.
func MustCloseTestServer(ctrl *server.Controller, srv *httptest.Server) {
	ctrl.Close()
	srv.Close()
}
