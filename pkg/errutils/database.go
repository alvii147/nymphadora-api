package errutils

import "errors"

// Database error codes for PostgreSQL.
const (
	// DatabaseErrCodeForeignKeyViolation is the PostgreSQL error code for a foreign key constraint violation.
	DatabaseErrCodeForeignKeyViolation = "23503"
	// DatabaseErrCodeUniqueViolation is the PostgreSQL error code for a unique key constraint violation.
	DatabaseErrCodeUniqueViolation = "23505"
)

// Database errors returned from repository layers.
var (
	// ErrDatabaseForeignKeyConstraintViolation is the error used for a foreign key constraint violation.
	ErrDatabaseForeignKeyConstraintViolation = errors.New("foreign key constaint violation")
	// ErrDatabaseUniqueViolation is the error used for a unique key contraint violation.
	ErrDatabaseUniqueViolation = errors.New("unique key constraint violation")
	// ErrDatabaseNoRowsReturned is the error used when no rows are found.
	ErrDatabaseNoRowsReturned = errors.New("no rows returned")
	// ErrDatabaseNoRowsAffected is the error used when no rows are affected.
	ErrDatabaseNoRowsAffected = errors.New("no rows affected")
)
