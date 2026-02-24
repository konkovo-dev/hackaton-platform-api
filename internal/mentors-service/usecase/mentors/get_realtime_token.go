package mentors

import (
	"context"
	"fmt"

	mentorspolicy "github.com/belikoooova/hackaton-platform-api/internal/mentors-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetRealtimeTokenIn struct {
	HackathonID string
}

type GetRealtimeTokenOut struct {
	Token     string
	ExpiresAt int64
}

func (s *Service) GetRealtimeToken(ctx context.Context, in GetRealtimeTokenIn) (*GetRealtimeTokenOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	hackathonID, err := uuid.Parse(in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid hackathon_id", ErrInvalidInput)
	}

	getRealtimeTokenPolicy := mentorspolicy.NewGetRealtimeTokenPolicy()
	pctx, err := getRealtimeTokenPolicy.LoadContext(ctx, mentorspolicy.GetRealtimeTokenParams{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load policy context: %w", err)
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)

	stage, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}
	pctx.SetHackathonStage(stage)

	roles, _, err := s.prClient.GetParticipationAndRoles(ctx, userID, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation and roles: %w", err)
	}
	pctx.SetActorRoles(roles)

	participates := len(roles) > 0
	pctx.SetParticipates(participates)

	decision := getRealtimeTokenPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	token, expiresAt, err := s.jwtHelper.GenerateConnectionToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate connection token: %w", err)
	}

	return &GetRealtimeTokenOut{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
