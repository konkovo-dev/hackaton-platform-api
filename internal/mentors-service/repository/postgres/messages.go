package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type MessageRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewMessageRepository(db queries.DBTX) *MessageRepository {
	return &MessageRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *MessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	row, err := r.Queries().GetMessageByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get message by id: %w", err)
	}

	return mapMessageToEntity(row), nil
}

func (r *MessageRepository) ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]*entity.Message, error) {
	rows, err := r.Queries().ListMessagesByTicket(ctx, queries.ListMessagesByTicketParams{
		TicketID: ticketID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list messages by ticket: %w", err)
	}

	messages := make([]*entity.Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, mapMessageToEntity(row))
	}

	return messages, nil
}

func (r *MessageRepository) ListByOwner(ctx context.Context, hackathonID uuid.UUID, ownerKind string, ownerID uuid.UUID, limit, offset int32) ([]*entity.Message, error) {
	rows, err := r.Queries().ListMessagesByOwner(ctx, queries.ListMessagesByOwnerParams{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to list messages by owner: %w", err)
	}

	messages := make([]*entity.Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, mapMessageToEntity(row))
	}

	return messages, nil
}

func (r *MessageRepository) Create(ctx context.Context, message *entity.Message) error {
	var clientMessageID pgtype.Text
	if message.ClientMessageID != "" {
		clientMessageID = pgtype.Text{
			String: message.ClientMessageID,
			Valid:  true,
		}
	}

	row, err := r.Queries().CreateMessage(ctx, queries.CreateMessageParams{
		ID:              message.ID,
		TicketID:        message.TicketID,
		AuthorUserID:    message.AuthorUserID,
		AuthorRole:      message.AuthorRole,
		Text:            message.Text,
		ClientMessageID: clientMessageID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create message: %w", err)
	}

	message.CreatedAt = row.CreatedAt

	return nil
}

func (r *MessageRepository) FindByClientMessageID(ctx context.Context, clientMessageID string) (*entity.Message, error) {
	row, err := r.Queries().FindMessageByClientID(ctx, pgtype.Text{
		String: clientMessageID,
		Valid:  true,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to find message by client id: %w", err)
	}

	return mapMessageToEntity(row), nil
}

func mapMessageToEntity(m queries.MentorsMessage) *entity.Message {
	var clientMsgIDStr string
	if m.ClientMessageID.Valid {
		clientMsgIDStr = m.ClientMessageID.String
	}

	return &entity.Message{
		ID:              m.ID,
		TicketID:        m.TicketID,
		AuthorUserID:    m.AuthorUserID,
		AuthorRole:      m.AuthorRole,
		Text:            m.Text,
		ClientMessageID: clientMsgIDStr,
		CreatedAt:       m.CreatedAt,
	}
}
