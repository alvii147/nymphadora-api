package auth

import (
	"context"
	"errors"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is used to access and update auth data.
//
//go:generate mockgen -package=authmocks -source=$GOFILE -destination=./mocks/repository.go
type Repository interface {
	CreateUser(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		user *User,
	) (*User, error)
	ActivateUserByUUID(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
	) error
	GetUserByEmail(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		email string,
	) (*User, error)
	GetUserByUUID(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
	) (*User, error)
	UpdateUser(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
		firstName *string,
		lastName *string,
	) (*User, error)
	CreateAPIKey(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		apiKey *APIKey,
	) (*APIKey, error)
	ListAPIKeysByUserUUID(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
	) ([]*APIKey, error)
	ListActiveAPIKeysByPrefix(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		prefix string,
	) ([]*APIKey, error)
	UpdateAPIKey(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
		apiKeyID int64,
		name *string,
		expiresAt jsonutils.Optional[time.Time],
	) (*APIKey, error)
	DeleteAPIKey(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
		apiKeyID int64,
	) error
}

// repository implements Repository.
type repository struct {
	timeProvider timekeeper.Provider
}

// NewRepository returns a new repository.
func NewRepository(timeProvider timekeeper.Provider) *repository {
	return &repository{
		timeProvider: timeProvider,
	}
}

// CreateUser creates a new user.
func (repo *repository) CreateUser(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	user *User,
) (*User, error) {
	now := repo.timeProvider.Now()
	createdUser := &User{}

	q := `
INSERT INTO "user" (
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at,
	updated_at
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9
)
RETURNING
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at,
	updated_at;
	`

	err := dbConn.QueryRow(
		ctx,
		q,
		user.UUID,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.IsSuperUser,
		now,
		now,
	).Scan(
		&createdUser.UUID,
		&createdUser.Email,
		&createdUser.Password,
		&createdUser.FirstName,
		&createdUser.LastName,
		&createdUser.IsActive,
		&createdUser.IsSuperUser,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)

	if ok && pgErr != nil && pgErr.Code == errutils.DatabaseErrCodeUniqueViolation {
		return nil, errutils.FormatError(errutils.ErrDatabaseUniqueViolation, "dbConn.Scan failed")
	}

	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return createdUser, nil
}

// ActivateUserByUUID activates a user.
// If no user is affected, error is returned.
func (repo *repository) ActivateUserByUUID(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
) error {
	q := `
UPDATE
	"user"
SET
	is_active = TRUE,
	updated_at = $1
WHERE
	uuid = $2
	AND is_active = FALSE;
	`

	ct, err := dbConn.Exec(ctx, q, repo.timeProvider.Now(), userUUID)
	if err != nil {
		return errutils.FormatError(err, "dbConn.Exec failed")
	}

	if ct.RowsAffected() == 0 {
		return errutils.FormatError(errutils.ErrDatabaseNoRowsAffected)
	}

	return nil
}

// GetUserByEmail fetches a user by email.
// If no user is found, error is returned.
func (repo *repository) GetUserByEmail(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	email string,
) (*User, error) {
	user := &User{}

	q := `
SELECT
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at,
	updated_at
FROM
	"user"
WHERE
	email = $1
	AND is_active = TRUE;
	`

	err := dbConn.QueryRow(ctx, q, email).Scan(
		&user.UUID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.IsSuperUser,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsReturned, "dbConn.Scan failed")
	}

	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return user, nil
}

// GetUserByUUID fetches a user by UUID.
// If no user is found, error is returned.
func (repo *repository) GetUserByUUID(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
) (*User, error) {
	user := &User{}

	q := `
SELECT
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at,
	updated_at
FROM
	"user"
WHERE
	uuid = $1
	AND is_active = TRUE;
	`

	err := dbConn.QueryRow(ctx, q, userUUID).Scan(
		&user.UUID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.IsSuperUser,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsReturned, "dbConn.Scan failed")
	}

	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return user, nil
}

// UpdateUser updates a user.
func (repo *repository) UpdateUser(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
	firstName *string,
	lastName *string,
) (*User, error) {
	if firstName == nil && lastName == nil {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsAffected, "all attributes are nil")
	}

	updatedUser := &User{}

	q := `
UPDATE
	"user"
SET
	first_name = COALESCE($1, first_name),
	last_name = COALESCE($2, last_name),
	updated_at = $3
WHERE
	uuid = $4
	AND is_active = TRUE
RETURNING
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at,
	updated_at;
	`

	err := dbConn.QueryRow(
		ctx,
		q,
		firstName,
		lastName,
		repo.timeProvider.Now(),
		userUUID,
	).Scan(
		&updatedUser.UUID,
		&updatedUser.Email,
		&updatedUser.Password,
		&updatedUser.FirstName,
		&updatedUser.LastName,
		&updatedUser.IsActive,
		&updatedUser.IsSuperUser,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsAffected, "dbConn.Scan failed")
	}

	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return updatedUser, nil
}

// CreateAPIKey creates an API key.
func (repo *repository) CreateAPIKey(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	apiKey *APIKey,
) (*APIKey, error) {
	now := repo.timeProvider.Now()
	createdAPIKey := &APIKey{}

	q := `
INSERT INTO api_key (
	user_uuid,
	prefix,
	hashed_key,
	name,
	expires_at,
	created_at,
	updated_at
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7
)
RETURNING
	id,
	user_uuid,
	prefix,
	hashed_key,
	name,
	expires_at,
	created_at,
	updated_at;
	`
	err := dbConn.QueryRow(
		ctx,
		q,
		apiKey.UserUUID,
		apiKey.Prefix,
		apiKey.HashedKey,
		apiKey.Name,
		apiKey.ExpiresAt,
		now,
		now,
	).Scan(
		&createdAPIKey.ID,
		&createdAPIKey.UserUUID,
		&createdAPIKey.Prefix,
		&createdAPIKey.HashedKey,
		&createdAPIKey.Name,
		&createdAPIKey.ExpiresAt,
		&createdAPIKey.CreatedAt,
		&createdAPIKey.UpdatedAt,
	)

	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)

	if ok && pgErr != nil {
		switch pgErr.Code {
		case errutils.DatabaseErrCodeForeignKeyViolation:
			return nil, errutils.FormatError(errutils.ErrDatabaseForeignKeyConstraintViolation, "dbConn.Scan failed")
		case errutils.DatabaseErrCodeUniqueViolation:
			return nil, errutils.FormatError(errutils.ErrDatabaseUniqueViolation, "dbConn.Scan failed")
		}
	}

	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return createdAPIKey, nil
}

// ListAPIKeysByUserUUID fetches API keys under a given user UUID.
func (repo *repository) ListAPIKeysByUserUUID(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
) ([]*APIKey, error) {
	apiKeys := make([]*APIKey, 0)

	q := `
SELECT
	k.id,
	k.user_uuid,
	k.prefix,
	k.hashed_key,
	k.name,
	k.expires_at,
	k.created_at,
	k.updated_at
FROM
	api_key k
INNER JOIN
	"user" u
ON
	k.user_uuid = u.uuid
WHERE
	k.user_uuid = $1
	AND u.is_active = TRUE;
	`

	rows, err := dbConn.Query(ctx, q, userUUID)
	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Query failed")
	}
	defer rows.Close()

	for rows.Next() {
		apiKey := &APIKey{}
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserUUID,
			&apiKey.Prefix,
			&apiKey.HashedKey,
			&apiKey.Name,
			&apiKey.ExpiresAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, errutils.FormatError(err, "rows.Scan failed")
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// ListActiveAPIKeysByPrefix fetches API keys with a given prefix.
func (repo *repository) ListActiveAPIKeysByPrefix(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	prefix string,
) ([]*APIKey, error) {
	apiKeys := make([]*APIKey, 0)

	q := `
SELECT
	k.id,
	k.user_uuid,
	k.prefix,
	k.hashed_key,
	k.name,
	k.expires_at,
	k.created_at,
	k.updated_at
FROM
	api_key k
INNER JOIN
	"user" u
ON
	k.user_uuid = u.uuid
WHERE
	k.prefix = $1
	AND (k.expires_at IS NULL OR k.expires_at > CURRENT_TIMESTAMP)
	AND u.is_active = TRUE;
	`

	rows, err := dbConn.Query(ctx, q, prefix)
	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Query failed")
	}
	defer rows.Close()

	for rows.Next() {
		apiKey := &APIKey{}
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserUUID,
			&apiKey.Prefix,
			&apiKey.HashedKey,
			&apiKey.Name,
			&apiKey.ExpiresAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, errutils.FormatError(err, "rows.Scan failed")
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// UpdateAPIKey updates an API key.
// If no API key is affected, error is returned.
func (repo *repository) UpdateAPIKey(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
	apiKeyID int64,
	name *string,
	expiresAt jsonutils.Optional[time.Time],
) (*APIKey, error) {
	if name == nil && !expiresAt.Valid {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsAffected, "all attributes are nil")
	}

	updatedAPIKey := &APIKey{}

	q := `
UPDATE
	api_key k
SET
	name = COALESCE($1, name),
	expires_at = CASE WHEN $2 THEN $3 ELSE expires_at END,
	updated_at = $4
FROM
	"user" u
WHERE
	k.id = $5
	AND k.user_uuid = $6
	AND k.user_uuid = u.uuid
	AND u.is_active = TRUE
RETURNING
	k.id,
	k.user_uuid,
	k.prefix,
	k.hashed_key,
	k.name,
	k.expires_at,
	k.created_at,
	k.updated_at;
	`

	err := dbConn.QueryRow(
		ctx,
		q,
		name,
		expiresAt.Valid,
		expiresAt.Value,
		repo.timeProvider.Now(),
		apiKeyID,
		userUUID,
	).Scan(
		&updatedAPIKey.ID,
		&updatedAPIKey.UserUUID,
		&updatedAPIKey.Prefix,
		&updatedAPIKey.HashedKey,
		&updatedAPIKey.Name,
		&updatedAPIKey.ExpiresAt,
		&updatedAPIKey.CreatedAt,
		&updatedAPIKey.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsAffected, "dbConn.Scan failed")
	}

	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return updatedAPIKey, nil
}

// DeleteAPIKey deletes an API key.
// If no API key is found, error is returned.
func (repo *repository) DeleteAPIKey(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
	apiKeyID int64,
) error {
	q := `
DELETE FROM
	api_key k
USING
	"user" u
WHERE
	k.id = $1
	AND k.user_uuid = $2
	AND k.user_uuid = u.uuid
	AND u.is_active = TRUE;
	`

	ct, err := dbConn.Exec(ctx, q, apiKeyID, userUUID)
	if err != nil {
		return errutils.FormatError(err, "dbConn.Exec failed")
	}

	if ct.RowsAffected() == 0 {
		return errutils.FormatError(errutils.ErrDatabaseNoRowsAffected)
	}

	return nil
}
