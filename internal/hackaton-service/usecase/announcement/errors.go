package announcement

import "errors"

var (
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrHackathonNotFound    = errors.New("hackathon not found")
	ErrAnnouncementNotFound = errors.New("announcement not found")
)
