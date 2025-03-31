package testkitinternal

import (
	"github.com/alvii147/nymphadora-api/internal/database"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
)

// MustNewDatabasePool creates, returns a new database connection pool, and panics on error.
func MustNewDatabasePool() database.Pool {
	cfg := MustCreateConfig()

	dbPool, err := database.NewPool(
		cfg.PostgresHostname,
		cfg.PostgresPort,
		cfg.PostgresUsername,
		cfg.PostgresPassword,
		cfg.PostgresDatabaseName,
	)
	if err != nil {
		panic(errutils.FormatError(err))
	}

	return dbPool
}
