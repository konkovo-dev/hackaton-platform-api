package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type ContactRepository struct {
	queries *queries.Queries
}

func NewContactRepository(db queries.DBTX) *ContactRepository {
	return &ContactRepository{
		queries: queries.New(db),
	}
}

func (r *ContactRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Contact, error) {
	rows, err := r.queries.ContactGetByUserID(ctx, pgxutil.UUIDToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user contacts: %w", err)
	}

	contacts := make([]*entity.Contact, 0, len(rows))
	for _, row := range rows {
		contacts = append(contacts, &entity.Contact{
			ID:         pgxutil.PgtypeToUUID(row.ID),
			UserID:     pgxutil.PgtypeToUUID(row.UserID),
			Type:       row.Type,
			Value:      row.Value,
			Visibility: domain.VisibilityLevel(row.Visibility),
		})
	}

	return contacts, nil
}

func (r *ContactRepository) Update(ctx context.Context, userID uuid.UUID, contacts []*entity.Contact) error {
	if err := r.queries.ContactDeleteByUserID(ctx, pgxutil.UUIDToPgtype(userID)); err != nil {
		return fmt.Errorf("failed to delete user contacts: %w", err)
	}

	for _, contact := range contacts {
		err := r.queries.ContactCreate(ctx, queries.ContactCreateParams{
			ID:         pgxutil.UUIDToPgtype(contact.ID),
			UserID:     pgxutil.UUIDToPgtype(userID),
			Type:       contact.Type,
			Value:      contact.Value,
			Visibility: string(contact.Visibility),
		})
		if err != nil {
			return fmt.Errorf("failed to create user contact: %w", err)
		}
	}

	return nil
}
