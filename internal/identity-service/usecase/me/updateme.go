package me

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/google/uuid"
)

type UpdateMeIn struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
	AvatarURL string
	Timezone  string
}

type UpdateMeOut struct {
	User *entity.User
}

func (s *Service) UpdateMe(ctx context.Context, in UpdateMeIn) (*UpdateMeOut, error) {
	if err := s.validateUpdateMeIn(in); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user.FirstName = in.FirstName
	user.LastName = in.LastName
	user.AvatarURL = in.AvatarURL
	user.Timezone = in.Timezone

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &UpdateMeOut{User: user}, nil
}

func (s *Service) validateUpdateMeIn(in UpdateMeIn) error {
	if in.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
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
