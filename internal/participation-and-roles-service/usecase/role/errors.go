package role

import "errors"

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidInput      = errors.New("invalid input")
	ErrHackathonNotFound = errors.New("hackathon not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
)
