package me

import (
	"context"
	"fmt"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/google/uuid"
)

type Service struct {
	userRepo UserRepository
}

func NewService(userRepo UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

type CreateMeInput struct {
	UserID    uuid.UUID
	Username  string
	FirstName string
	LastName  string
	Timezone  string
}

type CreateMeOutput struct {
	User *entity.User
}

func (s *Service) CreateMe(ctx context.Context, input CreateMeInput) (*CreateMeOutput, error) {
	if err := s.validateCreateMeInput(input); err != nil {
		return nil, err
	}

	existingUser, err := s.userRepo.GetByID(ctx, input.UserID)
	if err == nil && existingUser != nil {
		return &CreateMeOutput{User: existingUser}, nil
	}

	user := &entity.User{
		ID:        input.UserID,
		Username:  strings.ToLower(input.Username),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		AvatarURL: "",
		Timezone:  input.Timezone,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &CreateMeOutput{User: user}, nil
}

func (s *Service) validateCreateMeInput(input CreateMeInput) error {
	if input.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	if input.Username == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidInput)
	}

	if input.FirstName == "" {
		return fmt.Errorf("%w: first_name is required", ErrInvalidInput)
	}

	if input.LastName == "" {
		return fmt.Errorf("%w: last_name is required", ErrInvalidInput)
	}

	if input.Timezone == "" {
		return fmt.Errorf("%w: timezone is required", ErrInvalidInput)
	}

	return nil
}
