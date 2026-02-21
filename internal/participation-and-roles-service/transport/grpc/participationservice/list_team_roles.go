package participationservice

import (
	"context"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
)

func (a *API) ListTeamRoles(ctx context.Context, req *participationrolesv1.ListTeamRolesRequest) (*participationrolesv1.ListTeamRolesResponse, error) {
	roles, err := a.teamRoleRepo.ListAll(ctx)
	if err != nil {
		return nil, a.handleError(ctx, err, "list_team_roles")
	}

	protoRoles := make([]*participationrolesv1.TeamRole, 0, len(roles))
	for _, role := range roles {
		protoRoles = append(protoRoles, &participationrolesv1.TeamRole{
			Id:   role.ID.String(),
			Name: role.Name,
		})
	}

	return &participationrolesv1.ListTeamRolesResponse{
		TeamRoles: protoRoles,
	}, nil
}
