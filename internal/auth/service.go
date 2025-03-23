package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alvii147/nymphadora-api/internal/config"
	"github.com/alvii147/nymphadora-api/internal/templatesmanager"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/logging"
	"github.com/alvii147/nymphadora-api/pkg/mailclient"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FrontendActivationRoute is the frontend route for user activation.
const FrontendActivationRoute = "/signup/activate/%s"

// Service performs all auth-related business logic.
//
//go:generate mockgen -package=authmocks -source=$GOFILE -destination=./mocks/service.go
type Service interface {
	SendUserActivationMail(
		ctx context.Context,
		email string,
		data templatesmanager.ActivationEmailTemplateData,
	) error
	CreateUser(
		ctx context.Context,
		wg *sync.WaitGroup,
		email string,
		password string,
		firstName string,
		lastName string,
	) (*User, error)
	ActivateUser(
		ctx context.Context,
		token string,
	) error
	GetAuthenticatedUser(
		ctx context.Context,
	) (*User, error)
	UpdateAuthenticatedUser(
		ctx context.Context,
		firstName *string,
		lastName *string,
	) (*User, error)
	CreateJWT(
		ctx context.Context,
		email string,
		password string,
	) (string, string, error)
	RefreshJWT(
		ctx context.Context,
		token string,
	) (string, error)
	ValidateJWT(
		ctx context.Context,
		token string,
	) bool
	CreateAPIKey(ctx context.Context,
		name string,
		expiresAt *time.Time,
	) (*APIKey, string, error)
	ListAPIKeys(
		ctx context.Context,
	) ([]*APIKey, error)
	FindAPIKey(
		ctx context.Context,
		rawKey string,
	) (*APIKey, error)
	UpdateAPIKey(
		ctx context.Context,
		apiKeyID int64,
		name *string,
		expiresAt jsonutils.Optional[time.Time],
	) (*APIKey, error)
	DeleteAPIKey(
		ctx context.Context,
		apiKeyID int64,
	) error
}

// service implements Service.
type service struct {
	config       *config.Config
	timeProvider timekeeper.Provider
	dbPool       *pgxpool.Pool
	logger       logging.Logger
	crypto       cryptocore.Crypto
	mailClient   mailclient.Client
	tmplManager  templatesmanager.Manager
	repository   Repository
}

// NewService returns a new service.
func NewService(
	config *config.Config,
	timeProvider timekeeper.Provider,
	dbPool *pgxpool.Pool,
	logger logging.Logger,
	crypto cryptocore.Crypto,
	mailClient mailclient.Client,
	tmplManager templatesmanager.Manager,
	repo Repository,
) *service {
	return &service{
		config:       config,
		timeProvider: timeProvider,
		dbPool:       dbPool,
		logger:       logger,
		crypto:       crypto,
		mailClient:   mailClient,
		tmplManager:  tmplManager,
		repository:   repo,
	}
}

// SendUserActivationMail sends the activation email.
func (svc *service) SendUserActivationMail(
	ctx context.Context,
	email string,
	data templatesmanager.ActivationEmailTemplateData,
) error {
	textTmpl, htmlTmpl, err := svc.tmplManager.Load(templatesmanager.ActivationEmailTemplateName)
	if err != nil {
		return errutils.FormatError(err)
	}

	err = svc.mailClient.Send([]string{email}, templatesmanager.ActivationEmailSubject, textTmpl, htmlTmpl, data)
	if err != nil {
		return errutils.FormatError(err)
	}

	return nil
}

// CreateUser creates a new user.
func (svc *service) CreateUser(
	ctx context.Context,
	wg *sync.WaitGroup,
	email string,
	password string,
	firstName string,
	lastName string,
) (*User, error) {
	hashedPassword, err := svc.crypto.HashPassword(password)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	user := &User{
		UUID:        uuid.NewString(),
		Email:       email,
		Password:    hashedPassword,
		FirstName:   firstName,
		LastName:    lastName,
		IsActive:    false,
		IsSuperUser: false,
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	user, err = svc.repository.CreateUser(ctx, dbConn, user)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseUniqueViolation):
			err = errutils.FormatError(errutils.ErrUserAlreadyExists)
		default:
			err = errutils.FormatError(err)
		}

		return nil, err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		token, err := svc.crypto.CreateActivationJWT(user.UUID)
		if err != nil {
			svc.logger.LogError(errutils.FormatError(err))

			return
		}

		activationURL := fmt.Sprintf(svc.config.FrontendBaseURL+FrontendActivationRoute, token)
		data := templatesmanager.ActivationEmailTemplateData{
			RecipientEmail: user.Email,
			ActivationURL:  activationURL,
		}

		err = svc.SendUserActivationMail(
			ctx,
			user.Email,
			data,
		)
		if err != nil {
			svc.logger.LogError(errutils.FormatError(err))
		}
	}()

	return user, nil
}

// ActivateUser activates a user from a given activation JWT.
func (svc *service) ActivateUser(
	ctx context.Context,
	token string,
) error {
	claims, ok := svc.crypto.ValidateActivationJWT(token)
	if !ok {
		return errutils.FormatErrorf(errutils.ErrInvalidToken, "token %s", token)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	err = svc.repository.ActivateUserByUUID(ctx, dbConn, claims.Subject)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsAffected):
			err = errutils.FormatError(errutils.ErrUserNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return err
	}

	return nil
}

// GetAuthenticatedUser gets the currently authenticated user.
func (svc *service) GetAuthenticatedUser(
	ctx context.Context,
) (*User, error) {
	userUUID, err := GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	user, err := svc.repository.GetUserByUUID(ctx, dbConn, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrUserNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return nil, err
	}

	return user, nil
}

// UpdateAuthenticatedUser updates the current user.
func (svc *service) UpdateAuthenticatedUser(
	ctx context.Context,
	firstName *string,
	lastName *string,
) (*User, error) {
	userUUID, err := GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	user, err := svc.repository.UpdateUser(ctx, dbConn, userUUID, firstName, lastName)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsAffected):
			err = errutils.FormatError(errutils.ErrUserNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return nil, err
	}

	return user, nil
}

// CreateJWT authenticates a user and creates new access and refresh JWTs.
func (svc *service) CreateJWT(
	ctx context.Context,
	email string,
	password string,
) (string, string, error) {
	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return "", "", errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	user, err := svc.repository.GetUserByEmail(ctx, dbConn, email)
	if err != nil && !errors.Is(err, errutils.ErrDatabaseNoRowsReturned) {
		return "", "", errutils.FormatError(err)
	}

	dummyUser := &User{}
	dummyUser.UUID = "93c58b3c-f087-4e97-805a-1e4676cdd5ec"
	dummyUser.Password = "$2a$14$078K5mTZHgbMDQ.K/U656OEb7v5HIyB9cBPLoXiQREAoXmCmgsywW"
	failAuth := false

	if err != nil || user == nil || !user.IsActive {
		failAuth = true
		user = dummyUser
	}

	ok := svc.crypto.CheckPassword(user.Password, password)
	if !ok {
		failAuth = true
		user = dummyUser
	}

	accessToken, err := svc.crypto.CreateAuthJWT(
		user.UUID,
		cryptocore.JWTTypeAccess,
	)
	if err != nil {
		return "", "", errutils.FormatErrorf(err, "token type %s", cryptocore.JWTTypeAccess)
	}

	refreshToken, err := svc.crypto.CreateAuthJWT(
		user.UUID,
		cryptocore.JWTTypeRefresh,
	)
	if err != nil {
		return "", "", errutils.FormatErrorf(err, "token type %s", cryptocore.JWTTypeRefresh)
	}

	if failAuth {
		return "", "", errutils.FormatError(errutils.ErrInvalidCredentials)
	}

	return accessToken, refreshToken, nil
}

// RefreshJWT validates a refresh JWT and creates a new access JWT.
func (svc *service) RefreshJWT(
	ctx context.Context,
	token string,
) (string, error) {
	claims, ok := svc.crypto.ValidateAuthJWT(token, cryptocore.JWTTypeRefresh)
	if !ok {
		return "", errutils.FormatErrorf(errutils.ErrInvalidToken, "token %s", token)
	}

	accessToken, err := svc.crypto.CreateAuthJWT(
		claims.Subject,
		cryptocore.JWTTypeAccess,
	)
	if err != nil {
		return "", errutils.FormatErrorf(err, "token type %s", cryptocore.JWTTypeAccess)
	}

	return accessToken, nil
}

// ValidateJWT validates an access JWT.
func (svc *service) ValidateJWT(
	ctx context.Context,
	token string,
) bool {
	_, ok := svc.crypto.ValidateAuthJWT(token, cryptocore.JWTTypeAccess)

	return ok
}

// CreateAPIKey creates new API key.
func (svc *service) CreateAPIKey(
	ctx context.Context,
	name string,
	expiresAt *time.Time,
) (*APIKey, string, error) {
	userUUID, err := GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, "", errutils.FormatError(err)
	}

	prefix, rawKey, hashedKey, err := svc.crypto.CreateAPIKey()
	if err != nil {
		return nil, "", errutils.FormatError(err)
	}

	apiKey := &APIKey{
		UserUUID:  userUUID,
		Prefix:    prefix,
		HashedKey: hashedKey,
		Name:      name,
		ExpiresAt: expiresAt,
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, "", errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	apiKey, err = svc.repository.CreateAPIKey(ctx, dbConn, apiKey)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseUniqueViolation):
			err = errutils.FormatError(errutils.ErrAPIKeyAlreadyExists)
		default:
			err = errutils.FormatError(err)
		}

		return nil, "", err
	}

	return apiKey, rawKey, nil
}

// ListAPIKeys retrieves API keys for the currently authenticated user.
func (svc *service) ListAPIKeys(
	ctx context.Context,
) ([]*APIKey, error) {
	userUUID, err := GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	apiKeys, err := svc.repository.ListAPIKeysByUserUUID(ctx, dbConn, userUUID)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	return apiKeys, nil
}

// FindAPIKey parses and finds an API Key.
func (svc *service) FindAPIKey(
	ctx context.Context,
	rawKey string,
) (*APIKey, error) {
	prefix, _, err := svc.crypto.ParseAPIKey(rawKey)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	apiKeys, err := svc.repository.ListActiveAPIKeysByPrefix(ctx, dbConn, prefix)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	for _, apiKey := range apiKeys {
		ok := svc.crypto.CheckPassword(apiKey.HashedKey, rawKey)
		if ok {
			return apiKey, nil
		}
	}

	return nil, errutils.FormatError(errutils.ErrAPIKeyNotFound)
}

// UpdateAPIKey updates an API Key.
func (svc *service) UpdateAPIKey(
	ctx context.Context,
	apiKeyID int64,
	name *string,
	expiresAt jsonutils.Optional[time.Time],
) (*APIKey, error) {
	userUUID, err := GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	apiKey, err := svc.repository.UpdateAPIKey(
		ctx,
		dbConn,
		userUUID,
		apiKeyID,
		name,
		expiresAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsAffected):
			err = errutils.FormatError(errutils.ErrAPIKeyNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return nil, err
	}

	return apiKey, nil
}

// DeleteAPIKey deletes an API key.
func (svc *service) DeleteAPIKey(
	ctx context.Context,
	apiKeyID int64,
) error {
	userUUID, err := GetUserUUIDFromContext(ctx)
	if err != nil {
		return errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	err = svc.repository.DeleteAPIKey(ctx, dbConn, userUUID, apiKeyID)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsAffected):
			err = errutils.FormatError(errutils.ErrAPIKeyNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return err
	}

	return nil
}
