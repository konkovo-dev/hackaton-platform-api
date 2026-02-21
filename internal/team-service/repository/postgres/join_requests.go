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

type JoinRequestRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewJoinRequestRepository(db queries.DBTX) *JoinRequestRepository {
	return &JoinRequestRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *JoinRequestRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int32) ([]*entity.JoinRequest, error) {
	rows, err := r.Queries().ListJoinRequests(ctx, queries.ListJoinRequestsParams{
		TeamID: teamID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list join requests: %w", err)
	}

	requests := make([]*entity.JoinRequest, 0, len(rows))
	for _, row := range rows {
		var expiresAt *time.Time
		if row.ExpiresAt.Valid {
			t := row.ExpiresAt.Time
			expiresAt = &t
		}

		requests = append(requests, &entity.JoinRequest{
			ID:              row.ID,
			HackathonID:     row.HackathonID,
			TeamID:          row.TeamID,
			VacancyID:       row.VacancyID,
			RequesterUserID: row.RequesterUserID,
			Message:         row.Message,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ExpiresAt:       expiresAt,
		})
	}

	return requests, nil
}

func (r *JoinRequestRepository) CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountJoinRequests(ctx, teamID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count join requests: %w", err)
	}

	return count, nil
}

func (r *JoinRequestRepository) Create(ctx context.Context, request *entity.JoinRequest) error {
	var expiresAt pgtype.Timestamptz
	if request.ExpiresAt != nil {
		expiresAt = pgtype.Timestamptz{
			Time:  *request.ExpiresAt,
			Valid: true,
		}
	}

	err := r.Queries().CreateJoinRequest(ctx, queries.CreateJoinRequestParams{
		ID:              request.ID,
		HackathonID:     request.HackathonID,
		TeamID:          request.TeamID,
		VacancyID:       request.VacancyID,
		RequesterUserID: request.RequesterUserID,
		Message:         request.Message,
		Status:          request.Status,
		CreatedAt:       request.CreatedAt,
		UpdatedAt:       request.UpdatedAt,
		ExpiresAt:       expiresAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create join request: %w", err)
	}

	return nil
}

func (r *JoinRequestRepository) ListByRequesterUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.JoinRequest, error) {
	rows, err := r.Queries().ListMyJoinRequests(ctx, queries.ListMyJoinRequestsParams{
		RequesterUserID: userID,
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list my join requests: %w", err)
	}

	requests := make([]*entity.JoinRequest, 0, len(rows))
	for _, row := range rows {
		var expiresAt *time.Time
		if row.ExpiresAt.Valid {
			t := row.ExpiresAt.Time
			expiresAt = &t
		}

		requests = append(requests, &entity.JoinRequest{
			ID:              row.ID,
			HackathonID:     row.HackathonID,
			TeamID:          row.TeamID,
			VacancyID:       row.VacancyID,
			RequesterUserID: row.RequesterUserID,
			Message:         row.Message,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ExpiresAt:       expiresAt,
		})
	}

	return requests, nil
}

func (r *JoinRequestRepository) CountByRequesterUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountMyJoinRequests(ctx, userID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count my join requests: %w", err)
	}

	return count, nil
}

func (r *JoinRequestRepository) GetByID(ctx context.Context, requestID uuid.UUID) (*entity.JoinRequest, error) {
	row, err := r.Queries().GetJoinRequestByID(ctx, requestID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get join request: %w", err)
	}

	var expiresAt *time.Time
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		expiresAt = &t
	}

	return &entity.JoinRequest{
		ID:              row.ID,
		HackathonID:     row.HackathonID,
		TeamID:          row.TeamID,
		VacancyID:       row.VacancyID,
		RequesterUserID: row.RequesterUserID,
		Message:         row.Message,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ExpiresAt:       expiresAt,
	}, nil
}

func (r *JoinRequestRepository) UpdateStatus(ctx context.Context, requestID uuid.UUID, status string) error {
	_, err := r.Queries().UpdateJoinRequestStatus(ctx, queries.UpdateJoinRequestStatusParams{
		ID:     requestID,
		Status: status,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update join request status: %w", err)
	}

	return nil
}

func (r *JoinRequestRepository) CancelCompeting(ctx context.Context, userID, hackathonID uuid.UUID) error {
	err := r.Queries().CancelCompetingJoinRequests(ctx, queries.CancelCompetingJoinRequestsParams{
		RequesterUserID: userID,
		HackathonID:     hackathonID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to cancel competing join requests: %w", err)
	}

	return nil
}
