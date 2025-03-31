package auth_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/internal/auth"
	authmocks "github.com/alvii147/nymphadora-api/internal/auth/mocks"
	templatesmanagermocks "github.com/alvii147/nymphadora-api/internal/templatesmanager/mocks"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	mailclientmocks "github.com/alvii147/nymphadora-api/pkg/mailclient/mocks"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestJWTAuthMiddleware(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	userUUID := uuid.NewString()
	jti := uuid.NewString()
	oneDayAgo := timeProvider.Now().Add(-24 * time.Hour)
	secretKey := "deadbeef"
	crypto := cryptocore.NewCrypto(timeProvider, secretKey)

	validAccessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeAccess),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(secretKey))
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
	).SignedString([]byte(secretKey))
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
	).SignedString([]byte(secretKey))
	require.NoError(t, err)

	tokenWithInvalidClaim, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&struct {
			InvalidClaim string `json:"invalid_claim"`
			jwt.StandardClaims
		}{},
	).SignedString([]byte(secretKey))
	require.NoError(t, err)

	validResponse := map[string]any{
		"email":      testkit.GenerateFakeEmail(),
		"first_name": testkit.MustGenerateRandomString(8, true, true, false),
		"last_name":  testkit.MustGenerateRandomString(8, true, true, false),
		"created_at": timeProvider.Now().Format(time.RFC3339),
	}
	validStatusCode := 200

	testcases := map[string]struct {
		wantNextCall   bool
		wantErr        bool
		wantErrCode    string
		wantStatusCode int
		setAuthHeader  bool
		authHeader     string
	}{
		"Authentication with valid JWT is successful": {
			wantNextCall:   true,
			wantErr:        false,
			wantErrCode:    "",
			wantStatusCode: http.StatusOK,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Bearer %s", validAccessToken),
		},
		"Authentication with no authorization header is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  false,
			authHeader:     "",
		},
		"Authentication with invalid authentication type is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Invalidauthtype %s", validAccessToken),
		},
		"Authentication with invalid JWT is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     "Bearer ed0730889507fdb8549acfcd31548ee5",
		},
		"Authentication with expired JWT is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Bearer %s", expiredToken),
		},
		"Authentication with valid JWT of invalid type is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Bearer %s", tokenOfInvalidType),
		},
		"Authentication with JWT with invalid claim is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Bearer %s", tokenWithInvalidClaim),
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nextCallCount := 0
			var next httputils.HandlerFunc = func(w *httputils.ResponseWriter, r *http.Request) {
				require.Equal(t, userUUID, r.Context().Value(auth.AuthContextKeyUserUUID))
				w.WriteJSON(validResponse, validStatusCode)
				nextCallCount++
			}

			rec := httptest.NewRecorder()
			w := &httputils.ResponseWriter{
				ResponseWriter: rec,
				StatusCode:     -1,
			}
			r := httptest.NewRequest(http.MethodGet, "/auth/users/me", http.NoBody)

			if testcase.setAuthHeader {
				r.Header.Set("Authorization", testcase.authHeader)
			}

			auth.NewJWTAuthMiddleware(crypto)(next)(w, r)

			result := rec.Result()
			t.Cleanup(func() {
				err := result.Body.Close()
				require.NoError(t, err)
			})

			responseBodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			var responseBody map[string]any
			err = json.Unmarshal(responseBodyBytes, &responseBody)
			require.NoError(t, err)

			require.Equal(t, testcase.wantStatusCode, result.StatusCode)

			wantNextCallCount := 0
			if testcase.wantNextCall {
				wantNextCallCount = 1
			}

			require.Equal(t, wantNextCallCount, nextCallCount)

			if testcase.wantErr {
				errCode, ok := responseBody["code"]
				require.True(t, ok)
				require.Equal(t, testcase.wantErrCode, errCode)

				return
			}

			require.Equal(t, validResponse, responseBody)
		})
	}
}

func TestAPIKeyAuthMiddleware(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	_, validAPIKey := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	validResponse := map[string]any{
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"created_at": timeProvider.Now().Format(time.RFC3339),
	}
	validStatusCode := 200

	testcases := map[string]struct {
		wantNextCall   bool
		wantErr        bool
		wantErrCode    string
		wantStatusCode int
		setAuthHeader  bool
		authHeader     string
	}{
		"Authentication with valid API key is successful": {
			wantNextCall:   true,
			wantErr:        false,
			wantErrCode:    "",
			wantStatusCode: http.StatusOK,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("X-API-Key %s", validAPIKey),
		},
		"Authentication with no authorization header is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  false,
			authHeader:     "",
		},
		"Authentication with invalid authentication type is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Invalidauthtype %s", validAPIKey),
		},
		"Authentication with invalid API key is unauthorized": {
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     "X-API-Key DQGDG0Al.xoentiX0xPztDX6ybl6SNfveoCAT/M9Y6oXy96uMCGg=",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nextCallCount := 0
			var next httputils.HandlerFunc = func(w *httputils.ResponseWriter, r *http.Request) {
				require.Equal(t, user.UUID, r.Context().Value(auth.AuthContextKeyUserUUID))
				w.WriteJSON(validResponse, validStatusCode)
				nextCallCount++
			}

			rec := httptest.NewRecorder()
			w := &httputils.ResponseWriter{
				ResponseWriter: rec,
				StatusCode:     -1,
			}
			r := httptest.NewRequest(http.MethodGet, "/auth/users/me", http.NoBody)

			if testcase.setAuthHeader {
				r.Header.Set("Authorization", testcase.authHeader)
			}

			auth.NewAPIKeyAuthMiddleware(svc)(next)(w, r)

			result := rec.Result()
			t.Cleanup(func() {
				err := result.Body.Close()
				require.NoError(t, err)
			})

			responseBodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			var responseBody map[string]any
			err = json.Unmarshal(responseBodyBytes, &responseBody)
			require.NoError(t, err)

			require.Equal(t, testcase.wantStatusCode, result.StatusCode)

			wantNextCallCount := 0
			if testcase.wantNextCall {
				wantNextCallCount = 1
			}

			require.Equal(t, wantNextCallCount, nextCallCount)

			if testcase.wantErr {
				errCode, ok := responseBody["code"]
				require.True(t, ok)
				require.Equal(t, testcase.wantErrCode, errCode)

				return
			}

			require.Equal(t, validResponse, responseBody)
		})
	}
}

func TestAPIKeyAuthMiddlewareGenericServiceError(t *testing.T) {
	t.Parallel()

	nextCalled := false
	var next httputils.HandlerFunc = func(w *httputils.ResponseWriter, r *http.Request) {
		nextCalled = true
	}

	rec := httptest.NewRecorder()
	w := &httputils.ResponseWriter{
		ResponseWriter: rec,
		StatusCode:     -1,
	}
	r := httptest.NewRequest(http.MethodGet, "/auth/users/me", http.NoBody)
	r.Header.Set("Authorization", "X-API-Key TqxlYSSQ.Yj2j1jyAMC5407Nctsl51K7E8sOIPqYXn28SqT5Gnfg=")

	ctrl := gomock.NewController(t)
	svc := authmocks.NewMockService(ctrl)

	svc.
		EXPECT().
		FindAPIKey(gomock.Any(), "TqxlYSSQ.Yj2j1jyAMC5407Nctsl51K7E8sOIPqYXn28SqT5Gnfg=").
		Return(nil, errors.New("FindAPIKey failed")).
		Times(1)

	auth.NewAPIKeyAuthMiddleware(svc)(next)(w, r)

	result := rec.Result()
	t.Cleanup(func() {
		err := result.Body.Close()
		require.NoError(t, err)
	})

	responseBodyBytes, err := io.ReadAll(result.Body)
	require.NoError(t, err)

	var responseBody map[string]any
	err = json.Unmarshal(responseBodyBytes, &responseBody)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnauthorized, result.StatusCode)
	require.False(t, nextCalled)
	errCode, ok := responseBody["code"]
	require.True(t, ok)
	require.Equal(t, api.ErrCodeInvalidCredentials, errCode)
}
