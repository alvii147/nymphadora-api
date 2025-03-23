package testkitinternal

import (
	"context"

	"github.com/alvii147/nymphadora-api/internal/config"
	"github.com/alvii147/nymphadora-api/internal/database"
	"github.com/alvii147/nymphadora-api/pkg/env"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// RequireCreateDatabasePool creates and returns a new database connection pool.
// It also asserts no error is returned and declares clean up function to close the pool.
func RequireCreateDatabasePool(t testkit.TestingT) *pgxpool.Pool {
	cfg := &config.Config{}
	err := env.NewConfig(cfg)
	require.NoError(t, err)

	dbPool, err := database.CreatePool(
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

// RequireCreateDatabaseConn creates and returns a new database connection from a given connection pool.
// It also asserts no error is returned and declares clean up function to close the connection.
func RequireCreateDatabaseConn(t testkit.TestingT, dbPool *pgxpool.Pool, ctx context.Context) *pgxpool.Conn {
	dbConn, err := dbPool.Acquire(ctx)
	require.NoError(t, err)

	t.Cleanup(dbConn.Release)

	return dbConn
}
