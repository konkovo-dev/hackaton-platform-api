package pgxutil

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	pkgerrors "github.com/belikoooova/hackaton-platform-api/pkg/errors"
)

func MapDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return pkgerrors.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return pkgerrors.ErrConflict
		case "23503":
			return pkgerrors.ErrNotFound
		case "23502":
			return pkgerrors.ErrInvalidInput
		}
	}

	if strings.Contains(err.Error(), "duplicate key") {
		return pkgerrors.ErrConflict
	}

	if strings.Contains(err.Error(), "violates foreign key constraint") {
		return pkgerrors.ErrNotFound
	}

	if strings.Contains(err.Error(), "violates not-null constraint") {
		return pkgerrors.ErrInvalidInput
	}

	return err
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, pkgerrors.ErrNotFound)
}

func IsConflict(err error) bool {
	return errors.Is(err, pkgerrors.ErrConflict)
}

func IsInvalidInput(err error) bool {
	return errors.Is(err, pkgerrors.ErrInvalidInput)
}
