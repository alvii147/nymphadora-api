package database

import (
	"context"
	"fmt"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateConnString constructs PostgreSQL connection string from Config.
func CreateConnString(
	hostname string,
	port int,
	username string,
	password string,
	databaseName string,
) string {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		username,
		password,
		hostname,
		port,
		databaseName,
	)

	return connString
}

// CreatePool creates and returns a new database connection pool.
func CreatePool(
	hostname string,
	port int,
	username string,
	password string,
	databaseName string,
) (*pgxpool.Pool, error) {
	connString := CreateConnString(
		hostname,
		port,
		username,
		password,
		databaseName,
	)

	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, errutils.FormatErrorf(err, "pgxpool.New failed for connection string %s", connString)
	}

	return dbPool, nil
}
