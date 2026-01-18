package mappers

import (
	"time"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ProtoToCreateHackathonIn(req *hackathonv1.CreateHackathonRequest) hackathon.CreateHackathonIn {
	in := hackathon.CreateHackathonIn{
		Name:             req.Name,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		TeamSizeMax:      int32(req.Limits.TeamSizeMax),
	}

	if req.Location != nil {
		in.LocationOnline = req.Location.Online
		in.LocationCity = req.Location.City
		in.LocationCountry = req.Location.Country
		in.LocationVenue = req.Location.Venue
	}

	if req.Dates != nil {
		if req.Dates.StartsAt != nil {
			t := req.Dates.StartsAt.AsTime()
			in.StartsAt = &t
		}
		if req.Dates.EndsAt != nil {
			t := req.Dates.EndsAt.AsTime()
			in.EndsAt = &t
		}
		if req.Dates.RegistrationOpensAt != nil {
			t := req.Dates.RegistrationOpensAt.AsTime()
			in.RegistrationOpensAt = &t
		}
		if req.Dates.RegistrationClosesAt != nil {
			t := req.Dates.RegistrationClosesAt.AsTime()
			in.RegistrationClosesAt = &t
		}
		if req.Dates.SubmissionsOpensAt != nil {
			t := req.Dates.SubmissionsOpensAt.AsTime()
			in.SubmissionsOpensAt = &t
		}
		if req.Dates.SubmissionsClosesAt != nil {
			t := req.Dates.SubmissionsClosesAt.AsTime()
			in.SubmissionsClosesAt = &t
		}
		if req.Dates.JudgingEndsAt != nil {
			t := req.Dates.JudgingEndsAt.AsTime()
			in.JudgingEndsAt = &t
		}
	}

	if req.RegistrationPolicy != nil {
		in.AllowIndividual = req.RegistrationPolicy.AllowIndividual
		in.AllowTeam = req.RegistrationPolicy.AllowTeam
	}

	for _, link := range req.Links {
		in.Links = append(in.Links, hackathon.CreateHackathonLink{
			Title: link.Title,
			URL:   link.Url,
		})
	}

	return in
}

func timestampOrNil(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func mapDomainStageToProto(stage string) hackathonv1.HackathonStage {
	switch domain.HackathonStage(stage) {
	case domain.StageUpcoming:
		return hackathonv1.HackathonStage_HACKATHON_STAGE_UPCOMING
	case domain.StageRegistration:
		return hackathonv1.HackathonStage_HACKATHON_STAGE_REGISTRATION
	case domain.StagePreStart:
		return hackathonv1.HackathonStage_HACKATHON_STAGE_PRE_START
	case domain.StageRunning:
		return hackathonv1.HackathonStage_HACKATHON_STAGE_RUNNING
	case domain.StageJudging:
		return hackathonv1.HackathonStage_HACKATHON_STAGE_JUDGING
	case domain.StageFinished:
		return hackathonv1.HackathonStage_HACKATHON_STAGE_FINISHED
	default:
		return hackathonv1.HackathonStage_HACKATHON_STAGE_UNSPECIFIED
	}
}

func GetHackathonOutToProto(out *hackathon.GetHackathonOut) *hackathonv1.GetHackathonResponse {
	h := out.Hackathon

	resp := &hackathonv1.GetHackathonResponse{
		Hackathon: &hackathonv1.Hackathon{
			HackathonId:      h.ID.String(),
			Name:             h.Name,
			ShortDescription: h.ShortDescription,
			Description:      h.Description,
			Location: &hackathonv1.HackathonLocation{
				Online:  h.LocationOnline,
				City:    h.LocationCity,
				Country: h.LocationCountry,
				Venue:   h.LocationVenue,
			},
			Dates: &hackathonv1.HackathonDates{
				StartsAt:             timestampOrNil(h.StartsAt),
				EndsAt:               timestampOrNil(h.EndsAt),
				RegistrationOpensAt:  timestampOrNil(h.RegistrationOpensAt),
				RegistrationClosesAt: timestampOrNil(h.RegistrationClosesAt),
				SubmissionsOpensAt:   timestampOrNil(h.SubmissionsOpensAt),
				SubmissionsClosesAt:  timestampOrNil(h.SubmissionsClosesAt),
				JudgingEndsAt:        timestampOrNil(h.JudgingEndsAt),
			},
			Stage: mapDomainStageToProto(h.Stage),
			Limits: &hackathonv1.HackathonLimits{
				TeamSizeMax: uint32(h.TeamSizeMax),
			},
			RegistrationPolicy: &hackathonv1.HackathonRegistrationPolicy{
				AllowIndividual: h.AllowIndividual,
				AllowTeam:       h.AllowTeam,
			},
		},
	}

	for _, link := range out.Links {
		resp.Hackathon.Links = append(resp.Hackathon.Links, &hackathonv1.HackathonLink{
			Title: link.Title,
			Url:   link.URL,
		})
	}

	return resp
}
