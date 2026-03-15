package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *HackathonRepository) List(ctx context.Context, params hackathon.ListHackathonsRepoParams) ([]*entity.Hackathon, error) {
	query, args, err := r.buildListHackathonsQuery(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.DB().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", pgxutil.MapDBError(err))
	}
	defer rows.Close()

	hackathons := make([]*entity.Hackathon, 0)
	for rows.Next() {
		var h entity.Hackathon
		var startsAt, endsAt, registrationOpensAt, registrationClosesAt pgtype.Timestamptz
		var submissionsOpensAt, submissionsClosesAt, judgingEndsAt pgtype.Timestamptz
		var publishedAt, resultPublishedAt pgtype.Timestamptz

		err := rows.Scan(
			&h.ID,
			&h.Name,
			&h.ShortDescription,
			&h.Description,
			&h.LocationOnline,
			&h.LocationCity,
			&h.LocationCountry,
			&h.LocationVenue,
			&startsAt,
			&endsAt,
			&registrationOpensAt,
			&registrationClosesAt,
			&submissionsOpensAt,
			&submissionsClosesAt,
			&judgingEndsAt,
			&h.Stage,
			&h.State,
			&publishedAt,
			&resultPublishedAt,
			&h.Task,
			&h.Result,
			&h.TeamSizeMax,
			&h.AllowIndividual,
			&h.AllowTeam,
			&h.CreatedAt,
			&h.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		h.StartsAt = pgxutil.PgtypeTimestampToTimePtr(startsAt)
		h.EndsAt = pgxutil.PgtypeTimestampToTimePtr(endsAt)
		h.RegistrationOpensAt = pgxutil.PgtypeTimestampToTimePtr(registrationOpensAt)
		h.RegistrationClosesAt = pgxutil.PgtypeTimestampToTimePtr(registrationClosesAt)
		h.SubmissionsOpensAt = pgxutil.PgtypeTimestampToTimePtr(submissionsOpensAt)
		h.SubmissionsClosesAt = pgxutil.PgtypeTimestampToTimePtr(submissionsClosesAt)
		h.JudgingEndsAt = pgxutil.PgtypeTimestampToTimePtr(judgingEndsAt)
		h.PublishedAt = pgxutil.PgtypeTimestampToTimePtr(publishedAt)
		h.ResultPublishedAt = pgxutil.PgtypeTimestampToTimePtr(resultPublishedAt)

		hackathons = append(hackathons, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return hackathons, nil
}

func (r *HackathonRepository) buildListHackathonsQuery(params hackathon.ListHackathonsRepoParams) (string, []interface{}, error) {
	baseQuery := `SELECT 
		h.id,
		h.name,
		h.short_description,
		h.description,
		h.location_online,
		h.location_city,
		h.location_country,
		h.location_venue,
		h.starts_at,
		h.ends_at,
		h.registration_opens_at,
		h.registration_closes_at,
		h.submissions_opens_at,
		h.submissions_closes_at,
		h.judging_ends_at,
		h.stage,
		h.state,
		h.published_at,
		h.result_published_at,
		h.task,
		h.result,
		h.team_size_max,
		h.allow_individual,
		h.allow_team,
		h.created_at,
		h.updated_at
	FROM hackathon.hackathons h`

	qb := queryutil.NewQueryBuilder(baseQuery)

	qb.WithCustomWhere("h.state = ?", "published")

	if params.Filters != nil && len(params.Filters.HackathonIDsOnly) > 0 {
		placeholders := make([]string, len(params.Filters.HackathonIDsOnly))
		args := make([]interface{}, len(params.Filters.HackathonIDsOnly))
		for i, id := range params.Filters.HackathonIDsOnly {
			placeholders[i] = "?"
			args[i] = id
		}
		qb.WithCustomWhere(fmt.Sprintf("h.id IN (%s)", strings.Join(placeholders, ", ")), args...)
	}

	qb.WithSearch([]string{"h.name", "h.short_description"}, params.SearchQuery)

	if params.Filters != nil && len(params.Filters.FilterGroups) > 0 {
		qb.WithFilters(params.Filters.FilterGroups, params.FieldMapping)
	}

	query, args := qb.Build()

	if len(params.Sort) > 0 {
		sortFields := make([]string, 0, len(params.Sort))
		for _, sf := range params.Sort {
			mappedField := sf.Field
			if params.FieldMapping != nil {
				if mapped, ok := params.FieldMapping[sf.Field]; ok {
					mappedField = mapped
				}
			}
			sortFields = append(sortFields, fmt.Sprintf("%s %s", mappedField, sf.Direction))
		}
		query += " ORDER BY " + strings.Join(sortFields, ", ")
	} else {
		query += ` ORDER BY 
			CASE 
				WHEN h.stage IN ('registration', 'prestart', 'running') THEN 1
				WHEN h.stage = 'upcoming' THEN 2
				WHEN h.stage = 'judging' THEN 3
				WHEN h.stage = 'finished' THEN 4
				ELSE 5
			END,
			h.created_at DESC`
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.Limit, params.Offset)

	return query, args, nil
}

func (r *HackathonRepository) CountPublished(ctx context.Context) (int64, error) {
	count, err := r.Queries().CountPublishedHackathons(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count published hackathons: %w", pgxutil.MapDBError(err))
	}
	return count, nil
}
