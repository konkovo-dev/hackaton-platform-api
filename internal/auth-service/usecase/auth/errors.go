package auth

import (
	pkgerrors "github.com/belikoooova/hackaton-platform-api/pkg/errors"
)

var (
	ErrUserNotFound      = pkgerrors.ErrUserNotFound
	ErrUserAlreadyExists = pkgerrors.ErrUserAlreadyExists
	ErrInvalidUsername   = pkgerrors.ErrInvalidUsername
	ErrInvalidEmail      = pkgerrors.ErrInvalidEmail
	ErrInvalidPassword   = pkgerrors.ErrInvalidPassword

	ErrInvalidCredentials = pkgerrors.ErrInvalidCredentials
	ErrTokenExpired       = pkgerrors.ErrTokenExpired
	ErrTokenInvalid       = pkgerrors.ErrTokenInvalid
	ErrTokenRevoked       = pkgerrors.ErrTokenRevoked

	ErrEmptyUsername  = pkgerrors.ErrEmptyUsername
	ErrEmptyEmail     = pkgerrors.ErrEmptyEmail
	ErrEmptyPassword  = pkgerrors.ErrEmptyPassword
	ErrEmptyFirstName = pkgerrors.ErrEmptyFirstName
	ErrEmptyLastName  = pkgerrors.ErrEmptyLastName
	ErrEmptyLogin     = pkgerrors.ErrEmptyLogin
	ErrEmptyTimezone  = pkgerrors.ErrEmptyTimezone
)
