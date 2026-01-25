package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type HackathonLinkRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewHackathonLinkRepository(db queries.DBTX) *HackathonLinkRepository {
	return &HackathonLinkRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *HackathonLinkRepository) Create(ctx context.Context, link *entity.HackathonLink) error {
	err := r.Queries().CreateHackathonLink(ctx, queries.CreateHackathonLinkParams{
		ID:          link.ID,
		HackathonID: link.HackathonID,
		Title:       link.Title,
		Url:         link.URL,
	})

	if err != nil {
		return fmt.Errorf("failed to create hackathon link: %w", err)
	}

	return nil
}

func (r *HackathonLinkRepository) GetByHackathonID(ctx context.Context, hackathonID uuid.UUID) ([]*entity.HackathonLink, error) {
	rows, err := r.Queries().GetHackathonLinks(ctx, hackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon links: %w", err)
	}

	links := make([]*entity.HackathonLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, &entity.HackathonLink{
			ID:          row.ID,
			HackathonID: row.HackathonID,
			Title:       row.Title,
			URL:         row.Url,
		})
	}

	return links, nil
}

func (r *HackathonLinkRepository) GetByHackathonIDs(ctx context.Context, hackathonIDs []uuid.UUID) ([]*entity.HackathonLink, error) {
	rows, err := r.Queries().GetHackathonLinksByIDs(ctx, hackathonIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon links: %w", err)
	}

	links := make([]*entity.HackathonLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, &entity.HackathonLink{
			ID:          row.ID,
			HackathonID: row.HackathonID,
			Title:       row.Title,
			URL:         row.Url,
		})
	}

	return links, nil
}

func (r *HackathonLinkRepository) DeleteByHackathonID(ctx context.Context, hackathonID uuid.UUID) error {
	err := r.Queries().DeleteHackathonLinks(ctx, hackathonID)
	if err != nil {
		return fmt.Errorf("failed to delete hackathon links: %w", err)
	}
	return nil
}
