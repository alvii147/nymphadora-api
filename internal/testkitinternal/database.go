package testkitinternal

import (
	"context"

	"github.com/alvii147/nymphadora-api/internal/config"
	"github.com/alvii147/nymphadora-api/internal/database"
	"github.com/alvii147/nymphadora-api/pkg/env"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

// RequireNewDatabasePool creates and returns a new database connection pool.
// It also asserts no error is returned and declares clean up function to close the pool.
func RequireNewDatabasePool(t testkit.TestingT) database.Pool {
	cfg := &config.Config{}
	err := env.NewConfig(cfg)
	require.NoError(t, err)

	dbPool, err := database.NewPool(
		cfg.PostgresHostname,
		cfg.PostgresPort,
		cfg.PostgresUsername,
		cfg.PostgresPassword,
		cfg.PostgresDatabaseName,
	)
	require.NoError(t, err)

	t.Cleanup(dbPool.Close)

	return dbPool
}

// RequireNewDatabaseConn creates and returns a new database connection from a given connection pool.
// It also asserts no error is returned and declares clean up function to close the connection.
func RequireNewDatabaseConn(t testkit.TestingT, dbPool database.Pool, ctx context.Context) database.Conn {
	dbConn, err := dbPool.Acquire(ctx)
	require.NoError(t, err)

	t.Cleanup(dbConn.Release)

	return dbConn
}
