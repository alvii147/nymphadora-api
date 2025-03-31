package database

import (
	"context"
	"fmt"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectionFormatString is the string format for database connection strings.
const ConnectionFormatString = "postgres://%s:%s@%s:%d/%s"

// Pool represents a database connection pool.
//
//go:generate mockgen -package=databasemocks -source=$GOFILE -destination=./mocks/pool.go
type Pool interface {
	Acquire(ctx context.Context) (Conn, error)
	Close()
}

// pool implements Pool.
type pool struct {
	*pgxpool.Pool
}

// Acquire returns a database connection from the connection pool.
func (p *pool) Acquire(ctx context.Context) (Conn, error) {
	c, err := p.Pool.Acquire(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "p.Pool.Acquire failed")
	}

	return &conn{c}, nil
}

// NewPool creates and returns a new database connection pool.
func NewPool(
	hostname string,
	port int,
	username string,
	password string,
	databaseName string,
) (Pool, error) {
	connString := fmt.Sprintf(
		ConnectionFormatString,
		username,
		password,
		hostname,
		port,
		databaseName,
	)

	p, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, errutils.FormatErrorf(err, "pgxpool.New failed for connection string %s", connString)
	}

	return &pool{p}, nil
}
