package errors

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUsername   = errors.New("invalid username")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
)

var (
	ErrEmptyUsername  = errors.New("username cannot be empty")
	ErrEmptyEmail     = errors.New("email cannot be empty")
	ErrEmptyPassword  = errors.New("password cannot be empty")
	ErrEmptyFirstName = errors.New("first_name cannot be empty")
	ErrEmptyLastName  = errors.New("last_name cannot be empty")
	ErrEmptyLogin     = errors.New("login (email or username) cannot be empty")
	ErrEmptyTimezone  = errors.New("timezone cannot be empty")
	ErrInvalidInput   = errors.New("invalid input")
)

var (
	ErrNotFound    = errors.New("not found")
	ErrConflict    = errors.New("conflict")
	ErrInternal    = errors.New("internal error")
	ErrUnavailable = errors.New("service unavailable")
)
