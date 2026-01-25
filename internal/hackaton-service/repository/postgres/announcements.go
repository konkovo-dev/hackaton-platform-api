package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type AnnouncementRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewAnnouncementRepository(db queries.DBTX) *AnnouncementRepository {
	return &AnnouncementRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *AnnouncementRepository) Create(ctx context.Context, announcement *entity.HackathonAnnouncement) error {
	now := time.Now().UTC()
	announcement.CreatedAt = now
	announcement.UpdatedAt = now

	err := r.Queries().CreateAnnouncement(ctx, queries.CreateAnnouncementParams{
		ID:              announcement.ID,
		HackathonID:     announcement.HackathonID,
		Title:           announcement.Title,
		Body:            announcement.Body,
		CreatedByUserID: announcement.CreatedByUser,
		CreatedAt:       announcement.CreatedAt,
		UpdatedAt:       announcement.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create announcement: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *AnnouncementRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.HackathonAnnouncement, error) {
	row, err := r.Queries().GetAnnouncementByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get announcement: %w", err)
	}

	return &entity.HackathonAnnouncement{
		ID:            row.ID,
		HackathonID:   row.HackathonID,
		Title:         row.Title,
		Body:          row.Body,
		CreatedByUser: row.CreatedByUserID,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		DeletedAt:     pgxutil.PgtypeTimestampToTimePtr(row.DeletedAt),
	}, nil
}

func (r *AnnouncementRepository) ListByHackathonID(ctx context.Context, hackathonID uuid.UUID) ([]*entity.HackathonAnnouncement, error) {
	rows, err := r.Queries().ListAnnouncementsByHackathonID(ctx, hackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list announcements: %w", pgxutil.MapDBError(err))
	}

	announcements := make([]*entity.HackathonAnnouncement, 0, len(rows))
	for _, row := range rows {
		announcements = append(announcements, &entity.HackathonAnnouncement{
			ID:            row.ID,
			HackathonID:   row.HackathonID,
			Title:         row.Title,
			Body:          row.Body,
			CreatedByUser: row.CreatedByUserID,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
			DeletedAt:     pgxutil.PgtypeTimestampToTimePtr(row.DeletedAt),
		})
	}

	return announcements, nil
}

func (r *AnnouncementRepository) Update(ctx context.Context, announcement *entity.HackathonAnnouncement) error {
	announcement.UpdatedAt = time.Now().UTC()

	err := r.Queries().UpdateAnnouncement(ctx, queries.UpdateAnnouncementParams{
		ID:        announcement.ID,
		Title:     announcement.Title,
		Body:      announcement.Body,
		UpdatedAt: announcement.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to update announcement: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *AnnouncementRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	err := r.Queries().SoftDeleteAnnouncement(ctx, queries.SoftDeleteAnnouncementParams{
		ID:        id,
		DeletedAt: pgxutil.TimeToPgtype(now),
	})

	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", pgxutil.MapDBError(err))
	}

	return nil
}
