package users

import (
	pkgerrors "github.com/belikoooova/hackaton-platform-api/pkg/errors"
)

var (
	ErrUserNotFound = pkgerrors.ErrUserNotFound
	ErrInvalidInput = pkgerrors.ErrInvalidInput
	ErrTooManyUsers = pkgerrors.ErrInvalidInput
)
