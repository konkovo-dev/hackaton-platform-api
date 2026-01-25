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

type HackathonRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewHackathonRepository(db queries.DBTX) *HackathonRepository {
	return &HackathonRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *HackathonRepository) Create(ctx context.Context, hackathon *entity.Hackathon) error {
	now := time.Now().UTC()
	hackathon.CreatedAt = now
	hackathon.UpdatedAt = now

	err := r.Queries().CreateHackathon(ctx, queries.CreateHackathonParams{
		ID:                   hackathon.ID,
		Name:                 hackathon.Name,
		ShortDescription:     hackathon.ShortDescription,
		Description:          hackathon.Description,
		LocationOnline:       hackathon.LocationOnline,
		LocationCity:         hackathon.LocationCity,
		LocationCountry:      hackathon.LocationCountry,
		LocationVenue:        hackathon.LocationVenue,
		StartsAt:             pgxutil.TimePtrToPgtype(hackathon.StartsAt),
		EndsAt:               pgxutil.TimePtrToPgtype(hackathon.EndsAt),
		RegistrationOpensAt:  pgxutil.TimePtrToPgtype(hackathon.RegistrationOpensAt),
		RegistrationClosesAt: pgxutil.TimePtrToPgtype(hackathon.RegistrationClosesAt),
		SubmissionsOpensAt:   pgxutil.TimePtrToPgtype(hackathon.SubmissionsOpensAt),
		SubmissionsClosesAt:  pgxutil.TimePtrToPgtype(hackathon.SubmissionsClosesAt),
		JudgingEndsAt:        pgxutil.TimePtrToPgtype(hackathon.JudgingEndsAt),
		Stage:                hackathon.Stage,
		State:                hackathon.State,
		PublishedAt:          pgxutil.TimePtrToPgtype(hackathon.PublishedAt),
		ResultPublishedAt:    pgxutil.TimePtrToPgtype(hackathon.ResultPublishedAt),
		Task:                 hackathon.Task,
		Result:               hackathon.Result,
		TeamSizeMax:          hackathon.TeamSizeMax,
		AllowIndividual:      hackathon.AllowIndividual,
		AllowTeam:            hackathon.AllowTeam,
		CreatedAt:            hackathon.CreatedAt,
		UpdatedAt:            hackathon.UpdatedAt,
	})

	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsConflict(err) {
			return fmt.Errorf("hackathon already exists: %w", err)
		}
		return fmt.Errorf("failed to create hackathon: %w", err)
	}

	return nil
}

func (r *HackathonRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Hackathon, error) {
	row, err := r.Queries().GetHackathonByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, fmt.Errorf("hackathon not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	return &entity.Hackathon{
		ID:                   row.ID,
		Name:                 row.Name,
		ShortDescription:     row.ShortDescription,
		Description:          row.Description,
		LocationOnline:       row.LocationOnline,
		LocationCity:         row.LocationCity,
		LocationCountry:      row.LocationCountry,
		LocationVenue:        row.LocationVenue,
		StartsAt:             pgxutil.PgtypeTimestampToTimePtr(row.StartsAt),
		EndsAt:               pgxutil.PgtypeTimestampToTimePtr(row.EndsAt),
		RegistrationOpensAt:  pgxutil.PgtypeTimestampToTimePtr(row.RegistrationOpensAt),
		RegistrationClosesAt: pgxutil.PgtypeTimestampToTimePtr(row.RegistrationClosesAt),
		SubmissionsOpensAt:   pgxutil.PgtypeTimestampToTimePtr(row.SubmissionsOpensAt),
		SubmissionsClosesAt:  pgxutil.PgtypeTimestampToTimePtr(row.SubmissionsClosesAt),
		JudgingEndsAt:        pgxutil.PgtypeTimestampToTimePtr(row.JudgingEndsAt),
		Stage:                row.Stage,
		State:                row.State,
		PublishedAt:          pgxutil.PgtypeTimestampToTimePtr(row.PublishedAt),
		ResultPublishedAt:    pgxutil.PgtypeTimestampToTimePtr(row.ResultPublishedAt),
		Task:                 row.Task,
		Result:               row.Result,
		TeamSizeMax:          row.TeamSizeMax,
		AllowIndividual:      row.AllowIndividual,
		AllowTeam:            row.AllowTeam,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}

func (r *HackathonRepository) Update(ctx context.Context, hackathon *entity.Hackathon) error {
	hackathon.UpdatedAt = time.Now().UTC()

	err := r.Queries().UpdateHackathon(ctx, queries.UpdateHackathonParams{
		ID:                   hackathon.ID,
		Name:                 hackathon.Name,
		ShortDescription:     hackathon.ShortDescription,
		Description:          hackathon.Description,
		LocationOnline:       hackathon.LocationOnline,
		LocationCity:         hackathon.LocationCity,
		LocationCountry:      hackathon.LocationCountry,
		LocationVenue:        hackathon.LocationVenue,
		StartsAt:             pgxutil.TimePtrToPgtype(hackathon.StartsAt),
		EndsAt:               pgxutil.TimePtrToPgtype(hackathon.EndsAt),
		RegistrationOpensAt:  pgxutil.TimePtrToPgtype(hackathon.RegistrationOpensAt),
		RegistrationClosesAt: pgxutil.TimePtrToPgtype(hackathon.RegistrationClosesAt),
		SubmissionsOpensAt:   pgxutil.TimePtrToPgtype(hackathon.SubmissionsOpensAt),
		SubmissionsClosesAt:  pgxutil.TimePtrToPgtype(hackathon.SubmissionsClosesAt),
		JudgingEndsAt:        pgxutil.TimePtrToPgtype(hackathon.JudgingEndsAt),
		Stage:                hackathon.Stage,
		State:                hackathon.State,
		PublishedAt:          pgxutil.TimePtrToPgtype(hackathon.PublishedAt),
		ResultPublishedAt:    pgxutil.TimePtrToPgtype(hackathon.ResultPublishedAt),
		Task:                 hackathon.Task,
		Result:               hackathon.Result,
		TeamSizeMax:          hackathon.TeamSizeMax,
		AllowIndividual:      hackathon.AllowIndividual,
		AllowTeam:            hackathon.AllowTeam,
		UpdatedAt:            hackathon.UpdatedAt,
	})

	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return fmt.Errorf("hackathon not found: %w", err)
		}
		return fmt.Errorf("failed to update hackathon: %w", err)
	}

	return nil
}

func (r *HackathonRepository) List(ctx context.Context, limit, offset int32) ([]*entity.Hackathon, error) {
	rows, err := r.Queries().ListHackathons(ctx, queries.ListHackathonsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list hackathons: %w", pgxutil.MapDBError(err))
	}

	hackathons := make([]*entity.Hackathon, 0, len(rows))
	for _, row := range rows {
		hackathons = append(hackathons, &entity.Hackathon{
			ID:                   row.ID,
			Name:                 row.Name,
			ShortDescription:     row.ShortDescription,
			Description:          row.Description,
			LocationOnline:       row.LocationOnline,
			LocationCity:         row.LocationCity,
			LocationCountry:      row.LocationCountry,
			LocationVenue:        row.LocationVenue,
			StartsAt:             pgxutil.PgtypeTimestampToTimePtr(row.StartsAt),
			EndsAt:               pgxutil.PgtypeTimestampToTimePtr(row.EndsAt),
			RegistrationOpensAt:  pgxutil.PgtypeTimestampToTimePtr(row.RegistrationOpensAt),
			RegistrationClosesAt: pgxutil.PgtypeTimestampToTimePtr(row.RegistrationClosesAt),
			SubmissionsOpensAt:   pgxutil.PgtypeTimestampToTimePtr(row.SubmissionsOpensAt),
			SubmissionsClosesAt:  pgxutil.PgtypeTimestampToTimePtr(row.SubmissionsClosesAt),
			JudgingEndsAt:        pgxutil.PgtypeTimestampToTimePtr(row.JudgingEndsAt),
			Stage:                row.Stage,
			State:                row.State,
			PublishedAt:          pgxutil.PgtypeTimestampToTimePtr(row.PublishedAt),
			ResultPublishedAt:    pgxutil.PgtypeTimestampToTimePtr(row.ResultPublishedAt),
			Task:                 row.Task,
			Result:               row.Result,
			TeamSizeMax:          row.TeamSizeMax,
			AllowIndividual:      row.AllowIndividual,
			AllowTeam:            row.AllowTeam,
			CreatedAt:            row.CreatedAt,
			UpdatedAt:            row.UpdatedAt,
		})
	}

	return hackathons, nil
}

func (r *HackathonRepository) CountPublished(ctx context.Context) (int64, error) {
	count, err := r.Queries().CountPublishedHackathons(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count published hackathons: %w", pgxutil.MapDBError(err))
	}
	return count, nil
}
