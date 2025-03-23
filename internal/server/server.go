package server

import (
	"net/http"

	"github.com/alvii147/nymphadora-api/pkg/httputils"
)

// HandlePing handles server ping requests.
// Methods: GET
// URL: /ping.
func (ctrl *Controller) HandlePing(w *httputils.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("pong"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
