package code

import (
	"context"
	"errors"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is used to access, modify, and delete code space data.
//
//go:generate mockgen -package=codemocks -source=$GOFILE -destination=./mocks/repository.go
type Repository interface {
	CreateCodeSpace(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		codeSpace *CodeSpace,
	) (*CodeSpace, error)
	ListCodeSpaces(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
	) ([]*CodeSpace, []*CodeSpaceAccess, error)
	GetCodeSpace(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		codeSpaceID int64,
	) (*CodeSpace, error)
	GetCodeSpaceWithAccessByName(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
		name string,
	) (*CodeSpace, *CodeSpaceAccess, error)
	UpdateCodeSpace(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		codeSpaceID int64,
		contents *string,
	) (*CodeSpace, error)
	DeleteCodeSpace(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		codeSpaceID int64,
	) error
	CreateOrUpdateCodeSpaceAccess(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		codeSpaceAccess *CodeSpaceAccess,
	) (*CodeSpaceAccess, error)
	ListUsersWithCodeSpaceAccess(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		codeSpaceID int64,
	) ([]*auth.User, []*CodeSpaceAccess, error)
	DeleteCodeSpaceAccess(
		ctx context.Context,
		dbConn *pgxpool.Conn,
		userUUID string,
		codeSpaceID int64,
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

// CreateCodeSpace creates a new code space.
func (repo *repository) CreateCodeSpace(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	codeSpace *CodeSpace,
) (*CodeSpace, error) {
	now := repo.timeProvider.Now()
	createdCodeSpace := &CodeSpace{}

	q := `
INSERT INTO code_space (
	author_uuid,
	name,
	language,
	contents,
	created_at,
	updated_at
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6
)
RETURNING
	id,
	author_uuid,
	name,
	language,
	contents,
	created_at,
	updated_at;
	`

	err := dbConn.QueryRow(
		ctx,
		q,
		codeSpace.AuthorUUID,
		codeSpace.Name,
		codeSpace.Language,
		codeSpace.Contents,
		now,
		now,
	).Scan(
		&createdCodeSpace.ID,
		&createdCodeSpace.AuthorUUID,
		&createdCodeSpace.Name,
		&createdCodeSpace.Language,
		&createdCodeSpace.Contents,
		&createdCodeSpace.CreatedAt,
		&createdCodeSpace.UpdatedAt,
	)
	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return createdCodeSpace, nil
}

// ListCodeSpaces lists code spaces accessible by a given user.
func (repo *repository) ListCodeSpaces(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
) ([]*CodeSpace, []*CodeSpaceAccess, error) {
	codeSpaces := make([]*CodeSpace, 0)
	codeSpaceAccesses := make([]*CodeSpaceAccess, 0)

	q := `
SELECT
	c.id,
	c.author_uuid,
	c.name,
	c.language,
	c.contents,
	c.created_at,
	c.updated_at,
	a.id,
	a.user_uuid,
	a.code_space_id,
	a.level,
	a.created_at,
	a.updated_at
FROM
	code_space c
INNER JOIN
	code_space_access a
ON
	c.id = a.code_space_id
WHERE
	a.user_uuid = $1
	AND a.level >= $2;
	`

	rows, err := dbConn.Query(ctx, q, userUUID, CodeSpaceAccessLevelReadOnly)
	if err != nil {
		return nil, nil, errutils.FormatError(err, "dbConn.Query failed")
	}
	defer rows.Close()

	for rows.Next() {
		codeSpace := &CodeSpace{}
		codeSpaceAccess := &CodeSpaceAccess{}

		err := rows.Scan(
			&codeSpace.ID,
			&codeSpace.AuthorUUID,
			&codeSpace.Name,
			&codeSpace.Language,
			&codeSpace.Contents,
			&codeSpace.CreatedAt,
			&codeSpace.UpdatedAt,
			&codeSpaceAccess.ID,
			&codeSpaceAccess.UserUUID,
			&codeSpaceAccess.CodeSpaceID,
			&codeSpaceAccess.Level,
			&codeSpaceAccess.CreatedAt,
			&codeSpaceAccess.UpdatedAt,
		)
		if err != nil {
			return nil, nil, errutils.FormatError(err, "rows.Scan failed")
		}

		codeSpaces = append(codeSpaces, codeSpace)
		codeSpaceAccesses = append(codeSpaceAccesses, codeSpaceAccess)
	}

	return codeSpaces, codeSpaceAccesses, nil
}

// GetCodeSpace gets a given code space.
func (repo *repository) GetCodeSpace(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	codeSpaceID int64,
) (*CodeSpace, error) {
	codeSpace := &CodeSpace{}

	q := `
SELECT
	c.id,
	c.author_uuid,
	c.name,
	c.language,
	c.contents,
	c.created_at,
	c.updated_at
FROM
	code_space c
WHERE
	c.id = $1;
	`

	err := dbConn.QueryRow(ctx, q, codeSpaceID).Scan(
		&codeSpace.ID,
		&codeSpace.AuthorUUID,
		&codeSpace.Name,
		&codeSpace.Language,
		&codeSpace.Contents,
		&codeSpace.CreatedAt,
		&codeSpace.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsReturned, "dbConn.Scan failed")
	}

	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return codeSpace, nil
}

// GetCodeSpaceWithAccessByName gets a given code space and its corresponding code space access for a given user.
func (repo *repository) GetCodeSpaceWithAccessByName(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
	name string,
) (*CodeSpace, *CodeSpaceAccess, error) {
	codeSpace := &CodeSpace{}
	codeSpaceAccess := &CodeSpaceAccess{}

	q := `
SELECT
	c.id,
	c.author_uuid,
	c.name,
	c.language,
	c.contents,
	c.created_at,
	c.updated_at,
	a.id,
	a.user_uuid,
	a.code_space_id,
	a.level,
	a.created_at,
	a.updated_at
FROM
	code_space c
INNER JOIN
	code_space_access a
ON
	c.id = a.code_space_id
WHERE
	c.name = $1
	AND a.user_uuid = $2
	AND a.level >= $3;
	`

	err := dbConn.QueryRow(ctx, q, name, userUUID, CodeSpaceAccessLevelReadOnly).Scan(
		&codeSpace.ID,
		&codeSpace.AuthorUUID,
		&codeSpace.Name,
		&codeSpace.Language,
		&codeSpace.Contents,
		&codeSpace.CreatedAt,
		&codeSpace.UpdatedAt,
		&codeSpaceAccess.ID,
		&codeSpaceAccess.UserUUID,
		&codeSpaceAccess.CodeSpaceID,
		&codeSpaceAccess.Level,
		&codeSpaceAccess.CreatedAt,
		&codeSpaceAccess.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, errutils.FormatError(errutils.ErrDatabaseNoRowsReturned, "dbConn.Scan failed")
	}

	if err != nil {
		return nil, nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return codeSpace, codeSpaceAccess, nil
}

// UpdateCodeSpace updates a code space.
func (repo *repository) UpdateCodeSpace(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	codeSpaceID int64,
	contents *string,
) (*CodeSpace, error) {
	if contents == nil {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsAffected, "all attributes are nil")
	}

	updatedCodeSpace := &CodeSpace{}

	q := `
UPDATE
	code_space
SET
	contents = COALESCE($1, contents),
	updated_at = $2
WHERE
	id = $3
RETURNING
	id,
	author_uuid,
	name,
	language,
	contents,
	created_at,
	updated_at;
	`

	err := dbConn.QueryRow(
		ctx,
		q,
		contents,
		repo.timeProvider.Now(),
		codeSpaceID,
	).Scan(
		&updatedCodeSpace.ID,
		&updatedCodeSpace.AuthorUUID,
		&updatedCodeSpace.Name,
		&updatedCodeSpace.Language,
		&updatedCodeSpace.Contents,
		&updatedCodeSpace.CreatedAt,
		&updatedCodeSpace.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errutils.FormatError(errutils.ErrDatabaseNoRowsAffected, "dbConn.Scan failed")
	}

	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return updatedCodeSpace, nil
}

// DeleteCodeSpace deletes a code space.
// If no code space is found, error is returned.
func (repo *repository) DeleteCodeSpace(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	codeSpaceID int64,
) error {
	q := `
DELETE FROM
	code_space c
WHERE
	c.id = $1;
	`

	ct, err := dbConn.Exec(ctx, q, codeSpaceID)
	if err != nil {
		return errutils.FormatError(err, "dbConn.Exec failed")
	}

	if ct.RowsAffected() == 0 {
		return errutils.FormatError(errutils.ErrDatabaseNoRowsAffected)
	}

	return nil
}

// CreateOrUpdateCodeSpaceAccess creates a new code space access or updates the existing one.
func (repo *repository) CreateOrUpdateCodeSpaceAccess(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	codeSpaceAccess *CodeSpaceAccess,
) (*CodeSpaceAccess, error) {
	now := repo.timeProvider.Now()
	createdCodeSpaceAccess := &CodeSpaceAccess{}

	q := `
INSERT INTO code_space_access (
	user_uuid,
	code_space_id,
	level,
	created_at,
	updated_at
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
ON CONFLICT (user_uuid, code_space_id)
DO UPDATE SET
	level = EXCLUDED.level,
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	user_uuid,
	code_space_id,
	level,
	created_at,
	updated_at;
	`

	err := dbConn.QueryRow(
		ctx,
		q,
		codeSpaceAccess.UserUUID,
		codeSpaceAccess.CodeSpaceID,
		codeSpaceAccess.Level,
		now,
		now,
	).Scan(
		&createdCodeSpaceAccess.ID,
		&createdCodeSpaceAccess.UserUUID,
		&createdCodeSpaceAccess.CodeSpaceID,
		&createdCodeSpaceAccess.Level,
		&createdCodeSpaceAccess.CreatedAt,
		&createdCodeSpaceAccess.UpdatedAt,
	)
	if err != nil {
		return nil, errutils.FormatError(err, "dbConn.Scan failed")
	}

	return createdCodeSpaceAccess, nil
}

// ListUsersWithCodeSpaceAccess lists all users with access to a given code space.
func (repo *repository) ListUsersWithCodeSpaceAccess(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	codeSpaceID int64,
) ([]*auth.User, []*CodeSpaceAccess, error) {
	users := make([]*auth.User, 0)
	codeSpacesAccesses := make([]*CodeSpaceAccess, 0)

	q := `
SELECT
	u.uuid,
	u.email,
	u.password,
	u.first_name,
	u.last_name,
	u.is_active,
	u.is_superuser,
	u.created_at,
	u.updated_at,
	a.id,
	a.user_uuid,
	a.code_space_id,
	a.level,
	a.created_at,
	a.updated_at
FROM
	"user" u
INNER JOIN
	code_space_access a
ON
	u.uuid = a.user_uuid
WHERE
	a.code_space_id = $1
	AND a.level >= $2;
	`

	rows, err := dbConn.Query(
		ctx,
		q,
		codeSpaceID,
		CodeSpaceAccessLevelReadOnly,
	)
	if err != nil {
		return nil, nil, errutils.FormatError(err, "dbConn.Query failed")
	}
	defer rows.Close()

	for rows.Next() {
		user := &auth.User{}
		codeSpaceAccess := &CodeSpaceAccess{}

		err := rows.Scan(
			&user.UUID,
			&user.Email,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.IsActive,
			&user.IsSuperUser,
			&user.CreatedAt,
			&user.UpdatedAt,
			&codeSpaceAccess.ID,
			&codeSpaceAccess.UserUUID,
			&codeSpaceAccess.CodeSpaceID,
			&codeSpaceAccess.Level,
			&codeSpaceAccess.CreatedAt,
			&codeSpaceAccess.UpdatedAt,
		)
		if err != nil {
			return nil, nil, errutils.FormatError(err, "rows.Scan failed")
		}

		users = append(users, user)
		codeSpacesAccesses = append(codeSpacesAccesses, codeSpaceAccess)
	}

	return users, codeSpacesAccesses, nil
}

// DeleteCodeSpaceAccess deletes a code space access.
// If no code space access is found, error is returned.
func (repo *repository) DeleteCodeSpaceAccess(
	ctx context.Context,
	dbConn *pgxpool.Conn,
	userUUID string,
	codeSpaceID int64,
) error {
	q := `
DELETE FROM
	code_space_access a
WHERE
	a.code_space_id = $1
	AND a.user_uuid = $2;
	`

	ct, err := dbConn.Exec(ctx, q, codeSpaceID, userUUID)
	if err != nil {
		return errutils.FormatError(err, "dbConn.Exec failed")
	}

	if ct.RowsAffected() == 0 {
		return errutils.FormatError(errutils.ErrDatabaseNoRowsAffected)
	}

	return nil
}
