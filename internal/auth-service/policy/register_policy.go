package policy

import (
	"context"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

type RegisterParams struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Timezone  string
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type RegisterPolicy struct {
	policy.BasePolicy
	userRepo UserRepository
}

func NewRegisterPolicy(userRepo UserRepository) *RegisterPolicy {
	return &RegisterPolicy{
		BasePolicy: policy.NewBasePolicy(ActionRegister),
		userRepo:   userRepo,
	}
}

func (p *RegisterPolicy) LoadContext(ctx context.Context, params RegisterParams) (policy.PolicyContext, error) {
	return NewAuthPolicyContext(), nil
}

func (p *RegisterPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()

	return decision
}

func (p *RegisterPolicy) ValidateInput(ctx context.Context, params RegisterParams) *policy.Decision {
	decision := policy.NewDecision()

	if params.Username == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "username",
			Message: "username is required",
		})
	}

	if params.Email == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "email",
			Message: "email is required",
		})
	}

	if params.Password == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "password",
			Message: "password is required",
		})
	}

	if params.FirstName == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "first_name",
			Message: "first_name is required",
		})
	}

	if params.LastName == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "last_name",
			Message: "last_name is required",
		})
	}

	if params.Timezone == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "timezone",
			Message: "timezone is required",
		})
	}

	if params.Username != "" {
		username := strings.ToLower(params.Username)
		for _, r := range username {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
				decision.Deny(policy.Violation{
					Code:    policy.ViolationCodeFormat,
					Field:   "username",
					Message: "username must contain only lowercase letters, numbers, underscores, and hyphens",
				})
				break
			}
		}
	}

	if params.Password != "" && len(params.Password) < 8 {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeFormat,
			Field:   "password",
			Message: "password must be at least 8 characters long",
		})
	}

	if !decision.Allowed {
		return decision
	}

	username := strings.ToLower(params.Username)
	existingUser, err := p.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Field:   "username",
			Message: "username already exists",
		})
	}

	existingUser, err = p.userRepo.GetByEmail(ctx, params.Email)
	if err == nil && existingUser != nil {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Field:   "email",
			Message: "email already exists",
		})
	}

	return decision
}

var _ policy.Policy[RegisterParams] = (*RegisterPolicy)(nil)
