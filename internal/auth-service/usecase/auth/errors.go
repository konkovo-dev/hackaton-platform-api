package auth

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUsername   = errors.New("invalid username")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
	ErrTokenRevoked       = errors.New("token revoked")

	ErrEmptyUsername = errors.New("username cannot be empty")
	ErrEmptyEmail    = errors.New("email cannot be empty")
	ErrEmptyPassword = errors.New("password cannot be empty")
	ErrEmptyLogin    = errors.New("login (email or username) cannot be empty")
	ErrEmptyTimezone = errors.New("timezone cannot be empty")
)
