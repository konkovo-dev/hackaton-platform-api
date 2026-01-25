package me

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	identitypolicy "github.com/belikoooova/hackaton-platform-api/internal/identity-service/policy"
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
	updatePolicy := identitypolicy.NewUpdateMePolicy()
	pctx, err := updatePolicy.LoadContext(ctx, identitypolicy.UpdateMeParams{
		UserID: in.UserID,
	})
	if err != nil {
		return nil, err
	}

	decision := updatePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	inputDecision := updatePolicy.ValidateInput(identitypolicy.UpdateMeParams{
		UserID:    in.UserID,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Timezone:  in.Timezone,
	})
	if !inputDecision.Allowed {
		return nil, s.mapPolicyError(inputDecision)
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
