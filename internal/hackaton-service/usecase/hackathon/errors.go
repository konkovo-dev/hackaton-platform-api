package hackathon

import (
	"errors"
)

var (
	ErrHackathonNotFound = errors.New("hackathon not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidInput      = errors.New("invalid input")

	ErrEmptyName             = errors.New("name is required")
	ErrEmptyShortDescription = errors.New("short description is required")

	ErrInvalidLocation = errors.New("location is required: either online=true or city+country must be specified")

	ErrMissingStartsAt             = errors.New("starts_at is required")
	ErrMissingEndsAt               = errors.New("ends_at is required")
	ErrMissingRegistrationOpensAt  = errors.New("registration_opens_at is required")
	ErrMissingRegistrationClosesAt = errors.New("registration_closes_at is required")
	ErrMissingSubmissionsOpensAt   = errors.New("submissions_opens_at is required")
	ErrMissingSubmissionsClosesAt  = errors.New("submissions_closes_at is required")
	ErrMissingJudgingEndsAt        = errors.New("judging_ends_at is required")

	ErrInvalidDateSequence = errors.New("invalid date sequence: registration_opens_at < registration_closes_at < starts_at < submissions_opens_at < ends_at <= submissions_closes_at < judging_ends_at")

	ErrInvalidTeamSizeMax        = errors.New("team_size_max must be greater than 0")
	ErrInvalidRegistrationPolicy = errors.New("at least one of allow_individual or allow_team must be true")
	ErrInvalidLink               = errors.New("invalid link: title and valid URL required")
)
