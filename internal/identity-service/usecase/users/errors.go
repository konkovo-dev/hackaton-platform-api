package users

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrTooManyUsers = errors.New("too many users requested")
)
