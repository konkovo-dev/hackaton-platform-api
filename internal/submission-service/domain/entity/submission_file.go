package entity

import (
	"time"

	"github.com/google/uuid"
)

type SubmissionFile struct {
	ID           uuid.UUID
	SubmissionID uuid.UUID
	Filename     string
	SizeBytes    int64
	ContentType  string
	StorageKey   string
	UploadStatus string
	CreatedAt    time.Time
	CompletedAt  *time.Time
}
