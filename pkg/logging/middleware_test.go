package logging_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/alvii147/nymphadora-api/pkg/logging"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

func TestLogTraffic(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		statusCode int
		method     string
		wantLevel  string
		wantStderr bool
	}{
		{
			name:       "200 status code with GET request causes info level log",
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			wantLevel:  "I",
			wantStderr: false,
		},
		{
			name:       "200 status code with POST request causes info level log",
			statusCode: http.StatusOK,
			method:     http.MethodPost,
			wantLevel:  "I",
			wantStderr: false,
		},
		{
			name:       "201 status code causes info level log",
			statusCode: http.StatusCreated,
			method:     http.MethodGet,
			wantLevel:  "I",
			wantStderr: false,
		},
		{
			name:       "302 status code causes info level log",
			statusCode: http.StatusFound,
			method:     http.MethodGet,
			wantLevel:  "I",
			wantStderr: false,
		},
		{
			name:       "400 status code causes warn level log",
			statusCode: http.StatusBadRequest,
			method:     http.MethodGet,
			wantLevel:  "W",
			wantStderr: false,
		},
		{
			name:       "404 status code causes warn level log",
			statusCode: http.StatusNotFound,
			method:     http.MethodGet,
			wantLevel:  "W",
			wantStderr: false,
		},
		{
			name:       "500 status code causes error level log",
			statusCode: http.StatusInternalServerError,
			method:     http.MethodGet,
			wantLevel:  "E",
			wantStderr: true,
		},
		{
			name:       "501 status code causes error level log",
			statusCode: http.StatusNotImplemented,
			method:     http.MethodGet,
			wantLevel:  "E",
			wantStderr: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			bufOut, bufErr, logger := testkit.CreateInMemLogger()

			url := "/request/url/path"
			rec := httptest.NewRecorder()
			w := &httputils.ResponseWriter{
				ResponseWriter: rec,
				StatusCode:     testcase.statusCode,
			}
			r := httptest.NewRequest(testcase.method, url, http.NoBody)

			timeProvider := timekeeper.NewFrozenProvider()
			logging.LogTraffic(logger, w, r)

			stdoutMessages := strings.Split(strings.TrimSpace(bufOut.String()), "\n")
			stderrMessages := strings.Split(strings.TrimSpace(bufErr.String()), "\n")

			require.Len(t, stdoutMessages, 1)
			require.Len(t, stderrMessages, 1)

			var message string
			if testcase.wantStderr {
				message = stderrMessages[0]
			} else {
				message = stdoutMessages[0]
			}

			msgSplits := strings.Fields(message)
			require.Greater(t, len(msgSplits), 8)
			require.Equal(t, "["+testcase.wantLevel+"]", msgSplits[0])
			require.Contains(t, msgSplits[3], "pkg/logging/middleware.go")
			require.Equal(t, testcase.method, msgSplits[4])
			require.Equal(t, url, msgSplits[5])
			require.Equal(t, strconv.Itoa(testcase.statusCode), msgSplits[7])
			require.Equal(t, http.StatusText(testcase.statusCode), strings.Join(msgSplits[8:], " "))

			logTime, err := time.ParseInLocation("2006/01/02 15:04:05", msgSplits[1]+" "+msgSplits[2], time.UTC)
			require.NoError(t, err)
			require.WithinDuration(t, logTime, timeProvider.Now(), testkit.TimeToleranceTentative)
		})
	}
}

func TestLoggerMiddleware(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		statusCode int
		method     string
		wantLevel  string
		wantStderr bool
	}{
		{
			name:       "200 status code with GET request causes info level log",
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			wantLevel:  "I",
			wantStderr: false,
		},
		{
			name:       "200 status code with POST request causes info level log",
			statusCode: http.StatusOK,
			method:     http.MethodPost,
			wantLevel:  "I",
			wantStderr: false,
		},
		{
			name:       "201 status code causes info level log",
			statusCode: http.StatusCreated,
			method:     http.MethodGet,
			wantLevel:  "I",
			wantStderr: false,
		},
		{
			name:       "302 status code causes info level log",
			statusCode: http.StatusFound,
			method:     http.MethodGet,
			wantLevel:  "I",
			wantStderr: false,
		},
		{
			name:       "400 status code causes warn level log",
			statusCode: http.StatusBadRequest,
			method:     http.MethodGet,
			wantLevel:  "W",
			wantStderr: false,
		},
		{
			name:       "404 status code causes warn level log",
			statusCode: http.StatusNotFound,
			method:     http.MethodGet,
			wantLevel:  "W",
			wantStderr: false,
		},
		{
			name:       "500 status code causes error level log",
			statusCode: http.StatusInternalServerError,
			method:     http.MethodGet,
			wantLevel:  "E",
			wantStderr: true,
		},
		{
			name:       "501 status code causes error level log",
			statusCode: http.StatusNotImplemented,
			method:     http.MethodGet,
			wantLevel:  "E",
			wantStderr: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			nextCallCount := 0
			var next httputils.HandlerFunc = func(w *httputils.ResponseWriter, r *http.Request) {
				nextCallCount++
			}

			bufOut, bufErr, logger := testkit.CreateInMemLogger()

			url := "/request/url/path"
			rec := httptest.NewRecorder()
			w := &httputils.ResponseWriter{
				ResponseWriter: rec,
				StatusCode:     testcase.statusCode,
			}
			r := httptest.NewRequest(testcase.method, url, http.NoBody)

			timeProvider := timekeeper.NewFrozenProvider()
			logging.NewLoggerMiddleware(logger)(next)(w, r)
			require.Equal(t, 1, nextCallCount)

			stdoutMessages := strings.Split(strings.TrimSpace(bufOut.String()), "\n")
			stderrMessages := strings.Split(strings.TrimSpace(bufErr.String()), "\n")

			require.Len(t, stdoutMessages, 1)
			require.Len(t, stderrMessages, 1)

			var message string
			if testcase.wantStderr {
				message = stderrMessages[0]
			} else {
				message = stdoutMessages[0]
			}

			logLevel, logTime, logFile, logMsg := testkit.MustParseLogMessage(message)
			require.Equal(t, testcase.wantLevel, logLevel)
			require.WithinDuration(t, logTime, timeProvider.Now(), testkit.TimeToleranceTentative)
			require.Contains(t, logFile, "pkg/logging/middleware.go")

			pattern := `^` +
				testcase.method +
				`\s+` +
				url +
				`\s+HTTP/1.1\s+` +
				strconv.Itoa(testcase.statusCode) +
				`\s+` +
				http.StatusText(testcase.statusCode) +
				`$`
			require.Regexp(t, pattern, logMsg)
		})
	}
}
