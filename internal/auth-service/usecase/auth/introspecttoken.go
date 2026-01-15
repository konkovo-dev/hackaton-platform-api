package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type IntrospectTokenIn struct {
	AccessToken string
}

type IntrospectTokenOut struct {
	Active    bool
	UserID    uuid.UUID
	ExpiresAt time.Time
}

func (s *Service) IntrospectToken(ctx context.Context, in IntrospectTokenIn) (*IntrospectTokenOut, error) {
	if err := s.validateIntrospectTokenIn(in); err != nil {
		return nil, err
	}

	userID, expiresAt, err := s.jwtService.Verify(in.AccessToken)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	if time.Now().UTC().After(expiresAt) {
		return nil, ErrTokenExpired
	}

	return &IntrospectTokenOut{
		Active:    true,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) validateIntrospectTokenIn(in IntrospectTokenIn) error {
	if in.AccessToken == "" {
		return ErrTokenInvalid
	}

	return nil
}
