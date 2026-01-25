package mappers

import (
	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HackathonConversionOptions struct {
	IncludeDescription bool
	IncludeLimits      bool
}

func HackathonToProto(h *entity.Hackathon, opts HackathonConversionOptions) *hackathonv1.Hackathon {
	proto := &hackathonv1.Hackathon{
		HackathonId:      h.ID.String(),
		Name:             h.Name,
		ShortDescription: h.ShortDescription,
		Stage:            MapStageToProto(h.Stage),
		State:            MapStateToProto(h.State),
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

	return proto
}

func MapStageToProto(stage string) hackathonv1.HackathonStage {
	switch stage {
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

func MapStateToProto(state string) hackathonv1.HackathonState {
	switch state {
	case "draft":
		return hackathonv1.HackathonState_HACKATHON_STATE_DRAFT
	case "published":
		return hackathonv1.HackathonState_HACKATHON_STATE_PUBLISHED
	default:
		return hackathonv1.HackathonState_HACKATHON_STATE_UNSPECIFIED
	}
}
