package me

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/s3"
)

type Service struct {
	userRepo         UserRepository
	skillRepo        SkillRepository
	contactRepo      ContactRepository
	visibilityRepo   VisibilityRepository
	avatarUploadRepo AvatarUploadRepository
	s3Client         *s3.Client
	uow              UnitOfWork
}

func NewService(
	userRepo UserRepository,
	skillRepo SkillRepository,
	contactRepo ContactRepository,
	visibilityRepo VisibilityRepository,
	avatarUploadRepo AvatarUploadRepository,
	s3Client *s3.Client,
	uow UnitOfWork,
) *Service {
	return &Service{
		userRepo:         userRepo,
		skillRepo:        skillRepo,
		contactRepo:      contactRepo,
		visibilityRepo:   visibilityRepo,
		avatarUploadRepo: avatarUploadRepo,
		s3Client:         s3Client,
		uow:              uow,
	}
}
