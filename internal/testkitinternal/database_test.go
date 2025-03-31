package testkitinternal_test

import (
	"context"
	"errors"
	"testing"

	databasemocks "github.com/alvii147/nymphadora-api/internal/database/mocks"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRequireNewDatabasePool(t *testing.T) {
	t.Parallel()

	mockT := testkit.NewMockTestingT()
	testkitinternal.RequireNewDatabasePool(mockT)

	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Cleanups, 1)

	cleanupFunc := mockT.Cleanups[0]
	cleanupFunc()
}

func TestRequireNewDatabaseConnSuccess(t *testing.T) {
	t.Parallel()

	mockT := testkit.NewMockTestingT()
	dbPool := testkitinternal.RequireNewDatabasePool(t)
	ctx := context.Background()

	testkitinternal.RequireNewDatabaseConn(mockT, dbPool, ctx)

	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Cleanups, 1)

	cleanupFunc := mockT.Cleanups[0]
	cleanupFunc()
}

func TestRequireNewDatabaseConnCleanup(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockT := testkit.NewMockTestingT()
	dbPool := databasemocks.NewMockPool(ctrl)
	dbConn := databasemocks.NewMockConn(ctrl)
	ctx := context.Background()

	dbConn.
		EXPECT().
		Release().
		Times(1)

	dbPool.
		EXPECT().
		Acquire(gomock.Any()).
		Return(dbConn, nil).
		Times(1)

	testkitinternal.RequireNewDatabaseConn(mockT, dbPool, ctx)

	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Cleanups, 1)

	cleanupFunc := mockT.Cleanups[0]
	cleanupFunc()
}

func TestRequireCreateDatabaseConnError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockT := testkit.NewMockTestingT()
	dbPool := databasemocks.NewMockPool(ctrl)
	dbConn := databasemocks.NewMockConn(ctrl)
	ctx := context.Background()

	dbPool.
		EXPECT().
		Acquire(gomock.Any()).
		Return(dbConn, errors.New("Acquire failed")).
		Times(1)

	testkitinternal.RequireNewDatabaseConn(mockT, dbPool, ctx)

	require.True(t, mockT.HasFailed)
}
