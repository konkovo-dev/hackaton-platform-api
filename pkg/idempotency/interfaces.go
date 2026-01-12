package idempotency

import (
	"context"
	"time"
)

type Repository interface {
	Get(ctx context.Context, key, scope string) (*StoredResponse, error)
	Set(ctx context.Context, key, scope, requestHash string, responseBlob []byte, expiresAt time.Time) error
}

type StoredResponse struct {
	Key          string
	Scope        string
	RequestHash  string
	ResponseBlob []byte
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

var (
	ErrNotFound        = &NotFoundError{}
	ErrConflict        = &ConflictError{}
	ErrResponseInvalid = &ResponseInvalidError{}
)

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "idempotency key not found"
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "idempotency key conflict: request hash mismatch"
}

type ResponseInvalidError struct {
	Err error
}

func (e *ResponseInvalidError) Error() string {
	return "failed to unmarshal stored response: " + e.Err.Error()
}

func (e *ResponseInvalidError) Unwrap() error {
	return e.Err
}
