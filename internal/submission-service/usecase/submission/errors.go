package submission

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
)

var (
	ErrTooManySubmissions = errors.New("too many submissions: limit reached")
	ErrTooManyFiles       = errors.New("too many files: limit reached")
	ErrFileTooLarge       = errors.New("file too large")
	ErrTotalSizeTooLarge  = errors.New("total submission size too large")
	ErrInvalidFileType    = errors.New("invalid file type")
	ErrFileNotUploaded    = errors.New("file not uploaded to storage")
)
