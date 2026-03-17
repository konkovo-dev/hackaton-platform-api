package role

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/google/uuid"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 100
)

type ListHackathonInvitationsIn struct {
	HackathonID uuid.UUID
	PageSize    uint32
	PageToken   string
}

type ListHackathonInvitationsOut struct {
	Invitations   []*entity.StaffInvitation
	NextPageToken string
}

func (s *Service) ListHackathonInvitations(ctx context.Context, in ListHackathonInvitationsIn) (*ListHackathonInvitationsOut, error) {
	listPolicy := rolepolicy.NewListHackathonStaffInvitationsPolicy(s.roleRepo)
	pctx, err := listPolicy.LoadContext(ctx, rolepolicy.ListHackathonStaffInvitationsParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return nil, ErrInvalidInput
	}

	// Fetch one extra to check if there's a next page
	invitations, err := s.invitationRepo.ListByHackathonID(ctx, in.HackathonID, int32(pageSize+1), int32(offset))
	if err != nil {
		return nil, err
	}

	var nextPageToken string
	if len(invitations) > int(pageSize) {
		invitations = invitations[:pageSize]
		nextPageToken = encodePageToken(offset + int(pageSize))
	}

	return &ListHackathonInvitationsOut{
		Invitations:   invitations,
		NextPageToken: nextPageToken,
	}, nil
}
