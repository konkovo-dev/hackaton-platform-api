package auth

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

type CredentialsRepository interface {
	Create(ctx context.Context, creds *entity.Credentials) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Credentials, error)
	Update(ctx context.Context, creds *entity.Credentials) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string, revokedAt time.Time) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
}

type PasswordService interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

type JWTService interface {
	Sign(userID uuid.UUID, ttl time.Duration) (token string, expiresAt time.Time, err error)
	Verify(token string) (userID uuid.UUID, expiresAt time.Time, err error)
}
