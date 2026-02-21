package teammembersservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teammember"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) ListTeamMembers(ctx context.Context, req *teamv1.ListTeamMembersRequest) (*teamv1.ListTeamMembersResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "ListTeamMembers")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "ListTeamMembers")
	}

	out, err := a.teamMemberService.ListTeamMembers(ctx, teammember.ListTeamMembersIn{
		HackathonID: hackathonID,
		TeamID:      teamID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "ListTeamMembers")
	}

	protoMembers := make([]*teamv1.TeamMember, 0, len(out.Members))
	for _, member := range out.Members {
		protoMember := &teamv1.TeamMember{
			UserId:    member.UserID.String(),
			IsCaptain: member.IsCaptain,
			JoinedAt:  timestamppb.New(member.JoinedAt),
		}

		if member.AssignedVacancyID != nil {
			protoMember.AssignedVacancyId = member.AssignedVacancyID.String()
		}

		protoMembers = append(protoMembers, protoMember)
	}

	return &teamv1.ListTeamMembersResponse{
		Members: protoMembers,
	}, nil
}
