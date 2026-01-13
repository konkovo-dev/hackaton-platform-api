package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	cfg              *Config
	txManager        TxManager
	userRepo         UserRepository
	credentialsRepo  CredentialsRepository
	refreshTokenRepo RefreshTokenRepository
	passwordService  PasswordService
	jwtService       JWTService
	outboxRepo       OutboxRepository
}

func NewService(
	cfg *Config,
	txManager TxManager,
	userRepo UserRepository,
	credentialsRepo CredentialsRepository,
	refreshTokenRepo RefreshTokenRepository,
	passwordService PasswordService,
	jwtService JWTService,
	outboxRepo OutboxRepository,
) *Service {
	return &Service{
		cfg:              cfg,
		txManager:        txManager,
		userRepo:         userRepo,
		credentialsRepo:  credentialsRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordService:  passwordService,
		jwtService:       jwtService,
		outboxRepo:       outboxRepo,
	}
}

type RegisterInput struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Timezone  string
}

type RegisterOutput struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if err := s.validateRegisterInput(input); err != nil {
		return nil, err
	}

	username := strings.ToLower(input.Username)
	existingUser, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	existingUser, err = s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHash, err := s.passwordService.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		ID:       uuid.New(),
		Username: username,
		Email:    input.Email,
	}
	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		userRepoTx := s.userRepo.WithTx(tx)
		credRepoTx := s.credentialsRepo.WithTx(tx)
		outboxRepoTx := s.outboxRepo.WithTx(tx)

		if err := userRepoTx.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		credentials := &entity.Credentials{
			UserID:       user.ID,
			PasswordHash: passwordHash,
		}

		if err := credRepoTx.Create(ctx, credentials); err != nil {
			return fmt.Errorf("failed to create credentials: %w", err)
		}

		payload := map[string]string{
			"user_id":    user.ID.String(),
			"username":   user.Username,
			"first_name": input.FirstName,
			"last_name":  input.LastName,
			"timezone":   input.Timezone,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal outbox payload: %w", err)
		}

		event := outbox.NewEvent(user.ID.String(), "user", "user.registered", payloadBytes)

		if err := outboxRepoTx.Create(ctx, event); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) IntrospectToken(ctx context.Context, accessToken string) (active bool, userID uuid.UUID, expiresAt time.Time, err error) {
	if accessToken == "" {
		return false, uuid.Nil, time.Time{}, ErrTokenInvalid
	}

	userID, expiresAt, err = s.jwtService.Verify(accessToken)
	if err != nil {
		return false, uuid.Nil, time.Time{}, ErrTokenInvalid
	}

	if time.Now().UTC().After(expiresAt) {
		return false, userID, expiresAt, ErrTokenExpired
	}

	return true, userID, expiresAt, nil
}

func (s *Service) Login(ctx context.Context, login, password string) (*RegisterOutput, error) {
	if login == "" {
		return nil, ErrEmptyLogin
	}

	if password == "" {
		return nil, ErrEmptyPassword
	}

	var user *entity.User
	var err error

	if strings.Contains(login, "@") {
		user, err = s.userRepo.GetByEmail(ctx, login)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, strings.ToLower(login))
	}

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	credentials, err := s.credentialsRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := s.passwordService.Verify(password, credentials.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*RegisterOutput, error) {
	if refreshToken == "" {
		return nil, ErrTokenInvalid
	}

	tokenHash := s.hashRefreshToken(refreshToken)

	storedToken, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	if storedToken.RevokedAt != nil {
		return nil, ErrTokenRevoked
	}
	if time.Now().UTC().After(storedToken.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.refreshTokenRepo.Revoke(ctx, tokenHash, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrTokenInvalid
	}

	tokenHash := s.hashRefreshToken(refreshToken)

	if err := s.refreshTokenRepo.Revoke(ctx, tokenHash, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (s *Service) validateRegisterInput(input RegisterInput) error {
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

func (s *Service) generateTokens(ctx context.Context, userID uuid.UUID) (*RegisterOutput, error) {
	accessToken, accessExpiresAt, err := s.jwtService.Sign(userID, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	tokenHash := s.hashRefreshToken(refreshToken)

	refreshExpiresAt := time.Now().UTC().Add(s.cfg.RefreshTokenTTL)
	refreshTokenEntity := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: refreshExpiresAt,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &RegisterOutput{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *Service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *Service) hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
