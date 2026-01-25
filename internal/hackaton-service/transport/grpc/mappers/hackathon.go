package mappers

import (
	"time"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ProtoToCreateHackathonIn(req *hackathonv1.CreateHackathonRequest) hackathon.CreateHackathonIn {
	in := hackathon.CreateHackathonIn{
		Name:             req.Name,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
	}

	if req.Limits != nil {
		in.TeamSizeMax = int32(req.Limits.TeamSizeMax)
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

func ProtoToUpdateHackathonIn(req *hackathonv1.UpdateHackathonRequest) hackathon.UpdateHackathonIn {
	hackathonID, _ := uuid.Parse(req.HackathonId)

	in := hackathon.UpdateHackathonIn{
		HackathonID:      hackathonID,
		Name:             req.Name,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
	}

	if req.Limits != nil {
		in.TeamSizeMax = int32(req.Limits.TeamSizeMax)
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
			Stage: stageToProto(h.Stage),
			State: stateToProto(h.State),
			Limits: &hackathonv1.HackathonLimits{
				TeamSizeMax: uint32(h.TeamSizeMax),
			},
			RegistrationPolicy: &hackathonv1.HackathonRegistrationPolicy{
				AllowIndividual: h.AllowIndividual,
				AllowTeam:       h.AllowTeam,
			},
		},
	}

	if h.Task != "" {
		resp.Hackathon.Task = &h.Task
	}

	if h.Result != "" {
		resp.Hackathon.Result = &h.Result
	}

	if h.PublishedAt != nil {
		resp.Hackathon.PublishedAt = timestamppb.New(*h.PublishedAt)
	}

	if h.ResultPublishedAt != nil {
		resp.Hackathon.ResultPublishedAt = timestamppb.New(*h.ResultPublishedAt)
	}

	for _, link := range out.Links {
		resp.Hackathon.Links = append(resp.Hackathon.Links, &hackathonv1.HackathonLink{
			Title: link.Title,
			Url:   link.URL,
		})
	}

	return resp
}

type HackathonConversionOptions struct {
	IncludeDescription bool
	IncludeLimits      bool
	IncludeTask        bool
	IncludeResult      bool
}

func HackathonToProto(h *entity.Hackathon, opts HackathonConversionOptions) *hackathonv1.Hackathon {
	proto := &hackathonv1.Hackathon{
		HackathonId:      h.ID.String(),
		Name:             h.Name,
		ShortDescription: h.ShortDescription,
		Stage:            stageToProto(h.Stage),
		State:            stateToProto(h.State),
		Location: &hackathonv1.HackathonLocation{
			Online:  h.LocationOnline,
			City:    h.LocationCity,
			Country: h.LocationCountry,
			Venue:   h.LocationVenue,
		},
		Dates: &hackathonv1.HackathonDates{},
		RegistrationPolicy: &hackathonv1.HackathonRegistrationPolicy{
			AllowIndividual: h.AllowIndividual,
			AllowTeam:       h.AllowTeam,
		},
		CreatedAt: timestamppb.New(h.CreatedAt),
		UpdatedAt: timestamppb.New(h.UpdatedAt),
	}

	if opts.IncludeDescription {
		proto.Description = h.Description
	}

	if opts.IncludeLimits {
		proto.Limits = &hackathonv1.HackathonLimits{
			TeamSizeMax: uint32(h.TeamSizeMax),
		}
	}

	if opts.IncludeTask && h.Task != "" {
		proto.Task = &h.Task
	}

	if opts.IncludeResult && h.Result != "" {
		proto.Result = &h.Result
	}

	if h.StartsAt != nil {
		proto.Dates.StartsAt = timestamppb.New(*h.StartsAt)
	}
	if h.EndsAt != nil {
		proto.Dates.EndsAt = timestamppb.New(*h.EndsAt)
	}
	if h.RegistrationOpensAt != nil {
		proto.Dates.RegistrationOpensAt = timestamppb.New(*h.RegistrationOpensAt)
	}
	if h.RegistrationClosesAt != nil {
		proto.Dates.RegistrationClosesAt = timestamppb.New(*h.RegistrationClosesAt)
	}
	if h.SubmissionsOpensAt != nil {
		proto.Dates.SubmissionsOpensAt = timestamppb.New(*h.SubmissionsOpensAt)
	}
	if h.SubmissionsClosesAt != nil {
		proto.Dates.SubmissionsClosesAt = timestamppb.New(*h.SubmissionsClosesAt)
	}
	if h.JudgingEndsAt != nil {
		proto.Dates.JudgingEndsAt = timestamppb.New(*h.JudgingEndsAt)
	}
	if h.PublishedAt != nil {
		proto.PublishedAt = timestamppb.New(*h.PublishedAt)
	}

	if h.ResultPublishedAt != nil {
		proto.ResultPublishedAt = timestamppb.New(*h.ResultPublishedAt)
	}

	return proto
}

func timestampOrNil(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func stageToProto(stage string) hackathonv1.HackathonStage {
	switch stage {
	case "draft":
		return hackathonv1.HackathonStage_HACKATHON_STAGE_DRAFT
	case "upcoming":
		return hackathonv1.HackathonStage_HACKATHON_STAGE_UPCOMING
	case "registration":
		return hackathonv1.HackathonStage_HACKATHON_STAGE_REGISTRATION
	case "prestart":
		return hackathonv1.HackathonStage_HACKATHON_STAGE_PRE_START
	case "running":
		return hackathonv1.HackathonStage_HACKATHON_STAGE_RUNNING
	case "judging":
		return hackathonv1.HackathonStage_HACKATHON_STAGE_JUDGING
	case "finished":
		return hackathonv1.HackathonStage_HACKATHON_STAGE_FINISHED
	default:
		return hackathonv1.HackathonStage_HACKATHON_STAGE_UNSPECIFIED
	}
}

func stateToProto(state string) hackathonv1.HackathonState {
	switch state {
	case "draft":
		return hackathonv1.HackathonState_HACKATHON_STATE_DRAFT
	case "published":
		return hackathonv1.HackathonState_HACKATHON_STATE_PUBLISHED
	default:
		return hackathonv1.HackathonState_HACKATHON_STATE_UNSPECIFIED
	}
}
