package mentors

import (
	"log/slog"

	"github.com/belikoooova/hackaton-platform-api/pkg/centrifugo"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
)

type Service struct {
	ticketRepo        TicketRepository
	messageRepo       MessageRepository
	hackathonClient   HackathonClient
	prClient          ParticipationRolesClient
	teamClient        TeamClient
	txManager         *pgxutil.TxManager
	idempotencyHelper *idempotency.Helper
	idempotencyRepo   idempotency.Repository
	outboxRepo        outbox.EventRepository
	jwtHelper         *centrifugo.JWTHelper
	logger            *slog.Logger
}

func NewService(
	ticketRepo TicketRepository,
	messageRepo MessageRepository,
	hackathonClient HackathonClient,
	prClient ParticipationRolesClient,
	teamClient TeamClient,
	txManager *pgxutil.TxManager,
	idempotencyHelper *idempotency.Helper,
	idempotencyRepo idempotency.Repository,
	outboxRepo outbox.EventRepository,
	jwtHelper *centrifugo.JWTHelper,
	logger *slog.Logger,
) *Service {
	return &Service{
		ticketRepo:        ticketRepo,
		messageRepo:       messageRepo,
		hackathonClient:   hackathonClient,
		prClient:          prClient,
		teamClient:        teamClient,
		txManager:         txManager,
		idempotencyHelper: idempotencyHelper,
		idempotencyRepo:   idempotencyRepo,
		outboxRepo:        outboxRepo,
		jwtHelper:         jwtHelper,
		logger:            logger,
	}
}
