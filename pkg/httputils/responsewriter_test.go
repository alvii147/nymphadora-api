package httputils_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/stretchr/testify/require"
)

type mockResponseWriter struct {
	headers      map[string][]string
	writtenBytes []byte
	writeErr     error
	statusCode   int
}

func (w *mockResponseWriter) Header() http.Header {
	return w.headers
}

func (w *mockResponseWriter) Write(p []byte) (int, error) {
	w.writtenBytes = append(w.writtenBytes, p...)

	return len(p), w.writeErr
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestResponseWriterHeader(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	w := httputils.NewResponseWriter(rec)

	w.Header().Set("Content-Type", "application/json")
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestResponseWriterWriteSuccess(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	w := httputils.NewResponseWriter(rec)

	data := "DEADBEEF"

	_, err := w.Write([]byte(data))
	require.NoError(t, err)
	require.Equal(t, data, rec.Body.String())
}

func TestResponseWriterWriteError(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("Write failed")
	w := httputils.NewResponseWriter(&mockResponseWriter{
		headers:  map[string][]string{},
		writeErr: writeErr,
	})

	_, err := w.Write([]byte("DEADBEEF"))
	require.ErrorIs(t, err, writeErr)
}

func TestResponseWriterWriteHeader(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		statusCode int
	}{
		"Status code OK": {
			statusCode: http.StatusOK,
		},
		"Status code created": {
			statusCode: http.StatusCreated,
		},
		"Status code moved permanently": {
			statusCode: http.StatusMovedPermanently,
		},
		"Status code found": {
			statusCode: http.StatusFound,
		},
		"Status code not found": {
			statusCode: http.StatusNotFound,
		},
		"Status code internal server error": {
			statusCode: http.StatusInternalServerError,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			w := httputils.NewResponseWriter(rec)

			w.WriteHeader(testcase.statusCode)
			require.Equal(t, testcase.statusCode, w.StatusCode)
			require.Equal(t, testcase.statusCode, rec.Code)
		})
	}
}

func TestResponseWriterWriteJSONWithoutData(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	w := httputils.NewResponseWriter(rec)

	w.WriteJSON(nil, http.StatusOK)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestResponseWriterWriteJSONWithData(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	w := httputils.NewResponseWriter(rec)

	data := map[string]any{
		"number": float64(42),
		"string": "Hello",
		"null":   nil,
		"listOfNumbers": []any{
			float64(3),
			float64(1),
			float64(4),
			float64(1),
			float64(6),
		},
	}

	w.WriteJSON(data, http.StatusOK)
	require.Equal(t, http.StatusOK, rec.Code)

	writtenData := make(map[string]any)
	err := json.NewDecoder(rec.Body).Decode(&writtenData)
	require.NoError(t, err)
	require.Equal(t, data["number"], writtenData["number"])
	require.Equal(t, data["string"], writtenData["string"])
	require.Equal(t, data["null"], writtenData["null"])
	require.Equal(t, data["listOfNumbers"], writtenData["listOfNumbers"])
}

func TestResponseWriterWriteJSONError(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("Write failed")
	w := httputils.NewResponseWriter(&mockResponseWriter{
		headers:  map[string][]string{},
		writeErr: writeErr,
	})

	w.WriteJSON(map[string]any{}, http.StatusOK)
	require.Equal(t, http.StatusInternalServerError, w.StatusCode)
}

func TestResponseWriterMiddleware(t *testing.T) {
	t.Parallel()

	nextCallCount := 0
	var next httputils.HandlerFunc = func(w *httputils.ResponseWriter, r *http.Request) {
		nextCallCount++
	}

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/request/url/path", http.NoBody)

	httputils.ResponseWriterMiddleware(next).ServeHTTP(rec, r)
	require.Equal(t, 1, nextCallCount)
}
