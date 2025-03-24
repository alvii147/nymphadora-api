package code_test

import (
	"context"
	"testing"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/internal/code"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreateCodeSpaceSuccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := code.NewRepository(timeProvider)

	codeSpace := &code.CodeSpace{
		AuthorUUID: &author.UUID,
		Name:       "habitable-slaking-volatile-granger-mov",
		Language:   "python",
		Contents:   "print('hello')",
	}

	createdCodeSpace, err := repo.CreateCodeSpace(context.Background(), dbConn, codeSpace)
	require.NoError(t, err)

	require.NotNil(t, createdCodeSpace.AuthorUUID)
	require.Equal(t, author.UUID, *createdCodeSpace.AuthorUUID)
	require.Equal(t, codeSpace.Name, createdCodeSpace.Name)
	require.Equal(t, codeSpace.Language, createdCodeSpace.Language)
	require.Equal(t, codeSpace.Contents, createdCodeSpace.Contents)
	require.WithinDuration(t, timeProvider.Now(), createdCodeSpace.CreatedAt, testkit.TimeToleranceExact)
	require.WithinDuration(t, timeProvider.Now(), createdCodeSpace.UpdatedAt, testkit.TimeToleranceExact)
}

func TestRepositoryCreateCodeSpaceDuplicateName(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	existingCodeSpaceAuthor, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	existingCodeSpace, _ := testkitinternal.MustCreateCodeSpace(t, existingCodeSpaceAuthor.UUID, "python")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := code.NewRepository(timeProvider)

	codeSpace := &code.CodeSpace{
		AuthorUUID: &author.UUID,
		Name:       existingCodeSpace.Name,
		Language:   "python",
		Contents:   "print('hello')",
	}

	_, err := repo.CreateCodeSpace(context.Background(), dbConn, codeSpace)
	require.Error(t, err)
}

func TestRepositoryListCodeSpaces(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	sharedCodeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")
	privateCodeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	editor, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		editor.UUID,
		sharedCodeSpace.ID,
		code.CodeSpaceAccessLevelReadWrite,
	)

	viewer, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		viewer.UUID,
		sharedCodeSpace.ID,
		code.CodeSpaceAccessLevelReadOnly,
	)

	thirdPartyUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	timeProvider := timekeeper.NewFrozenProvider()
	repo := code.NewRepository(timeProvider)

	testcases := map[string]struct {
		userUUID               string
		wantCodeSpaceAccessMap map[int64]code.CodeSpaceAccessLevel
	}{
		"Author can access both shared and private code spaces": {
			userUUID: author.UUID,
			wantCodeSpaceAccessMap: map[int64]code.CodeSpaceAccessLevel{
				sharedCodeSpace.ID:  code.CodeSpaceAccessLevelReadWrite,
				privateCodeSpace.ID: code.CodeSpaceAccessLevelReadWrite,
			},
		},
		"Editor can access shared code space": {
			userUUID: editor.UUID,
			wantCodeSpaceAccessMap: map[int64]code.CodeSpaceAccessLevel{
				sharedCodeSpace.ID: code.CodeSpaceAccessLevelReadWrite,
			},
		},
		"Viewer can access shared code space": {
			userUUID: viewer.UUID,
			wantCodeSpaceAccessMap: map[int64]code.CodeSpaceAccessLevel{
				sharedCodeSpace.ID: code.CodeSpaceAccessLevelReadOnly,
			},
		},
		"Third party user cannot access any code space": {
			userUUID:               thirdPartyUser.UUID,
			wantCodeSpaceAccessMap: map[int64]code.CodeSpaceAccessLevel{},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())

			codeSpaces, codeSpaceAccesses, err := repo.ListCodeSpaces(context.Background(), dbConn, testcase.userUUID)
			require.NoError(t, err)

			require.Equal(t, len(testcase.wantCodeSpaceAccessMap), len(codeSpaces))
			require.Equal(t, len(testcase.wantCodeSpaceAccessMap), len(codeSpaceAccesses))

			wantCodeSpaceIDs := make([]int64, len(testcase.wantCodeSpaceAccessMap))
			i := 0
			for codeSpaceID := range testcase.wantCodeSpaceAccessMap {
				wantCodeSpaceIDs[i] = codeSpaceID
				i++
			}

			codeSpaceIDs := make([]int64, len(codeSpaces))
			for i, codeSpace := range codeSpaces {
				codeSpaceIDs[i] = codeSpace.ID
				require.Contains(t, testcase.wantCodeSpaceAccessMap, codeSpace.ID)
				require.Equal(
					t,
					testcase.wantCodeSpaceAccessMap[codeSpace.ID],
					codeSpaceAccesses[i].Level,
				)
			}

			require.ElementsMatch(t, wantCodeSpaceIDs, codeSpaceIDs)
		})
	}
}

func TestRepositoryGetCodeSpaceWithAccessByName(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	sharedCodeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")
	privateCodeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	editor, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		editor.UUID,
		sharedCodeSpace.ID,
		code.CodeSpaceAccessLevelReadWrite,
	)

	viewer, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		viewer.UUID,
		sharedCodeSpace.ID,
		code.CodeSpaceAccessLevelReadOnly,
	)

	thirdPartyUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := code.NewRepository(timeProvider)

	testcases := map[string]struct {
		userUUID        string
		codeSpace       *code.CodeSpace
		wantAccessible  bool
		wantAccessLevel code.CodeSpaceAccessLevel
	}{
		"Author can access shared code space": {
			userUUID:        author.UUID,
			codeSpace:       sharedCodeSpace,
			wantAccessible:  true,
			wantAccessLevel: code.CodeSpaceAccessLevelReadWrite,
		},
		"Author can access private code space": {
			userUUID:        author.UUID,
			codeSpace:       privateCodeSpace,
			wantAccessible:  true,
			wantAccessLevel: code.CodeSpaceAccessLevelReadWrite,
		},
		"Editor can access shared code space": {
			userUUID:        editor.UUID,
			codeSpace:       sharedCodeSpace,
			wantAccessible:  true,
			wantAccessLevel: code.CodeSpaceAccessLevelReadWrite,
		},
		"Editor cannot access private code space": {
			userUUID:        editor.UUID,
			codeSpace:       privateCodeSpace,
			wantAccessible:  false,
			wantAccessLevel: 0,
		},
		"Viewer can access shared code space": {
			userUUID:        viewer.UUID,
			codeSpace:       sharedCodeSpace,
			wantAccessible:  true,
			wantAccessLevel: code.CodeSpaceAccessLevelReadOnly,
		},
		"Viewer cannot access private code space": {
			userUUID:        viewer.UUID,
			codeSpace:       privateCodeSpace,
			wantAccessible:  false,
			wantAccessLevel: 0,
		},
		"Third party user cannot access shared code space": {
			userUUID:        thirdPartyUser.UUID,
			codeSpace:       sharedCodeSpace,
			wantAccessible:  false,
			wantAccessLevel: 0,
		},
		"Third party user cannot access private code space": {
			userUUID:        thirdPartyUser.UUID,
			codeSpace:       privateCodeSpace,
			wantAccessible:  false,
			wantAccessLevel: 0,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())

			codeSpace, codeSpaceAccess, err := repo.GetCodeSpaceWithAccessByName(
				context.Background(),
				dbConn,
				testcase.userUUID,
				testcase.codeSpace.Name,
			)

			if !testcase.wantAccessible {
				require.Error(t, err)
				require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.codeSpace.ID, codeSpace.ID)
			require.NotNil(t, codeSpace.AuthorUUID)
			require.Equal(t, author.UUID, *codeSpace.AuthorUUID)
			require.Equal(t, testcase.codeSpace.Name, codeSpace.Name)
			require.Equal(t, testcase.codeSpace.Language, codeSpace.Language)
			require.Equal(t, testcase.codeSpace.Contents, codeSpace.Contents)
			require.WithinDuration(t, testcase.codeSpace.CreatedAt, codeSpace.CreatedAt, testkit.TimeToleranceExact)
			require.WithinDuration(t, testcase.codeSpace.UpdatedAt, codeSpace.UpdatedAt, testkit.TimeToleranceExact)

			require.Equal(t, testcase.userUUID, codeSpaceAccess.UserUUID)
			require.Equal(t, testcase.codeSpace.ID, codeSpaceAccess.CodeSpaceID)
			require.Equal(t, testcase.wantAccessLevel, codeSpaceAccess.Level)
		})
	}
}

func TestRepositoryUpdateCodeSpaceSuccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	timeProvider := timekeeper.NewFrozenProvider()
	now := timeProvider.Now()
	tomorrow := now.AddDate(0, 0, 1)
	timeProvider.SetTime(tomorrow)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := code.NewRepository(timeProvider)

	updatedContents := "print('FizzBuzz')"
	updatedCodeSpace, err := repo.UpdateCodeSpace(context.Background(), dbConn, codeSpace.ID, &updatedContents)
	require.NoError(t, err)

	require.NotNil(t, codeSpace.AuthorUUID)
	require.Equal(t, author.UUID, *codeSpace.AuthorUUID)
	require.Equal(t, codeSpace.ID, updatedCodeSpace.ID)
	require.Equal(t, codeSpace.Name, updatedCodeSpace.Name)
	require.Equal(t, codeSpace.Language, updatedCodeSpace.Language)
	require.Equal(t, updatedContents, updatedCodeSpace.Contents)
	require.WithinDuration(t, now, updatedCodeSpace.CreatedAt, testkit.TimeToleranceTentative)
	require.WithinDuration(t, tomorrow, updatedCodeSpace.UpdatedAt, testkit.TimeToleranceTentative)
}

func TestRepositoryUpdateCodeSpaceError(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := code.NewRepository(timeProvider)
	updatedContents := "print('FizzBuzz')"

	testcases := map[string]struct {
		codeSpaceID     int64
		updatedContents *string
	}{
		"Update non-existent code space": {
			codeSpaceID:     314159265,
			updatedContents: &updatedContents,
		},
		"Update blank update": {
			codeSpaceID:     codeSpace.ID,
			updatedContents: nil,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())

			_, err := repo.UpdateCodeSpace(context.Background(), dbConn, testcase.codeSpaceID, testcase.updatedContents)
			require.Error(t, err)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
		})
	}
}

func TestRepositoryDeleteCodeSpaceSuccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := code.NewRepository(timeProvider)

	err := repo.DeleteCodeSpace(
		context.Background(),
		dbConn,
		codeSpace.ID,
	)
	require.NoError(t, err)

	_, _, err = repo.GetCodeSpaceWithAccessByName(context.Background(), dbConn, author.UUID, codeSpace.Name)
	require.Error(t, err)
	require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
}

func TestRepositoryDeleteCodeSpaceNoRowsAffected(t *testing.T) {
	t.Parallel()

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	timeProvider := timekeeper.NewFrozenProvider()
	repo := code.NewRepository(timeProvider)

	err := repo.DeleteCodeSpace(
		context.Background(),
		dbConn,
		314159265,
	)
	require.Error(t, err)
	require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
}

func TestRepositoryCreateOrUpdateCodeSpaceAccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	editor, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		editor.UUID,
		codeSpace.ID,
		code.CodeSpaceAccessLevelReadWrite,
	)

	viewer, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		viewer.UUID,
		codeSpace.ID,
		code.CodeSpaceAccessLevelReadOnly,
	)

	invitee, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := code.NewRepository(timeProvider)

	testcases := map[string]struct {
		userUUID    string
		codeSpaceID int64
		accessLevel code.CodeSpaceAccessLevel
		wantErr     bool
	}{
		"Create code space access for invitee without access": {
			userUUID:    invitee.UUID,
			codeSpaceID: codeSpace.ID,
			accessLevel: code.CodeSpaceAccessLevelReadOnly,
			wantErr:     false,
		},
		"Update code space access for editor to revoke write access": {
			userUUID:    editor.UUID,
			codeSpaceID: codeSpace.ID,
			accessLevel: code.CodeSpaceAccessLevelReadOnly,
			wantErr:     false,
		},
		"Update code space access for viewer to grant write access": {
			userUUID:    viewer.UUID,
			codeSpaceID: codeSpace.ID,
			accessLevel: code.CodeSpaceAccessLevelReadWrite,
			wantErr:     false,
		},
		"Non-existent code space ID": {
			userUUID:    author.UUID,
			codeSpaceID: 314159265,
			accessLevel: code.CodeSpaceAccessLevelReadOnly,
			wantErr:     true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			codeSpaceAccess := &code.CodeSpaceAccess{
				UserUUID:    testcase.userUUID,
				CodeSpaceID: testcase.codeSpaceID,
				Level:       testcase.accessLevel,
			}

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			codeSpaceAccess, err := repo.CreateOrUpdateCodeSpaceAccess(context.Background(), dbConn, codeSpaceAccess)

			if testcase.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.userUUID, codeSpaceAccess.UserUUID)
			require.Equal(t, testcase.codeSpaceID, codeSpaceAccess.CodeSpaceID)
			require.Equal(t, testcase.accessLevel, codeSpaceAccess.Level)
			require.WithinDuration(t, timeProvider.Now(), codeSpaceAccess.CreatedAt, testkit.TimeToleranceTentative)
			require.WithinDuration(t, timeProvider.Now(), codeSpaceAccess.UpdatedAt, testkit.TimeToleranceTentative)
		})
	}
}

func TestRepositoryListUsersWithCodeSpaceAccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	editor, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		editor.UUID,
		codeSpace.ID,
		code.CodeSpaceAccessLevelReadWrite,
	)

	viewer, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		viewer.UUID,
		codeSpace.ID,
		code.CodeSpaceAccessLevelReadOnly,
	)

	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := code.NewRepository(timeProvider)

	users, codeSpaceAccesses, err := repo.ListUsersWithCodeSpaceAccess(context.Background(), dbConn, codeSpace.ID)
	require.NoError(t, err)
	require.Len(t, users, 3)
	require.Len(t, codeSpaceAccesses, 3)

	userAccessLevelMap := map[string]code.CodeSpaceAccessLevel{
		author.UUID: code.CodeSpaceAccessLevelReadWrite,
		editor.UUID: code.CodeSpaceAccessLevelReadWrite,
		viewer.UUID: code.CodeSpaceAccessLevelReadOnly,
	}

	for i, codeSpaceAccess := range codeSpaceAccesses {
		wantAccessLevel, ok := userAccessLevelMap[codeSpaceAccess.UserUUID]
		require.True(t, ok)

		delete(userAccessLevelMap, codeSpaceAccess.UserUUID)

		require.Equal(t, users[i].UUID, codeSpaceAccess.UserUUID)
		require.Equal(t, codeSpace.ID, codeSpaceAccess.CodeSpaceID)
		require.Equal(t, wantAccessLevel, codeSpaceAccess.Level)
	}
}

func TestRepositoryDeleteCodeSpaceAccessSuccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	invitee, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		invitee.UUID,
		codeSpace.ID,
		code.CodeSpaceAccessLevelReadOnly,
	)

	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := code.NewRepository(timeProvider)

	err := repo.DeleteCodeSpaceAccess(context.Background(), dbConn, invitee.UUID, codeSpace.ID)
	require.NoError(t, err)

	_, _, err = repo.GetCodeSpaceWithAccessByName(context.Background(), dbConn, invitee.UUID, codeSpace.Name)
	require.Error(t, err)
	require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
}

func TestRepositoryDeleteCodeSpaceAccessNoRowsAffected(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := code.NewRepository(timeProvider)

	err := repo.DeleteCodeSpaceAccess(context.Background(), dbConn, uuid.NewString(), 314159265)
	require.Error(t, err)
	require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
}
