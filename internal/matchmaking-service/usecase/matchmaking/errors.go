package matchmaking

import "errors"

var (
	ErrHackathonNotFound     = errors.New("hackathon not found")
	ErrParticipationNotFound = errors.New("participation not found")
	ErrTeamNotFound          = errors.New("team not found")
	ErrVacancyNotFound       = errors.New("vacancy not found")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidHackathonStage = errors.New("invalid hackathon stage")
	ErrNotLookingForTeam     = errors.New("user is not looking for team")
	ErrNotTeamCaptain        = errors.New("user is not team captain")
)
