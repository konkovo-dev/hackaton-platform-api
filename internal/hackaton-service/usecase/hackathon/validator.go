package hackathon

import (
	"net/url"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
)

type HackathonValidator struct{}

func NewHackathonValidator() *HackathonValidator {
	return &HackathonValidator{}
}

func (v *HackathonValidator) ValidateForPublish(h *entity.Hackathon, links []CreateHackathonLink) []domain.ValidationError {
	var errors []domain.ValidationError

	if h.Name == "" {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeRequired,
			Field:   "name",
			Message: "name is required for publishing",
		})
	}

	if !h.LocationOnline && h.LocationCity == "" && h.LocationCountry == "" {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeRequired,
			Field:   "location",
			Message: "location is required for publishing",
		})
	}

	if h.RegistrationOpensAt == nil {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeRequired,
			Field:   "registration_opens_at",
			Message: "registration_opens_at is required",
		})
	}

	if h.RegistrationClosesAt == nil {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeRequired,
			Field:   "registration_closes_at",
			Message: "registration_closes_at is required",
		})
	}

	if h.StartsAt == nil {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeRequired,
			Field:   "starts_at",
			Message: "starts_at is required",
		})
	}

	if h.EndsAt == nil {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeRequired,
			Field:   "ends_at",
			Message: "ends_at is required",
		})
	}

	if h.JudgingEndsAt == nil {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeRequired,
			Field:   "judging_ends_at",
			Message: "judging_ends_at is required",
		})
	}

	if h.Task == "" {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeRequired,
			Field:   "task",
			Message: "task is required for publishing",
		})
	}

	if h.RegistrationOpensAt != nil && h.RegistrationClosesAt != nil && h.StartsAt != nil && h.EndsAt != nil && h.JudgingEndsAt != nil {
		timeErrors := v.ValidateTimeRule(h)
		errors = append(errors, timeErrors...)
	}

	if !h.AllowIndividual && !h.AllowTeam {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodePolicyRule,
			Field:   "registration_policy",
			Message: "at least one registration type must be allowed",
		})
	}

	for i, link := range links {
		if link.Title == "" {
			errors = append(errors, domain.ValidationError{
				Code:    domain.ValidationCodeRequired,
				Field:   "links",
				Message: "link title is required",
				Meta:    map[string]string{"index": string(rune(i))},
			})
		}
		if _, err := url.ParseRequestURI(link.URL); err != nil {
			errors = append(errors, domain.ValidationError{
				Code:    domain.ValidationCodeFormat,
				Field:   "links",
				Message: "invalid link URL format",
				Meta:    map[string]string{"index": string(rune(i))},
			})
		}
	}

	return errors
}

func (v *HackathonValidator) ValidateTimeRule(h *entity.Hackathon) []domain.ValidationError {
	var errors []domain.ValidationError

	if h.RegistrationOpensAt == nil || h.RegistrationClosesAt == nil || h.StartsAt == nil || h.EndsAt == nil || h.JudgingEndsAt == nil {
		return errors
	}

	oneHour := time.Hour

	if !h.RegistrationOpensAt.Add(oneHour).Before(*h.RegistrationClosesAt) {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeTimeRule,
			Field:   "dates",
			Message: "registration_opens_at must be at least 1 hour before registration_closes_at",
		})
	}

	if !h.RegistrationClosesAt.Before(*h.StartsAt) && !h.RegistrationClosesAt.Equal(*h.StartsAt) {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeTimeRule,
			Field:   "dates",
			Message: "registration_closes_at must be before or equal to starts_at",
		})
	}

	if !h.StartsAt.Before(*h.EndsAt) {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeTimeRule,
			Field:   "dates",
			Message: "starts_at must be before ends_at",
		})
	}

	if !h.EndsAt.Add(oneHour).Before(*h.JudgingEndsAt) {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodeTimeRule,
			Field:   "dates",
			Message: "ends_at must be at least 1 hour before judging_ends_at",
		})
	}

	return errors
}

func (v *HackathonValidator) ValidateUpdate(old, new *entity.Hackathon, newLinks []CreateHackathonLink, now time.Time) []domain.ValidationError {
	var errors []domain.ValidationError

	stage := domain.HackathonStage(old.Stage)

	if stage != domain.StageDraft {
		if new.Name == "" {
			errors = append(errors, domain.ValidationError{
				Code:    domain.ValidationCodeRequired,
				Field:   "name",
				Message: "name is required for published hackathon",
			})
		}

		if !new.LocationOnline && new.LocationCity == "" && new.LocationCountry == "" {
			errors = append(errors, domain.ValidationError{
				Code:    domain.ValidationCodeRequired,
				Field:   "location",
				Message: "location is required for published hackathon",
			})
		}

		if stage == domain.StageRunning || stage == domain.StageJudging || stage == domain.StageFinished {
			if old.LocationCity != new.LocationCity || old.LocationCountry != new.LocationCountry || old.LocationOnline != new.LocationOnline {
				errors = append(errors, domain.ValidationError{
					Code:    domain.ValidationCodeTimeLocked,
					Field:   "location",
					Message: "location cannot be changed after PRESTART",
				})
			}
		}
	}

	if stage == domain.StageRunning || stage == domain.StageJudging || stage == domain.StageFinished {
		if old.TeamSizeMax != new.TeamSizeMax {
			errors = append(errors, domain.ValidationError{
				Code:    domain.ValidationCodeTimeLocked,
				Field:   "team_size_max",
				Message: "team_size_max can only be changed in DRAFT or UPCOMING",
			})
		}
	}

	if old.AllowIndividual != new.AllowIndividual || old.AllowTeam != new.AllowTeam {
		if old.AllowIndividual && !new.AllowIndividual {
			if stage != domain.StageDraft && stage != domain.StageUpcoming {
				errors = append(errors, domain.ValidationError{
					Code:    domain.ValidationCodeTimeLocked,
					Field:   "registration_policy",
					Message: "disabling registration type allowed only in DRAFT or UPCOMING",
				})
			}
		}

		if !old.AllowIndividual && new.AllowIndividual {
			if stage != domain.StageDraft {
				errors = append(errors, domain.ValidationError{
					Code:    domain.ValidationCodeTimeLocked,
					Field:   "registration_policy",
					Message: "enabling registration type allowed only in DRAFT",
				})
			}
		}

		if old.AllowTeam && !new.AllowTeam {
			if stage != domain.StageDraft && stage != domain.StageUpcoming {
				errors = append(errors, domain.ValidationError{
					Code:    domain.ValidationCodeTimeLocked,
					Field:   "registration_policy",
					Message: "disabling registration type allowed only in DRAFT or UPCOMING",
				})
			}
		}

		if !old.AllowTeam && new.AllowTeam {
			if stage != domain.StageDraft {
				errors = append(errors, domain.ValidationError{
					Code:    domain.ValidationCodeTimeLocked,
					Field:   "registration_policy",
					Message: "enabling registration type allowed only in DRAFT",
				})
			}
		}
	}

	if !new.AllowIndividual && !new.AllowTeam {
		errors = append(errors, domain.ValidationError{
			Code:    domain.ValidationCodePolicyRule,
			Field:   "registration_policy",
			Message: "at least one registration type must be allowed",
		})
	}

	if new.RegistrationOpensAt != nil && new.RegistrationClosesAt != nil && new.StartsAt != nil && new.EndsAt != nil && new.JudgingEndsAt != nil {
		timeErrors := v.ValidateTimeRule(new)
		errors = append(errors, timeErrors...)

		if stage != domain.StageDraft {
			if old.RegistrationOpensAt != nil && new.RegistrationOpensAt != nil {
				if !old.RegistrationOpensAt.Equal(*new.RegistrationOpensAt) {
					if !now.Before(*old.RegistrationOpensAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeLocked,
							Field:   "registration_opens_at",
							Message: "registration_opens_at can only be changed before it occurs",
						})
					}
					if !now.Before(*new.RegistrationOpensAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeLocked,
							Field:   "registration_opens_at",
							Message: "registration_opens_at cannot be set to past",
						})
					}
				}
			}

			if old.RegistrationClosesAt != nil && new.RegistrationClosesAt != nil {
				if !old.RegistrationClosesAt.Equal(*new.RegistrationClosesAt) {
					if !now.Before(*old.RegistrationClosesAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeLocked,
							Field:   "registration_closes_at",
							Message: "registration_closes_at can only be changed before it occurs",
						})
					}
					if !old.RegistrationClosesAt.Before(*new.RegistrationClosesAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeRule,
							Field:   "registration_closes_at",
							Message: "registration_closes_at can only be extended forward",
						})
					}
				}
			}

			if old.StartsAt != nil && new.StartsAt != nil {
				if !old.StartsAt.Equal(*new.StartsAt) {
					if !now.Before(*old.StartsAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeLocked,
							Field:   "starts_at",
							Message: "starts_at can only be changed before it occurs",
						})
					}
					if !old.StartsAt.Before(*new.StartsAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeRule,
							Field:   "starts_at",
							Message: "starts_at can only be extended forward",
						})
					}
				}
			}

			if old.EndsAt != nil && new.EndsAt != nil {
				if !old.EndsAt.Equal(*new.EndsAt) {
					if !now.Before(*old.EndsAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeLocked,
							Field:   "ends_at",
							Message: "ends_at can only be changed before it occurs",
						})
					}
					if !old.EndsAt.Before(*new.EndsAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeRule,
							Field:   "ends_at",
							Message: "ends_at can only be extended forward",
						})
					}
				}
			}

			if old.JudgingEndsAt != nil && new.JudgingEndsAt != nil {
				if !old.JudgingEndsAt.Equal(*new.JudgingEndsAt) {
					if !now.Before(*old.JudgingEndsAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeLocked,
							Field:   "judging_ends_at",
							Message: "judging_ends_at can only be changed before it occurs",
						})
					}
					if !now.Before(*new.JudgingEndsAt) {
						errors = append(errors, domain.ValidationError{
							Code:    domain.ValidationCodeTimeLocked,
							Field:   "judging_ends_at",
							Message: "judging_ends_at cannot be set to past",
						})
					}
				}
			}
		}
	}

	for i, link := range newLinks {
		if link.Title == "" {
			errors = append(errors, domain.ValidationError{
				Code:    domain.ValidationCodeRequired,
				Field:   "links",
				Message: "link title is required",
				Meta:    map[string]string{"index": string(rune(i))},
			})
		}
		if _, err := url.ParseRequestURI(link.URL); err != nil {
			errors = append(errors, domain.ValidationError{
				Code:    domain.ValidationCodeFormat,
				Field:   "links",
				Message: "invalid link URL format",
				Meta:    map[string]string{"index": string(rune(i))},
			})
		}
	}

	return errors
}
