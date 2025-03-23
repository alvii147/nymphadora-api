package auth

import (
	"context"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
)

// User represents the database table "user".
type User struct {
	UUID        string    `db:"uuid"`
	Email       string    `db:"email"`
	Password    string    `db:"password"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	IsActive    bool      `db:"is_active"`
	IsSuperUser bool      `db:"is_superuser"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// APIKey represents the database table "api_key".
type APIKey struct {
	ID        int64      `db:"id"`
	UserUUID  string     `db:"user_uuid"`
	Prefix    string     `db:"prefix"`
	HashedKey string     `db:"hashed_key"`
	Name      string     `db:"name"`
	ExpiresAt *time.Time `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

// AuthContextKey is a string representing auth-related context keys.
type AuthContextKey string

// AuthContextKeyUserUUID is the key in context where user UUID is stored after authentication.
const AuthContextKeyUserUUID AuthContextKey = "userUUID"

// GetUserUUIDFromContext extracts the user UUID from a given context.
func GetUserUUIDFromContext(ctx context.Context) (string, error) {
	userUUID, ok := ctx.Value(AuthContextKeyUserUUID).(string)
	if !ok {
		return "", errutils.FormatError(nil, "user UUID missing in context")
	}

	return userUUID, nil
}
