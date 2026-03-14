package entity

import (
	"time"

	"github.com/google/uuid"
)

type Evaluation struct {
	ID           uuid.UUID
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
	JudgeUserID  uuid.UUID
	Score        int32
	Comment      string
	EvaluatedAt  time.Time
	UpdatedAt    time.Time
}
