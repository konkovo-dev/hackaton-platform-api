package me

import (
	"context"
	"fmt"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/google/uuid"
)

type CreateMeIn struct {
	UserID    uuid.UUID
	Username  string
	FirstName string
	LastName  string
	Timezone  string
}

type CreateMeOut struct {
	User *entity.User
}

func (s *Service) CreateMe(ctx context.Context, in CreateMeIn) (*CreateMeOut, error) {
	if err := s.validateCreateMeIn(in); err != nil {
		return nil, err
	}

	existingUser, err := s.userRepo.GetByID(ctx, in.UserID)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	user := &entity.User{
		ID:        in.UserID,
		Username:  strings.ToLower(in.Username),
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Timezone:  in.Timezone,
	}

	err = s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
		if err := txRepos.Users.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		visibility := entity.DefaultVisibility(user.ID)
		if err := txRepos.Visibility.Create(ctx, visibility); err != nil {
			return fmt.Errorf("failed to create default user's visibility: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &CreateMeOut{User: user}, nil
}

func (s *Service) validateCreateMeIn(in CreateMeIn) error {
	if in.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	if in.Username == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidInput)
	}

	if in.FirstName == "" {
		return fmt.Errorf("%w: first_name is required", ErrInvalidInput)
	}

	if in.LastName == "" {
		return fmt.Errorf("%w: last_name is required", ErrInvalidInput)
	}

	if in.Timezone == "" {
		return fmt.Errorf("%w: timezone is required", ErrInvalidInput)
	}

	return nil
}
