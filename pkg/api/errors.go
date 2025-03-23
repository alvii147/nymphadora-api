package api

// Error codes for API responses.
const (
	// ErrCodeInvalidRequest is the error code returned for invalid requests.
	// Used typically with status code 400.
	ErrCodeInvalidRequest = "invalid_request"
	// ErrCodeResourceExists is the error code returned when a resource already exists.
	// Used typically with status code 400.
	ErrCodeResourceExists = "resource_exists"
	// ErrCodeResourceNotFound is the error code returned when the resource is not found.
	// Used typically with status code 404.
	ErrCodeResourceNotFound = "resource_not_found"
	// ErrCodeMissingCredentials is the error code returned when authentication credentials are missing.
	// Used typically with status code 401.
	ErrCodeMissingCredentials = "missing_credentials"
	// ErrCodeInvalidCredentials is the error code returned when authentication credentials are invalid.
	// Used typically with status code 401.
	ErrCodeInvalidCredentials = "invalid_credentials"
	// ErrCodeAccessDenied is the error code returned when access is denied.
	// Used typically with status code 403.
	ErrCodeAccessDenied = "access_denied"
	// ErrCodeInternalServerError is the error code returned when an internal server error occurs.
	// Used typically with status code 500.
	ErrCodeInternalServerError = "internal_server_error"
)

// Error details for API responses.
const (
	// ErrDetailInvalidRequestData is the error detail returned when the request data is invalid.
	ErrDetailInvalidRequestData = "Invalid or malformed request data."
	// ErrDetailInternalServerError is the error detail returned when an internal server error occurs.
	ErrDetailInternalServerError = "Internal server error occurred."
	// ErrDetailUserExists is the error detail returned when a user already exists.
	ErrDetailUserExists = "User already exists"
	// ErrDetailUserNotFound is the error detail returned when the user is not found.
	ErrDetailUserNotFound = "User not found"
	// ErrDetailInvalidEmailOrPassword is the error detail returned when the given email or password is incorrect.
	ErrDetailInvalidEmailOrPassword = "Incorrect email or password."
	// ErrDetailInvalidToken is the error detail returned when the given token is invalid.
	ErrDetailInvalidToken = "Provided token is invalid"
	// ErrDetailMissingToken is the error detail returned when the token is missing.
	ErrDetailMissingToken = "No token was provided"
	// ErrDetailAPIKeyExists is the error detail returned when an API key already exists.
	ErrDetailAPIKeyExists = "API key already exists"
	// ErrDetailAPIKeyNotFound is the error detail returned when the API key is not found.
	ErrDetailAPIKeyNotFound = "API key not found"
	// ErrDetailCodeSpaceExists is the error detail returned when a code space already exists.
	ErrDetailCodeSpaceExists = "Code space already exists"
	// ErrDetailCodeSpaceNotFound is the error detail returned when the code space is not found.
	ErrDetailCodeSpaceNotFound = "Code space not found"
	// ErrDetailCodeSpaceAccessDenied is the error detail returned when access to a code space is denied.
	ErrDetailCodeSpaceAccessDenied = "Code space access denied"
)

// ErrorResponse represents the general error response body.
type ErrorResponse struct {
	Code               string              `json:"code"`
	Detail             string              `json:"detail"`
	ValidationFailures map[string][]string `json:"failures,omitempty"`
}
