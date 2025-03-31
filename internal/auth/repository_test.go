package auth_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreateUserSuccess(t *testing.T) {
	t.Parallel()

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	user := &auth.User{
		UUID:        uuid.NewString(),
		Email:       testkit.GenerateFakeEmail(),
		Password:    testkitinternal.MustHashPassword(testkit.GenerateFakePassword()),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    false,
		IsSuperUser: false,
	}

	createdUser, err := repo.CreateUser(context.Background(), dbConn, user)
	require.NoError(t, err)

	require.Equal(t, user.Email, createdUser.Email)
	require.Equal(t, user.Password, createdUser.Password)
	require.Equal(t, user.FirstName, createdUser.FirstName)
	require.Equal(t, user.LastName, createdUser.LastName)
	require.Equal(t, user.IsActive, createdUser.IsActive)
	require.Equal(t, user.IsSuperUser, createdUser.IsSuperUser)
	require.WithinDuration(t, timeProvider.Now(), createdUser.CreatedAt, testkit.TimeToleranceExact)
	require.WithinDuration(t, timeProvider.Now(), createdUser.UpdatedAt, testkit.TimeToleranceExact)
}

func TestRepositoryCreateUserDuplicateEmail(t *testing.T) {
	t.Parallel()

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	email := testkit.GenerateFakeEmail()
	user1 := &auth.User{
		UUID:        uuid.NewString(),
		Email:       email,
		Password:    testkitinternal.MustHashPassword(testkit.GenerateFakePassword()),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    false,
		IsSuperUser: false,
	}

	_, err := repo.CreateUser(context.Background(), dbConn, user1)
	require.NoError(t, err)

	user2 := &auth.User{
		UUID:        uuid.NewString(),
		Email:       email,
		Password:    testkitinternal.MustHashPassword(testkit.GenerateFakePassword()),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    false,
		IsSuperUser: false,
	}

	_, err = repo.CreateUser(context.Background(), dbConn, user2)
	require.ErrorIs(t, err, errutils.ErrDatabaseUniqueViolation)
}

func TestRepositoryActivateUserByUUIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	_, err := repo.GetUserByEmail(context.Background(), dbConn, user.Email)
	require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)

	timeProvider.AddDate(0, 0, 1)

	err = repo.ActivateUserByUUID(context.Background(), dbConn, user.UUID)
	require.NoError(t, err)

	fetchedUser, err := repo.GetUserByEmail(context.Background(), dbConn, user.Email)
	require.NoError(t, err)

	require.True(t, fetchedUser.IsActive)
	require.WithinDuration(t, timeProvider.Now(), fetchedUser.UpdatedAt, testkit.TimeToleranceExact)
}

func TestRepositoryActivateUserByUUIDError(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	testcases := map[string]struct {
		userUUID string
	}{
		"No user under given UUID": {
			userUUID: uuid.NewString(),
		},
		"No inactive user under given UUID": {
			userUUID: activeUser.UUID,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
			err := repo.ActivateUserByUUID(context.Background(), dbConn, testcase.userUUID)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
		})
	}
}

func TestRepositoryGetUserByEmailSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	fetchedUser, err := repo.GetUserByEmail(context.Background(), dbConn, user.Email)
	require.NoError(t, err)

	require.Equal(t, user.UUID, fetchedUser.UUID)
	require.Equal(t, user.Email, fetchedUser.Email)
	require.Equal(t, user.Password, fetchedUser.Password)
	require.Equal(t, user.FirstName, fetchedUser.FirstName)
	require.Equal(t, user.LastName, fetchedUser.LastName)
	require.Equal(t, user.IsActive, fetchedUser.IsActive)
	require.Equal(t, user.IsSuperUser, fetchedUser.IsSuperUser)
	require.Equal(t, user.CreatedAt, fetchedUser.CreatedAt)
	require.Equal(t, user.UpdatedAt, fetchedUser.UpdatedAt)
}

func TestRepositoryGetUserByEmailError(t *testing.T) {
	t.Parallel()

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	testcases := map[string]struct {
		email string
	}{
		"No user": {
			email: testkit.GenerateFakeEmail(),
		},
		"No active user": {
			email: inactiveUser.Email,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
			_, err := repo.GetUserByEmail(context.Background(), dbConn, testcase.email)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
		})
	}
}

func TestRepositoryGetUserByUUIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	fetchedUser, err := repo.GetUserByUUID(context.Background(), dbConn, user.UUID)
	require.NoError(t, err)

	require.Equal(t, user.UUID, fetchedUser.UUID)
	require.Equal(t, user.Email, fetchedUser.Email)
	require.Equal(t, user.Password, fetchedUser.Password)
	require.Equal(t, user.FirstName, fetchedUser.FirstName)
	require.Equal(t, user.LastName, fetchedUser.LastName)
	require.Equal(t, user.IsActive, fetchedUser.IsActive)
	require.Equal(t, user.IsSuperUser, fetchedUser.IsSuperUser)
	require.Equal(t, user.CreatedAt, fetchedUser.CreatedAt)
	require.Equal(t, user.UpdatedAt, fetchedUser.UpdatedAt)
}

func TestRepositoryGetUserByUUIDError(t *testing.T) {
	t.Parallel()

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	testcases := map[string]struct {
		userUUID string
	}{
		"No user under UUID": {
			userUUID: uuid.NewString(),
		},
		"No active user": {
			userUUID: inactiveUser.UUID,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
			_, err := repo.GetUserByUUID(context.Background(), dbConn, testcase.userUUID)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
		})
	}
}

func TestRepositoryUpdateUserSuccess(t *testing.T) {
	t.Parallel()

	startingFirstName := "Firstname"
	startingLastName := "Lastname"
	updatedFirstName := "Updatedfirstname"
	updatedLastName := "Updatedlastname"

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	timeProvider := timekeeper.NewFrozenProvider()
	timeProvider.AddDate(0, 0, 1)
	repo := auth.NewRepository(timeProvider)

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
		"Update first name": {
			firstName:     &updatedFirstName,
			lastName:      nil,
			wantFirstName: updatedFirstName,
			wantLastName:  startingLastName,
		},
		"Update last name": {
			firstName:     nil,
			lastName:      &updatedLastName,
			wantFirstName: startingFirstName,
			wantLastName:  updatedLastName,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
				u.FirstName = startingFirstName
				u.LastName = startingLastName
				u.IsActive = true
			})

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())

			updatedUser, err := repo.UpdateUser(context.Background(), dbConn, user.UUID, testcase.firstName, testcase.lastName)
			require.NoError(t, err)

			require.Equal(t, user.UUID, updatedUser.UUID)
			require.Equal(t, user.Email, updatedUser.Email)
			require.Equal(t, user.Password, updatedUser.Password)
			require.Equal(t, testcase.wantFirstName, updatedUser.FirstName)
			require.Equal(t, testcase.wantLastName, updatedUser.LastName)
			require.Equal(t, user.IsActive, updatedUser.IsActive)
			require.Equal(t, user.IsSuperUser, updatedUser.IsSuperUser)
			require.Equal(t, user.CreatedAt, updatedUser.CreatedAt)
			require.WithinDuration(t, timeProvider.Now(), updatedUser.UpdatedAt, testkit.TimeToleranceExact)
		})
	}
}

func TestRepositoryUpdateUserError(t *testing.T) {
	t.Parallel()

	updatedFirstName := "Updatedfirstname"
	updatedLastName := "Updatedlastname"

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	testcases := map[string]struct {
		userUUID  string
		firstName *string
		lastName  *string
	}{
		"Update neither first nor last name": {
			userUUID:  activeUser.UUID,
			firstName: nil,
			lastName:  nil,
		},
		"Update non-existent user": {
			userUUID:  uuid.NewString(),
			firstName: &updatedFirstName,
			lastName:  &updatedLastName,
		},
		"Update inactive user": {
			userUUID:  inactiveUser.UUID,
			firstName: &updatedFirstName,
			lastName:  &updatedLastName,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())

			_, err := repo.UpdateUser(context.Background(), dbConn, testcase.userUUID, testcase.firstName, testcase.lastName)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
		})
	}
}

func TestRepositoryCreateAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	apiKey := &auth.APIKey{
		UserUUID:  user.UUID,
		Prefix:    testkit.MustGenerateRandomString(8, true, true, true),
		HashedKey: testkitinternal.MustHashPassword(testkit.MustGenerateRandomString(16, true, true, true)),
		Name:      "My API Key",
		ExpiresAt: nil,
	}

	createdAPIKey, err := repo.CreateAPIKey(context.Background(), dbConn, apiKey)
	require.NoError(t, err)

	require.Equal(t, user.UUID, createdAPIKey.UserUUID)
	require.Equal(t, apiKey.Prefix, createdAPIKey.Prefix)
	require.Equal(t, apiKey.HashedKey, createdAPIKey.HashedKey)
	require.Equal(t, apiKey.Name, createdAPIKey.Name)
	require.Nil(t, apiKey.ExpiresAt)
	require.WithinDuration(t, timeProvider.Now(), createdAPIKey.CreatedAt, testkit.TimeToleranceExact)
	require.WithinDuration(t, timeProvider.Now(), createdAPIKey.UpdatedAt, testkit.TimeToleranceExact)
}

func TestRepositoryCreateAPIKeyError(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testkitinternal.MustCreateUserAPIKey(t, user.UUID, func(k *auth.APIKey) {
		k.Name = "My API Key"
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	testcases := map[string]struct {
		userUUID   string
		apiKeyName string
		wantErr    error
	}{
		"No user": {
			userUUID:   uuid.NewString(),
			apiKeyName: "My Other API Key",
			wantErr:    errutils.ErrDatabaseForeignKeyConstraintViolation,
		},
		"Duplicate API Key": {
			userUUID:   user.UUID,
			apiKeyName: "My API Key",
			wantErr:    errutils.ErrDatabaseUniqueViolation,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			apiKey := &auth.APIKey{
				UserUUID:  testcase.userUUID,
				Prefix:    testkit.MustGenerateRandomString(8, true, true, true),
				HashedKey: testkit.MustGenerateRandomString(16, true, true, true),
				Name:      testcase.apiKeyName,
				ExpiresAt: nil,
			}

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())

			_, err := repo.CreateAPIKey(context.Background(), dbConn, apiKey)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestRepositoryListAPIKeysByUserUUIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey1, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)
	apiKey2, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	wantAPIKeys := []*auth.APIKey{apiKey1, apiKey2}
	sort.Slice(wantAPIKeys, func(i, j int) bool {
		return wantAPIKeys[i].Prefix < wantAPIKeys[j].Prefix
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	apiKeys, err := repo.ListAPIKeysByUserUUID(context.Background(), dbConn, user.UUID)
	require.NoError(t, err)
	require.Len(t, apiKeys, 2)

	sort.Slice(apiKeys, func(i, j int) bool {
		return apiKeys[i].Prefix < apiKeys[j].Prefix
	})

	for i, apiKey := range apiKeys {
		wantKey := wantAPIKeys[i]
		require.Equal(t, wantKey.ID, apiKey.ID)
		require.Equal(t, wantKey.UserUUID, apiKey.UserUUID)
		require.Equal(t, wantKey.Prefix, apiKey.Prefix)
		require.Equal(t, wantKey.Name, apiKey.Name)
		require.Equal(t, wantKey.ExpiresAt, apiKey.ExpiresAt)
		require.Equal(t, wantKey.CreatedAt, apiKey.CreatedAt)
		require.Equal(t, wantKey.UpdatedAt, apiKey.UpdatedAt)
	}
}

func TestRepositoryListAPIKeysByUserUUIDInactiveUser(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	apiKeys, err := repo.ListAPIKeysByUserUUID(context.Background(), dbConn, user.UUID)
	require.NoError(t, err)
	require.Empty(t, apiKeys)
}

func TestRepositoryListActiveAPIKeysByPrefixSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	apiKeys, err := repo.ListActiveAPIKeysByPrefix(context.Background(), dbConn, apiKey.Prefix)
	require.NoError(t, err)
	require.Len(t, apiKeys, 1)

	require.NotNil(t, apiKeys[0])
	require.Equal(t, apiKey.ID, apiKeys[0].ID)
	require.Equal(t, apiKey.Prefix, apiKeys[0].Prefix)
	require.Equal(t, apiKey.HashedKey, apiKeys[0].HashedKey)
	require.Equal(t, apiKey.Name, apiKeys[0].Name)
	require.Equal(t, apiKey.ExpiresAt, apiKeys[0].ExpiresAt)
	require.Equal(t, apiKey.CreatedAt, apiKeys[0].CreatedAt)
	require.Equal(t, apiKey.UpdatedAt, apiKeys[0].UpdatedAt)
}

func TestRepositoryListActiveAPIKeysByPrefixEmpty(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	timeProvider := timekeeper.NewFrozenProvider()
	yesterday := timeProvider.Now().AddDate(0, 0, -1)
	expiredAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.ExpiresAt = &yesterday
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	validAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	repo := auth.NewRepository(timeProvider)

	testcases := map[string]struct {
		apiKey *auth.APIKey
	}{
		"Active user with expired API key": {
			apiKey: expiredAPIKey,
		},
		"Inactive user with valid API key": {
			apiKey: validAPIKey,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
			apiKeys, err := repo.ListActiveAPIKeysByPrefix(context.Background(), dbConn, testcase.apiKey.Prefix)
			require.NoError(t, err)
			require.Empty(t, apiKeys)
		})
	}
}

func TestRepositoryUpdateAPIKeySuccess(t *testing.T) {
	t.Parallel()

	startingName := "APIKeyName"
	updatedName := "UpdatedAPIKeyName"

	timeProvider := timekeeper.NewFrozenProvider()
	nextMonth := timeProvider.Now().AddDate(0, 1, 0)
	nextYear := timeProvider.Now().AddDate(1, 0, 0)

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	timeProvider.AddDate(0, 0, 1)
	repo := auth.NewRepository(timeProvider)

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

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())

			updatedAPIKey, err := repo.UpdateAPIKey(
				context.Background(),
				dbConn,
				user.UUID,
				apiKey.ID,
				testcase.updatedName,
				testcase.updatedExpiresAt,
			)
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

func TestRepositoryUpdateAPIKeyError(t *testing.T) {
	t.Parallel()

	updatedName := "UpdatedAPIKeyName"

	timeProvider := timekeeper.NewFrozenProvider()
	nextMonth := timeProvider.Now().AddDate(0, 1, 0)
	nextYear := timeProvider.Now().AddDate(1, 0, 0)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	activeUserAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.Name = "APIKeyName"
		k.ExpiresAt = &nextMonth
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUserAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, func(k *auth.APIKey) {
		k.Name = "APIKeyName"
		k.ExpiresAt = &nextMonth
	})

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	repo := auth.NewRepository(timeProvider)

	testcases := map[string]struct {
		apiKeyID         int64
		userUUID         string
		updatedName      *string
		updatedExpiresAt jsonutils.Optional[time.Time]
	}{
		"Update neither name nor expires at": {
			apiKeyID:    activeUserAPIKey.ID,
			userUUID:    activeUser.UUID,
			updatedName: nil,
			updatedExpiresAt: jsonutils.Optional[time.Time]{
				Valid: false,
			},
		},
		"Update non-existent API key": {
			apiKeyID:    314159,
			userUUID:    activeUser.UUID,
			updatedName: &updatedName,
			updatedExpiresAt: jsonutils.Optional[time.Time]{
				Valid: true,
				Value: &nextYear,
			},
		},
		"Update API key for inactive user": {
			apiKeyID:    inactiveUserAPIKey.ID,
			userUUID:    inactiveUser.UUID,
			updatedName: &updatedName,
			updatedExpiresAt: jsonutils.Optional[time.Time]{
				Valid: true,
				Value: &nextYear,
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())

			_, err := repo.UpdateAPIKey(
				context.Background(),
				dbConn,
				testcase.userUUID,
				testcase.apiKeyID,
				testcase.updatedName,
				testcase.updatedExpiresAt,
			)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected, err)
		})
	}
}

func TestRepositoryDeleteAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	err := repo.DeleteAPIKey(
		context.Background(),
		dbConn,
		user.UUID,
		apiKey.ID,
	)
	require.NoError(t, err)

	apiKeys, err := repo.ListAPIKeysByUserUUID(context.Background(), dbConn, user.UUID)
	require.NoError(t, err)
	require.Empty(t, apiKeys)
}

func TestRepositoryDeleteAPIKeyError(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUserAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	dbPool := testkitinternal.RequireNewDatabasePool(t)
	timeProvider := timekeeper.NewFrozenProvider()
	repo := auth.NewRepository(timeProvider)

	testcases := map[string]struct {
		apiKeyID int64
		userUUID string
	}{
		"Active user with no API keys": {
			apiKeyID: 314159,
			userUUID: activeUser.UUID,
		},
		"Inactive user with valid API key": {
			apiKeyID: inactiveUserAPIKey.ID,
			userUUID: inactiveUser.UUID,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireNewDatabaseConn(t, dbPool, context.Background())
			err := repo.DeleteAPIKey(
				context.Background(),
				dbConn,
				testcase.userUUID,
				testcase.apiKeyID,
			)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
		})
	}
}
