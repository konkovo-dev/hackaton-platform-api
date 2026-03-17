package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type StaffInvitationRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewStaffInvitationRepository(db queries.DBTX) *StaffInvitationRepository {
	return &StaffInvitationRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *StaffInvitationRepository) Create(ctx context.Context, invitation *entity.StaffInvitation) error {
	err := r.Queries().CreateStaffInvitation(ctx, queries.CreateStaffInvitationParams{
		ID:              invitation.ID,
		HackathonID:     invitation.HackathonID,
		TargetUserID:    invitation.TargetUserID,
		RequestedRole:   invitation.RequestedRole,
		CreatedByUserID: invitation.CreatedByUserID,
		Message:         invitation.Message,
		Status:          invitation.Status,
		CreatedAt:       invitation.CreatedAt,
		UpdatedAt:       invitation.UpdatedAt,
		ExpiresAt:       pgxutil.TimePtrToPgtype(invitation.ExpiresAt),
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create staff invitation: %w", err)
	}
	return nil
}

func (r *StaffInvitationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.StaffInvitation, error) {
	row, err := r.Queries().GetStaffInvitationByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get staff invitation: %w", err)
	}

	return &entity.StaffInvitation{
		ID:              row.ID,
		HackathonID:     row.HackathonID,
		TargetUserID:    row.TargetUserID,
		RequestedRole:   row.RequestedRole,
		CreatedByUserID: row.CreatedByUserID,
		Message:         row.Message,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ExpiresAt:       pgxutil.PgtypeTimestampToTimePtr(row.ExpiresAt),
	}, nil
}

func (r *StaffInvitationRepository) GetPendingInvitationForUser(ctx context.Context, hackathonID, userID uuid.UUID, role string) (*entity.StaffInvitation, error) {
	row, err := r.Queries().GetPendingInvitationForUser(ctx, queries.GetPendingInvitationForUserParams{
		HackathonID:   hackathonID,
		TargetUserID:  userID,
		RequestedRole: role,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pending invitation: %w", err)
	}

	return &entity.StaffInvitation{
		ID:              row.ID,
		HackathonID:     row.HackathonID,
		TargetUserID:    row.TargetUserID,
		RequestedRole:   row.RequestedRole,
		CreatedByUserID: row.CreatedByUserID,
		Message:         row.Message,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ExpiresAt:       pgxutil.PgtypeTimestampToTimePtr(row.ExpiresAt),
	}, nil
}

func (r *StaffInvitationRepository) GetByTargetUserID(ctx context.Context, userID uuid.UUID) ([]*entity.StaffInvitation, error) {
	rows, err := r.Queries().GetStaffInvitationsByTargetUser(ctx, userID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get staff invitations by target user: %w", err)
	}

	invitations := make([]*entity.StaffInvitation, 0, len(rows))
	for _, row := range rows {
		invitations = append(invitations, &entity.StaffInvitation{
			ID:              row.ID,
			HackathonID:     row.HackathonID,
			TargetUserID:    row.TargetUserID,
			RequestedRole:   row.RequestedRole,
			CreatedByUserID: row.CreatedByUserID,
			Message:         row.Message,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ExpiresAt:       pgxutil.PgtypeTimestampToTimePtr(row.ExpiresAt),
		})
	}

	return invitations, nil
}

func (r *StaffInvitationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, updatedAt time.Time) error {
	err := r.Queries().UpdateStaffInvitationStatus(ctx, queries.UpdateStaffInvitationStatusParams{
		ID:        id,
		Status:    status,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update staff invitation status: %w", err)
	}
	return nil
}

func (r *StaffInvitationRepository) GetStatusAndHackathonID(ctx context.Context, id uuid.UUID) (string, uuid.UUID, error) {
	invitation, err := r.GetByID(ctx, id)
	if err != nil {
		return "", uuid.Nil, err
	}
	return invitation.Status, invitation.HackathonID, nil
}

func (r *StaffInvitationRepository) GetDetails(ctx context.Context, id uuid.UUID) (bool, string, uuid.UUID, uuid.UUID, string, error) {
	invitation, err := r.GetByID(ctx, id)
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return false, "", uuid.Nil, uuid.Nil, "", nil
		}
		return false, "", uuid.Nil, uuid.Nil, "", err
	}
	return true, invitation.Status, invitation.TargetUserID, invitation.HackathonID, invitation.RequestedRole, nil
}

func (r *StaffInvitationRepository) GetBasicInfo(ctx context.Context, id uuid.UUID) (bool, string, uuid.UUID, error) {
	invitation, err := r.GetByID(ctx, id)
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return false, "", uuid.Nil, nil
		}
		return false, "", uuid.Nil, err
	}
	return true, invitation.Status, invitation.TargetUserID, nil
}

func (r *StaffInvitationRepository) ListByHackathonID(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*entity.StaffInvitation, error) {
	rows, err := r.Queries().ListStaffInvitationsByHackathon(ctx, queries.ListStaffInvitationsByHackathonParams{
		HackathonID: hackathonID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list staff invitations by hackathon: %w", err)
	}

	invitations := make([]*entity.StaffInvitation, 0, len(rows))
	for _, row := range rows {
		invitations = append(invitations, &entity.StaffInvitation{
			ID:              row.ID,
			HackathonID:     row.HackathonID,
			TargetUserID:    row.TargetUserID,
			RequestedRole:   row.RequestedRole,
			CreatedByUserID: row.CreatedByUserID,
			Message:         row.Message,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ExpiresAt:       pgxutil.PgtypeTimestampToTimePtr(row.ExpiresAt),
		})
	}

	return invitations, nil
}
