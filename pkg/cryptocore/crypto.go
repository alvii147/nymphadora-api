package cryptocore

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/random"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// JWTType is a string representing type of JWT.
// Allowed strings are "access", "refresh", and "activation".
type JWTType string

const (
	// HashingCost is the cost for cryptographic hashing.
	HashingCost = 14
	// JWTTypeAccess represents access JWTs.
	JWTTypeAccess JWTType = "access"
	// JWTTypeRefresh represents refresh JWTs.
	JWTTypeRefresh JWTType = "refresh"
	// JWTTypeActivation represents activation JWTs.
	JWTTypeActivation JWTType = "activation"
	// JWTTypeCodeSpaceInvitation represents code space invitation JWTs.
	JWTTypeCodeSpaceInvitation JWTType = "codespaceinvitation"
	// JWTLifetimeAccess is the lifetime of an access JWT.
	JWTLifetimeAccess = time.Hour
	// JWTLifetimeRefresh is the lifetime of a refresh JWT.
	JWTLifetimeRefresh = 30 * 24 * time.Hour
	// JWTLifetimeActivation is the lifetime of an activation JWT.
	JWTLifetimeActivation = 30 * 24 * time.Hour
	// JWTLifetimeCodeSpaceInvitation is the lifetime of a code space invitation JWT.
	JWTLifetimeCodeSpaceInvitation = 7 * 24 * time.Hour
	// APIKeyPrefixLength is the length of API key prefixes.
	APIKeyPrefixLength = 8
	// APIKeySecretNBytes is the number of bytes in API key secrets.
	APIKeySecretNBytes = 32
)

// AuthJWTClaims represents claims in JWTs used for user authentication.
type AuthJWTClaims struct {
	Subject   string                  `json:"sub"`
	TokenType string                  `json:"token_type"`
	IssuedAt  jsonutils.UnixTimestamp `json:"iat"`
	ExpiresAt jsonutils.UnixTimestamp `json:"exp"`
	JWTID     string                  `json:"jti"`
	jwt.StandardClaims
}

// ActivationJWTClaims represents claims in JWTs used for user activation.
type ActivationJWTClaims struct {
	Subject   string                  `json:"sub"`
	TokenType string                  `json:"token_type"`
	IssuedAt  jsonutils.UnixTimestamp `json:"iat"`
	ExpiresAt jsonutils.UnixTimestamp `json:"exp"`
	JWTID     string                  `json:"jti"`
	jwt.StandardClaims
}

// CodeSpaceInvitationJWTClaims represents claims in JWTs used for code space invitation.
type CodeSpaceInvitationJWTClaims struct {
	Subject      string                  `json:"sub"`
	InviteeEmail string                  `json:"invitee_email"`
	CodeSpaceID  int64                   `json:"code_space_id"`
	AccessLevel  int                     `json:"access_level"`
	TokenType    string                  `json:"token_type"`
	IssuedAt     jsonutils.UnixTimestamp `json:"iat"`
	ExpiresAt    jsonutils.UnixTimestamp `json:"exp"`
	JWTID        string                  `json:"jti"`
	jwt.StandardClaims
}

// Crypto performs all cryptography-related computations and logic.
//
//go:generate mockgen -package=cryptocoremocks -source=$GOFILE -destination=./mocks/crypto.go
type Crypto interface {
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword string, password string) bool
	CreateAuthJWT(userUUID string, tokenType JWTType) (string, error)
	ValidateAuthJWT(token string, tokenType JWTType) (*AuthJWTClaims, bool)
	CreateActivationJWT(userUUID string) (string, error)
	ValidateActivationJWT(token string) (*ActivationJWTClaims, bool)
	CreateAPIKey() (string, string, string, error)
	ParseAPIKey(key string) (string, string, error)
	CreateCodeSpaceInvitationJWT(
		userUUID string,
		inviteeEmail string,
		codeSpaceID int64,
		accessLevel int,
	) (string, error)
	ValidateCodeSpaceInvitationJWT(token string) (*CodeSpaceInvitationJWTClaims, bool)
}

// crypto implements Crypto.
type crypto struct {
	timeProvider timekeeper.Provider
	secretKey    string
}

// NewCrypto returns a new crypto.
func NewCrypto(timeProvider timekeeper.Provider, secretKey string) *crypto {
	return &crypto{
		timeProvider: timeProvider,
		secretKey:    secretKey,
	}
}

// HashPassword hashes a given password.
func (c *crypto) HashPassword(password string) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), HashingCost)
	if err != nil {
		return "", errutils.FormatError(err, "bcrypt.GenerateFromPassword failed")
	}
	hashedPassword := string(hashedPasswordBytes)

	return hashedPassword, nil
}

// CheckPassword checks if a given hashed password matches a given plaintext password.
func (c *crypto) CheckPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	return err == nil
}

// CreateAuthJWT creates JWTs for User authentication of given type.
// Returns error when token type is not access or refresh.
func (c *crypto) CreateAuthJWT(userUUID string, tokenType JWTType) (string, error) {
	var lifetime time.Duration
	switch tokenType {
	case JWTTypeAccess:
		lifetime = JWTLifetimeAccess
	case JWTTypeRefresh:
		lifetime = JWTLifetimeRefresh
	default:
		return "", errutils.FormatErrorf(
			nil,
			"invalid JWT type %s, expected JWT type %s or %s",
			tokenType,
			JWTTypeAccess,
			JWTTypeRefresh,
		)
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(tokenType),
			IssuedAt:  jsonutils.UnixTimestamp(c.timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(c.timeProvider.Now().Add(lifetime)),
			JWTID:     uuid.NewString(),
		},
	)
	signedToken, err := token.SignedString([]byte(c.secretKey))
	if err != nil {
		return "", errutils.FormatErrorf(
			err,
			"jwt.Token.SignedString failed for user.UUID %s of token type %s",
			userUUID,
			tokenType,
		)
	}

	return signedToken, nil
}

// ValidateAuthJWT validates JWT for user authentication,
// checks that the JWT is not expired,
// and returns parsed JWT claims.
func (c *crypto) ValidateAuthJWT(token string, tokenType JWTType) (*AuthJWTClaims, bool) {
	claims := &AuthJWTClaims{}
	ok := true

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(c.secretKey), nil
	})
	if err != nil {
		ok = false
	}

	if parsedToken == nil || !parsedToken.Valid {
		ok = false
	}

	if subtle.ConstantTimeCompare([]byte(claims.TokenType), []byte(tokenType)) == 0 {
		ok = false
	}

	if c.timeProvider.Now().After(time.Time(claims.ExpiresAt)) {
		ok = false
	}

	if !ok {
		return nil, false
	}

	return claims, true
}

// CreateActivationJWT creates JWT for user activation.
func (c *crypto) CreateActivationJWT(userUUID string) (string, error) {
	now := c.timeProvider.Now()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(JWTTypeActivation),
			IssuedAt:  jsonutils.UnixTimestamp(now),
			ExpiresAt: jsonutils.UnixTimestamp(now.Add(JWTLifetimeActivation)),
			JWTID:     uuid.NewString(),
		},
	)
	signedToken, err := token.SignedString([]byte(c.secretKey))
	if err != nil {
		return "", errutils.FormatErrorf(
			err,
			"jwt.Token.SignedString failed for user.UUID %s of token type %s",
			userUUID,
			JWTTypeActivation,
		)
	}

	return signedToken, nil
}

// ValidateActivationJWT validates JWT for user activation using secret key,
// checks that the JWT is not expired, and returns parsed JWT claims.
func (c *crypto) ValidateActivationJWT(token string) (*ActivationJWTClaims, bool) {
	claims := &ActivationJWTClaims{}
	ok := true

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(c.secretKey), nil
	})
	if err != nil {
		ok = false
	}

	if parsedToken == nil || !parsedToken.Valid {
		ok = false
	}

	if subtle.ConstantTimeCompare([]byte(claims.TokenType), []byte(JWTTypeActivation)) == 0 {
		ok = false
	}

	if c.timeProvider.Now().After(time.Time(claims.ExpiresAt)) {
		ok = false
	}

	if !ok {
		return nil, false
	}

	return claims, true
}

// CreateAPIKey creates prefix, secret, and hashed key for API key.
func (c *crypto) CreateAPIKey() (string, string, string, error) {
	prefix, err := random.GenerateString(APIKeyPrefixLength, true, true, true)
	if err != nil {
		return "", "", "", errutils.FormatError(err)
	}

	secretBytes := make([]byte, APIKeySecretNBytes)
	_, err = rand.Read(secretBytes)
	if err != nil {
		return "", "", "", errutils.FormatError(err)
	}

	secret := base64.StdEncoding.EncodeToString(secretBytes)

	rawKey := fmt.Sprintf("%s.%s", prefix, secret)

	hashedKey, err := c.HashPassword(rawKey)
	if err != nil {
		return "", "", "", errutils.FormatError(err)
	}

	return prefix, rawKey, hashedKey, nil
}

// ParseAPIKey parses API key and returns prefix and secret.
func (c *crypto) ParseAPIKey(key string) (string, string, error) {
	prefix, secret, ok := strings.Cut(key, ".")
	if !ok {
		return "", "", errutils.FormatErrorf(nil, "failed to parse prefix and secret for API key %s", key)
	}

	return prefix, secret, nil
}

// CreateCodeSpaceInvitationJWT creates JWT for code space invitation.
func (c *crypto) CreateCodeSpaceInvitationJWT(
	userUUID string,
	inviteeEmail string,
	codeSpaceID int64,
	accessLevel int,
) (string, error) {
	now := c.timeProvider.Now()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&CodeSpaceInvitationJWTClaims{
			Subject:      userUUID,
			InviteeEmail: inviteeEmail,
			CodeSpaceID:  codeSpaceID,
			AccessLevel:  accessLevel,
			TokenType:    string(JWTTypeCodeSpaceInvitation),
			IssuedAt:     jsonutils.UnixTimestamp(now),
			ExpiresAt:    jsonutils.UnixTimestamp(now.Add(JWTLifetimeCodeSpaceInvitation)),
			JWTID:        uuid.NewString(),
		},
	)
	signedToken, err := token.SignedString([]byte(c.secretKey))
	if err != nil {
		return "", errutils.FormatErrorf(
			err,
			"jwt.Token.SignedString failed for user.UUID %s of token type %s",
			userUUID,
			JWTTypeCodeSpaceInvitation,
		)
	}

	return signedToken, nil
}

// ValidateCodeSpaceInvitationJWT validates JWT for code space invitation using secret key,
// checks that the JWT is not expired, and returns parsed JWT claims.
func (c *crypto) ValidateCodeSpaceInvitationJWT(token string) (*CodeSpaceInvitationJWTClaims, bool) {
	claims := &CodeSpaceInvitationJWTClaims{}
	ok := true

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(c.secretKey), nil
	})
	if err != nil {
		ok = false
	}

	if parsedToken == nil || !parsedToken.Valid {
		ok = false
	}

	if subtle.ConstantTimeCompare([]byte(claims.TokenType), []byte(JWTTypeCodeSpaceInvitation)) == 0 {
		ok = false
	}

	if c.timeProvider.Now().After(time.Time(claims.ExpiresAt)) {
		ok = false
	}

	if !ok {
		return nil, false
	}

	return claims, true
}
