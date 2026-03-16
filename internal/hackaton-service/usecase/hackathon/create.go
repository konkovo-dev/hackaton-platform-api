package hackathon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
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
	createPolicy := hackathonpolicy.NewCreateHackathonPolicy()
	pctx, err := createPolicy.LoadContext(ctx, hackathonpolicy.CreateHackathonParams{})
	if err != nil {
		return nil, err
	}

	decision := createPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	hctx := pctx.(*hackathonpolicy.HackathonPolicyContext)
	userID := hctx.ActorUserID().String()

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
		State:                string(domain.StateDraft),
		Stage:                string(domain.StageDraft),
		PublishedAt:          nil,
		Task:                 "",
		Result:               "",
	}

	err = s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
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

func (s *Service) mapPolicyError(decision *policy.Decision) error {
	if len(decision.Violations) == 0 {
		return ErrUnauthorized
	}

	v := decision.Violations[0]
	if v.Code == policy.ViolationCodeForbidden {
		return ErrUnauthorized
	}

	return ErrUnauthorized
}
