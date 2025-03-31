package database_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/internal/database"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/stretchr/testify/require"
)

func TestNewPoolSuccess(t *testing.T) {
	t.Parallel()

	cfg := testkitinternal.MustCreateConfig()

	dbPool, err := database.NewPool(
		cfg.PostgresHostname,
		cfg.PostgresPort,
		cfg.PostgresUsername,
		cfg.PostgresPassword,
		cfg.PostgresDatabaseName,
	)
	require.NoError(t, err)
	require.NotNil(t, dbPool)
}

func TestNewPoolBadConnString(t *testing.T) {
	t.Parallel()

	_, err := database.NewPool("", 0, "", "", "")
	require.Error(t, err)
}
