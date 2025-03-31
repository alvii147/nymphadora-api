package database

import (
	"context"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Conn represents a database connection.
//
//go:generate mockgen -package=databasemocks -source=$GOFILE -destination=./mocks/conn.go
type Conn interface {
	Begin(ctx context.Context) (Tx, error)
	Release()
	Querier
}

// conn implements Conn.
type conn struct {
	*pgxpool.Conn
}

// Begin initiates a database transaction.
func (c *conn) Begin(ctx context.Context) (Tx, error) {
	tx, err := c.Conn.Begin(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "c.Conn.Begin failed")
	}

	return tx, nil
}
