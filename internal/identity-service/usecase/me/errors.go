package me

import (
	pkgerrors "github.com/belikoooova/hackaton-platform-api/pkg/errors"
)

var (
	ErrUserNotFound      = pkgerrors.ErrUserNotFound
	ErrUserAlreadyExists = pkgerrors.ErrUserAlreadyExists
	ErrInvalidInput      = pkgerrors.ErrInvalidInput
	ErrUnauthorized      = pkgerrors.ErrUnauthorized
	ErrForbidden         = pkgerrors.ErrForbidden
)
