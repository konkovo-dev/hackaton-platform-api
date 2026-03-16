package hackathon

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type UpdateHackathonIn struct {
	HackathonID uuid.UUID

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

type UpdateHackathonOut struct {
}

func (s *Service) UpdateHackathon(ctx context.Context, in UpdateHackathonIn) (*UpdateHackathonOut, error) {
	updatePolicy := hackathonpolicy.NewUpdateHackathonPolicy(s.hackathonRepo, s.parClient)
	pctx, err := updatePolicy.LoadContext(ctx, hackathonpolicy.UpdateHackathonParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := updatePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	oldHackathon, err := s.hackathonRepo.GetByID(ctx, in.HackathonID)
	if err != nil {
		return nil, err
	}
	if oldHackathon == nil {
		return nil, ErrHackathonNotFound
	}

	newHackathon := &entity.Hackathon{
		ID:                   oldHackathon.ID,
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
		Stage:                oldHackathon.Stage,
		State:                oldHackathon.State,
		PublishedAt:          oldHackathon.PublishedAt,
		Task:                 oldHackathon.Task,
		Result:               oldHackathon.Result,
		ResultPublishedAt:    oldHackathon.ResultPublishedAt,
		CreatedAt:            oldHackathon.CreatedAt,
	}

	err = s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
		if err := txRepos.Hackathons.Update(ctx, newHackathon); err != nil {
			return err
		}

		if err := txRepos.Links.DeleteByHackathonID(ctx, in.HackathonID); err != nil {
			return err
		}

		for _, link := range in.Links {
			hackathonLink := &entity.HackathonLink{
				ID:          uuid.New(),
				HackathonID: in.HackathonID,
				Title:       link.Title,
				URL:         link.URL,
			}
			if err := txRepos.Links.Create(ctx, hackathonLink); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &UpdateHackathonOut{}, nil
}
