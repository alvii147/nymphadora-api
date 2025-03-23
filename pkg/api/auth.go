package api

import (
	"time"

	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/validate"
)

// CreateUserRequest represents the request body for user creation requests.
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Validate validates fields in CreateUserRequest.
func (r *CreateUserRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringEmail("email", r.Email)
	v.ValidateStringNotBlank("email", r.Email)
	v.ValidateStringNotBlank("password", r.Password)
	v.ValidateStringNotBlank("first_name", r.FirstName)
	v.ValidateStringNotBlank("last_name", r.LastName)

	return v.Passed(), v.Failures()
}

// CreateUserResponse represents the response body for user creation requests.
type CreateUserResponse struct {
	UUID      string    `json:"uuid"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ActivateUserRequest represents the request body for user activation requests.
type ActivateUserRequest struct {
	Token string `json:"token"`
}

// Validate validates fields in ActivateUserRequest.
func (r *ActivateUserRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringNotBlank("token", r.Token)

	return v.Passed(), v.Failures()
}

// GetUserMeResponse represents the response body for get current user requests.
type GetUserMeResponse struct {
	UUID      string    `json:"uuid"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateUserRequest represents the request body for user update requests.
type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

// Validate validates fields in UpdateUserRequest.
func (r *UpdateUserRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	if r.FirstName != nil {
		v.ValidateStringNotBlank("first_name", *r.FirstName)
	}
	if r.LastName != nil {
		v.ValidateStringNotBlank("last_name", *r.LastName)
	}

	return v.Passed(), v.Failures()
}

// UpdateUserResponse represents the response body for user update requests.
type UpdateUserResponse struct {
	UUID      string    `json:"uuid"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateTokenRequest represents the request body for create token requests.
type CreateTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate validates fields in CreateTokenRequest.
func (r *CreateTokenRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringEmail("email", r.Email)
	v.ValidateStringNotBlank("email", r.Email)
	v.ValidateStringNotBlank("password", r.Password)

	return v.Passed(), v.Failures()
}

// CreateTokenResponse represents the response body for create token requests.
type CreateTokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// RefreshTokenRequest represents the request body for refresh token requests.
type RefreshTokenRequest struct {
	Refresh string `json:"refresh"`
}

// Validate validates fields in RefreshTokenRequest.
func (r *RefreshTokenRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringNotBlank("refresh", r.Refresh)

	return v.Passed(), v.Failures()
}

// RefreshTokenResponse represents the response body for refresh token requests.
type RefreshTokenResponse struct {
	Access string `json:"access"`
}

// ValidateTokenRequest represents the request body for refresh token requests.
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// Validate validates fields in ValidateTokenRequest.
func (r *ValidateTokenRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringNotBlank("token", r.Token)

	return v.Passed(), v.Failures()
}

// ValidateTokenResponse represents the response body for refresh token requests.
type ValidateTokenResponse struct {
	Valid bool `json:"valid"`
}

// CreateAPIKeyRequest represents the request body for API key creation requests.
type CreateAPIKeyRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Validate validates fields in CreateAPIKeyRequest.
func (r *CreateAPIKeyRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringNotBlank("name", r.Name)

	return v.Passed(), v.Failures()
}

// CreateAPIKeyResponse represents the response body for API key creation requests.
type CreateAPIKeyResponse struct {
	ID        int64      `json:"id"`
	RawKey    string     `json:"raw_key"`
	UserUUID  string     `json:"user_uuid"`
	Prefix    string     `json:"prefix"`
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// GetAPIKeyResponse represents the response body for a single API key in API key retrieval requests.
type GetAPIKeyResponse struct {
	ID        int64      `json:"id"`
	UserUUID  string     `json:"user_uuid"`
	Prefix    string     `json:"prefix"`
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ListAPIKeysResponse represents the response body for API key retrieval requests.
type ListAPIKeysResponse struct {
	Keys []*GetAPIKeyResponse `json:"keys"`
}

// UpdateAPIKeyRequest represents the request body for API key update requests.
type UpdateAPIKeyRequest struct {
	Name      *string                       `json:"name"`
	ExpiresAt jsonutils.Optional[time.Time] `json:"expires_at"`
}

// Validate validates fields in UpdateAPIKeyRequest.
func (r *UpdateAPIKeyRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	if r.Name != nil {
		v.ValidateStringNotBlank("name", *r.Name)
	}

	return v.Passed(), v.Failures()
}

// UpdateAPIKeyResponse represents the response body for API key update requests.
type UpdateAPIKeyResponse struct {
	ID        int64      `json:"id"`
	UserUUID  string     `json:"user_uuid"`
	Prefix    string     `json:"prefix"`
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
