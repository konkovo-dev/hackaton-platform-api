package auth

import (
	"context"
	"time"

	authpolicy "github.com/belikoooova/hackaton-platform-api/internal/auth-service/policy"
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
	introspectPolicy := authpolicy.NewIntrospectPolicy()
	decision := introspectPolicy.ValidateInput(authpolicy.IntrospectParams{
		AccessToken: in.AccessToken,
	})

	if !decision.Allowed {
		return nil, s.mapPolicyDecisionToError(decision)
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
