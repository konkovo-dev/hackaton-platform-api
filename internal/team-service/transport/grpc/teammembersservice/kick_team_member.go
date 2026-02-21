package teammembersservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teammember"
	"github.com/google/uuid"
)

func (a *API) KickTeamMember(ctx context.Context, req *teamv1.KickTeamMemberRequest) (*teamv1.KickTeamMemberResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "KickTeamMember")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "KickTeamMember")
	}

	targetUserID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, a.handleError(ctx, err, "KickTeamMember")
	}

	_, err = a.teamMemberService.KickTeamMember(ctx, teammember.KickTeamMemberIn{
		HackathonID:  hackathonID,
		TeamID:       teamID,
		TargetUserID: targetUserID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "KickTeamMember")
	}

	return &teamv1.KickTeamMemberResponse{}, nil
}
