package testkitinternal

import (
	"context"
	"os"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/internal/code"
	"github.com/alvii147/nymphadora-api/internal/templatesmanager"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/alvii147/nymphadora-api/pkg/mailclient"
	"github.com/alvii147/nymphadora-api/pkg/piston"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

// MustCreateCodeSpace creates and returns a new code space for a given author UUID and panics on error.
func MustCreateCodeSpace(
	t testkit.TestingT,
	authorUUID string,
	language string,
) (*code.CodeSpace, *code.CodeSpaceAccess) {
	cfg := MustCreateConfig()
	timeProvider := timekeeper.NewFrozenProvider()

	dbPool := MustNewDatabasePool()
	defer dbPool.Close()

	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)
	mailClient := mailclient.NewConsoleClient("support@nymphadora.com", timeProvider, os.Stdout)
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

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, authorUUID)
	codeSpace, codeSpaceAccess, err := svc.CreateCodeSpace(
		ctx,
		language,
	)
	if err != nil {
		panic(errutils.FormatError(err))
	}

	return codeSpace, codeSpaceAccess
}

// MustCreateCodeSpaceAccess creates and returns a new code space access for a given user UUID and code space.
func MustCreateCodeSpaceAccess(
	t testkit.TestingT,
	userUUID string,
	codeSpaceID int64,
	accessLevel code.CodeSpaceAccessLevel,
) *code.CodeSpaceAccess {
	timeProvider := timekeeper.NewFrozenProvider()
	dbPool := MustNewDatabasePool()
	defer dbPool.Close()

	dbConn, err := dbPool.Acquire(context.Background())
	require.NoError(t, err)
	defer dbConn.Release()

	repo := code.NewRepository(timeProvider)

	codeSpaceAccess := &code.CodeSpaceAccess{
		UserUUID:    userUUID,
		CodeSpaceID: codeSpaceID,
		Level:       accessLevel,
	}

	codeSpaceAccess, err = repo.CreateOrUpdateCodeSpaceAccess(context.Background(), dbConn, codeSpaceAccess)
	if err != nil {
		panic(errutils.FormatError(err))
	}

	return codeSpaceAccess
}
