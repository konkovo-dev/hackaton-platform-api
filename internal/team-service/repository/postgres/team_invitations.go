package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TeamInvitationRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewTeamInvitationRepository(db queries.DBTX) *TeamInvitationRepository {
	return &TeamInvitationRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *TeamInvitationRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int32) ([]*entity.TeamInvitation, error) {
	rows, err := r.Queries().ListTeamInvitations(ctx, queries.ListTeamInvitationsParams{
		TeamID: teamID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list team invitations: %w", err)
	}

	invitations := make([]*entity.TeamInvitation, 0, len(rows))
	for _, row := range rows {
		var expiresAt *time.Time
		if row.ExpiresAt.Valid {
			t := row.ExpiresAt.Time
			expiresAt = &t
		}

		invitations = append(invitations, &entity.TeamInvitation{
			ID:              row.ID,
			HackathonID:     row.HackathonID,
			TeamID:          row.TeamID,
			VacancyID:       row.VacancyID,
			TargetUserID:    row.TargetUserID,
			CreatedByUserID: row.CreatedByUserID,
			Message:         row.Message,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ExpiresAt:       expiresAt,
		})
	}

	return invitations, nil
}

func (r *TeamInvitationRepository) CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountTeamInvitations(ctx, teamID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count team invitations: %w", err)
	}

	return count, nil
}

func (r *TeamInvitationRepository) Create(ctx context.Context, invitation *entity.TeamInvitation) error {
	var expiresAt pgtype.Timestamptz
	if invitation.ExpiresAt != nil {
		expiresAt = pgtype.Timestamptz{
			Time:  *invitation.ExpiresAt,
			Valid: true,
		}
	}

	row, err := r.Queries().CreateTeamInvitation(ctx, queries.CreateTeamInvitationParams{
		ID:              invitation.ID,
		HackathonID:     invitation.HackathonID,
		TeamID:          invitation.TeamID,
		VacancyID:       invitation.VacancyID,
		TargetUserID:    invitation.TargetUserID,
		CreatedByUserID: invitation.CreatedByUserID,
		Message:         invitation.Message,
		Status:          invitation.Status,
		CreatedAt:       invitation.CreatedAt,
		UpdatedAt:       invitation.UpdatedAt,
		ExpiresAt:       expiresAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create team invitation: %w", err)
	}

	invitation.CreatedAt = row.CreatedAt
	invitation.UpdatedAt = row.UpdatedAt
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		invitation.ExpiresAt = &t
	}

	return nil
}

func (r *TeamInvitationRepository) CheckUserInTeam(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	exists, err := r.Queries().CheckUserInTeam(ctx, queries.CheckUserInTeamParams{
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return false, fmt.Errorf("failed to check user in team: %w", err)
	}

	return exists, nil
}

func (r *TeamInvitationRepository) GetByID(ctx context.Context, invitationID uuid.UUID) (*entity.TeamInvitation, error) {
	row, err := r.Queries().GetTeamInvitationByID(ctx, invitationID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get team invitation: %w", err)
	}

	var expiresAt *time.Time
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		expiresAt = &t
	}

	return &entity.TeamInvitation{
		ID:              row.ID,
		HackathonID:     row.HackathonID,
		TeamID:          row.TeamID,
		VacancyID:       row.VacancyID,
		TargetUserID:    row.TargetUserID,
		CreatedByUserID: row.CreatedByUserID,
		Message:         row.Message,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ExpiresAt:       expiresAt,
	}, nil
}

func (r *TeamInvitationRepository) UpdateStatus(ctx context.Context, invitationID uuid.UUID, status string) error {
	_, err := r.Queries().UpdateTeamInvitationStatus(ctx, queries.UpdateTeamInvitationStatusParams{
		ID:     invitationID,
		Status: status,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update invitation status: %w", err)
	}

	return nil
}

func (r *TeamInvitationRepository) ListByTargetUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.TeamInvitation, error) {
	rows, err := r.Queries().ListMyTeamInvitations(ctx, queries.ListMyTeamInvitationsParams{
		TargetUserID: userID,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list my team invitations: %w", err)
	}

	invitations := make([]*entity.TeamInvitation, 0, len(rows))
	for _, row := range rows {
		var expiresAt *time.Time
		if row.ExpiresAt.Valid {
			t := row.ExpiresAt.Time
			expiresAt = &t
		}

		invitations = append(invitations, &entity.TeamInvitation{
			ID:              row.ID,
			HackathonID:     row.HackathonID,
			TeamID:          row.TeamID,
			VacancyID:       row.VacancyID,
			TargetUserID:    row.TargetUserID,
			CreatedByUserID: row.CreatedByUserID,
			Message:         row.Message,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ExpiresAt:       expiresAt,
		})
	}

	return invitations, nil
}

func (r *TeamInvitationRepository) CountByTargetUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountMyTeamInvitations(ctx, userID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count my team invitations: %w", err)
	}

	return count, nil
}

func (r *TeamInvitationRepository) CancelCompeting(ctx context.Context, userID, hackathonID uuid.UUID) error {
	err := r.Queries().CancelCompetingInvitations(ctx, queries.CancelCompetingInvitationsParams{
		TargetUserID: userID,
		HackathonID:  hackathonID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to cancel competing invitations: %w", err)
	}

	return nil
}
