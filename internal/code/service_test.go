package code_test

import (
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"regexp"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/alvii147/nymphadora-api/internal/auth"
	authmocks "github.com/alvii147/nymphadora-api/internal/auth/mocks"
	"github.com/alvii147/nymphadora-api/internal/code"
	codemocks "github.com/alvii147/nymphadora-api/internal/code/mocks"
	"github.com/alvii147/nymphadora-api/internal/templatesmanager"
	templatesmanagermocks "github.com/alvii147/nymphadora-api/internal/templatesmanager/mocks"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	cryptocoremocks "github.com/alvii147/nymphadora-api/pkg/cryptocore/mocks"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	mailclientmocks "github.com/alvii147/nymphadora-api/pkg/mailclient/mocks"
	"github.com/alvii147/nymphadora-api/pkg/piston"
	pistonmocks "github.com/alvii147/nymphadora-api/pkg/piston/mocks"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/alvii147/nymphadora-api/pkg/validate"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServiceGenerateCodeSpaceName(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	pistonClient := piston.NewClient(nil, httputils.NewHTTPClient(nil))
	repo := codemocks.NewMockRepository(ctrl)
	authRepo := authmocks.NewMockRepository(ctrl)

	svc := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		repo,
		authRepo,
	)

	codeSpaceName, err := svc.GenerateCodeSpaceName()
	require.NoError(t, err)

	v := validate.NewValidator()
	v.ValidateStringSlug("name", codeSpaceName)

	require.True(t, v.Passed())
}

func TestServiceCreateCodeSpaceSuccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	cfg := testkitinternal.MustCreateConfig()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	pistonClient := piston.NewClient(nil, httputils.NewHTTPClient(nil))
	repo := code.NewRepository(timeProvider)
	authRepo := auth.NewRepository(timeProvider)

	svc := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		repo,
		authRepo,
	)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, author.UUID)

	codeSpace, codeSpaceAccess, err := svc.CreateCodeSpace(ctx, "python")
	require.NoError(t, err)

	require.NotNil(t, codeSpace.AuthorUUID)
	require.Equal(t, author.UUID, *codeSpace.AuthorUUID)
	require.Equal(t, "python", codeSpace.Language)

	require.Equal(t, author.UUID, codeSpaceAccess.UserUUID)
	require.Equal(t, codeSpace.ID, codeSpaceAccess.CodeSpaceID)
	require.Equal(t, code.CodeSpaceAccessLevelReadWrite, codeSpaceAccess.Level)
}

func TestServiceCreateCodeSpaceError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	genericCreateCodeSpaceRepoErr := errors.New("CreateCodeSpace failed")
	genericCreateCodeSpaceAccessRepoErr := errors.New("CreateOrUpdateCodeSpaceAccess failed")

	testcases := map[string]struct {
		ctx                          context.Context
		language                     string
		createCodeSpaceRepoErr       error
		createCodeSpaceAccessRepoErr error
		wantErr                      error
	}{
		"No user in context": {
			ctx:                          context.Background(),
			language:                     "python",
			createCodeSpaceRepoErr:       nil,
			createCodeSpaceAccessRepoErr: nil,
			wantErr:                      nil,
		},
		"Unsupported language": {
			ctx:                          context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, author.UUID),
			language:                     "parseltongue",
			createCodeSpaceRepoErr:       nil,
			createCodeSpaceAccessRepoErr: nil,
			wantErr:                      errutils.ErrCodeSpaceUnsupportedLanguage,
		},
		"CreateCodeSpace fails": {
			ctx:                          context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, author.UUID),
			language:                     "python",
			createCodeSpaceRepoErr:       genericCreateCodeSpaceRepoErr,
			createCodeSpaceAccessRepoErr: nil,
			wantErr:                      genericCreateCodeSpaceRepoErr,
		},
		"CreateOrUpdateCodeSpaceAccess fails": {
			ctx:                          context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, author.UUID),
			language:                     "python",
			createCodeSpaceRepoErr:       nil,
			createCodeSpaceAccessRepoErr: genericCreateCodeSpaceAccessRepoErr,
			wantErr:                      genericCreateCodeSpaceAccessRepoErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			repo.
				EXPECT().
				CreateCodeSpace(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(
					&code.CodeSpace{
						ID:         42,
						AuthorUUID: &author.UUID,
						Name:       "habitable-slaking-volatile-granger-mov",
						Language:   "python",
						Contents:   "print('hello')",
						CreatedAt:  timeProvider.Now(),
						UpdatedAt:  timeProvider.Now(),
					},
					testcase.createCodeSpaceRepoErr,
				).
				MaxTimes(1)

			repo.
				EXPECT().
				CreateOrUpdateCodeSpaceAccess(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(
					&code.CodeSpaceAccess{
						ID:          314159,
						UserUUID:    author.UUID,
						CodeSpaceID: 42,
						Level:       code.CodeSpaceAccessLevelReadWrite,
						CreatedAt:   timeProvider.Now(),
						UpdatedAt:   timeProvider.Now(),
					},
					testcase.createCodeSpaceAccessRepoErr,
				).
				MaxTimes(1)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			_, _, err := svc.CreateCodeSpace(testcase.ctx, testcase.language)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceListCodeSpacesSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

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

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := code.NewRepository(timeProvider)
			authRepo := auth.NewRepository(timeProvider)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, testcase.userUUID)
			codeSpaces, codeSpaceAccesses, err := svc.ListCodeSpaces(ctx)
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

func TestServiceListCodeSpacesError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	genericRepoErr := errors.New("ListCodeSpaces failed")

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
		"ListCodeSpaces fails": {
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
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			repo.
				EXPECT().
				ListCodeSpaces(gomock.Any(), gomock.Any(), userUUID).
				Return(nil, nil, testcase.repoErr).
				MaxTimes(1)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			_, _, err := svc.ListCodeSpaces(testcase.ctx)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceGetCodeSpace(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

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

	testcases := map[string]struct {
		userUUID        string
		codeSpace       *code.CodeSpace
		wantErr         error
		wantAccessLevel code.CodeSpaceAccessLevel
	}{
		"Author can access shared code space": {
			userUUID:        author.UUID,
			codeSpace:       sharedCodeSpace,
			wantErr:         nil,
			wantAccessLevel: code.CodeSpaceAccessLevelReadWrite,
		},
		"Author can access private code space": {
			userUUID:        author.UUID,
			codeSpace:       privateCodeSpace,
			wantErr:         nil,
			wantAccessLevel: code.CodeSpaceAccessLevelReadWrite,
		},
		"Editor can access shared code space": {
			userUUID:        editor.UUID,
			codeSpace:       sharedCodeSpace,
			wantErr:         nil,
			wantAccessLevel: code.CodeSpaceAccessLevelReadWrite,
		},
		"Editor cannot access private code space": {
			userUUID:        editor.UUID,
			codeSpace:       privateCodeSpace,
			wantErr:         errutils.ErrCodeSpaceNotFound,
			wantAccessLevel: 0,
		},
		"Viewer can access shared code space": {
			userUUID:        viewer.UUID,
			codeSpace:       sharedCodeSpace,
			wantErr:         nil,
			wantAccessLevel: code.CodeSpaceAccessLevelReadOnly,
		},
		"Viewer cannot access private code space": {
			userUUID:        viewer.UUID,
			codeSpace:       privateCodeSpace,
			wantErr:         errutils.ErrCodeSpaceNotFound,
			wantAccessLevel: 0,
		},
		"Third party user cannot access shared code space": {
			userUUID:        thirdPartyUser.UUID,
			codeSpace:       sharedCodeSpace,
			wantErr:         errutils.ErrCodeSpaceNotFound,
			wantAccessLevel: 0,
		},
		"Third party user cannot access private code space": {
			userUUID:        thirdPartyUser.UUID,
			codeSpace:       privateCodeSpace,
			wantErr:         errutils.ErrCodeSpaceNotFound,
			wantAccessLevel: 0,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := code.NewRepository(timeProvider)
			authRepo := auth.NewRepository(timeProvider)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, testcase.userUUID)
			codeSpace, codeSpaceAccess, err := svc.GetCodeSpace(ctx, testcase.codeSpace.Name)

			if testcase.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, testcase.wantErr)

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

func TestServiceGetCodeSpaceError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	userUUID := uuid.NewString()
	codeSpaceName := "habitable-slaking-volatile-granger-mov"
	genericRepoErr := errors.New("GetCodeSpaceWithAccessByName failed")

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
		"GetCodeSpaceWithAccessByName fails": {
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
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			repo.
				EXPECT().
				GetCodeSpaceWithAccessByName(gomock.Any(), gomock.Any(), userUUID, codeSpaceName).
				Return(nil, nil, testcase.repoErr).
				MaxTimes(1)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			_, _, err := svc.GetCodeSpace(testcase.ctx, codeSpaceName)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceUpdateCodeSpaceAuthorSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	pistonClient := pistonmocks.NewMockClient(ctrl)
	repo := code.NewRepository(timeProvider)
	authRepo := auth.NewRepository(timeProvider)

	now := timeProvider.Now()
	tomorrow := now.AddDate(0, 0, 1)
	timeProvider.SetTime(tomorrow)

	svc := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		repo,
		authRepo,
	)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, author.UUID)
	updatedContents := "print('FizzBuzz')"
	updatedCodeSpace, codeSpaceAccess, err := svc.UpdateCodeSpace(ctx, codeSpace.Name, &updatedContents)
	require.NoError(t, err)

	require.NotNil(t, codeSpace.AuthorUUID)
	require.Equal(t, author.UUID, *codeSpace.AuthorUUID)
	require.Equal(t, codeSpace.ID, updatedCodeSpace.ID)
	require.Equal(t, codeSpace.Name, updatedCodeSpace.Name)
	require.Equal(t, codeSpace.Language, updatedCodeSpace.Language)
	require.Equal(t, updatedContents, updatedCodeSpace.Contents)
	require.WithinDuration(t, now, updatedCodeSpace.CreatedAt, testkit.TimeToleranceTentative)
	require.WithinDuration(t, tomorrow, updatedCodeSpace.UpdatedAt, testkit.TimeToleranceTentative)
	require.Equal(t, code.CodeSpaceAccessLevelReadWrite, codeSpaceAccess.Level)
}

func TestServiceUpdateCodeSpaceEditorSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

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

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	pistonClient := pistonmocks.NewMockClient(ctrl)
	repo := code.NewRepository(timeProvider)
	authRepo := auth.NewRepository(timeProvider)

	now := timeProvider.Now()
	tomorrow := now.AddDate(0, 0, 1)
	timeProvider.SetTime(tomorrow)

	svc := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		repo,
		authRepo,
	)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, editor.UUID)
	updatedContents := "print('FizzBuzz')"
	updatedCodeSpace, codeSpaceAccess, err := svc.UpdateCodeSpace(ctx, codeSpace.Name, &updatedContents)
	require.NoError(t, err)

	require.NotNil(t, codeSpace.AuthorUUID)
	require.Equal(t, author.UUID, *codeSpace.AuthorUUID)
	require.Equal(t, codeSpace.ID, updatedCodeSpace.ID)
	require.Equal(t, codeSpace.Name, updatedCodeSpace.Name)
	require.Equal(t, codeSpace.Language, updatedCodeSpace.Language)
	require.Equal(t, updatedContents, updatedCodeSpace.Contents)
	require.WithinDuration(t, now, updatedCodeSpace.CreatedAt, testkit.TimeToleranceTentative)
	require.WithinDuration(t, tomorrow, updatedCodeSpace.UpdatedAt, testkit.TimeToleranceTentative)
	require.Equal(t, code.CodeSpaceAccessLevelReadWrite, codeSpaceAccess.Level)
}

func TestServiceUpdateCodeSpaceFails(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	viewer, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		viewer.UUID,
		codeSpace.ID,
		code.CodeSpaceAccessLevelReadOnly,
	)

	thirdPartyUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testcases := map[string]struct {
		userUUID string
		wantErr  error
	}{
		"Viewer cannot update code space": {
			userUUID: viewer.UUID,
			wantErr:  errutils.ErrCodeSpaceAccessDenied,
		},
		"Third party user cannot update code space": {
			userUUID: thirdPartyUser.UUID,
			wantErr:  errutils.ErrCodeSpaceNotFound,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := code.NewRepository(timeProvider)
			authRepo := auth.NewRepository(timeProvider)

			now := timeProvider.Now()
			tomorrow := now.AddDate(0, 0, 1)
			timeProvider.SetTime(tomorrow)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, testcase.userUUID)
			updatedContents := "print('FizzBuzz')"
			_, _, err := svc.UpdateCodeSpace(ctx, codeSpace.Name, &updatedContents)
			require.Error(t, err)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceUpdateCodeSpaceError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	authorUUID := uuid.NewString()
	codeSpace := &code.CodeSpace{
		ID:         42,
		AuthorUUID: &authorUUID,
		Name:       "habitable-slaking-volatile-granger-mov",
		Language:   "python",
		Contents:   "print('hello')",
	}
	codeSpaceAccess := &code.CodeSpaceAccess{
		ID:          314,
		UserUUID:    authorUUID,
		CodeSpaceID: codeSpace.ID,
		Level:       code.CodeSpaceAccessLevelReadWrite,
	}

	updatedContents := "print('FizzBuzz')"
	genericRepoGetErr := errors.New("GetCodeSpaceWithAccessByName failed")
	genericRepoUpdateErr := errors.New("UpdateCodeSpace failed")

	testcases := map[string]struct {
		ctx           context.Context
		repoGetErr    error
		repoUpdateErr error
		wantErr       error
	}{
		"No user UUID in context": {
			ctx:           context.Background(),
			repoGetErr:    nil,
			repoUpdateErr: nil,
			wantErr:       nil,
		},
		"GetCodeSpaceWithAccessByName fails, no rows returned": {
			ctx:           context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:    errutils.ErrDatabaseNoRowsReturned,
			repoUpdateErr: nil,
			wantErr:       errutils.ErrCodeSpaceNotFound,
		},
		"GetCodeSpaceWithAccessByName fails, generic error": {
			ctx:           context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:    genericRepoGetErr,
			repoUpdateErr: nil,
			wantErr:       genericRepoGetErr,
		},
		"UpdateCodeSpace fails": {
			ctx:           context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:    nil,
			repoUpdateErr: genericRepoUpdateErr,
			wantErr:       genericRepoUpdateErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			repo.
				EXPECT().
				GetCodeSpaceWithAccessByName(gomock.Any(), gomock.Any(), authorUUID, codeSpace.Name).
				Return(codeSpace, codeSpaceAccess, testcase.repoGetErr).
				MaxTimes(1)

			repo.
				EXPECT().
				UpdateCodeSpace(gomock.Any(), gomock.Any(), codeSpace.ID, &updatedContents).
				Return(codeSpace, testcase.repoUpdateErr).
				MaxTimes(1)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			_, _, err := svc.UpdateCodeSpace(testcase.ctx, codeSpace.Name, &updatedContents)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceDeleteCodeSpaceAuthorSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	pistonClient := pistonmocks.NewMockClient(ctrl)
	repo := code.NewRepository(timeProvider)
	authRepo := auth.NewRepository(timeProvider)

	svc := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		repo,
		authRepo,
	)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, author.UUID)
	err := svc.DeleteCodeSpace(ctx, codeSpace.Name)
	require.NoError(t, err)

	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	_, err = repo.GetCodeSpace(context.Background(), dbConn, codeSpace.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
}

func TestServiceDeleteCodeSpaceEditorSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

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

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := mailclientmocks.NewMockClient(ctrl)
	tmplManager := templatesmanagermocks.NewMockManager(ctrl)
	pistonClient := pistonmocks.NewMockClient(ctrl)
	repo := code.NewRepository(timeProvider)
	authRepo := auth.NewRepository(timeProvider)

	svc := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		repo,
		authRepo,
	)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, editor.UUID)
	err := svc.DeleteCodeSpace(ctx, codeSpace.Name)
	require.NoError(t, err)

	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	_, err = repo.GetCodeSpace(context.Background(), dbConn, codeSpace.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
}

func TestServiceDeleteCodeSpaceFails(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	viewer, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	testkitinternal.MustCreateCodeSpaceAccess(
		t,
		viewer.UUID,
		codeSpace.ID,
		code.CodeSpaceAccessLevelReadOnly,
	)

	thirdPartyUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testcases := map[string]struct {
		userUUID string
		wantErr  error
	}{
		"Viewer cannot delete code space": {
			userUUID: viewer.UUID,
			wantErr:  errutils.ErrCodeSpaceAccessDenied,
		},
		"Third party user cannot delete code space": {
			userUUID: thirdPartyUser.UUID,
			wantErr:  errutils.ErrCodeSpaceNotFound,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := code.NewRepository(timeProvider)
			authRepo := auth.NewRepository(timeProvider)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, testcase.userUUID)
			err := svc.DeleteCodeSpace(ctx, codeSpace.Name)
			require.Error(t, err)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceDeleteCodeSpaceError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	authorUUID := uuid.NewString()
	codeSpace := &code.CodeSpace{
		ID:         42,
		AuthorUUID: &authorUUID,
		Name:       "habitable-slaking-volatile-granger-mov",
		Language:   "python",
		Contents:   "print('hello')",
	}
	codeSpaceAccess := &code.CodeSpaceAccess{
		ID:          314,
		UserUUID:    authorUUID,
		CodeSpaceID: codeSpace.ID,
		Level:       code.CodeSpaceAccessLevelReadWrite,
	}

	genericRepoGetErr := errors.New("GetCodeSpaceWithAccessByName failed")
	genericRepoDeleteErr := errors.New("DeleteCodeSpace failed")

	testcases := map[string]struct {
		ctx           context.Context
		repoGetErr    error
		repoDeleteErr error
		wantErr       error
	}{
		"No user UUID in context": {
			ctx:           context.Background(),
			repoGetErr:    nil,
			repoDeleteErr: nil,
			wantErr:       nil,
		},
		"GetCodeSpaceWithAccessByName fails, no rows returned": {
			ctx:           context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:    errutils.ErrDatabaseNoRowsReturned,
			repoDeleteErr: nil,
			wantErr:       errutils.ErrCodeSpaceNotFound,
		},
		"GetCodeSpaceWithAccessByName fails, generic error": {
			ctx:           context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:    genericRepoGetErr,
			repoDeleteErr: nil,
			wantErr:       genericRepoGetErr,
		},
		"DeleteCodeSpace fails": {
			ctx:           context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:    nil,
			repoDeleteErr: genericRepoDeleteErr,
			wantErr:       genericRepoDeleteErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			repo.
				EXPECT().
				GetCodeSpaceWithAccessByName(gomock.Any(), gomock.Any(), authorUUID, codeSpace.Name).
				Return(codeSpace, codeSpaceAccess, testcase.repoGetErr).
				MaxTimes(1)

			repo.
				EXPECT().
				DeleteCodeSpace(gomock.Any(), gomock.Any(), codeSpace.ID).
				Return(testcase.repoDeleteErr).
				MaxTimes(1)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			err := svc.DeleteCodeSpace(testcase.ctx, codeSpace.Name)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceRunCodeSpaceSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	exitCodeZero := 0

	testcases := map[string]struct {
		language            string
		wantVersion         string
		wantCompileResponse *api.PistonResults
		wantRunResponse     api.PistonResults
	}{
		"Run C code space": {
			language:    api.PistonLanguageC,
			wantVersion: api.PistonVersionC,
			wantCompileResponse: &api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "",
				Stderr: "",
				Output: "",
			},
			wantRunResponse: api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "Yello!\n",
				Stderr: "",
				Output: "Yello!\n",
			},
		},
		"Run C++ code space": {
			language:    api.PistonLanguageCPlusPlus,
			wantVersion: api.PistonVersionCPlusPlus,
			wantCompileResponse: &api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "",
				Stderr: "",
				Output: "",
			},
			wantRunResponse: api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "Yello!\n",
				Stderr: "",
				Output: "Yello!\n",
			},
		},
		"Run Go code space": {
			language:            api.PistonLanguageGo,
			wantVersion:         api.PistonVersionGo,
			wantCompileResponse: nil,
			wantRunResponse: api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "Yello!\n",
				Stderr: "",
				Output: "Yello!\n",
			},
		},
		"Run Java code space": {
			language:            api.PistonLanguageJava,
			wantVersion:         api.PistonVersionJava,
			wantCompileResponse: nil,
			wantRunResponse: api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "Yello!\n",
				Stderr: "",
				Output: "Yello!\n",
			},
		},
		"Run JavaScript code space": {
			language:            api.PistonLanguageJavaScript,
			wantVersion:         api.PistonVersionJavaScript,
			wantCompileResponse: nil,
			wantRunResponse: api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "Yello!\n",
				Stderr: "",
				Output: "Yello!\n",
			},
		},
		"Run Python code space": {
			language:            api.PistonLanguagePython,
			wantVersion:         api.PistonVersionPython,
			wantCompileResponse: nil,
			wantRunResponse: api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "Yello!\n",
				Stderr: "",
				Output: "Yello!\n",
			},
		},
		"Run Rust code space": {
			language:    api.PistonLanguageRust,
			wantVersion: api.PistonVersionRust,
			wantCompileResponse: &api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "",
				Stderr: "",
				Output: "",
			},
			wantRunResponse: api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "Yello!\n",
				Stderr: "",
				Output: "Yello!\n",
			},
		},
		"Run TypeScript code space": {
			language:            api.PistonLanguageTypeScript,
			wantVersion:         api.PistonVersionTypeScript,
			wantCompileResponse: nil,
			wantRunResponse: api.PistonResults{
				Code:   &exitCodeZero,
				Signal: nil,
				Stdout: "Yello!\n",
				Stderr: "",
				Output: "Yello!\n",
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, testcase.language)

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := piston.NewClient(nil, httputils.NewHTTPClient(nil))
			repo := code.NewRepository(timeProvider)
			authRepo := auth.NewRepository(timeProvider)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, author.UUID)
			pistonResponse, err := svc.RunCodeSpace(ctx, codeSpace.Name)
			require.NoError(t, err)
			require.Equal(t, testcase.language, pistonResponse.Language)
			require.Equal(t, testcase.wantVersion, pistonResponse.Version)
			require.Equal(t, testcase.wantCompileResponse, pistonResponse.Compile)
			require.Equal(t, testcase.wantRunResponse, pistonResponse.Run)
		})
	}
}

func TestServiceRunCodeSpaceError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	authorUUID := uuid.NewString()

	genericRepoErr := errors.New("GetCodeSpaceWithAccessByName failed")
	genericPistonErr := errors.New("Execute failed")

	testcases := map[string]struct {
		ctx       context.Context
		language  string
		repoErr   error
		pistonErr error
		wantErr   error
	}{
		"No user UUID in context": {
			ctx:       context.Background(),
			language:  "python",
			repoErr:   nil,
			pistonErr: nil,
			wantErr:   nil,
		},
		"GetCodeSpaceWithAccessByName fails, no rows returned": {
			ctx:       context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			language:  "python",
			repoErr:   errutils.ErrDatabaseNoRowsReturned,
			pistonErr: nil,
			wantErr:   errutils.ErrCodeSpaceNotFound,
		},
		"GetCodeSpaceWithAccessByName fails, generic error": {
			ctx:       context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			language:  "python",
			repoErr:   genericRepoErr,
			pistonErr: nil,
			wantErr:   genericRepoErr,
		},
		"Unknown language": {
			ctx:       context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			language:  "unknown",
			repoErr:   nil,
			pistonErr: nil,
			wantErr:   nil,
		},
		"Execute fails": {
			ctx:       context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			language:  "python",
			repoErr:   nil,
			pistonErr: genericPistonErr,
			wantErr:   genericPistonErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			pistonClient.
				EXPECT().
				Execute(gomock.Any()).
				Return(nil, testcase.pistonErr).
				MaxTimes(1)

			codeSpace := &code.CodeSpace{
				ID:         42,
				AuthorUUID: &authorUUID,
				Name:       "habitable-slaking-volatile-granger-mov",
				Language:   testcase.language,
				Contents:   "print('hello')",
			}
			codeSpaceAccess := &code.CodeSpaceAccess{
				ID:          314,
				UserUUID:    authorUUID,
				CodeSpaceID: codeSpace.ID,
				Level:       code.CodeSpaceAccessLevelReadWrite,
			}

			repo.
				EXPECT().
				GetCodeSpaceWithAccessByName(gomock.Any(), gomock.Any(), authorUUID, codeSpace.Name).
				Return(codeSpace, codeSpaceAccess, testcase.repoErr).
				MaxTimes(1)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			_, err := svc.RunCodeSpace(testcase.ctx, codeSpace.Name)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceListCodeSpaceUsers(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

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

	thirdPartyUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testcases := map[string]struct {
		userUUID string
		wantErr  error
	}{
		"Author can list codespace users": {
			userUUID: author.UUID,
			wantErr:  nil,
		},
		"Editor can list codespace users": {
			userUUID: editor.UUID,
			wantErr:  nil,
		},
		"Viewer gets access denied": {
			userUUID: viewer.UUID,
			wantErr:  errutils.ErrCodeSpaceAccessDenied,
		},
		"Third party user gets code space not found": {
			userUUID: thirdPartyUser.UUID,
			wantErr:  errutils.ErrCodeSpaceNotFound,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := code.NewRepository(timeProvider)
			authRepo := auth.NewRepository(timeProvider)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, testcase.userUUID)
			users, codeSpaceAccesses, err := svc.ListCodeSpaceUsers(ctx, codeSpace.Name)

			if testcase.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, testcase.wantErr)

				return
			}

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
		})
	}
}

func TestServiceListCodeSpaceUsersError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	authorUUID := uuid.NewString()

	genericRepoGetErr := errors.New("GetCodeSpaceWithAccessByName failed")
	genericRepoListErr := errors.New("ListUsersWithCodeSpaceAccess failed")

	testcases := map[string]struct {
		ctx         context.Context
		repoGetErr  error
		repoListErr error
		wantErr     error
	}{
		"No user UUID in context": {
			ctx:         context.Background(),
			repoGetErr:  nil,
			repoListErr: nil,
			wantErr:     nil,
		},
		"GetCodeSpaceWithAccessByName fails, no rows returned": {
			ctx:         context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:  errutils.ErrDatabaseNoRowsReturned,
			repoListErr: nil,
			wantErr:     errutils.ErrCodeSpaceNotFound,
		},
		"GetCodeSpaceWithAccessByName fails, generic error": {
			ctx:         context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:  genericRepoGetErr,
			repoListErr: nil,
			wantErr:     genericRepoGetErr,
		},
		"ListUsersWithCodeSpaceAccess fails": {
			ctx:         context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			repoGetErr:  nil,
			repoListErr: genericRepoListErr,
			wantErr:     genericRepoListErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			codeSpace := &code.CodeSpace{
				ID:         42,
				AuthorUUID: &authorUUID,
				Name:       "habitable-slaking-volatile-granger-mov",
				Language:   "python",
				Contents:   "print('hello')",
			}
			codeSpaceAccess := &code.CodeSpaceAccess{
				ID:          314,
				UserUUID:    authorUUID,
				CodeSpaceID: codeSpace.ID,
				Level:       code.CodeSpaceAccessLevelReadWrite,
			}

			repo.
				EXPECT().
				GetCodeSpaceWithAccessByName(gomock.Any(), gomock.Any(), authorUUID, codeSpace.Name).
				Return(codeSpace, codeSpaceAccess, testcase.repoGetErr).
				MaxTimes(1)

			repo.
				EXPECT().
				ListUsersWithCodeSpaceAccess(gomock.Any(), gomock.Any(), codeSpace.ID).
				Return(nil, nil, testcase.repoListErr).
				MaxTimes(1)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			_, _, err := svc.ListCodeSpaceUsers(testcase.ctx, codeSpace.Name)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceSendCodeSpaceInvitationMailSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	email := testkit.GenerateFakeEmail()

	ctrl := gomock.NewController(t)
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	crypto := cryptocoremocks.NewMockCrypto(ctrl)
	mailClient := testkit.NewInMemMailClient("support@nymphadora.com", timeProvider)
	tmplManager := templatesmanager.NewManager()
	pistonClient := pistonmocks.NewMockClient(ctrl)
	repo := codemocks.NewMockRepository(ctrl)
	authRepo := authmocks.NewMockRepository(ctrl)

	svc := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		repo,
		authRepo,
	)

	invitationURL := "http://localhost:3000/code/space/" +
		"habitable-slaking-volatile-granger-mov/invitation/" +
		"1nv1t4t10njwt"
	err := svc.SendCodeSpaceInvitationMail(
		context.Background(),
		email,
		templatesmanager.CodeSpaceInvitationEmailTemplateData{
			InvitationURL: invitationURL,
		},
	)
	require.NoError(t, err)
	require.Len(t, mailClient.Logs, 1)

	lastMail := mailClient.Logs[len(mailClient.Logs)-1]
	require.Equal(t, []string{email}, lastMail.To)
	require.Equal(t, "You've been invited to collaborate on a code space!", lastMail.Subject)
	require.WithinDuration(t, timeProvider.Now(), lastMail.SentAt, testkit.TimeToleranceExact)

	mailMessage := string(lastMail.Message)
	require.Contains(t, mailMessage, "Nymphadora - You've been invited to collaborate on a code space!")
	require.Contains(t, mailMessage, "Accept Invitation")
	require.Contains(t, mailMessage, invitationURL)
}

func TestServiceSendCodeSpaceInvitationMailError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	invitationURL := "http://localhost:3000/code/space/" +
		"habitable-slaking-volatile-granger-mov/invitation/" +
		"1nv1t4t10njwt"
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
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			tmplManager.
				EXPECT().
				Load("codespaceinvitation").
				Return(texttemplate.New("text"), htmltemplate.New("html"), testcase.tmplLoadErr).
				MaxTimes(1)

			mailClient.
				EXPECT().
				Send(
					[]string{email},
					"You've been invited to collaborate on a code space!",
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(testcase.mailSendErr).
				MaxTimes(1)

			err := svc.SendCodeSpaceInvitationMail(
				context.Background(),
				email,
				templatesmanager.CodeSpaceInvitationEmailTemplateData{
					InvitationURL: invitationURL,
				},
			)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceInviteCodeSpaceUserSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	editor, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	viewer, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testcases := map[string]struct {
		inviteeEmail string
		accessLevel  code.CodeSpaceAccessLevel
	}{
		"Author invites editor": {
			inviteeEmail: editor.Email,
			accessLevel:  code.CodeSpaceAccessLevelReadWrite,
		},
		"Author invites viewer": {
			inviteeEmail: viewer.Email,
			accessLevel:  code.CodeSpaceAccessLevelReadOnly,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
			mailClient := testkit.NewInMemMailClient("support@nymphadora.com", timeProvider)
			tmplManager := templatesmanager.NewManager()
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := code.NewRepository(timeProvider)
			authRepo := auth.NewRepository(timeProvider)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			mailCount := len(mailClient.Logs)

			ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, author.UUID)
			err := svc.InviteCodeSpaceUser(ctx, codeSpace.Name, testcase.inviteeEmail, testcase.accessLevel)
			require.NoError(t, err)

			require.Len(t, mailClient.Logs, mailCount+1)

			lastMail := mailClient.Logs[len(mailClient.Logs)-1]
			require.Equal(t, []string{testcase.inviteeEmail}, lastMail.To)
			require.Equal(t, "You've been invited to collaborate on a code space!", lastMail.Subject)
			require.WithinDuration(t, timeProvider.Now(), lastMail.SentAt, testkit.TimeToleranceExact)

			mailMessage := string(lastMail.Message)
			require.Contains(t, mailMessage, "Nymphadora - You've been invited to collaborate on a code space!")
			require.Contains(t, mailMessage, "Accept Invitation")

			pattern := fmt.Sprintf(
				cfg.FrontendBaseURL+code.FrontendCodeSpaceInvitationRoute,
				codeSpace.Name,
				`(\S+)`,
			)
			r, err := regexp.Compile(pattern)
			require.NoError(t, err)

			matches := r.FindStringSubmatch(mailMessage)
			require.Len(t, matches, 2)

			invitationToken := matches[1]
			claims := &cryptocore.CodeSpaceInvitationJWTClaims{}
			parsedToken, err := jwt.ParseWithClaims(invitationToken, claims, func(t *jwt.Token) (any, error) {
				return []byte(cfg.SecretKey), nil
			})
			require.NoError(t, err)

			require.NotNil(t, parsedToken)
			require.True(t, parsedToken.Valid)
			require.Equal(t, author.UUID, claims.Subject)
			require.Equal(t, testcase.inviteeEmail, claims.InviteeEmail)
			require.Equal(t, string(cryptocore.JWTTypeCodeSpaceInvitation), claims.TokenType)
			require.WithinDuration(t, timeProvider.Now(), time.Time(claims.IssuedAt), testkit.TimeToleranceExact)
			require.WithinDuration(
				t,
				timeProvider.Now().Add(cryptocore.JWTLifetimeCodeSpaceInvitation),
				time.Time(claims.ExpiresAt),
				testkit.TimeToleranceExact,
			)
		})
	}
}

func TestServiceInviteCodeSpaceUserError(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	authorUUID := uuid.NewString()
	inviteeEmail := testkit.GenerateFakeEmail()

	genericRepoErr := errors.New("GetCodeSpaceWithAccessByName failed")
	createJWTErr := errors.New("CreateCodeSpaceInvitationJWT failed")
	mailClientErr := errors.New("Send failed")

	testcases := map[string]struct {
		ctx               context.Context
		authorAccessLevel code.CodeSpaceAccessLevel
		repoErr           error
		createJWTErr      error
		mailClientErr     error
		wantErr           error
	}{
		"No user UUID in context": {
			ctx:               context.Background(),
			authorAccessLevel: code.CodeSpaceAccessLevelReadWrite,
			repoErr:           nil,
			createJWTErr:      nil,
			mailClientErr:     nil,
			wantErr:           nil,
		},
		"Access denied without write access": {
			ctx:               context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			authorAccessLevel: code.CodeSpaceAccessLevelReadOnly,
			repoErr:           nil,
			createJWTErr:      nil,
			mailClientErr:     nil,
			wantErr:           errutils.ErrCodeSpaceAccessDenied,
		},
		"GetCodeSpaceWithAccessByName fails, no rows returned": {
			ctx:               context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			authorAccessLevel: code.CodeSpaceAccessLevelReadWrite,
			repoErr:           errutils.ErrDatabaseNoRowsReturned,
			createJWTErr:      nil,
			mailClientErr:     nil,
			wantErr:           errutils.ErrCodeSpaceNotFound,
		},
		"GetCodeSpaceWithAccessByName fails, generic error": {
			ctx:               context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			authorAccessLevel: code.CodeSpaceAccessLevelReadWrite,
			repoErr:           genericRepoErr,
			createJWTErr:      nil,
			mailClientErr:     nil,
			wantErr:           genericRepoErr,
		},
		"CreateCodeSpaceInvitationJWT fails": {
			ctx:               context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			authorAccessLevel: code.CodeSpaceAccessLevelReadWrite,
			repoErr:           nil,
			createJWTErr:      createJWTErr,
			mailClientErr:     nil,
			wantErr:           createJWTErr,
		},
		"mailClient.Send fails": {
			ctx:               context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID),
			authorAccessLevel: code.CodeSpaceAccessLevelReadWrite,
			repoErr:           nil,
			createJWTErr:      nil,
			mailClientErr:     mailClientErr,
			wantErr:           mailClientErr,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			timeProvider := timekeeper.NewFrozenProvider()
			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			crypto := cryptocoremocks.NewMockCrypto(ctrl)
			mailClient := mailclientmocks.NewMockClient(ctrl)
			tmplManager := templatesmanagermocks.NewMockManager(ctrl)
			pistonClient := pistonmocks.NewMockClient(ctrl)
			repo := codemocks.NewMockRepository(ctrl)
			authRepo := authmocks.NewMockRepository(ctrl)

			codeSpace := &code.CodeSpace{
				ID:         42,
				AuthorUUID: &authorUUID,
				Name:       "habitable-slaking-volatile-granger-mov",
				Language:   "python",
				Contents:   "print('hello')",
			}
			codeSpaceAccess := &code.CodeSpaceAccess{
				ID:          314,
				UserUUID:    authorUUID,
				CodeSpaceID: codeSpace.ID,
				Level:       testcase.authorAccessLevel,
			}

			repo.
				EXPECT().
				GetCodeSpaceWithAccessByName(gomock.Any(), gomock.Any(), authorUUID, codeSpace.Name).
				Return(codeSpace, codeSpaceAccess, testcase.repoErr).
				MaxTimes(1)

			crypto.
				EXPECT().
				CreateCodeSpaceInvitationJWT(
					authorUUID,
					inviteeEmail,
					codeSpace.ID,
					int(code.CodeSpaceAccessLevelReadWrite),
				).
				Return("1nv1t4t10nt0k3n", testcase.createJWTErr).
				MaxTimes(1)

			tmplManager.
				EXPECT().
				Load("codespaceinvitation").
				Return(texttemplate.New("text"), htmltemplate.New("html"), nil).
				MaxTimes(1)

			mailClient.
				EXPECT().
				Send(
					[]string{inviteeEmail},
					"You've been invited to collaborate on a code space!",
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(testcase.mailClientErr).
				MaxTimes(1)

			svc := code.NewService(
				cfg,
				timeProvider,
				dbPool,
				crypto,
				mailClient,
				tmplManager,
				pistonClient,
				repo,
				authRepo,
			)

			err := svc.InviteCodeSpaceUser(
				testcase.ctx,
				codeSpace.Name,
				inviteeEmail,
				code.CodeSpaceAccessLevelReadWrite,
			)
			require.Error(t, err)

			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceCodeSpace(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	invitee, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	cfg := testkitinternal.MustCreateConfig()

	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := testkit.NewInMemMailClient("support@nymphadora.com", timeProvider)
	tmplManager := templatesmanager.NewManager()
	pistonClient := piston.NewClient(nil, httputils.NewHTTPClient(nil))
	repo := code.NewRepository(timeProvider)
	authRepo := auth.NewRepository(timeProvider)

	svc := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		repo,
		authRepo,
	)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)

	codeSpace, codeSpaceAccess, err := svc.CreateCodeSpace(ctx, "python")
	require.NoError(t, err)

	require.NotNil(t, codeSpace.AuthorUUID)
	require.Equal(t, user.UUID, *codeSpace.AuthorUUID)
	require.Equal(t, "python", codeSpace.Language)

	require.Equal(t, user.UUID, codeSpaceAccess.UserUUID)
	require.Equal(t, codeSpace.ID, codeSpaceAccess.CodeSpaceID)
	require.Equal(t, code.CodeSpaceAccessLevelReadWrite, codeSpaceAccess.Level)

	codeSpaces, _, err := svc.ListCodeSpaces(ctx)
	require.NoError(t, err)

	require.Len(t, codeSpaces, 1)
	require.Equal(t, codeSpace.ID, codeSpaces[0].ID)

	contents := "print('FUCK YEAAAA')\n"

	codeSpace, _, err = svc.UpdateCodeSpace(ctx, codeSpace.Name, &contents)
	require.NoError(t, err)

	require.Equal(t, contents, codeSpace.Contents)

	_, err = svc.RunCodeSpace(ctx, codeSpace.Name)
	require.NoError(t, err)

	codeSpaceUsers, codeSpaceAccesses, err := svc.ListCodeSpaceUsers(ctx, codeSpace.Name)
	require.NoError(t, err)

	require.Len(t, codeSpaceUsers, 1)
	require.Len(t, codeSpaceAccesses, 1)

	require.Equal(t, user.UUID, codeSpaceUsers[0].UUID)
	require.Equal(t, codeSpaceAccess.ID, codeSpaceAccesses[0].ID)

	mailCount := len(mailClient.Logs)

	err = svc.InviteCodeSpaceUser(ctx, codeSpace.Name, invitee.Email, code.CodeSpaceAccessLevelReadOnly)
	require.NoError(t, err)

	require.Len(t, mailClient.Logs, mailCount+1)

	lastMail := mailClient.Logs[len(mailClient.Logs)-1]
	require.Equal(t, []string{invitee.Email}, lastMail.To)
	require.Equal(t, "You've been invited to collaborate on a code space!", lastMail.Subject)
	require.WithinDuration(t, timeProvider.Now(), lastMail.SentAt, testkit.TimeToleranceExact)

	mailMessage := string(lastMail.Message)
	require.Contains(t, mailMessage, "You've been invited to collaborate on a code space!")

	pattern := fmt.Sprintf(cfg.FrontendBaseURL+code.FrontendCodeSpaceInvitationRoute, codeSpace.Name, `(\S+)`)
	r, err := regexp.Compile(pattern)
	require.NoError(t, err)

	matches := r.FindStringSubmatch(mailMessage)
	require.Len(t, matches, 2)

	invitationToken := matches[1]
	claims := &cryptocore.CodeSpaceInvitationJWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(invitationToken, claims, func(t *jwt.Token) (any, error) {
		return []byte(cfg.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedToken)
	require.True(t, parsedToken.Valid)
	require.Equal(t, user.UUID, claims.Subject)
	require.Equal(t, invitee.Email, claims.InviteeEmail)
	require.Equal(t, codeSpace.ID, claims.CodeSpaceID)
	require.EqualValues(t, code.CodeSpaceAccessLevelReadOnly, claims.AccessLevel)
	require.Equal(t, string(cryptocore.JWTTypeCodeSpaceInvitation), claims.TokenType)
	require.WithinDuration(t, timeProvider.Now(), time.Time(claims.IssuedAt), testkit.TimeToleranceExact)
	require.WithinDuration(
		t,
		timeProvider.Now().Add(cryptocore.JWTLifetimeCodeSpaceInvitation),
		time.Time(claims.ExpiresAt),
		testkit.TimeToleranceExact,
	)

	_, inviteeCodeSpaceAccess, err := svc.AcceptCodeSpaceUserInvitation(ctx, codeSpace.Name, invitationToken)
	require.NoError(t, err)

	require.Equal(t, invitee.UUID, inviteeCodeSpaceAccess.UserUUID)
	require.Equal(t, codeSpace.ID, inviteeCodeSpaceAccess.CodeSpaceID)
	require.Equal(t, code.CodeSpaceAccessLevelReadOnly, inviteeCodeSpaceAccess.Level)

	_, err = svc.RunCodeSpace(
		context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, invitee.UUID),
		codeSpace.Name,
	)
	require.NoError(t, err)

	err = svc.RemoveCodeSpaceUser(ctx, codeSpace.Name, invitee.UUID)
	require.NoError(t, err)

	_, err = svc.RunCodeSpace(
		context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, invitee.UUID),
		codeSpace.Name,
	)
	require.Error(t, err)

	err = svc.DeleteCodeSpace(ctx, codeSpace.Name)
	require.NoError(t, err)

	_, err = svc.RunCodeSpace(ctx, codeSpace.Name)
	require.ErrorIs(t, err, errutils.ErrCodeSpaceNotFound)
}
