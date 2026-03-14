package entity

import (
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID           uuid.UUID
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
	JudgeUserID  uuid.UUID
	AssignedAt   time.Time
}
