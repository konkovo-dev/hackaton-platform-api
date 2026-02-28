package mentors

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	mentorspolicy "github.com/belikoooova/hackaton-platform-api/internal/mentors-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetMyChatMessagesIn struct {
	HackathonID string
	Limit       int32
	Offset      int32
}

type GetMyChatMessagesOut struct {
	Messages []*entity.Message
	HasMore  bool
}

func (s *Service) GetMyChatMessages(ctx context.Context, in GetMyChatMessagesIn) (*GetMyChatMessagesOut, error) {
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
		return nil, fmt.Errorf("invalid hackathon_id: %w", err)
	}

	getMyChatMessagesPolicy := mentorspolicy.NewGetMyChatMessagesPolicy()
	pctx, err := getMyChatMessagesPolicy.LoadContext(ctx, mentorspolicy.GetMyChatMessagesParams{
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

	decision := getMyChatMessagesPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	_, teamIDPtr, err := s.prClient.GetParticipationAndRoles(ctx, userID, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation and roles: %w", err)
	}

	var ownerKind string
	var ownerID uuid.UUID

	if teamIDPtr != nil && *teamIDPtr != "" {
		teamID, err := uuid.Parse(*teamIDPtr)
		if err == nil {
			ownerKind = domain.OwnerKindTeam
			ownerID = teamID
		} else {
			ownerKind = domain.OwnerKindUser
			ownerID = userUUID
		}
	} else {
		ownerKind = domain.OwnerKindUser
		ownerID = userUUID
	}

	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := s.messageRepo.ListByOwner(ctx, hackathonID, ownerKind, ownerID, limit+1, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	hasMore := false
	if len(messages) > int(limit) {
		hasMore = true
		messages = messages[:limit]
	}

	return &GetMyChatMessagesOut{
		Messages: messages,
		HasMore:  hasMore,
	}, nil
}
