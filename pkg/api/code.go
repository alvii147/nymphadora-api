package api

import (
	"time"

	"github.com/alvii147/nymphadora-api/pkg/validate"
)

// SupportedCodingLanguages is the list of supported coding languages.
var SupportedCodingLanguages = []string{
	PistonLanguageC,
	PistonLanguageCPlusPlus,
	PistonLanguageGo,
	PistonLanguageJava,
	PistonLanguageJavaScript,
	PistonLanguagePython,
	PistonLanguageRust,
	PistonLanguageTypeScript,
}

const (
	// CodeSpaceAccessLevelReadOnly represents read-only access on code spaces.
	CodeSpaceAccessLevelReadOnly = "R"
	// CodeSpaceAccessLevelReadOnly represents read-write access on code spaces.
	CodeSpaceAccessLevelReadWrite = "W"
)

// CreateCodeSpaceRequest represents the request body for code space creation requests.
type CreateCodeSpaceRequest struct {
	Language string `json:"language"`
}

// Validate validates fields in CreateCodeSpaceRequest.
func (r *CreateCodeSpaceRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringOptions("language", r.Language, SupportedCodingLanguages, false)

	return v.Passed(), v.Failures()
}

// CreateCodeSpaceResponse represents the response body for code space creation requests.
type CreateCodeSpaceResponse struct {
	ID          int64     `json:"id"`
	AuthorUUID  *string   `json:"author_uuid"`
	Name        string    `json:"name"`
	Language    string    `json:"language"`
	Contents    string    `json:"contents"`
	AccessLevel string    `json:"access_level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetCodeSpaceResponse represents the response body for a single code space in code space retrieval requests.
type GetCodeSpaceResponse struct {
	ID          int64     `json:"id"`
	AuthorUUID  *string   `json:"author_uuid"`
	Name        string    `json:"name"`
	Language    string    `json:"language"`
	Contents    string    `json:"contents"`
	AccessLevel string    `json:"access_level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListCodeSpacesResponse represents the response body for code space retrieval requests.
type ListCodeSpacesResponse struct {
	CodeSpaces []*GetCodeSpaceResponse `json:"code_spaces"`
}

// UpdateCodeSpaceRequest represents the request body for code space update requests.
type UpdateCodeSpaceRequest struct {
	Contents *string `json:"contents"`
}

// Validate validates fields in UpdateCodeSpaceRequest.
func (r *UpdateCodeSpaceRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()

	return v.Passed(), v.Failures()
}

// UpdateCodeSpaceResponse represents the response body for code space update requests.
type UpdateCodeSpaceResponse struct {
	ID          int64     `json:"id"`
	AuthorUUID  *string   `json:"author_uuid"`
	Name        string    `json:"name"`
	Language    string    `json:"language"`
	Contents    string    `json:"contents"`
	AccessLevel string    `json:"access_level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RunCodeSpaceResultsResponse represents code execution results for code space run requests.
type RunCodeSpaceResultsResponse struct {
	Stdout string  `json:"stdout"`
	Stderr string  `json:"stderr"`
	Code   *int    `json:"code"`
	Signal *string `json:"signal"`
}

// RunCodeSpaceResponse represents the response body for code space run requests.
type RunCodeSpaceResponse struct {
	Compile *RunCodeSpaceResultsResponse `json:"compile"`
	Run     RunCodeSpaceResultsResponse  `json:"run"`
}

// GetCodespaceUserResponse represents the response body for a single user's code space access
// for list code space users requests.
type GetCodespaceUserResponse struct {
	UserUUID    string `json:"user_uuid"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	CodeSpaceID int64  `json:"code_space_id"`
	AccessLevel string `json:"access_level"`
}

// ListCodespaceUsersResponse represents the response body for list code space users requests.
type ListCodespaceUsersResponse struct {
	Users []*GetCodespaceUserResponse `json:"users"`
}

// InviteCodeSpaceUserRequest represents the request body for code space user invitation requests.
type InviteCodeSpaceUserRequest struct {
	Email       string `json:"string"`
	AccessLevel string `json:"access_level"`
}

// Validate validates fields in InviteCodeSpaceUserRequest.
func (r *InviteCodeSpaceUserRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringEmail("email", r.Email)
	v.ValidateStringNotBlank("email", r.Email)
	v.ValidateStringOptions(
		"access_level",
		r.AccessLevel,
		[]string{CodeSpaceAccessLevelReadOnly, CodeSpaceAccessLevelReadWrite},
		false,
	)

	return v.Passed(), v.Failures()
}

// RemoveCodeSpaceUserRequest represents the request body for code space user removal requests.
type RemoveCodeSpaceUserRequest struct {
	UserUUID string `json:"user_uuid"`
}

// Validate validates fields in RemoveCodeSpaceUserRequest.
func (r *RemoveCodeSpaceUserRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringNotBlank("user_uuid", r.UserUUID)

	return v.Passed(), v.Failures()
}

// AcceptCodeSpaceUserInvitationRequest represents the request body for
// code space invitation acceptance requests.
type AcceptCodeSpaceUserInvitationRequest struct {
	Token string `json:"token"`
}

// Validate validates fields in AcceptCodeSpaceUserInvitationRequest.
func (r *AcceptCodeSpaceUserInvitationRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringNotBlank("token", r.Token)

	return v.Passed(), v.Failures()
}

// AcceptCodeSpaceUserInvitationResponse represents the response body for
// code space invitation acceptance requests.
type AcceptCodeSpaceUserInvitationResponse struct {
	ID          int64     `json:"id"`
	AuthorUUID  *string   `json:"author_uuid"`
	Name        string    `json:"name"`
	Language    string    `json:"language"`
	Contents    string    `json:"contents"`
	AccessLevel string    `json:"access_level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
