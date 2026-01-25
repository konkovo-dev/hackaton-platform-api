package participationandrolesservice

import (
	"context"
	"log/slog"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/role"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) ListHackathonStaff(ctx context.Context, req *participationrolesv1.ListHackathonStaffRequest) (*participationrolesv1.ListHackathonStaffResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	result, err := a.roleService.ListStaff(ctx, role.ListStaffIn{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "list_hackathon_staff")
	}

	staff := make([]*participationrolesv1.HackathonStaffMember, 0, len(result.Staff))
	for _, member := range result.Staff {
		protoRoles := make([]participationrolesv1.HackathonRole, 0, len(member.Roles))
		for _, roleStr := range member.Roles {
			protoRole := domain.MapDomainRoleToProto(domain.HackathonRole(roleStr))
			if protoRole != participationrolesv1.HackathonRole_HACKATHON_ROLE_UNSPECIFIED {
				protoRoles = append(protoRoles, protoRole)
			}
		}

		staff = append(staff, &participationrolesv1.HackathonStaffMember{
			UserId: member.UserID.String(),
			Roles:  protoRoles,
		})
	}

	a.logger.InfoContext(ctx, "list_hackathon_staff: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.Int("staff_count", len(staff)),
	)

	return &participationrolesv1.ListHackathonStaffResponse{
		Staff: staff,
		Page:  &commonv1.PageResponse{},
	}, nil
}
