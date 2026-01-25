package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	authpolicy "github.com/belikoooova/hackaton-platform-api/internal/auth-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
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
	// Validate input using policy
	registerPolicy := authpolicy.NewRegisterPolicy(s.userRepo)
	decision := registerPolicy.ValidateInput(ctx, authpolicy.RegisterParams{
		Email:     in.Email,
		Username:  in.Username,
		Password:  in.Password,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Timezone:  in.Timezone,
	})

	if !decision.Allowed {
		return nil, s.mapPolicyDecisionToError(decision)
	}

	// Normalize username to lowercase
	username := strings.ToLower(in.Username)

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

// mapPolicyDecisionToError maps policy violations to legacy errors for backward compatibility
func (s *Service) mapPolicyDecisionToError(decision *policy.Decision) error {
	if len(decision.Violations) == 0 {
		return ErrInvalidCredentials
	}

	// Map first violation to legacy error
	v := decision.Violations[0]
	switch v.Field {
	case "username":
		if v.Code == "REQUIRED" {
			return ErrEmptyUsername
		}
		if v.Code == "FORMAT" {
			return ErrInvalidUsername
		}
		if v.Code == "CONFLICT" {
			return ErrUserAlreadyExists
		}
	case "email":
		if v.Code == "REQUIRED" {
			return ErrEmptyEmail
		}
		if v.Code == "CONFLICT" {
			return ErrUserAlreadyExists
		}
	case "password":
		if v.Code == "REQUIRED" {
			return ErrEmptyPassword
		}
		if v.Code == "FORMAT" {
			return ErrInvalidPassword
		}
	case "first_name":
		return ErrEmptyFirstName
	case "last_name":
		return ErrEmptyLastName
	case "timezone":
		return ErrEmptyTimezone
	case "login":
		return ErrEmptyLogin
	case "refresh_token", "access_token":
		return ErrTokenInvalid
	}

	return ErrInvalidCredentials
}
