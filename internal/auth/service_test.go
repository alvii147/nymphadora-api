package auth_test

import (
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/alvii147/nymphadora-api/internal/auth"
	authmocks "github.com/alvii147/nymphadora-api/internal/auth/mocks"
	databasemocks "github.com/alvii147/nymphadora-api/internal/database/mocks"
	"github.com/alvii147/nymphadora-api/internal/templatesmanager"
	templatesmanagermocks "github.com/alvii147/nymphadora-api/internal/templatesmanager/mocks"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	cryptocoremocks "github.com/alvii147/nymphadora-api/pkg/cryptocore/mocks"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	mailclientmocks "github.com/alvii147/nymphadora-api/pkg/mailclient/mocks"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestServiceSendUserActivationMailSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	email := testkit.GenerateFakeEmail()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	mailClient := testkit.NewInMemMailClient("support@nymphadora.com", timeProvider)
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	tmplManager := templatesmanager.NewManager()
	repo := authmocks.NewMockRepository(ctrl)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	activationURL := "http://localhost:3000/signup/activate/4ct1v4t10njwt"
	err := svc.SendUserActivationMail(
		context.Background(),
		email,
		templatesmanager.ActivationEmailTemplateData{
			RecipientEmail: email,
			ActivationURL:  activationURL,
		},
	)
	require.NoError(t, err)
	require.Len(t, mailClient.Logs, 1)

	lastMail := mailClient.Logs[len(mailClient.Logs)-1]
	require.Equal(t, []string{email}, lastMail.To)
	require.Equal(t, "Welcome to Nymphadora!", lastMail.Subject)
	require.WithinDuration(t, timeProvider.Now(), lastMail.SentAt, testkit.TimeToleranceExact)

	mailMessage := string(lastMail.Message)
	require.Contains(t, mailMessage, "Welcome to Nymphadora!")
	require.Contains(t, mailMessage, "Nymphadora - Activate Your Account")
	require.Contains(t, mailMessage, activationURL)
}

func TestServiceSendUserActivationMailError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	activationURL := "http://localhost:3000/signup/activate/4ct1v4t10njwt"
	email := testkit.GenerateFakeEmail()
	tmplLoadErr := errors.New("Load failed")
	mailSendErr := errors.New("Send failed")

	testcases := map[string]struct {
		tmplLoadErr error
		mailSendErr error
		wantErr     error
	}{
		"template Load fails": {
			tmplLoadErr: tmplLoadErr,
			mailSendErr: nil,
			wantErr:     tmplLoadErr,
		},
		"mail Send fails": {
			tmplLoadErr: nil,
			mailSendErr: mailSendErr,
			wantErr:     mailSendErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			mailClient := mailclientmocks.NewMockClient(ctrl)
			dbPool := databasemocks.NewMockPool(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)
			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			tmplManager.
				EXPECT().
				Load("activation").
				Return(texttemplate.New("text"), htmltemplate.New("html"), testcase.tmplLoadErr).
				MaxTimes(1)

			mailClient.
				EXPECT().
				Send(
					[]string{email},
					"Welcome to Nymphadora!",
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(testcase.mailSendErr).
				MaxTimes(1)

			err := svc.SendUserActivationMail(
				context.Background(),
				email,
				templatesmanager.ActivationEmailTemplateData{
					RecipientEmail: email,
					ActivationURL:  activationURL,
				},
			)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceCreateUserSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	timeProvider := timekeeper.NewFrozenProvider()
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := testkit.NewInMemMailClient("support@nymphadora.com", timeProvider)
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	email := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()
	firstName := testkit.MustGenerateRandomString(8, true, true, false)
	lastName := testkit.MustGenerateRandomString(8, true, true, false)

	mailCount := len(mailClient.Logs)

	var wg sync.WaitGroup
	user, err := svc.CreateUser(context.Background(), &wg, email, password, firstName, lastName)
	require.NoError(t, err)

	require.Equal(t, email, user.Email)
	require.Equal(t, firstName, user.FirstName)
	require.Equal(t, lastName, user.LastName)
	require.False(t, user.IsActive)
	require.False(t, user.IsSuperUser)
	require.WithinDuration(t, timeProvider.Now(), user.CreatedAt, testkit.TimeToleranceExact)
	require.WithinDuration(t, timeProvider.Now(), user.UpdatedAt, testkit.TimeToleranceExact)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	require.NoError(t, err)

	wg.Wait()

	require.Len(t, mailClient.Logs, mailCount+1)

	lastMail := mailClient.Logs[len(mailClient.Logs)-1]
	require.Equal(t, []string{user.Email}, lastMail.To)
	require.Equal(t, "Welcome to Nymphadora!", lastMail.Subject)
	require.WithinDuration(t, timeProvider.Now(), lastMail.SentAt, testkit.TimeToleranceExact)

	mailMessage := string(lastMail.Message)
	require.Contains(t, mailMessage, "Welcome to Nymphadora!")
	require.Contains(t, mailMessage, "Nymphadora - Activate Your Account")

	pattern := fmt.Sprintf(cfg.FrontendBaseURL+auth.FrontendActivationRoute, `(\S+)`)
	r, err := regexp.Compile(pattern)
	require.NoError(t, err)

	matches := r.FindStringSubmatch(mailMessage)
	require.Len(t, matches, 2)

	activationToken := matches[1]
	claims := &cryptocore.ActivationJWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(activationToken, claims, func(t *jwt.Token) (any, error) {
		return []byte(cfg.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedToken)
	require.True(t, parsedToken.Valid)
	require.Equal(t, user.UUID, claims.Subject)
	require.Equal(t, string(cryptocore.JWTTypeActivation), claims.TokenType)
	require.WithinDuration(t, timeProvider.Now(), time.Time(claims.IssuedAt), testkit.TimeToleranceExact)
	require.WithinDuration(
		t,
		timeProvider.Now().Add(cryptocore.JWTLifetimeActivation),
		time.Time(claims.ExpiresAt),
		testkit.TimeToleranceExact,
	)
}

func TestServiceCreateUserEmailExists(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := authmocks.NewMockRepository(ctrl)

	user := &auth.User{
		UUID:        uuid.NewString(),
		Email:       testkit.GenerateFakeEmail(),
		Password:    testkitinternal.MustHashPassword(testkit.GenerateFakePassword()),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    false,
		IsSuperUser: false,
	}

	repo.
		EXPECT().
		CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(user, errutils.ErrDatabaseUniqueViolation).
		Times(1)

	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	var wg sync.WaitGroup
	_, err := svc.CreateUser(
		context.Background(),
		&wg,
		user.Email,
		testkit.GenerateFakePassword(),
		testkit.MustGenerateRandomString(8, true, true, false),
		testkit.MustGenerateRandomString(8, true, true, false),
	)
	require.ErrorIs(t, err, errutils.ErrUserAlreadyExists)
}

func TestServiceCreateUserError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	password := testkit.GenerateFakePassword()
	hashedPassword := testkitinternal.MustHashPassword(password)
	hashPasswordErr := errors.New("HashPassword failed")
	genericRepoErr := errors.New("CreateUser failed")

	testcases := map[string]struct {
		hashPasswordErr error
		repoErr         error
		wantErr         error
	}{
		"HashPassword fails": {
			hashPasswordErr: hashPasswordErr,
			repoErr:         nil,
			wantErr:         hashPasswordErr,
		},
		"Generic repo error": {
			hashPasswordErr: nil,
			repoErr:         genericRepoErr,
			wantErr:         genericRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			user := &auth.User{
				UUID:        uuid.NewString(),
				Email:       testkit.GenerateFakeEmail(),
				Password:    hashedPassword,
				FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
				LastName:    testkit.MustGenerateRandomString(8, true, true, false),
				IsActive:    false,
				IsSuperUser: false,
			}

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			crypto.
				EXPECT().
				HashPassword(password).
				Return(hashedPassword, testcase.hashPasswordErr).
				MaxTimes(1)

			repo.
				EXPECT().
				CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(user, testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			var wg sync.WaitGroup
			_, err := svc.CreateUser(
				context.Background(),
				&wg,
				user.Email,
				password,
				user.FirstName,
				user.LastName,
			)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceCreateUserEmailSendFails(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	password := testkit.GenerateFakePassword()
	hashedPassword := testkitinternal.MustHashPassword(password)

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	mailClient := mailclientmocks.NewMockClient(ctrl)
	dbPool := databasemocks.NewMockPool(ctrl)
	dbConn := databasemocks.NewMockConn(ctrl)
	_, bufErr, logger := testkit.CreateInMemLogger()
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	tmplManager := templatesmanager.NewManager()
	repo := authmocks.NewMockRepository(ctrl)

	user := &auth.User{
		UUID:        userUUID,
		Email:       testkit.GenerateFakeEmail(),
		Password:    hashedPassword,
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    false,
		IsSuperUser: false,
	}

	dbConn.
		EXPECT().
		Release().
		MaxTimes(1)

	dbPool.
		EXPECT().
		Acquire(gomock.Any()).
		Return(dbConn, nil).
		MaxTimes(1)

	crypto.
		EXPECT().
		HashPassword(password).
		Return(hashedPassword, nil).
		Times(1)

	crypto.
		EXPECT().
		CreateActivationJWT(userUUID).
		Return("DEADBEEF", nil).
		MaxTimes(1)

	repo.
		EXPECT().
		CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(user, nil).
		Times(1)

	mailClient.
		EXPECT().
		Send(
			[]string{user.Email},
			"Welcome to Nymphadora!",
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Return(errors.New("Send failed")).
		Times(1)

	svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

	var wg sync.WaitGroup
	_, err := svc.CreateUser(
		context.Background(),
		&wg,
		user.Email,
		password,
		user.FirstName,
		user.LastName,
	)
	require.NoError(t, err)

	wg.Wait()

	logErrorMessages := strings.Split(strings.TrimSpace(bufErr.String()), "\n")
	require.Len(t, logErrorMessages, 1)

	stdErrMessage := logErrorMessages[0]
	logLevel, logTime, _, logMsg := testkit.MustParseLogMessage(stdErrMessage)
	require.Equal(t, "E", logLevel)
	require.WithinDuration(t, logTime, timeProvider.Now(), testkit.TimeToleranceTentative)
	require.Contains(t, logMsg, "Send failed")
}

func TestServiceActivateUserSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	timeProvider.AddDate(0, 0, 1)

	dbConn, err := TestDBPool.Acquire(context.Background())
	require.NoError(t, err)
	defer dbConn.Release()

	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	jti := uuid.NewString()
	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.ActivationJWTClaims{
			Subject:   user.UUID,
			TokenType: string(cryptocore.JWTTypeActivation),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	err = svc.ActivateUser(context.Background(), token)
	require.NoError(t, err)

	activatedUser, err := repo.GetUserByEmail(context.Background(), dbConn, user.Email)
	require.NoError(t, err)

	require.Equal(t, user.Email, activatedUser.Email)
	require.Equal(t, user.Password, activatedUser.Password)
	require.Equal(t, user.FirstName, activatedUser.FirstName)
	require.Equal(t, user.LastName, activatedUser.LastName)
	require.True(t, activatedUser.IsActive)
	require.Equal(t, user.IsSuperUser, activatedUser.IsSuperUser)
	require.Equal(t, user.CreatedAt, activatedUser.CreatedAt)
	require.WithinDuration(t, timeProvider.Now(), activatedUser.UpdatedAt, testkit.TimeToleranceExact)
}

func TestServiceActivateUserError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	timeProvider := timekeeper.NewFrozenProvider()
	userUUID := uuid.NewString()
	genericRepoErr := errors.New("ActivateUserByUUID failed")

	invalidToken := "ed0730889507fdb8549acfcd31548ee5"
	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeActivation),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now().Add(-2 * time.Hour)),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(-time.Hour)),
			JWTID:     uuid.NewString(),
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)
	validToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeActivation),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now()),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(time.Hour)),
			JWTID:     uuid.NewString(),
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	testcases := map[string]struct {
		token   string
		repoErr error
		wantErr error
	}{
		"Invalid token": {
			token:   invalidToken,
			repoErr: nil,
			wantErr: errutils.ErrInvalidToken,
		},
		"Expired token": {
			token:   expiredToken,
			repoErr: nil,
			wantErr: errutils.ErrInvalidToken,
		},
		"User not found": {
			repoErr: errutils.ErrDatabaseNoRowsAffected,
			token:   validToken,
			wantErr: errutils.ErrUserNotFound,
		},
		"Activate user fails": {
			repoErr: genericRepoErr,
			token:   validToken,
			wantErr: genericRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			repo.
				EXPECT().
				ActivateUserByUUID(gomock.Any(), gomock.Any(), userUUID).
				Return(testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			err = svc.ActivateUser(context.Background(), testcase.token)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceGetAuthenticatedUserSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
	fetchedUser, err := svc.GetAuthenticatedUser(ctx)
	require.NoError(t, err)

	require.Equal(t, user.Email, fetchedUser.Email)
	require.Equal(t, user.Password, fetchedUser.Password)
	require.Equal(t, user.FirstName, fetchedUser.FirstName)
	require.Equal(t, user.LastName, fetchedUser.LastName)
	require.Equal(t, user.IsActive, fetchedUser.IsActive)
	require.Equal(t, user.IsSuperUser, fetchedUser.IsSuperUser)
	require.Equal(t, user.CreatedAt, fetchedUser.CreatedAt)
	require.Equal(t, user.UpdatedAt, fetchedUser.UpdatedAt)
}

func TestServiceGetAuthenticatedUserError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	genericRepoErr := errors.New("GetUserByUUID failed")

	testcases := map[string]struct {
		ctx     context.Context
		repoErr error
		wantErr error
	}{
		"No user UUID in context": {
			ctx:     context.Background(),
			repoErr: nil,
			wantErr: nil,
		},
		"User not found": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: errutils.ErrDatabaseNoRowsReturned,
			wantErr: errutils.ErrUserNotFound,
		},
		"Generic repo error": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: genericRepoErr,
			wantErr: genericRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			repo.
				EXPECT().
				GetUserByUUID(gomock.Any(), gomock.Any(), userUUID).
				Return(nil, testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			_, err := svc.GetAuthenticatedUser(testcase.ctx)
			require.Error(t, err)
			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceUpdateAuthenticatedUserSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	startingFirstName := "startingFirstName"
	startingLastName := "startingLastName"
	updatedFirstName := "updatedFirstName"
	updatedLastName := "updatedLastName"

	testcases := map[string]struct {
		firstName     *string
		lastName      *string
		wantFirstName string
		wantLastName  string
	}{
		"Update both first and last names": {
			firstName:     &updatedFirstName,
			lastName:      &updatedLastName,
			wantFirstName: updatedFirstName,
			wantLastName:  updatedLastName,
		},
		"Update only first name": {
			firstName:     &updatedFirstName,
			lastName:      nil,
			wantFirstName: updatedFirstName,
			wantLastName:  startingLastName,
		},
		"Update only last name": {
			firstName:     nil,
			lastName:      &updatedLastName,
			wantFirstName: startingFirstName,
			wantLastName:  updatedLastName,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			timeProvider.AddDate(0, 0, 1)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := auth.NewRepository(timeProvider)

			svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

			user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
				u.FirstName = startingFirstName
				u.LastName = startingLastName
				u.IsActive = true
			})
			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)

			updatedUser, err := svc.UpdateAuthenticatedUser(ctx, testcase.firstName, testcase.lastName)
			require.NoError(t, err)

			require.Equal(t, user.UUID, updatedUser.UUID)
			require.Equal(t, testcase.wantFirstName, updatedUser.FirstName)
			require.Equal(t, testcase.wantLastName, updatedUser.LastName)
			require.WithinDuration(t, timeProvider.Now(), updatedUser.UpdatedAt, testkit.TimeToleranceExact)
		})
	}
}

func TestServiceUpdateAuthenticatedUserError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	updatedFirstName := "updatedFirstName"
	updatedLastName := "updatedLastName"
	genericRepoErr := errors.New("UpdateUser failed")

	testcases := map[string]struct {
		ctx     context.Context
		repoErr error
		wantErr error
	}{
		"No user UUID in context": {
			ctx:     context.Background(),
			repoErr: nil,
			wantErr: nil,
		},
		"User not found": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: errutils.ErrDatabaseNoRowsAffected,
			wantErr: errutils.ErrUserNotFound,
		},
		"Generic repo error": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: genericRepoErr,
			wantErr: genericRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			repo.
				EXPECT().
				UpdateUser(gomock.Any(), gomock.Any(), userUUID, &updatedFirstName, &updatedLastName).
				Return(nil, testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			_, err := svc.UpdateAuthenticatedUser(testcase.ctx, &updatedFirstName, &updatedLastName)
			require.Error(t, err)
			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceCreateJWTSuccess(t *testing.T) {
	t.Parallel()

	user, password := testkitinternal.MustCreateUser(t, func(u *auth.User) {
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

	accessToken, refreshToken, err := svc.CreateJWT(context.Background(), user.Email, password)
	require.NoError(t, err)

	accessClaims := &cryptocore.AuthJWTClaims{}
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, accessClaims, func(t *jwt.Token) (any, error) {
		return []byte(cfg.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedAccessToken)
	require.True(t, parsedAccessToken.Valid)
	require.Equal(t, user.UUID, accessClaims.Subject)
	require.Equal(t, string(cryptocore.JWTTypeAccess), accessClaims.TokenType)
	require.WithinDuration(t, timeProvider.Now(), time.Time(accessClaims.IssuedAt), testkit.TimeToleranceExact)
	require.WithinDuration(
		t,
		timeProvider.Now().Add(cryptocore.JWTLifetimeAccess),
		time.Time(accessClaims.ExpiresAt),
		testkit.TimeToleranceExact,
	)

	refreshClaims := &cryptocore.AuthJWTClaims{}
	parsedRefreshToken, err := jwt.ParseWithClaims(refreshToken, refreshClaims, func(t *jwt.Token) (any, error) {
		return []byte(cfg.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedRefreshToken)
	require.True(t, parsedRefreshToken.Valid)
	require.Equal(t, user.UUID, refreshClaims.Subject)
	require.Equal(t, string(cryptocore.JWTTypeRefresh), refreshClaims.TokenType)
	require.WithinDuration(t, timeProvider.Now(), time.Time(refreshClaims.IssuedAt), testkit.TimeToleranceExact)
	require.WithinDuration(
		t,
		timeProvider.Now().Add(cryptocore.JWTLifetimeRefresh),
		time.Time(refreshClaims.ExpiresAt),
		testkit.TimeToleranceExact,
	)
}

func TestServiceCreateJWTIncorrectCredentials(t *testing.T) {
	t.Parallel()

	user, password := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	cfg := testkitinternal.MustCreateConfig()

	testcases := map[string]struct {
		email    string
		password string
	}{
		"Incorrect email": {
			email:    testkit.GenerateFakeEmail(),
			password: password,
		},
		"Incorrect password": {
			email:    user.Email,
			password: testkit.GenerateFakePassword(),
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := auth.NewRepository(timeProvider)
			svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

			_, _, err := svc.CreateJWT(context.Background(), testcase.email, testcase.password)
			require.ErrorIs(t, err, errutils.ErrInvalidCredentials)
		})
	}
}

func TestServiceCreateJWTError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	email := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()
	hashedPassword := testkitinternal.MustHashPassword(password)
	testkitinternal.MustCreateUserAuthJWTs(userUUID)
	genericRepoErr := errors.New("GetUserByEmail failed")
	createAccessJWTErr := errors.New("CreateAuthJWT failed for access token")
	createRefreshJWTErr := errors.New("CreateAuthJWT failed for refresh token")

	testcases := map[string]struct {
		userIsActive        bool
		passwordCorrect     bool
		repoErr             error
		createAccessJWTErr  error
		createRefreshJWTErr error
		wantErr             error
	}{
		"Inactive user": {
			userIsActive:        false,
			passwordCorrect:     true,
			repoErr:             nil,
			createAccessJWTErr:  nil,
			createRefreshJWTErr: nil,
			wantErr:             errutils.ErrInvalidCredentials,
		},
		"Generic repo error": {
			userIsActive:        true,
			passwordCorrect:     true,
			repoErr:             genericRepoErr,
			createAccessJWTErr:  nil,
			createRefreshJWTErr: nil,
			wantErr:             genericRepoErr,
		},
		"Incorrect password": {
			userIsActive:        true,
			passwordCorrect:     false,
			repoErr:             nil,
			createAccessJWTErr:  nil,
			createRefreshJWTErr: nil,
			wantErr:             errutils.ErrInvalidCredentials,
		},
		"CreateAuthJWT fails for access token": {
			userIsActive:        true,
			passwordCorrect:     true,
			repoErr:             nil,
			createAccessJWTErr:  createAccessJWTErr,
			createRefreshJWTErr: nil,
			wantErr:             createAccessJWTErr,
		},
		"CreateAuthJWT fails for refresh token": {
			userIsActive:        true,
			passwordCorrect:     true,
			repoErr:             nil,
			createAccessJWTErr:  nil,
			createRefreshJWTErr: createRefreshJWTErr,
			wantErr:             createRefreshJWTErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			user := &auth.User{
				UUID:     userUUID,
				Email:    email,
				Password: hashedPassword,
				IsActive: testcase.userIsActive,
			}

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			repo.
				EXPECT().
				GetUserByEmail(gomock.Any(), gomock.Any(), email).
				Return(user, testcase.repoErr).
				Times(1)

			crypto.
				EXPECT().
				CheckPassword(gomock.Any(), password).
				Return(testcase.passwordCorrect).
				MaxTimes(1)

			crypto.
				EXPECT().
				CreateAuthJWT(gomock.Any(), cryptocore.JWTTypeAccess).
				Return("4cc355t0k3n", testcase.createAccessJWTErr).
				MaxTimes(1)

			crypto.
				EXPECT().
				CreateAuthJWT(gomock.Any(), cryptocore.JWTTypeRefresh).
				Return("r3fr35ht0k3n", testcase.createRefreshJWTErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			_, _, err := svc.CreateJWT(context.Background(), email, password)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceRefreshJWTSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := databasemocks.NewMockPool(ctrl)
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := authmocks.NewMockRepository(ctrl)

	svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

	userUUID := uuid.NewString()
	_, refreshToken := testkitinternal.MustCreateUserAuthJWTs(userUUID)

	accessToken, err := svc.RefreshJWT(context.Background(), refreshToken)
	require.NoError(t, err)

	claims := &cryptocore.AuthJWTClaims{}
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (any, error) {
		return []byte(cfg.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedAccessToken)
	require.True(t, parsedAccessToken.Valid)
	require.Equal(t, userUUID, claims.Subject)
	require.Equal(t, string(cryptocore.JWTTypeAccess), claims.TokenType)

	require.WithinDuration(t, timeProvider.Now(), time.Time(claims.IssuedAt), testkit.TimeToleranceExact)
	require.WithinDuration(
		t,
		timeProvider.Now().Add(cryptocore.JWTLifetimeAccess),
		time.Time(claims.ExpiresAt),
		testkit.TimeToleranceExact,
	)
}

func TestServiceRefreshJWTValidateError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := databasemocks.NewMockPool(ctrl)
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := authmocks.NewMockRepository(ctrl)

	svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

	userUUID := uuid.NewString()
	jti := uuid.NewString()

	invalidToken := "ed0730889507fdb8549acfcd31548ee5"
	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&cryptocore.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(cryptocore.JWTTypeRefresh),
			IssuedAt:  jsonutils.UnixTimestamp(timeProvider.Now().Add(-2 * time.Hour)),
			ExpiresAt: jsonutils.UnixTimestamp(timeProvider.Now().Add(-time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(cfg.SecretKey))
	require.NoError(t, err)

	testcases := map[string]struct {
		token string
	}{
		"Invalid refresh token": {
			token: invalidToken,
		},
		"Expired refresh token": {
			token: expiredToken,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := svc.RefreshJWT(context.Background(), testcase.token)
			require.ErrorIs(t, err, errutils.ErrInvalidToken)
		})
	}
}

func TestServiceRefreshJWTError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	accessJWT, refreshJWT := testkitinternal.MustCreateUserAuthJWTs(userUUID)
	createJWTErr := errors.New("CreateAuthJWT failed")

	testcases := map[string]struct {
		validationOk bool
		createJWTErr error
		wantErr      error
	}{
		"ValidateAuthJWT fails": {
			validationOk: false,
			createJWTErr: nil,
			wantErr:      errutils.ErrInvalidToken,
		},
		"CreateAuthJWT": {
			validationOk: true,
			createJWTErr: createJWTErr,
			wantErr:      createJWTErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			crypto.
				EXPECT().
				ValidateAuthJWT(refreshJWT, cryptocore.JWTTypeRefresh).
				Return(
					&cryptocore.AuthJWTClaims{
						Subject:   userUUID,
						TokenType: string(cryptocore.JWTTypeRefresh),
					},
					testcase.validationOk,
				).
				Times(1)

			crypto.
				EXPECT().
				CreateAuthJWT(userUUID, cryptocore.JWTTypeAccess).
				Return(
					accessJWT,
					testcase.createJWTErr,
				).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			_, err := svc.RefreshJWT(context.Background(), refreshJWT)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceValidateJWT(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	timeProvider := timekeeper.NewFrozenProvider()
	userUUID := uuid.NewString()
	jti := uuid.NewString()
	oneDayAgo := timeProvider.Now().Add(-24 * time.Hour)

	ctrl := gomock.NewController(t)
	dbPool := databasemocks.NewMockPool(ctrl)
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := authmocks.NewMockRepository(ctrl)

	svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

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
		token     string
		wantValid bool
	}{
		"Valid access token": {
			token:     validAccessToken,
			wantValid: true,
		},
		"Valid refresh token": {
			token:     validRefreshToken,
			wantValid: false,
		},
		"Invalid secret key": {
			token:     tokenWithInvalidSecretKey,
			wantValid: false,
		},
		"Token of invalid type": {
			token:     tokenOfInvalidType,
			wantValid: false,
		},
		"Invalid token": {
			token:     "ed0730889507fdb8549acfcd31548ee5",
			wantValid: false,
		},
		"Expired token": {
			token:     expiredToken,
			wantValid: false,
		},
		"Token with invalid claim": {
			token:     tokenWithInvalidClaim,
			wantValid: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			valid := svc.ValidateJWT(context.Background(), testcase.token)
			require.Equal(t, testcase.wantValid, valid)
		})
	}
}

func TestServiceCreateAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	_, _, logger := testkit.CreateInMemLogger()
	mailClient := mailclientmocks.NewMockClient(ctrl)
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
	name := "My API Key"
	apiKey, rawKey, err := svc.CreateAPIKey(ctx, name, nil)
	require.NoError(t, err)

	require.NotNil(t, apiKey)
	require.Equal(t, apiKey.UserUUID, user.UUID)
	require.Equal(t, name, apiKey.Name)
	require.Nil(t, apiKey.ExpiresAt)
	require.WithinDuration(t, timeProvider.Now(), apiKey.CreatedAt, testkit.TimeToleranceExact)
	require.WithinDuration(t, timeProvider.Now(), apiKey.UpdatedAt, testkit.TimeToleranceExact)

	err = bcrypt.CompareHashAndPassword([]byte(apiKey.HashedKey), []byte(rawKey))
	require.NoError(t, err)

	r := regexp.MustCompile(`^(\S+)\.(\S+)$`)
	matches := r.FindStringSubmatch(rawKey)

	require.Len(t, matches, 3)
	require.Equal(t, apiKey.Prefix, matches[1])
}

func TestServiceCreateAPIKeyError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	name := "My API Key"
	apiKey := &auth.APIKey{
		UserUUID:  userUUID,
		Prefix:    "pr3f1x",
		HashedKey: "h45h3dk3y",
		Name:      name,
		ExpiresAt: nil,
	}
	cryptoCreateAPIKeyErr := errors.New("crypto.CreateAPIKey failed")
	genericRepoErr := errors.New("repo.CreateAPIKey failed")

	testcases := map[string]struct {
		ctx                   context.Context
		cryptoCreateAPIKeyErr error
		repoErr               error
		wantErr               error
	}{
		"No user UUID in context": {
			ctx:                   context.Background(),
			cryptoCreateAPIKeyErr: nil,
			repoErr:               nil,
			wantErr:               nil,
		},
		"crypto.CreateAPIKey fails": {
			ctx:                   context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			cryptoCreateAPIKeyErr: cryptoCreateAPIKeyErr,
			repoErr:               nil,
			wantErr:               cryptoCreateAPIKeyErr,
		},
		"API Key already exists": {
			ctx:                   context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			cryptoCreateAPIKeyErr: nil,
			repoErr:               errutils.ErrDatabaseUniqueViolation,
			wantErr:               errutils.ErrAPIKeyAlreadyExists,
		},
		"Generic repo error": {
			ctx:                   context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			cryptoCreateAPIKeyErr: nil,
			repoErr:               genericRepoErr,
			wantErr:               genericRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			crypto.
				EXPECT().
				CreateAPIKey().
				Return(apiKey.Prefix, "R4W4P1K3Y", apiKey.HashedKey, testcase.cryptoCreateAPIKeyErr).
				MaxTimes(1)

			repo.
				EXPECT().
				CreateAPIKey(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(apiKey, testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			_, _, err := svc.CreateAPIKey(testcase.ctx, name, nil)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceListAPIKeysSuccess(t *testing.T) {
	t.Parallel()

	user1, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	user2, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	timeProvider := timekeeper.NewFrozenProvider()
	tomorrow := timeProvider.Now().AddDate(0, 0, 1)
	user1Key1, _ := testkitinternal.MustCreateUserAPIKey(t, user1.UUID, func(k *auth.APIKey) {
		k.Name = "User1APIKey1"
		k.ExpiresAt = nil
	})

	user1Key2, _ := testkitinternal.MustCreateUserAPIKey(t, user1.UUID, func(k *auth.APIKey) {
		k.Name = "User1APIKey2"
		k.ExpiresAt = &tomorrow
	})

	user2Key, _ := testkitinternal.MustCreateUserAPIKey(t, user2.UUID, func(k *auth.APIKey) {
		k.Name = "User2APIKey"
		k.ExpiresAt = &tomorrow
	})

	user1Keys := []*auth.APIKey{user1Key1, user1Key2}
	user2Keys := []*auth.APIKey{user2Key}

	sort.Slice(user1Keys, func(i, j int) bool {
		return user1Keys[i].Prefix < user1Keys[j].Prefix
	})

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user1.UUID)
	fetchedUser1Keys, err := svc.ListAPIKeys(ctx)
	require.NoError(t, err)

	sort.Slice(fetchedUser1Keys, func(i, j int) bool {
		return fetchedUser1Keys[i].Prefix < fetchedUser1Keys[j].Prefix
	})

	ctx = context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user2.UUID)
	fetchedUser2Keys, err := svc.ListAPIKeys(ctx)
	require.NoError(t, err)

	require.Len(t, fetchedUser1Keys, 2)
	require.Len(t, fetchedUser2Keys, 1)

	for i, fetchedKey := range fetchedUser1Keys {
		wantKey := user1Keys[i]
		require.Equal(t, wantKey.ID, fetchedKey.ID)
		require.Equal(t, wantKey.UserUUID, fetchedKey.UserUUID)
		require.Equal(t, wantKey.Prefix, fetchedKey.Prefix)
		require.Equal(t, wantKey.Name, fetchedKey.Name)
		require.Equal(t, wantKey.ExpiresAt, fetchedKey.ExpiresAt)
		require.Equal(t, wantKey.CreatedAt, fetchedKey.CreatedAt)
		require.Equal(t, wantKey.UpdatedAt, fetchedKey.UpdatedAt)
	}

	for i, fetchedKey := range fetchedUser2Keys {
		wantKey := user2Keys[i]
		require.Equal(t, wantKey.ID, fetchedKey.ID)
		require.Equal(t, wantKey.UserUUID, fetchedKey.UserUUID)
		require.Equal(t, wantKey.Prefix, fetchedKey.Prefix)
		require.Equal(t, wantKey.Name, fetchedKey.Name)
		require.Equal(t, wantKey.ExpiresAt, fetchedKey.ExpiresAt)
		require.Equal(t, wantKey.CreatedAt, fetchedKey.CreatedAt)
		require.Equal(t, wantKey.UpdatedAt, fetchedKey.UpdatedAt)
	}
}

func TestServiceListAPIKeysError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	genericRepoErr := errors.New("ListAPIKeysByUserUUID failed")

	testcases := map[string]struct {
		ctx     context.Context
		repoErr error
		wantErr error
	}{
		"No user UUID in context": {
			ctx:     context.Background(),
			repoErr: nil,
			wantErr: nil,
		},
		"Generic repo error": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: genericRepoErr,
			wantErr: genericRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			repo.
				EXPECT().
				ListAPIKeysByUserUUID(gomock.Any(), gomock.Any(), userUUID).
				Return(nil, testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			_, err := svc.ListAPIKeys(testcase.ctx)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceFindAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey, rawKey := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	foundAPIKey, err := svc.FindAPIKey(context.Background(), rawKey)
	require.NoError(t, err)

	require.Equal(t, apiKey.ID, foundAPIKey.ID)
	require.Equal(t, apiKey.UserUUID, foundAPIKey.UserUUID)
	require.Equal(t, apiKey.Prefix, foundAPIKey.Prefix)
	require.Equal(t, apiKey.Name, foundAPIKey.Name)
	require.Equal(t, apiKey.ExpiresAt, foundAPIKey.ExpiresAt)
	require.Equal(t, apiKey.CreatedAt, foundAPIKey.CreatedAt)
	require.Equal(t, apiKey.UpdatedAt, foundAPIKey.UpdatedAt)
}

func TestServiceFindAPIKeyError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	validAPIKey := "TqxlYSSQ.Yj2j1jyAMC5407Nctsl51K7E8sOIPqYXn28SqT5Gnfg="
	invalidAPIKey := "deadbeef"
	genericRepoErr := errors.New("ListActiveAPIKeysByPrefix failed")

	testcases := map[string]struct {
		rawKey  string
		repoErr error
		wantErr error
	}{
		"Invalid API key": {
			rawKey:  invalidAPIKey,
			repoErr: nil,
			wantErr: nil,
		},
		"Generic repo error": {
			rawKey:  validAPIKey,
			repoErr: genericRepoErr,
			wantErr: genericRepoErr,
		},
		"API key not found": {
			rawKey:  validAPIKey,
			repoErr: nil,
			wantErr: errutils.ErrAPIKeyNotFound,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			repo.
				EXPECT().
				ListActiveAPIKeysByPrefix(gomock.Any(), gomock.Any(), "TqxlYSSQ").
				Return(nil, testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			_, err := svc.FindAPIKey(context.Background(), testcase.rawKey)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceUpdateAPIKeySuccess(t *testing.T) {
	t.Parallel()

	startingName := "APIKeyName"
	updatedName := "UpdatedAPIKeyName"
	timeProvider := timekeeper.NewFrozenProvider()
	nextMonth := timeProvider.Now().AddDate(0, 1, 0)
	nextYear := timeProvider.Now().AddDate(1, 0, 0)

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider.AddDate(0, 0, 1)
	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	testcases := map[string]struct {
		startingName      string
		updatedName       *string
		startingExpiresAt *time.Time
		updatedExpiresAt  jsonutils.Optional[time.Time]
		wantName          string
		wantExpiresAt     *time.Time
	}{
		"Update name, update expires at from valid date to valid date": {
			startingName:      startingName,
			updatedName:       &updatedName,
			startingExpiresAt: &nextMonth,
			updatedExpiresAt: jsonutils.Optional[time.Time]{
				Valid: true,
				Value: &nextYear,
			},
			wantName:      updatedName,
			wantExpiresAt: &nextYear,
		},
		"Update name only": {
			startingName:      startingName,
			updatedName:       &updatedName,
			startingExpiresAt: &nextMonth,
			updatedExpiresAt: jsonutils.Optional[time.Time]{
				Valid: false,
			},
			wantName:      updatedName,
			wantExpiresAt: &nextMonth,
		},
		"Update name, update expires at from null to valid date": {
			startingName:      startingName,
			updatedName:       &updatedName,
			startingExpiresAt: nil,
			updatedExpiresAt: jsonutils.Optional[time.Time]{
				Valid: true,
				Value: &nextYear,
			},
			wantName:      updatedName,
			wantExpiresAt: &nextYear,
		},
		"Update name, update expires at from valid date to null": {
			startingName:      startingName,
			updatedName:       &updatedName,
			startingExpiresAt: &nextMonth,
			updatedExpiresAt: jsonutils.Optional[time.Time]{
				Valid: true,
				Value: nil,
			},
			wantName:      updatedName,
			wantExpiresAt: nil,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
				u.IsActive = true
			})

			apiKey, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, func(k *auth.APIKey) {
				k.Name = testcase.startingName
				k.ExpiresAt = testcase.startingExpiresAt
			})

			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
			updatedAPIKey, err := svc.UpdateAPIKey(ctx, apiKey.ID, testcase.updatedName, testcase.updatedExpiresAt)
			require.NoError(t, err)

			require.Equal(t, apiKey.ID, updatedAPIKey.ID)
			require.Equal(t, user.UUID, updatedAPIKey.UserUUID)
			require.Equal(t, apiKey.Prefix, updatedAPIKey.Prefix)
			require.Equal(t, apiKey.HashedKey, updatedAPIKey.HashedKey)
			require.Equal(t, testcase.wantName, updatedAPIKey.Name)
			if testcase.wantExpiresAt != nil {
				require.NotNil(t, updatedAPIKey.ExpiresAt)
				require.WithinDuration(t, *testcase.wantExpiresAt, *updatedAPIKey.ExpiresAt, testkit.TimeToleranceExact)
			} else {
				require.Nil(t, updatedAPIKey.ExpiresAt)
			}

			require.Equal(t, apiKey.CreatedAt, updatedAPIKey.CreatedAt)
			require.WithinDuration(t, timeProvider.Now(), updatedAPIKey.UpdatedAt, testkit.TimeToleranceExact)
		})
	}
}

func TestServiceUpdateAPIKeyError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	var apiKeyID int64 = 314159
	updatedAPIKeyName := "updatedName"
	expiresAt := jsonutils.Optional[time.Time]{
		Valid: false,
	}
	genericRepoErr := errors.New("UpdateAPIKey failed")

	testcases := map[string]struct {
		ctx     context.Context
		repoErr error
		wantErr error
	}{
		"No user UUID in context": {
			ctx:     context.Background(),
			repoErr: nil,
			wantErr: nil,
		},
		"API key not found": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: errutils.ErrDatabaseNoRowsAffected,
			wantErr: errutils.ErrAPIKeyNotFound,
		},
		"Generic repo error": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: genericRepoErr,
			wantErr: genericRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			repo.
				EXPECT().
				UpdateAPIKey(gomock.Any(), gomock.Any(), userUUID, apiKeyID, &updatedAPIKeyName, expiresAt).
				Return(nil, testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			_, err := svc.UpdateAPIKey(testcase.ctx, 314159, &updatedAPIKeyName, expiresAt)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceDeleteAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()

	dbConn, err := TestDBPool.Acquire(context.Background())
	require.NoError(t, err)
	defer dbConn.Release()

	_, _, logger := testkit.CreateInMemLogger()
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	repo := auth.NewRepository(timeProvider)
	svc := auth.NewService(cfg, timeProvider, TestDBPool, logger, crypto, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
	err = svc.DeleteAPIKey(ctx, apiKey.ID)
	require.NoError(t, err)

	apiKeys, err := repo.ListAPIKeysByUserUUID(context.Background(), dbConn, user.UUID)
	require.NoError(t, err)
	require.Empty(t, apiKeys)
}

func TestServiceDeleteAPIKeyError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	var apiKeyID int64 = 314159
	genericRepoErr := errors.New("DeleteAPIKey failed")

	testcases := map[string]struct {
		ctx     context.Context
		repoErr error
		wantErr error
	}{
		"No user UUID in context": {
			ctx:     context.Background(),
			repoErr: nil,
			wantErr: nil,
		},
		"API key not found": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: errutils.ErrDatabaseNoRowsAffected,
			wantErr: errutils.ErrAPIKeyNotFound,
		},
		"Generic repo error": {
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			repoErr: genericRepoErr,
			wantErr: genericRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := databasemocks.NewMockPool(ctrl)
			dbConn := databasemocks.NewMockConn(ctrl)
			_, _, logger := testkit.CreateInMemLogger()
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			repo := authmocks.NewMockRepository(ctrl)

			dbConn.
				EXPECT().
				Release().
				MaxTimes(1)

			dbPool.
				EXPECT().
				Acquire(gomock.Any()).
				Return(dbConn, nil).
				MaxTimes(1)

			repo.
				EXPECT().
				DeleteAPIKey(gomock.Any(), gomock.Any(), userUUID, apiKeyID).
				Return(testcase.repoErr).
				MaxTimes(1)

			svc := auth.NewService(cfg, timeProvider, dbPool, logger, crypto, mailClient, tmplManager, repo)

			err := svc.DeleteAPIKey(testcase.ctx, apiKeyID)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}
