package testkitinternal

import (
	"context"
	"fmt"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// MustHashPassword hashes a given password and panics on error.
func MustHashPassword(password string) string {
	cfg := MustCreateConfig()

	timeProvider := timekeeper.NewFrozenProvider()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)

	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		panic(errutils.FormatError(err))
	}

	return hashedPassword
}

// MustCreateUser creates and returns a new user and panics on error.
func MustCreateUser(t testkit.TestingT, modifier func(u *auth.User)) (*auth.User, string) {
	dbPool := RequireNewDatabasePool(t)
	dbConn := RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	userUUID := uuid.NewString()
	password := testkit.GenerateFakePassword()
	user := &auth.User{
		UUID:        userUUID,
		Email:       testkit.GenerateFakeEmail(),
		Password:    MustHashPassword(password),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    true,
		IsSuperUser: false,
	}

	if modifier != nil {
		modifier(user)
	}

	user, err := repo.CreateUser(context.Background(), dbConn, user)
	if err != nil {
		panic(errutils.FormatError(err))
	}

	return user, password
}

// MustCreateUserAuthJWTs creates access and refresh JWTs for a given user UUID and panics on error.
func MustCreateUserAuthJWTs(userUUID string) (string, string) {
	cfg := MustCreateConfig()

	timeProvider := timekeeper.NewFrozenProvider()
	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeAccess),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(
				timeProvider.Now().Add(cryptocore.JWTLifetimeAccess),
			),
			JWTID: uuid.NewString(),
		},
	).SignedString([]byte(cfg.SecretKey))
	if err != nil {
		panic(errutils.FormatError(err, "jwt.Token.SignedString failed"))
	}

	refreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeRefresh),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(
				timeProvider.Now().Add(cryptocore.JWTLifetimeRefresh),
			),
			JWTID: uuid.NewString(),
		},
	).SignedString([]byte(cfg.SecretKey))
	if err != nil {
		panic(errutils.FormatError(err, "jwt.Token.SignedString failed"))
	}

	return accessToken, refreshToken
}

// MustCreateUserAPIKey creates and returns a new API key for a given user UUID and panics on error.
func MustCreateUserAPIKey(t testkit.TestingT, userUUID string, modifier func(k *auth.APIKey)) (*auth.APIKey, string) {
	dbPool := RequireNewDatabasePool(t)
	dbConn := RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	name := testkit.MustGenerateRandomString(12, true, true, true)
	prefix := testkit.MustGenerateRandomString(8, true, true, true)
	secret := testkit.MustGenerateRandomString(32, true, true, true)
	rawKey := fmt.Sprintf("%s.%s", prefix, secret)
	hashedKey := MustHashPassword(rawKey)

	apiKey := &auth.APIKey{
		UserUUID:  userUUID,
		Prefix:    prefix,
		HashedKey: hashedKey,
		Name:      name,
		ExpiresAt: nil,
	}

	if modifier != nil {
		modifier(apiKey)
	}

	apiKey, err := repo.CreateAPIKey(context.Background(), dbConn, apiKey)
	if err != nil {
		panic(errutils.FormatError(err))
	}

	return apiKey, rawKey
}
