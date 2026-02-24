package entity

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID                uuid.UUID
	TicketID          uuid.UUID
	AuthorUserID      uuid.UUID
	AuthorRole        string
	Text              string
	ClientMessageID   string
	CreatedAt         time.Time
}
