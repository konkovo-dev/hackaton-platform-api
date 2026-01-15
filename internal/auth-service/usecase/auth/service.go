package auth

type Service struct {
	cfg              *Config
	uow              UnitOfWork
	userRepo         UserRepository
	credentialsRepo  CredentialsRepository
	refreshTokenRepo RefreshTokenRepository
	passwordService  PasswordService
	jwtService       JWTService
}

func NewService(
	cfg *Config,
	uow UnitOfWork,
	userRepo UserRepository,
	credentialsRepo CredentialsRepository,
	refreshTokenRepo RefreshTokenRepository,
	passwordService PasswordService,
	jwtService JWTService,
) *Service {
	return &Service{
		cfg:              cfg,
		uow:              uow,
		userRepo:         userRepo,
		credentialsRepo:  credentialsRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordService:  passwordService,
		jwtService:       jwtService,
	}
}
