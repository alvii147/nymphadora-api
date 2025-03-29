package errutils

import "errors"

// General shared errors.
var (
	ErrInvalidToken                 = errors.New("invalid token")
	ErrInvalidCredentials           = errors.New("invalid credentials")
	ErrUserAlreadyExists            = errors.New("user already exists")
	ErrUserNotFound                 = errors.New("user not found")
	ErrAPIKeyAlreadyExists          = errors.New("api key already exists")
	ErrAPIKeyNotFound               = errors.New("api key not found")
	ErrCodeSpaceAlreadyExists       = errors.New("code space already exists")
	ErrCodeSpaceNotFound            = errors.New("code space not found")
	ErrCodeSpaceAccessNotFound      = errors.New("code space access not found")
	ErrCodeSpaceAccessDenied        = errors.New("code space access denied")
	ErrCodeSpaceUnsupportedLanguage = errors.New("code space language not supported")
)
