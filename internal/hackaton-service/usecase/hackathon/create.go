package hackathon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type CreateHackathonIn struct {
	Name             string
	ShortDescription string
	Description      string

	LocationOnline  bool
	LocationCity    string
	LocationCountry string
	LocationVenue   string

	StartsAt             *time.Time
	EndsAt               *time.Time
	RegistrationOpensAt  *time.Time
	RegistrationClosesAt *time.Time
	SubmissionsOpensAt   *time.Time
	SubmissionsClosesAt  *time.Time
	JudgingEndsAt        *time.Time

	TeamSizeMax int32

	AllowIndividual bool
	AllowTeam       bool

	Links []CreateHackathonLink
}

type CreateHackathonLink struct {
	Title string
	URL   string
}

type CreateHackathonOut struct {
	HackathonID uuid.UUID
}

func (s *Service) CreateHackathon(ctx context.Context, in CreateHackathonIn) (*CreateHackathonOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	if err := s.validateCreateHackathonIn(in); err != nil {
		return nil, err
	}

	hackathonID := uuid.New()

	hackathon := &entity.Hackathon{
		ID:                   hackathonID,
		Name:                 in.Name,
		ShortDescription:     in.ShortDescription,
		Description:          in.Description,
		LocationOnline:       in.LocationOnline,
		LocationCity:         in.LocationCity,
		LocationCountry:      in.LocationCountry,
		LocationVenue:        in.LocationVenue,
		StartsAt:             in.StartsAt,
		EndsAt:               in.EndsAt,
		RegistrationOpensAt:  in.RegistrationOpensAt,
		RegistrationClosesAt: in.RegistrationClosesAt,
		SubmissionsOpensAt:   in.SubmissionsOpensAt,
		SubmissionsClosesAt:  in.SubmissionsClosesAt,
		JudgingEndsAt:        in.JudgingEndsAt,
		TeamSizeMax:          in.TeamSizeMax,
		AllowIndividual:      in.AllowIndividual,
		AllowTeam:            in.AllowTeam,
	}

	hackathon.Stage = computeStage(time.Now().UTC(), hackathon)

	err := s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
		if err := txRepos.Hackathons.Create(ctx, hackathon); err != nil {
			return fmt.Errorf("failed to create hackathon: %w", err)
		}

		for _, linkIn := range in.Links {
			link := &entity.HackathonLink{
				ID:          uuid.New(),
				HackathonID: hackathonID,
				Title:       linkIn.Title,
				URL:         linkIn.URL,
			}
			if err := txRepos.Links.Create(ctx, link); err != nil {
				return fmt.Errorf("failed to create hackathon link: %w", err)
			}
		}

		payload := map[string]string{
			"hackathon_id": hackathonID.String(),
			"user_id":      userID,
			"role":         string(domain.RoleOwner),
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal outbox payload: %w", err)
		}

		event := outbox.NewEvent(hackathonID.String(), "hackathon", "hackathon.owner_assigned", payloadBytes)

		if err := txRepos.Outbox.Create(ctx, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &CreateHackathonOut{
		HackathonID: hackathonID,
	}, nil
}

func (s *Service) validateCreateHackathonIn(in CreateHackathonIn) error {
	if in.Name == "" {
		return ErrEmptyName
	}

	if in.ShortDescription == "" {
		return ErrEmptyShortDescription
	}

	if !in.LocationOnline && in.LocationCity == "" && in.LocationCountry == "" {
		return ErrInvalidLocation
	}

	if in.StartsAt == nil {
		return ErrMissingStartsAt
	}
	if in.EndsAt == nil {
		return ErrMissingEndsAt
	}
	if in.RegistrationOpensAt == nil {
		return ErrMissingRegistrationOpensAt
	}
	if in.RegistrationClosesAt == nil {
		return ErrMissingRegistrationClosesAt
	}
	if in.SubmissionsOpensAt == nil {
		return ErrMissingSubmissionsOpensAt
	}
	if in.SubmissionsClosesAt == nil {
		return ErrMissingSubmissionsClosesAt
	}
	if in.JudgingEndsAt == nil {
		return ErrMissingJudgingEndsAt
	}

	if !in.RegistrationOpensAt.Before(*in.RegistrationClosesAt) {
		return ErrInvalidDateSequence
	}
	if !in.RegistrationClosesAt.Before(*in.StartsAt) {
		return ErrInvalidDateSequence
	}
	if !in.StartsAt.Before(*in.SubmissionsOpensAt) {
		return ErrInvalidDateSequence
	}
	if !in.EndsAt.Before(*in.SubmissionsClosesAt) && !in.EndsAt.Equal(*in.SubmissionsClosesAt) {
		return ErrInvalidDateSequence
	}
	if !in.SubmissionsClosesAt.Before(*in.JudgingEndsAt) {
		return ErrInvalidDateSequence
	}

	if in.TeamSizeMax <= 0 {
		return ErrInvalidTeamSizeMax
	}

	if !in.AllowIndividual && !in.AllowTeam {
		return ErrInvalidRegistrationPolicy
	}

	for _, link := range in.Links {
		if link.Title == "" {
			return ErrInvalidLink
		}
		if _, err := url.ParseRequestURI(link.URL); err != nil {
			return ErrInvalidLink
		}
	}

	return nil
}

func computeStage(now time.Time, dates *entity.Hackathon) string {
	if now.Before(*dates.RegistrationOpensAt) {
		return string(domain.StageUpcoming)
	}
	if now.Before(*dates.RegistrationClosesAt) {
		return string(domain.StageRegistration)
	}
	if now.Before(*dates.StartsAt) {
		return string(domain.StagePreStart)
	}
	if now.Before(*dates.EndsAt) {
		return string(domain.StageRunning)
	}
	if now.Before(*dates.JudgingEndsAt) {
		return string(domain.StageJudging)
	}
	return string(domain.StageFinished)
}
