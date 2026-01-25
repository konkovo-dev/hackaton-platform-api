package policy

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

func loadHackathonPolicyContext(
	ctx context.Context,
	hackathonID uuid.UUID,
	hackathonRepo HackathonRepository,
	parClient ParticipationAndRolesClient,
) (*HackathonPolicyContext, error) {
	pctx := NewHackathonPolicyContext()

	hackathon, err := hackathonRepo.GetByID(ctx, hackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to load hackathon: %w", err)
	}
	if hackathon == nil {
		return nil, fmt.Errorf("hackathon not found")
	}

	pctx.SetHackathonID(hackathon.ID)
	pctx.SetStage(hackathon.Stage)
	pctx.SetState(hackathon.State)
	pctx.SetPublishedAt(hackathon.PublishedAt)

	userID, ok := auth.GetUserID(ctx)
	if ok {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			pctx.SetAuthenticated(true)
			pctx.SetActorUserID(userUUID)

			_, participationStatus, teamID, roles, err := parClient.GetHackathonContext(ctx, hackathonID.String())
			if err == nil {
				pctx.SetRoles(roles)
				pctx.SetParticipationKind(participationStatus)
				pctx.SetTeamID(teamID)
			}
		}
	}

	return pctx, nil
}
