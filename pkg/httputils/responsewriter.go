package httputils

import (
	"encoding/json"
	"net/http"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
)

// ResponseWriter stores an http.ResponseWriter and the HTTP status code.
// This is used for retaining the status code after the handler is executed.
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// NewResponseWriter returns a new ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

// Header returns the headers in ResponseWriter.
func (w *ResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Header writes bytes data to ResponseWriter.
func (w *ResponseWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	if err != nil {
		return 0, errutils.FormatErrorf(err, "ResponseWriter.Write failed")
	}

	return n, nil
}

// Header writes status code to ResponseWriter.
func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Header writes status code and JSON data to ResponseWriter.
func (w *ResponseWriter) WriteJSON(data any, statusCode int) {
	w.WriteHeader(statusCode)
	if data == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ResponseWriterMiddleware converts a HandlerFunc to an http.Handler.
// This should be the top-level middleware when setting up routes.
func ResponseWriterMiddleware(next HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)
	})
}
