package users

import (
	"errors"

	pkgerrors "github.com/belikoooova/hackaton-platform-api/pkg/errors"
)

var (
	ErrUserNotFound = pkgerrors.ErrUserNotFound
	ErrInvalidInput = pkgerrors.ErrInvalidInput
	ErrTooManyUsers = errors.New("too many users requested")
)
