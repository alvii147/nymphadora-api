package database

import "context"

// Tx represents a database transaction.
//
//go:generate mockgen -package=databasemocks -source=$GOFILE -destination=./mocks/tx.go
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Querier
}
