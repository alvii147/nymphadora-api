package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/internal/server"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetAPIKeyIDParam(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		pathValues   map[string]string
		wantAPIKeyID int64
		wantErr      bool
	}{
		"Valid API key ID": {
			pathValues: map[string]string{
				"id": "42",
			},
			wantAPIKeyID: 42,
			wantErr:      false,
		},
		"No API key ID": {
			pathValues: map[string]string{
				"dead": "beef",
			},
			wantAPIKeyID: 0,
			wantErr:      true,
		},
		"Invalid API key ID": {
			pathValues: map[string]string{
				"id": "deadbeef",
			},
			wantAPIKeyID: 0,
			wantErr:      true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := &http.Request{}
			for name, value := range testcase.pathValues {
				req.SetPathValue(name, value)
			}

			apiKeyID, err := server.GetAPIKeyIDParam(req)
			if testcase.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wantAPIKeyID, apiKeyID)
		})
	}
}

func TestHandleCreateUser(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	existingUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	email := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()
	firstName := testkit.MustGenerateRandomString(8, true, true, false)
	lastName := testkit.MustGenerateRandomString(8, true, true, false)

	testcases := map[string]struct {
		requestBody    string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		"Valid request": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, email, password, firstName, lastName),
			wantStatusCode: http.StatusCreated,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Existing email": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, existingUser.Email, password, firstName, lastName),
			wantStatusCode: http.StatusConflict,
			wantErrCode:    api.ErrCodeResourceExists,
			wantErrDetail:  api.ErrDetailUserExists,
		},
		"Invalid email": {
			requestBody: fmt.Sprintf(`
				{
					"email": "1nv4l1d3m41l",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, password, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Missing email": {
			requestBody: fmt.Sprintf(`
				{
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, password, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Empty email": {
			requestBody: fmt.Sprintf(`
				{
					"email": "",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, password, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Missing password": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, email, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Empty password": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, email, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Missing first name": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"last_name": "%s"
				}
			`, email, password, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Empty first name": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "",
					"last_name": "%s"
				}
			`, email, password, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Missing last name": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
				}
			`, email, password, firstName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Empty last name": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
					"last_name": ""
				}
			`, email, password, firstName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/users",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

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

			var createUserResp api.CreateUserResponse
			err = json.NewDecoder(res.Body).Decode(&createUserResp)
			require.NoError(t, err)

			require.Equal(t, email, createUserResp.Email)
			require.Equal(t, firstName, createUserResp.FirstName)
			require.Equal(t, lastName, createUserResp.LastName)
			require.WithinDuration(t, timeProvider.Now(), createUserResp.CreatedAt, testkit.TimeToleranceTentative)
			require.WithinDuration(t, timeProvider.Now(), createUserResp.UpdatedAt, testkit.TimeToleranceTentative)
		})
	}
}

func TestHandleActivateUser(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	timeProvider := timekeeper.NewFrozenProvider()
	activeUserToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.ActivationJWTClaims{
			Subject:   activeUser.UUID,
			TokenType: string(cryptocore.JWTTypeActivation),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(
				timeProvider.Now().Add(
					cryptocore.JWTLifetimeActivation,
				),
			),
			JWTID: uuid.NewString(),
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUserToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.ActivationJWTClaims{
			Subject:   inactiveUser.UUID,
			TokenType: string(cryptocore.JWTTypeActivation),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(
				timeProvider.Now().Add(
					cryptocore.JWTLifetimeActivation,
				),
			),
			JWTID: uuid.NewString(),
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	testcases := map[string]struct {
		requestBody    string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		"Inactive user": {
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, inactiveUserToken),
			wantStatusCode: http.StatusOK,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Already active user": {
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, activeUserToken),
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailUserNotFound,
		},
		"Invalid token": {
			requestBody: `
				{
					"token": "1nV4LiDT0k3n"
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidToken,
		},
		"Missing token": {
			requestBody: `
				{}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/users/activate",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

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
			}
		})
	}
}

func TestHandleGetUserMe(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	_, activeUserRawAPIKey := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, nil)

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	_, inactiveUserRawAPIKey := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	testcases := map[string]struct {
		path           string
		headers        map[string]string
		user           *auth.User
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		"Get active user using JWT": {
			path: "/auth/users/me",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			user:           activeUser,
			wantStatusCode: http.StatusOK,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Get active user using API key": {
			path: "/api/v1/auth/users/me",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("X-API-Key %s", activeUserRawAPIKey),
			},
			user:           activeUser,
			wantStatusCode: http.StatusOK,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Get inactive user using JWT": {
			path: "/auth/users/me",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			user:           inactiveUser,
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailUserNotFound,
		},
		"Get inactive user using API key": {
			path: "/api/v1/auth/users/me",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("X-API-Key %s", inactiveUserRawAPIKey),
			},
			user:           inactiveUser,
			wantStatusCode: http.StatusUnauthorized,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantErrDetail:  api.ErrDetailInvalidToken,
		},
		"Get user without authentication": {
			path:           "/auth/users/me",
			headers:        map[string]string{},
			user:           nil,
			wantStatusCode: http.StatusUnauthorized,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingToken,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, TestServerURL+testcase.path, http.NoBody)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

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

			var getUserMeResp api.GetUserMeResponse
			err = json.NewDecoder(res.Body).Decode(&getUserMeResp)
			require.NoError(t, err)

			require.Equal(t, testcase.user.UUID, getUserMeResp.UUID)
			require.Equal(t, testcase.user.Email, getUserMeResp.Email)
			require.Equal(t, testcase.user.FirstName, getUserMeResp.FirstName)
			require.Equal(t, testcase.user.LastName, getUserMeResp.LastName)
			require.Equal(t, testcase.user.CreatedAt, getUserMeResp.CreatedAt)
			require.Equal(t, testcase.user.UpdatedAt, getUserMeResp.UpdatedAt)
		})
	}
}

func TestHandleUpdateUser(t *testing.T) {
	t.Parallel()

	startingFirstName := "startingFirstName"
	startingLastName := "startingLastName"
	updatedFirstName := "updatedFirstName"
	updatedLastName := "updatedLastName"

	httpClient := httputils.NewHTTPClient(nil)

	activeUser1, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.FirstName = startingFirstName
		u.LastName = startingLastName
		u.IsActive = true
	})
	activeUserAccessJWT1, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser1.UUID)

	activeUser2, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.FirstName = startingFirstName
		u.LastName = startingLastName
		u.IsActive = true
	})
	activeUserAccessJWT2, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser2.UUID)

	activeUser3, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.FirstName = startingFirstName
		u.LastName = startingLastName
		u.IsActive = true
	})
	activeUserAccessJWT3, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser3.UUID)

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.FirstName = startingFirstName
		u.LastName = startingLastName
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)

	testcases := map[string]struct {
		headers        map[string]string
		requestBody    string
		wantStatusCode int
		wantFirstName  string
		wantLastName   string
		wantErrCode    string
		wantErrDetail  string
	}{
		"Update active user first and last names": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT1),
			},
			requestBody: fmt.Sprintf(`
				{
					"first_name": "%s",
					"last_name": "%s"
				}
			`, updatedFirstName, updatedLastName),
			wantStatusCode: http.StatusOK,
			wantFirstName:  updatedFirstName,
			wantLastName:   updatedLastName,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Update active user first name": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT2),
			},
			requestBody: fmt.Sprintf(`
				{
					"first_name": "%s"
				}
			`, updatedFirstName),
			wantStatusCode: http.StatusOK,
			wantFirstName:  updatedFirstName,
			wantLastName:   activeUser2.LastName,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Update active user last name": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT3),
			},
			requestBody: fmt.Sprintf(`
				{
					"last_name": "%s"
				}
			`, updatedLastName),
			wantStatusCode: http.StatusOK,
			wantFirstName:  activeUser3.FirstName,
			wantLastName:   updatedLastName,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Update inactive user first name and last name": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			requestBody: fmt.Sprintf(`
				{
					"first_name": "%s",
					"last_name": "%s"
				}
			`, updatedFirstName, updatedLastName),
			wantStatusCode: http.StatusNotFound,
			wantFirstName:  "",
			wantLastName:   "",
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailUserNotFound,
		},
		"Update user with invalid first and last names": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT1),
			},
			requestBody: fmt.Sprintf(`
				{
					"first_name": %d,
					"last_name": %v
				}
			`, 42, true),
			wantStatusCode: http.StatusBadRequest,
			wantFirstName:  "",
			wantLastName:   "",
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Update user with empty first and last names": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT1),
			},
			requestBody: `
				{
					"first_name": "",
					"last_name": ""
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantFirstName:  "",
			wantLastName:   "",
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Update user first and last names without authentication": {
			headers: map[string]string{},
			requestBody: fmt.Sprintf(`
				{
					"first_name": "%s",
					"last_name": "%s"
				}
			`, updatedFirstName, updatedLastName),
			wantStatusCode: http.StatusUnauthorized,
			wantFirstName:  "",
			wantLastName:   "",
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingToken,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			timeProvider := timekeeper.NewFrozenProvider()
			req, err := http.NewRequest(
				http.MethodPatch,
				TestServerURL+"/auth/users/me",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

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

			var updateUserResp api.UpdateUserResponse
			err = json.NewDecoder(res.Body).Decode(&updateUserResp)
			require.NoError(t, err)

			require.Equal(t, testcase.wantFirstName, updateUserResp.FirstName)
			require.Equal(t, testcase.wantLastName, updateUserResp.LastName)
			require.WithinDuration(t, timeProvider.Now(), updateUserResp.UpdatedAt, testkit.TimeToleranceTentative)
		})
	}
}

func TestHandleCreateJWT(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	httpClient := httputils.NewHTTPClient(nil)

	user, password := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testcases := map[string]struct {
		requestBody    string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		"Valid request": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s"
				}
			`, user.Email, password),
			wantStatusCode: http.StatusCreated,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Incorrect password": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "1nc0rr3CTP455w0Rd"
				}
			`, user.Email),
			wantStatusCode: http.StatusUnauthorized,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantErrDetail:  api.ErrDetailInvalidEmailOrPassword,
		},
		"Invalid email": {
			requestBody: fmt.Sprintf(`
				{
					"email": "1nv4l1d3M4iL",
					"password": "%s"
				}
			`, password),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Missing email": {
			requestBody: fmt.Sprintf(`
				{
					"password": "%s"
				}
			`, password),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Empty email": {
			requestBody: fmt.Sprintf(`
				{
					"email": "",
					"password": "%s"
				}
			`, password),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Missing password": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s"
				}
			`, user.Email),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Empty password": {
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": ""
				}
			`, user.Email),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/tokens",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

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

			var createTokenResp api.CreateTokenResponse
			err = json.NewDecoder(res.Body).Decode(&createTokenResp)
			require.NoError(t, err)

			accessClaims := &cryptocore.AuthJWTClaims{}
			parsedAccessToken, err := jwt.ParseWithClaims(createTokenResp.Access, accessClaims, func(t *jwt.Token) (any, error) {
				return []byte(cfg.SecretKey), nil
			})
			require.NoError(t, err)

			require.NotNil(t, parsedAccessToken)
			require.True(t, parsedAccessToken.Valid)
			require.Equal(t, user.UUID, accessClaims.Subject)
			require.Equal(t, string(cryptocore.JWTTypeAccess), accessClaims.TokenType)

			require.WithinDuration(t, timeProvider.Now(), time.Time(accessClaims.IssuedAt), testkit.TimeToleranceTentative)
			require.WithinDuration(
				t,
				timeProvider.Now().Add(cryptocore.JWTLifetimeAccess),
				time.Time(accessClaims.ExpiresAt),
				testkit.TimeToleranceTentative,
			)

			refreshClaims := &cryptocore.AuthJWTClaims{}
			parsedRefreshToken, err := jwt.ParseWithClaims(
				createTokenResp.Refresh,
				refreshClaims,
				func(t *jwt.Token) (any, error) {
					return []byte(cfg.SecretKey), nil
				},
			)
			require.NoError(t, err)

			require.NotNil(t, parsedRefreshToken)
			require.True(t, parsedRefreshToken.Valid)
			require.Equal(t, user.UUID, refreshClaims.Subject)
			require.Equal(t, string(cryptocore.JWTTypeRefresh), refreshClaims.TokenType)

			require.WithinDuration(t, timeProvider.Now(), time.Time(refreshClaims.IssuedAt), testkit.TimeToleranceTentative)
			require.WithinDuration(
				t,
				timeProvider.Now().Add(cryptocore.JWTLifetimeRefresh),
				time.Time(refreshClaims.ExpiresAt),
				testkit.TimeToleranceTentative,
			)
		})
	}
}

func TestHandleRefreshJWT(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	httpClient := httputils.NewHTTPClient(nil)

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	_, validRefreshToken := testkitinternal.MustCreateUserAuthJWTs(user.UUID)

	testcases := map[string]struct {
		requestBody    string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		"Valid request": {
			requestBody: fmt.Sprintf(`
				{
					"refresh": "%s"
				}
			`, validRefreshToken),
			wantStatusCode: http.StatusCreated,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Invalid token": {
			requestBody: `
				{
					"refresh": "iNv4liDT0k3N"
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Missing token": {
			requestBody: `
				{}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/tokens/refresh",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

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

			var refreshTokenResp api.RefreshTokenResponse
			err = json.NewDecoder(res.Body).Decode(&refreshTokenResp)
			require.NoError(t, err)

			claims := &cryptocore.AuthJWTClaims{}
			parsedToken, err := jwt.ParseWithClaims(refreshTokenResp.Access, claims, func(t *jwt.Token) (any, error) {
				return []byte(cfg.SecretKey), nil
			})
			require.NoError(t, err)

			require.NotNil(t, parsedToken)
			require.True(t, parsedToken.Valid)
			require.Equal(t, user.UUID, claims.Subject)
			require.Equal(t, string(cryptocore.JWTTypeAccess), claims.TokenType)

			require.WithinDuration(t, timeProvider.Now(), time.Time(claims.IssuedAt), testkit.TimeToleranceTentative)
			require.WithinDuration(
				t,
				timeProvider.Now().Add(cryptocore.JWTLifetimeAccess),
				time.Time(claims.ExpiresAt),
				testkit.TimeToleranceTentative,
			)
		})
	}
}

func TestHandleValidateJWT(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	httpClient := httputils.NewHTTPClient(nil)

	timeProvider := timekeeper.NewFrozenProvider()
	userUUID := uuid.NewString()
	jti := uuid.NewString()
	oneDayAgo := timeProvider.Now().Add(-24 * time.Hour)

	validAccessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeAccess),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	validRefreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeRefresh),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	tokenWithInvalidSecretKey, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeAccess),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte("1nV4LiDS3cR3Tk3Y"))
	require.NoError(t, err)

	tokenOfInvalidType, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string("invalidtype"),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeAccess),
			IssuedAt:  jsonutils.UnixTimestamp(oneDayAgo),
			ExpiresAt: jsonutils.UnixTimestamp(oneDayAgo.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	tokenWithInvalidClaim, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&struct {
			InvalidClaim string `json:"invalid_claim"`
			jwt.StandardClaims
		}{},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	testcases := map[string]struct {
		requestBody    string
		wantStatusCode int
		wantValid      bool
		wantErrCode    string
		wantErrDetail  string
	}{
		"Valid access token": {
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, validAccessToken),
			wantStatusCode: http.StatusOK,
			wantValid:      true,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Valid refresh token": {
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, validRefreshToken),
			wantStatusCode: http.StatusOK,
			wantValid:      false,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Invalid secret key": {
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, tokenWithInvalidSecretKey),
			wantStatusCode: http.StatusOK,
			wantValid:      false,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Token of invalid type": {
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, tokenOfInvalidType),
			wantStatusCode: http.StatusOK,
			wantValid:      false,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Invalid token": {
			requestBody: `
				{
					"token": "ed0730889507fdb8549acfcd31548ee5"
				}
			`,
			wantStatusCode: http.StatusOK,
			wantValid:      false,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Expired token": {
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, expiredToken),
			wantStatusCode: http.StatusOK,
			wantValid:      false,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Token with invalid claim": {
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, tokenWithInvalidClaim),
			wantStatusCode: http.StatusOK,
			wantValid:      false,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Empty token": {
			requestBody: `
				{
					"token": ""
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantValid:      false,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Missing token": {
			requestBody: `
				{}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantValid:      false,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Invalid request data": {
			requestBody: `
				{
					"token": 42
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantValid:      false,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/tokens/validate",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

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

			var validateTokenResp api.ValidateTokenResponse
			err = json.NewDecoder(res.Body).Decode(&validateTokenResp)
			require.NoError(t, err)
			require.Equal(t, testcase.wantValid, validateTokenResp.Valid)
		})
	}
}

func TestHandleCreateAPIKey(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	userAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(user.UUID)
	userExistingAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	expirationDate := time.Date(2038, 1, 19, 3, 14, 8, 0, time.UTC)
	expirationDateString := "2038-01-19T03:14:08Z"

	testcases := map[string]struct {
		headers            map[string]string
		requestBody        string
		wantStatusCode     int
		wantAPIKeyName     string
		wantExpirationDate *time.Time
		wantErrCode        string
		wantErrDetail      string
	}{
		"Valid request with no expiration date": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"name": "My non-expiring API key"
				}
			`,
			wantStatusCode:     http.StatusCreated,
			wantAPIKeyName:     "My non-expiring API key",
			wantExpirationDate: nil,
			wantErrCode:        "",
			wantErrDetail:      "",
		},
		"Valid request with expiration date": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: fmt.Sprintf(`
				{
					"name": "My expiring API key",
					"expires_at": "%s"
				}
			`, expirationDateString),
			wantStatusCode:     http.StatusCreated,
			wantAPIKeyName:     "My expiring API key",
			wantExpirationDate: &expirationDate,
			wantErrCode:        "",
			wantErrDetail:      "",
		},
		"Name missing": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: fmt.Sprintf(`
				{
					"expires_at": "%s"
				}
			`, expirationDateString),
			wantStatusCode:     http.StatusBadRequest,
			wantAPIKeyName:     "",
			wantExpirationDate: &expirationDate,
			wantErrCode:        api.ErrCodeInvalidRequest,
			wantErrDetail:      api.ErrDetailInvalidRequestData,
		},
		"Invalid expiration date": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"name": "My invalidly-expiring API key",
					"expires_at": "1nv4l1dd4t3"
				}
			`,
			wantStatusCode:     http.StatusBadRequest,
			wantAPIKeyName:     "",
			wantExpirationDate: nil,
			wantErrCode:        api.ErrCodeInvalidRequest,
			wantErrDetail:      api.ErrDetailInvalidRequestData,
		},
		"Unauthenticated request": {
			headers: map[string]string{},
			requestBody: `
				{
					"name": "My unauthenticated API key"
				}
			`,
			wantStatusCode:     http.StatusUnauthorized,
			wantAPIKeyName:     "",
			wantExpirationDate: nil,
			wantErrCode:        api.ErrCodeMissingCredentials,
			wantErrDetail:      api.ErrDetailMissingToken,
		},
		"API key exists": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: fmt.Sprintf(`
				{
					"name": "%s"
				}
			`, userExistingAPIKey.Name),
			wantStatusCode:     http.StatusConflict,
			wantAPIKeyName:     "",
			wantExpirationDate: nil,
			wantErrCode:        api.ErrCodeResourceExists,
			wantErrDetail:      api.ErrDetailAPIKeyExists,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/api-keys",
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

			var createAPIKeyResp api.CreateAPIKeyResponse
			err = json.NewDecoder(res.Body).Decode(&createAPIKeyResp)
			require.NoError(t, err)

			require.Equal(t, user.UUID, createAPIKeyResp.UserUUID)
			require.Equal(t, testcase.wantAPIKeyName, createAPIKeyResp.Name)
			require.Equal(t, testcase.wantExpirationDate, createAPIKeyResp.ExpiresAt)
			require.WithinDuration(t, timeProvider.Now(), createAPIKeyResp.CreatedAt, testkit.TimeToleranceTentative)
			require.WithinDuration(t, timeProvider.Now(), createAPIKeyResp.UpdatedAt, testkit.TimeToleranceTentative)
		})
	}
}

func TestHandleListAPIKeys(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	activeUserAPIKey, activeUserRawAPIKey := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, nil)

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	testcases := map[string]struct {
		headers               map[string]string
		wantStatusCode        int
		wantAPIKeysInResponse bool
		wantErrCode           string
		wantErrDetail         string
	}{
		"List API keys for active user": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode:        http.StatusOK,
			wantAPIKeysInResponse: true,
			wantErrCode:           "",
			wantErrDetail:         "",
		},
		"List API keys for inactive user": {
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode:        http.StatusOK,
			wantAPIKeysInResponse: false,
			wantErrCode:           api.ErrCodeResourceNotFound,
			wantErrDetail:         api.ErrDetailUserNotFound,
		},
		"List API keys without authentication": {
			headers:               map[string]string{},
			wantStatusCode:        http.StatusUnauthorized,
			wantAPIKeysInResponse: false,
			wantErrCode:           api.ErrCodeMissingCredentials,
			wantErrDetail:         api.ErrDetailMissingToken,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodGet,
				TestServerURL+"/auth/api-keys",
				http.NoBody,
			)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

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

			var listAPIKeysResp api.ListAPIKeysResponse
			err = json.NewDecoder(res.Body).Decode(&listAPIKeysResp)
			require.NoError(t, err)

			if !testcase.wantAPIKeysInResponse {
				require.Empty(t, listAPIKeysResp.Keys)

				return
			}

			require.Len(t, listAPIKeysResp.Keys, 1)
			require.Equal(t, activeUserAPIKey.ID, listAPIKeysResp.Keys[0].ID)
			require.Equal(t, activeUser.UUID, listAPIKeysResp.Keys[0].UserUUID)
			require.True(t, strings.HasPrefix(activeUserRawAPIKey, listAPIKeysResp.Keys[0].Prefix))
			require.Equal(t, activeUserAPIKey.Name, listAPIKeysResp.Keys[0].Name)
			require.Equal(t, activeUserAPIKey.ExpiresAt, listAPIKeysResp.Keys[0].ExpiresAt)
			require.Equal(t, activeUserAPIKey.CreatedAt, listAPIKeysResp.Keys[0].CreatedAt)
			require.Equal(t, activeUserAPIKey.UpdatedAt, listAPIKeysResp.Keys[0].UpdatedAt)
		})
	}
}

func TestHandleUpdateAPIKey(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	startingExpiresAt := timeProvider.Now().AddDate(0, 1, 0)
	updatedExpiresAt := timeProvider.Now().AddDate(1, 0, 0)

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	activeUserAPIKey1, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.Name = "apiKey1Name"
		k.ExpiresAt = &startingExpiresAt
	})
	activeUserAPIKey2, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.Name = "apiKey2Name"
		k.ExpiresAt = &startingExpiresAt
	})
	activeUserAPIKey3, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.Name = "apiKey3Name"
		k.ExpiresAt = &startingExpiresAt
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	inactiveUserAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	testcases := map[string]struct {
		path           string
		headers        map[string]string
		requestBody    string
		wantStatusCode int
		wantName       string
		wantExpiresAt  *time.Time
		wantErrCode    string
		wantErrDetail  string
	}{
		"Update name, update expires at from valid date to valid date": {
			path: fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey1.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: fmt.Sprintf(`
				{
					"name": "updatedAPIKey1Name",
					"expires_at": "%s"
				}
			`, updatedExpiresAt.Format(time.RFC3339)),
			wantStatusCode: http.StatusOK,
			wantName:       "updatedAPIKey1Name",
			wantExpiresAt:  &updatedExpiresAt,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Update name, update expires at from valid date to null": {
			path: fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey2.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"name": "updatedAPIKey2Name",
					"expires_at": null
				}
			`,
			wantStatusCode: http.StatusOK,
			wantName:       "updatedAPIKey2Name",
			wantExpiresAt:  nil,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Update name only": {
			path: fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey3.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"name": "updatedAPIKey3Name"
				}
			`,
			wantStatusCode: http.StatusOK,
			wantName:       "updatedAPIKey3Name",
			wantExpiresAt:  activeUserAPIKey3.ExpiresAt,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Update name to empty string fails": {
			path: fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey1.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"name": ""
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantName:       "",
			wantExpiresAt:  nil,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Updating using invalid API Key ID fails": {
			path: fmt.Sprintf("/auth/api-keys/%s", "invalidID"),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"name": "updatedAPIKey1Name"
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantName:       "",
			wantExpiresAt:  nil,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		"Updating using non-existent API Key ID fails": {
			path: fmt.Sprintf("/auth/api-keys/%d", 314159),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"name": "updatedAPIKey1Name"
				}
			`,
			wantStatusCode: http.StatusNotFound,
			wantName:       "",
			wantExpiresAt:  nil,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailAPIKeyNotFound,
		},
		"Updating without authentication fails": {
			path:    fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey1.ID),
			headers: map[string]string{},
			requestBody: `
				{
					"name": "updatedAPIKey1Name"
				}
			`,
			wantStatusCode: http.StatusUnauthorized,
			wantName:       "",
			wantExpiresAt:  nil,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingToken,
		},
		"Updating API Key for inactive user fails": {
			path: fmt.Sprintf("/auth/api-keys/%d", inactiveUserAPIKey.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			requestBody: `
				{
					"name": "updatedAPIKey1Name"
				}
			`,
			wantStatusCode: http.StatusNotFound,
			wantName:       "",
			wantExpiresAt:  nil,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailAPIKeyNotFound,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			timeProvider := timekeeper.NewFrozenProvider()
			req, err := http.NewRequest(
				http.MethodPatch,
				TestServerURL+testcase.path,
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

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

			var updateAPIKeyResp api.UpdateAPIKeyResponse
			err = json.NewDecoder(res.Body).Decode(&updateAPIKeyResp)
			require.NoError(t, err)

			require.Equal(t, testcase.wantName, updateAPIKeyResp.Name)
			if testcase.wantExpiresAt != nil {
				require.NotNil(t, updateAPIKeyResp.ExpiresAt)
				require.WithinDuration(t, *testcase.wantExpiresAt, *updateAPIKeyResp.ExpiresAt, testkit.TimeToleranceExact)
			} else {
				require.Nil(t, updateAPIKeyResp.ExpiresAt)
			}

			require.WithinDuration(t, timeProvider.Now(), updateAPIKeyResp.UpdatedAt, testkit.TimeToleranceTentative)
		})
	}
}

func TestHandleDeleteAPIKey(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	activeUserAPIKey1, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.Name = "MyAPIKey1"
	})
	activeUserAPIKey2, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.Name = "MyAPIKey2"
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	inactiveUserAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	testcases := map[string]struct {
		path           string
		headers        map[string]string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		"Delete API key for active user": {
			path: fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey1.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode: http.StatusNoContent,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		"Delete API key for inactive user": {
			path: fmt.Sprintf("/auth/api-keys/%d", inactiveUserAPIKey.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailAPIKeyNotFound,
		},
		"Delete API key for another user": {
			path: fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey2.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailAPIKeyNotFound,
		},
		"Delete non-existent API key": {
			path: fmt.Sprintf("/auth/api-keys/%d", 314159),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailAPIKeyNotFound,
		},
		"Delete API key without authentication": {
			path:           fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey2.ID),
			headers:        map[string]string{},
			wantStatusCode: http.StatusUnauthorized,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingToken,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodDelete, TestServerURL+testcase.path, http.NoBody)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

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
			}
		})
	}
}
