package submission

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain/entity"
	submissionpolicy "github.com/belikoooova/hackaton-platform-api/internal/submission-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type ListSubmissionsIn struct {
	HackathonID uuid.UUID
	OwnerKind   string
	OwnerID     uuid.UUID
	Limit       int32
	Offset      int32
}

type ListSubmissionsOut struct {
	Submissions []*entity.Submission
	HasMore     bool
}

func (s *Service) ListSubmissions(ctx context.Context, in ListSubmissionsIn) (*ListSubmissionsOut, error) {
	// Service-to-service calls bypass policy checks
	if auth.IsServiceCall(ctx) {
		limit := in.Limit
		if limit <= 0 {
			limit = 50
		}
		if limit > 100 {
			limit = 100
		}

		// For service calls, list all submissions in hackathon (no owner filter)
		submissions, err := s.submissionRepo.ListByHackathon(ctx, in.HackathonID, limit+1, in.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list submissions: %w", err)
		}

		hasMore := false
		if len(submissions) > int(limit) {
			hasMore = true
			submissions = submissions[:limit]
		}

		return &ListSubmissionsOut{
			Submissions: submissions,
			HasMore:     hasMore,
		}, nil
	}

	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	actorUserID, _, roles, teamID, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	actorUUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actor user id: %w", err)
	}

	// Determine actor's owner kind and ID
	actorOwnerKind := domain.OwnerKindUser
	actorOwnerID := actorUUID
	if teamID != "" {
		actorOwnerKind = domain.OwnerKindTeam
		teamUUID, err := uuid.Parse(teamID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse team id: %w", err)
		}
		actorOwnerID = teamUUID
	}

	targetOwnerKind := in.OwnerKind
	targetOwnerID := in.OwnerID

	// If no target specified, default to actor's owner
	if targetOwnerKind == "" || targetOwnerID == uuid.Nil {
		targetOwnerKind = actorOwnerKind
		targetOwnerID = actorOwnerID
	}

	listPolicy := submissionpolicy.NewListSubmissionsPolicy()
	pctx, err := listPolicy.LoadContext(ctx, submissionpolicy.ListSubmissionsParams{
		HackathonID:     in.HackathonID,
		TargetOwnerKind: targetOwnerKind,
		TargetOwnerID:   targetOwnerID,
		ActorOwnerKind:  actorOwnerKind,
		ActorOwnerID:    actorOwnerID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetActorRoles(roles)

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	submissions, err := s.submissionRepo.ListByOwner(ctx, in.HackathonID, targetOwnerKind, targetOwnerID, limit+1, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}

	hasMore := false
	if len(submissions) > int(limit) {
		hasMore = true
		submissions = submissions[:limit]
	}

	return &ListSubmissionsOut{
		Submissions: submissions,
		HasMore:     hasMore,
	}, nil
}
