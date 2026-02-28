package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Message struct {
	ID                uuid.UUID
	TicketID          uuid.UUID
	AuthorUserID      pgtype.UUID
	AuthorRole        string
	Text              string
	ClientMessageID   string
	CreatedAt         time.Time
}
