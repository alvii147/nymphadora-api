package httputils_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/stretchr/testify/require"
)

func TestGetAuthorizationHeader(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		header    http.Header
		authType  string
		wantToken string
		wantOk    bool
	}{
		"Valid header with valid auth type": {
			header: map[string][]string{
				"Authorization": {"Bearer 0xdeadbeef"},
			},
			authType:  "Bearer",
			wantToken: "0xdeadbeef",
			wantOk:    true,
		},
		"No header": {
			header:    map[string][]string{},
			authType:  "Bearer",
			wantToken: "0xdeadbeef",
			wantOk:    false,
		},
		"Invalid auth type": {
			header: map[string][]string{
				"Authorization": {"Bearer 0xdeadbeef"},
			},
			authType:  "Basic",
			wantToken: "0xdeadbeef",
			wantOk:    false,
		},
		"Valid header with spaces": {
			header: map[string][]string{
				"Authorization": {"  Bearer   0xdeadbeef    "},
			},
			authType:  "Bearer",
			wantToken: "0xdeadbeef",
			wantOk:    true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			token, ok := httputils.GetAuthorizationHeader(testcase.header, testcase.authType)
			require.Equal(t, testcase.wantOk, ok)
			if testcase.wantOk {
				require.Equal(t, testcase.wantToken, token)
			}
		})
	}
}

func TestIsHTTPSuccess(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		statusCode  int
		wantSuccess bool
	}{
		"200 OK is success": {
			statusCode:  http.StatusOK,
			wantSuccess: true,
		},
		"201 Created is success": {
			statusCode:  http.StatusCreated,
			wantSuccess: true,
		},
		"204 No content is success": {
			statusCode:  http.StatusNoContent,
			wantSuccess: true,
		},
		"302 Found is not success": {
			statusCode:  http.StatusFound,
			wantSuccess: false,
		},
		"400 Bad request is not success": {
			statusCode:  http.StatusBadRequest,
			wantSuccess: false,
		},
		"401 Unauthorized is not success": {
			statusCode:  http.StatusUnauthorized,
			wantSuccess: false,
		},
		"403 Forbidden is not success": {
			statusCode:  http.StatusForbidden,
			wantSuccess: false,
		},
		"404 Not found is not success": {
			statusCode:  http.StatusNotFound,
			wantSuccess: false,
		},
		"405 Method not allowed is not success": {
			statusCode:  http.StatusMethodNotAllowed,
			wantSuccess: false,
		},
		"500 Internal server error is not success": {
			statusCode:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testcase.wantSuccess, httputils.IsHTTPSuccess(testcase.statusCode))
		})
	}
}

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		modifier    func(c *http.Client)
		wantTimeout time.Duration
	}{
		"No modifier": {
			modifier:    nil,
			wantTimeout: httputils.HTTPClientDefaultTimeout,
		},
		"Empty modifier": {
			modifier:    func(c *http.Client) {},
			wantTimeout: httputils.HTTPClientDefaultTimeout,
		},
		"Timeout modifier": {
			modifier: func(c *http.Client) {
				c.Timeout = 5 * time.Second
			},
			wantTimeout: 5 * time.Second,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			httpClient := httputils.NewHTTPClient(testcase.modifier)
			require.Equal(t, testcase.wantTimeout, httpClient.Timeout)
		})
	}
}
