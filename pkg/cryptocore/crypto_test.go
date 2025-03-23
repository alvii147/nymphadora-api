package cryptocore_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestCryptoHashPassword(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	c := cryptocore.NewCrypto(timeProvider, "deadbeef")

	password := testkit.GenerateFakePassword()
	hashedPassword, err := c.HashPassword(password)
	require.NoError(t, err)

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	require.NoError(t, err)
}

func TestCryptoCheckPassword(t *testing.T) {
	t.Parallel()

	secretKey := "deadbeef"
	correctPassword := testkit.GenerateFakePassword()
	incorrectPassword := testkit.GenerateFakePassword()

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(correctPassword), cryptocore.HashingCost)
	require.NoError(t, err)

	hashedPassword := string(hashedPasswordBytes)

	testcases := []struct {
		name     string
		password string
		wantOk   bool
	}{
		{
			name:     "Correct password",
			password: correctPassword,
			wantOk:   true,
		},
		{
			name:     "Incorrect password",
			password: incorrectPassword,
			wantOk:   false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			timeProvider := timekeeper.NewFrozenProvider()
			c := cryptocore.NewCrypto(timeProvider, secretKey)

			ok := c.CheckPassword(hashedPassword, testcase.password)
			require.Equal(t, testcase.wantOk, ok)
		})
	}
}

func TestCryptoCreateAuthJWTSuccess(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	secretKey := "deadbeef"

	c := cryptocore.NewCrypto(timeProvider, secretKey)

	testcases := []struct {
		name         string
		tokenType    cryptocore.JWTType
		wantLifetime time.Duration
	}{
		{
			name:         "Access token",
			tokenType:    cryptocore.JWTTypeAccess,
			wantLifetime: cryptocore.JWTLifetimeAccess,
		},
		{
			name:         "Refresh token",
			tokenType:    cryptocore.JWTTypeRefresh,
			wantLifetime: cryptocore.JWTLifetimeRefresh,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			userUUID := uuid.NewString()
			token, err := c.CreateAuthJWT(
				userUUID,
				testcase.tokenType,
			)
			require.NoError(t, err)

			claims := &cryptocore.AuthJWTClaims{}
			parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
				return []byte(secretKey), nil
			})
			require.NoError(t, err)

			require.NotNil(t, parsedToken)
			require.True(t, parsedToken.Valid)
			require.Equal(t, userUUID, claims.Subject)
			require.Equal(t, string(testcase.tokenType), claims.TokenType)
			require.WithinDuration(t, timeProvider.Now(), time.Time(claims.IssuedAt), testkit.TimeToleranceExact)
			require.WithinDuration(
				t,
				timeProvider.Now().Add(testcase.wantLifetime),
				time.Time(claims.ExpiresAt),
				testkit.TimeToleranceExact,
			)
		})
	}
}

func TestCryptoCreateAuthJWTInvalidType(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	c := cryptocore.NewCrypto(timeProvider, "deadbeef")

	_, err := c.CreateAuthJWT(
		uuid.NewString(),
		cryptocore.JWTType("invalidtype"),
	)
	require.Error(t, err)
}

func TestCryptoValidateAuthJWT(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	userUUID := uuid.NewString()
	jti := uuid.NewString()
	oneDayAgo := timeProvider.Now().Add(-24 * time.Hour)
	validSecretKey := "deadbeef"

	validAccessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeAccess),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
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
	).SignedString([]byte(validSecretKey))
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
	).SignedString([]byte(validSecretKey))
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
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenWithInvalidClaim, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&struct {
			InvalidClaim string `json:"invalid_claim"`
			jwt.StandardClaims
		}{},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	testcases := []struct {
		name      string
		token     string
		tokenType cryptocore.JWTType
		secretKey string
		wantOk    bool
	}{
		{
			name:      "Valid access token",
			token:     validAccessToken,
			tokenType: cryptocore.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    true,
		},
		{
			name:      "Valid refresh token",
			token:     validRefreshToken,
			tokenType: cryptocore.JWTTypeRefresh,
			secretKey: validSecretKey,
			wantOk:    true,
		},
		{
			name:      "Invalid secret key",
			token:     validAccessToken,
			tokenType: cryptocore.JWTTypeAccess,
			secretKey: "invalidsecretkey",
			wantOk:    false,
		},
		{
			name:      "Token of invalid type",
			token:     tokenOfInvalidType,
			tokenType: cryptocore.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Invalid token",
			token:     "ed0730889507fdb8549acfcd31548ee5",
			tokenType: cryptocore.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Expired token",
			token:     expiredToken,
			tokenType: cryptocore.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Token with invalid claim",
			token:     tokenWithInvalidClaim,
			tokenType: cryptocore.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			c := cryptocore.NewCrypto(timeProvider, testcase.secretKey)

			claims, ok := c.ValidateAuthJWT(testcase.token, testcase.tokenType)
			require.Equal(t, testcase.wantOk, ok)

			if testcase.wantOk {
				require.Equal(t, userUUID, claims.Subject)
				require.Equal(t, string(testcase.tokenType), claims.TokenType)
			}
		})
	}
}

func TestCryptoCreateActivationJWT(t *testing.T) {
	t.Parallel()

	userUUID := uuid.NewString()
	secretKey := "deadbeef"
	timeProvider := timekeeper.NewFrozenProvider()

	c := cryptocore.NewCrypto(timeProvider, secretKey)

	token, err := c.CreateActivationJWT(userUUID)
	require.NoError(t, err)

	claims := &cryptocore.ActivationJWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedToken)
	require.True(t, parsedToken.Valid)
	require.Equal(t, userUUID, claims.Subject)
	require.Equal(t, string(cryptocore.JWTTypeActivation), claims.TokenType)
	require.WithinDuration(t, timeProvider.Now(), time.Time(claims.IssuedAt), testkit.TimeToleranceExact)
	require.WithinDuration(
		t,
		timeProvider.Now().Add(cryptocore.JWTLifetimeActivation),
		time.Time(claims.ExpiresAt),
		testkit.TimeToleranceExact,
	)
}

func TestCryptoValidateActivationJWT(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	userUUID := uuid.NewString()
	jti := uuid.NewString()
	oneDayAgo := timeProvider.Now().Add(-24 * time.Hour)
	validSecretKey := "deadbeef"

	validToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeActivation),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenOfInvalidType, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string("invalidtype"),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeActivation),
			IssuedAt:  jsonutils.UnixTimestamp(oneDayAgo),
			ExpiresAt: jsonutils.UnixTimestamp(oneDayAgo.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenWithInvalidClaim, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&struct {
			InvalidClaim string `json:"invalid_claim"`
			jwt.StandardClaims
		}{},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	testcases := []struct {
		name      string
		token     string
		secretKey string
		wantOk    bool
	}{
		{
			name:      "Valid token of correct type",
			token:     validToken,
			secretKey: validSecretKey,
			wantOk:    true,
		},
		{
			name:      "Invalid secret key",
			token:     validToken,
			secretKey: "invalidsecretkey",
			wantOk:    false,
		},
		{
			name:      "Token of incorrect type",
			token:     tokenOfInvalidType,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Invalid token",
			token:     "ed0730889507fdb8549acfcd31548ee5",
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Expired token",
			token:     expiredToken,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Token with invalid claim",
			token:     tokenWithInvalidClaim,
			secretKey: validSecretKey,
			wantOk:    false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			c := cryptocore.NewCrypto(timeProvider, testcase.secretKey)

			claims, ok := c.ValidateActivationJWT(testcase.token)
			require.Equal(t, testcase.wantOk, ok)

			if testcase.wantOk {
				require.Equal(t, userUUID, claims.Subject)
				require.Equal(t, string(cryptocore.JWTTypeActivation), claims.TokenType)
			}
		})
	}
}

func TestCryptoCreateAPIKey(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	c := cryptocore.NewCrypto(timeProvider, "deadbeef")

	prefix, rawKey, hashedKey, err := c.CreateAPIKey()
	require.NoError(t, err)

	err = bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(rawKey))
	require.NoError(t, err)

	r := regexp.MustCompile(`^(\S+)\.(\S+)$`)
	matches := r.FindStringSubmatch(rawKey)

	require.Len(t, matches, 3)
	require.Equal(t, prefix, matches[1])
}

func TestCryptoParseAPIKey(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()

	testcases := []struct {
		name       string
		key        string
		wantPrefix string
		wantSecret string
		wantErr    bool
	}{
		{
			name:       "Valid API key",
			key:        "TqxlYSSQ.Yj2j1jyAMC5407Nctsl51K7E8sOIPqYXn28SqT5Gnfg=",
			wantPrefix: "TqxlYSSQ",
			wantSecret: "Yj2j1jyAMC5407Nctsl51K7E8sOIPqYXn28SqT5Gnfg=",
			wantErr:    false,
		},
		{
			name:       "Invalid API key",
			key:        "DeAdBeEf",
			wantPrefix: "",
			wantSecret: "",
			wantErr:    true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			c := cryptocore.NewCrypto(timeProvider, "deadbeef")

			prefix, secret, err := c.ParseAPIKey(testcase.key)
			if testcase.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wantPrefix, prefix)
			require.Equal(t, testcase.wantSecret, secret)
		})
	}
}

func TestCryptoCreateCodeSpaceInvitationJWT(t *testing.T) {
	t.Parallel()

	userUUID := uuid.NewString()
	inviteeEmail := testkit.GenerateFakeEmail()
	var codeSpaceID int64 = 314159
	accessLevel := 2
	secretKey := "deadbeef"
	timeProvider := timekeeper.NewFrozenProvider()

	c := cryptocore.NewCrypto(timeProvider, secretKey)

	token, err := c.CreateCodeSpaceInvitationJWT(userUUID, inviteeEmail, codeSpaceID, accessLevel)
	require.NoError(t, err)

	claims := &cryptocore.CodeSpaceInvitationJWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedToken)
	require.True(t, parsedToken.Valid)
	require.Equal(t, userUUID, claims.Subject)
	require.Equal(t, inviteeEmail, claims.InviteeEmail)
	require.Equal(t, codeSpaceID, claims.CodeSpaceID)
	require.Equal(t, accessLevel, claims.AccessLevel)
	require.Equal(t, string(cryptocore.JWTTypeCodeSpaceInvitation), claims.TokenType)
	require.WithinDuration(t, timeProvider.Now(), time.Time(claims.IssuedAt), testkit.TimeToleranceExact)
	require.WithinDuration(
		t,
		timeProvider.Now().Add(cryptocore.JWTLifetimeCodeSpaceInvitation),
		time.Time(claims.ExpiresAt),
		testkit.TimeToleranceExact,
	)
}

func TestCryptoValidateCodeSpaceInvitationJWT(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	userUUID := uuid.NewString()
	inviteeEmail := testkit.GenerateFakeEmail()
	var codeSpaceID int64 = 314159
	accessLevel := 2
	jti := uuid.NewString()
	oneDayAgo := timeProvider.Now().Add(-24 * time.Hour)
	validSecretKey := "deadbeef"

	validToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.CodeSpaceInvitationJWTClaims{
			Subject:      userUUID,
			InviteeEmail: inviteeEmail,
			CodeSpaceID:  codeSpaceID,
			AccessLevel:  accessLevel,
			TokenType:    string(cryptocore.JWTTypeCodeSpaceInvitation),
			IssuedAt:     jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt:    jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:        jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenOfInvalidType, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.CodeSpaceInvitationJWTClaims{
			Subject:      userUUID,
			InviteeEmail: inviteeEmail,
			CodeSpaceID:  codeSpaceID,
			AccessLevel:  accessLevel,
			TokenType:    string("invalidtype"),
			IssuedAt:     jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt:    jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:        jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.CodeSpaceInvitationJWTClaims{
			Subject:      userUUID,
			InviteeEmail: inviteeEmail,
			CodeSpaceID:  codeSpaceID,
			AccessLevel:  accessLevel,
			TokenType:    string(cryptocore.JWTTypeCodeSpaceInvitation),
			IssuedAt:     jsonutils.UnixTimestamp(oneDayAgo),
			ExpiresAt:    jsonutils.UnixTimestamp(oneDayAgo.Add(time.Hour)),
			JWTID:        jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenWithInvalidClaim, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&struct {
			InvalidClaim string `json:"invalid_claim"`
			jwt.StandardClaims
		}{},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	testcases := []struct {
		name      string
		token     string
		secretKey string
		wantOk    bool
	}{
		{
			name:      "Valid token of correct type",
			token:     validToken,
			secretKey: validSecretKey,
			wantOk:    true,
		},
		{
			name:      "Invalid secret key",
			token:     validToken,
			secretKey: "invalidsecretkey",
			wantOk:    false,
		},
		{
			name:      "Token of incorrect type",
			token:     tokenOfInvalidType,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Invalid token",
			token:     "ed0730889507fdb8549acfcd31548ee5",
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Expired token",
			token:     expiredToken,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Token with invalid claim",
			token:     tokenWithInvalidClaim,
			secretKey: validSecretKey,
			wantOk:    false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			c := cryptocore.NewCrypto(timeProvider, testcase.secretKey)

			claims, ok := c.ValidateCodeSpaceInvitationJWT(testcase.token)
			require.Equal(t, testcase.wantOk, ok)

			if testcase.wantOk {
				require.Equal(t, userUUID, claims.Subject)
				require.Equal(t, string(cryptocore.JWTTypeCodeSpaceInvitation), claims.TokenType)
			}
		})
	}
}
