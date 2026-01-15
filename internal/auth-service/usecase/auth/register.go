package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type RegisterIn struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Timezone  string
}

func (s *Service) Register(ctx context.Context, in RegisterIn) (*AuthOut, error) {
	if err := s.validateRegisterIn(in); err != nil {
		return nil, err
	}

	username := strings.ToLower(in.Username)
	existingUser, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	existingUser, err = s.userRepo.GetByEmail(ctx, in.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHash, err := s.passwordService.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		ID:       uuid.New(),
		Username: username,
		Email:    in.Email,
	}
	err = s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
		if err := txRepos.Users.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		credentials := &entity.Credentials{
			UserID:       user.ID,
			PasswordHash: passwordHash,
		}

		if err := txRepos.Credentials.Create(ctx, credentials); err != nil {
			return fmt.Errorf("failed to create credentials: %w", err)
		}

		payload := map[string]string{
			"user_id":    user.ID.String(),
			"username":   user.Username,
			"first_name": in.FirstName,
			"last_name":  in.LastName,
			"timezone":   in.Timezone,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal outbox payload: %w", err)
		}

		event := outbox.NewEvent(user.ID.String(), "user", "user.registered", payloadBytes)

		if err := txRepos.Outbox.Create(ctx, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) validateRegisterIn(input RegisterIn) error {
	if input.Username == "" {
		return ErrEmptyUsername
	}

	if input.Email == "" {
		return ErrEmptyEmail
	}

	if input.Password == "" {
		return ErrEmptyPassword
	}

	if input.FirstName == "" {
		return ErrEmptyFirstName
	}

	if input.LastName == "" {
		return ErrEmptyLastName
	}

	if input.Timezone == "" {
		return ErrEmptyTimezone
	}

	username := strings.ToLower(input.Username)
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return ErrInvalidUsername
		}
	}

	if len(input.Password) < 8 {
		return ErrInvalidPassword
	}

	return nil
}
