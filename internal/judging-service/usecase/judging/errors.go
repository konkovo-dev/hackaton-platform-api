package judging

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
)

var (
	ErrNotAssigned        = errors.New("submission is not assigned to this judge")
	ErrInvalidScore       = errors.New("score must be between 0 and 10")
	ErrInvalidComment     = errors.New("comment is required and must be between 1 and 5000 characters")
	ErrWrongStage         = errors.New("operation not allowed in current hackathon stage")
	ErrResultNotPublished = errors.New("results have not been published yet")
	ErrAlreadyAssigned    = errors.New("submissions already assigned to judges")
	ErrNoJudges           = errors.New("no judges found for this hackathon")
	ErrNoSubmissions      = errors.New("no final submissions found for this hackathon")
)
