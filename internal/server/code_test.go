package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

func TestHandleCreateCodeSpace(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	userAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(user.UUID)

	testcases := []struct {
		name           string
		headers        map[string]string
		requestBody    string
		wantStatusCode int
		wantLanguage   string
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Valid request for Go code space",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"language": "go"
				}
			`,
			wantStatusCode: http.StatusCreated,
			wantLanguage:   "go",
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Valid request for Python code space",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"language": "python"
				}
			`,
			wantStatusCode: http.StatusCreated,
			wantLanguage:   "python",
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Unsupported language",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"language": "un5upp0rt3d"
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantLanguage:   "",
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "No language",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantLanguage:   "",
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name:    "Unauthenticated request",
			headers: map[string]string{},
			requestBody: `
				{
					"language": "python"
				}
			`,
			wantStatusCode: http.StatusUnauthorized,
			wantLanguage:   "python",
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingToken,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/code/space",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

			timeProvider := timekeeper.NewFrozenProvider()
			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if !httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)

				return
			}

			var createCodeSpaceResp api.CreateCodeSpaceResponse
			err = json.NewDecoder(res.Body).Decode(&createCodeSpaceResp)
			require.NoError(t, err)

			require.NotNil(t, createCodeSpaceResp.AuthorUUID)
			require.Equal(t, user.UUID, *createCodeSpaceResp.AuthorUUID)
			require.Equal(t, testcase.wantLanguage, createCodeSpaceResp.Language)
			require.Equal(t, api.CodeSpaceAccessLevelReadWrite, createCodeSpaceResp.AccessLevel)
			require.WithinDuration(t, timeProvider.Now(), createCodeSpaceResp.CreatedAt, testkit.TimeToleranceTentative)
			require.WithinDuration(t, timeProvider.Now(), createCodeSpaceResp.UpdatedAt, testkit.TimeToleranceTentative)
		})
	}
}
