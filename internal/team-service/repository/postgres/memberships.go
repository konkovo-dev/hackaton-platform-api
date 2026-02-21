package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type MembershipRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewMembershipRepository(db queries.DBTX) *MembershipRepository {
	return &MembershipRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *MembershipRepository) Create(ctx context.Context, membership *entity.Membership) error {
	var assignedVacancyID pgtype.UUID
	if membership.AssignedVacancyID != nil {
		assignedVacancyID = pgtype.UUID{
			Bytes: *membership.AssignedVacancyID,
			Valid: true,
		}
	}

	row, err := r.Queries().CreateMembership(ctx, queries.CreateMembershipParams{
		TeamID:            membership.TeamID,
		UserID:            membership.UserID,
		IsCaptain:         membership.IsCaptain,
		AssignedVacancyID: assignedVacancyID,
		JoinedAt:          membership.JoinedAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create membership: %w", err)
	}

	membership.JoinedAt = row.JoinedAt

	return nil
}

func (r *MembershipRepository) CheckIsCaptain(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	isCaptain, err := r.Queries().CheckIsCaptain(ctx, queries.CheckIsCaptainParams{
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return false, fmt.Errorf("failed to check captain status: %w", err)
	}

	return isCaptain, nil
}

func (r *MembershipRepository) CountMembers(ctx context.Context, teamID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountMembers(ctx, teamID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count members: %w", err)
	}

	return count, nil
}

func (r *MembershipRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]*entity.Membership, error) {
	rows, err := r.Queries().ListTeamMembers(ctx, teamID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list team members: %w", err)
	}

	members := make([]*entity.Membership, 0, len(rows))
	for _, row := range rows {
		var assignedVacancyID *uuid.UUID
		if row.AssignedVacancyID.Valid {
			id := uuid.UUID(row.AssignedVacancyID.Bytes)
			assignedVacancyID = &id
		}

		members = append(members, &entity.Membership{
			TeamID:            row.TeamID,
			UserID:            row.UserID,
			IsCaptain:         row.IsCaptain,
			AssignedVacancyID: assignedVacancyID,
			JoinedAt:          row.JoinedAt,
		})
	}

	return members, nil
}

func (r *MembershipRepository) GetByTeamAndUser(ctx context.Context, teamID, userID uuid.UUID) (*entity.Membership, error) {
	row, err := r.Queries().GetMembership(ctx, queries.GetMembershipParams{
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get membership: %w", err)
	}

	var assignedVacancyID *uuid.UUID
	if row.AssignedVacancyID.Valid {
		id := uuid.UUID(row.AssignedVacancyID.Bytes)
		assignedVacancyID = &id
	}

	return &entity.Membership{
		TeamID:            row.TeamID,
		UserID:            row.UserID,
		IsCaptain:         row.IsCaptain,
		AssignedVacancyID: assignedVacancyID,
		JoinedAt:          row.JoinedAt,
	}, nil
}

func (r *MembershipRepository) Delete(ctx context.Context, teamID, userID uuid.UUID) error {
	err := r.Queries().DeleteMembership(ctx, queries.DeleteMembershipParams{
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	return nil
}

func (r *MembershipRepository) UpdateCaptainStatus(ctx context.Context, teamID, userID uuid.UUID, isCaptain bool) error {
	err := r.Queries().UpdateCaptainStatus(ctx, queries.UpdateCaptainStatusParams{
		TeamID:    teamID,
		UserID:    userID,
		IsCaptain: isCaptain,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update captain status: %w", err)
	}

	return nil
}
