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
