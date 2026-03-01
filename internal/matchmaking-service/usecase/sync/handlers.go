package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain/entity"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid uuid %s: %w", id, err)
		}
		result = append(result, parsed)
	}
	return result, nil
}

type UserSkillsUpdatedPayload struct {
	UserID           string   `json:"user_id"`
	Username         string   `json:"username"`
	AvatarURL        string   `json:"avatar_url"`
	CatalogSkillIDs  []string `json:"catalog_skill_ids"`
	CustomSkillNames []string `json:"custom_skill_names"`
	UpdatedAt        string   `json:"updated_at"`
}

type TeamCreatedPayload struct {
	TeamID      string `json:"team_id"`
	HackathonID string `json:"hackathon_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsJoinable  bool   `json:"is_joinable"`
	CreatedAt   string `json:"created_at"`
}

type TeamUpdatedPayload struct {
	TeamID      string `json:"team_id"`
	HackathonID string `json:"hackathon_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsJoinable  bool   `json:"is_joinable"`
	UpdatedAt   string `json:"updated_at"`
}

type TeamDeletedPayload struct {
	TeamID      string `json:"team_id"`
	HackathonID string `json:"hackathon_id"`
	DeletedAt   string `json:"deleted_at"`
}

type VacancyCreatedPayload struct {
	VacancyID       string   `json:"vacancy_id"`
	TeamID          string   `json:"team_id"`
	HackathonID     string   `json:"hackathon_id"`
	Description     string   `json:"description"`
	DesiredRoleIDs  []string `json:"desired_role_ids"`
	DesiredSkillIDs []string `json:"desired_skill_ids"`
	SlotsOpen       int32    `json:"slots_open"`
	CreatedAt       string   `json:"created_at"`
}

type VacancyUpdatedPayload struct {
	VacancyID       string   `json:"vacancy_id"`
	TeamID          string   `json:"team_id"`
	HackathonID     string   `json:"hackathon_id"`
	Description     string   `json:"description"`
	DesiredRoleIDs  []string `json:"desired_role_ids"`
	DesiredSkillIDs []string `json:"desired_skill_ids"`
	SlotsOpen       int32    `json:"slots_open"`
	UpdatedAt       string   `json:"updated_at"`
}

type ParticipationRegisteredPayload struct {
	HackathonID    string   `json:"hackathon_id"`
	UserID         string   `json:"user_id"`
	Status         string   `json:"status"`
	WishedRoleIDs  []string `json:"wished_role_ids"`
	MotivationText string   `json:"motivation_text"`
	RegisteredAt   string   `json:"registered_at"`
}

type ParticipationUpdatedPayload struct {
	HackathonID    string   `json:"hackathon_id"`
	UserID         string   `json:"user_id"`
	Status         string   `json:"status"`
	WishedRoleIDs  []string `json:"wished_role_ids"`
	MotivationText string   `json:"motivation_text"`
	UpdatedAt      string   `json:"updated_at"`
}

type ParticipationStatusChangedPayload struct {
	HackathonID string `json:"hackathon_id"`
	UserID      string `json:"user_id"`
	OldStatus   string `json:"old_status"`
	NewStatus   string `json:"new_status"`
	ChangedAt   string `json:"changed_at"`
}

type ParticipationTeamAssignedPayload struct {
	HackathonID string `json:"hackathon_id"`
	UserID      string `json:"user_id"`
	TeamID      string `json:"team_id"`
	IsCaptain   bool   `json:"is_captain"`
	AssignedAt  string `json:"assigned_at"`
}

type ParticipationTeamRemovedPayload struct {
	HackathonID string `json:"hackathon_id"`
	UserID      string `json:"user_id"`
	TeamID      string `json:"team_id"`
	RemovedAt   string `json:"removed_at"`
}

type UserRepository interface {
	Upsert(ctx context.Context, user *entity.User) error
}

type ParticipationRepository interface {
	Upsert(ctx context.Context, participation *entity.Participation) error
	Delete(ctx context.Context, hackathonID, userID uuid.UUID) error
}

type TeamRepository interface {
	Upsert(ctx context.Context, team *entity.Team) error
	Delete(ctx context.Context, teamID uuid.UUID) error
}

type VacancyRepository interface {
	Upsert(ctx context.Context, vacancy *entity.Vacancy) error
	Delete(ctx context.Context, vacancyID uuid.UUID) error
	DeleteByTeamID(ctx context.Context, teamID uuid.UUID) error
}

type UserSkillsUpdatedHandler struct {
	userRepo UserRepository
	logger   *slog.Logger
}

func NewUserSkillsUpdatedHandler(
	userRepo UserRepository,
	logger *slog.Logger,
) *UserSkillsUpdatedHandler {
	return &UserSkillsUpdatedHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (h *UserSkillsUpdatedHandler) Subject() string {
	return "identity.user.skills.updated.>"
}

func (h *UserSkillsUpdatedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload UserSkillsUpdatedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, payload.UpdatedAt)
	if err != nil {
		return fmt.Errorf("invalid updated_at: %w", err)
	}

	catalogSkillIDs, err := parseUUIDs(payload.CatalogSkillIDs)
	if err != nil {
		return fmt.Errorf("invalid catalog_skill_ids: %w", err)
	}

	user := &entity.User{
		UserID:           userID,
		Username:         payload.Username,
		AvatarURL:        payload.AvatarURL,
		CatalogSkillIDs:  catalogSkillIDs,
		CustomSkillNames: payload.CustomSkillNames,
		UpdatedAt:        updatedAt,
	}

	if err := h.userRepo.Upsert(ctx, user); err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	h.logger.InfoContext(ctx, "synced user skills",
		"user_id", payload.UserID,
	)

	return nil
}

type TeamCreatedHandler struct {
	teamRepo TeamRepository
	logger   *slog.Logger
}

func NewTeamCreatedHandler(
	teamRepo TeamRepository,
	logger *slog.Logger,
) *TeamCreatedHandler {
	return &TeamCreatedHandler{
		teamRepo: teamRepo,
		logger:   logger,
	}
}

func (h *TeamCreatedHandler) Subject() string {
	return "team.created.>"
}

func (h *TeamCreatedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload TeamCreatedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	teamID, err := uuid.Parse(payload.TeamID)
	if err != nil {
		return fmt.Errorf("invalid team_id: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, payload.CreatedAt)
	if err != nil {
		return fmt.Errorf("invalid created_at: %w", err)
	}

	team := &entity.Team{
		TeamID:      teamID,
		HackathonID: hackathonID,
		Name:        payload.Name,
		Description: payload.Description,
		IsJoinable:  payload.IsJoinable,
		UpdatedAt:   createdAt,
	}

	if err := h.teamRepo.Upsert(ctx, team); err != nil {
		return fmt.Errorf("failed to upsert team: %w", err)
	}

	h.logger.InfoContext(ctx, "synced team created",
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type TeamUpdatedHandler struct {
	teamRepo TeamRepository
	logger   *slog.Logger
}

func NewTeamUpdatedHandler(
	teamRepo TeamRepository,
	logger *slog.Logger,
) *TeamUpdatedHandler {
	return &TeamUpdatedHandler{
		teamRepo: teamRepo,
		logger:   logger,
	}
}

func (h *TeamUpdatedHandler) Subject() string {
	return "team.updated.>"
}

func (h *TeamUpdatedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload TeamUpdatedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	teamID, err := uuid.Parse(payload.TeamID)
	if err != nil {
		return fmt.Errorf("invalid team_id: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, payload.UpdatedAt)
	if err != nil {
		return fmt.Errorf("invalid updated_at: %w", err)
	}

	team := &entity.Team{
		TeamID:      teamID,
		HackathonID: hackathonID,
		Name:        payload.Name,
		Description: payload.Description,
		IsJoinable:  payload.IsJoinable,
		UpdatedAt:   updatedAt,
	}

	if err := h.teamRepo.Upsert(ctx, team); err != nil {
		return fmt.Errorf("failed to upsert team: %w", err)
	}

	h.logger.InfoContext(ctx, "synced team updated",
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type TeamDeletedHandler struct {
	teamRepo    TeamRepository
	vacancyRepo VacancyRepository
	logger      *slog.Logger
}

func NewTeamDeletedHandler(
	teamRepo TeamRepository,
	vacancyRepo VacancyRepository,
	logger *slog.Logger,
) *TeamDeletedHandler {
	return &TeamDeletedHandler{
		teamRepo:    teamRepo,
		vacancyRepo: vacancyRepo,
		logger:      logger,
	}
}

func (h *TeamDeletedHandler) Subject() string {
	return "team.deleted.>"
}

func (h *TeamDeletedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload TeamDeletedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	teamID, err := uuid.Parse(payload.TeamID)
	if err != nil {
		return fmt.Errorf("invalid team_id: %w", err)
	}

	if err := h.vacancyRepo.DeleteByTeamID(ctx, teamID); err != nil {
		return fmt.Errorf("failed to delete vacancies: %w", err)
	}

	if err := h.teamRepo.Delete(ctx, teamID); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	h.logger.InfoContext(ctx, "synced team deleted",
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type VacancyCreatedHandler struct {
	vacancyRepo VacancyRepository
	logger      *slog.Logger
}

func NewVacancyCreatedHandler(
	vacancyRepo VacancyRepository,
	logger *slog.Logger,
) *VacancyCreatedHandler {
	return &VacancyCreatedHandler{
		vacancyRepo: vacancyRepo,
		logger:      logger,
	}
}

func (h *VacancyCreatedHandler) Subject() string {
	return "team.vacancy.created.>"
}

func (h *VacancyCreatedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload VacancyCreatedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	vacancyID, err := uuid.Parse(payload.VacancyID)
	if err != nil {
		return fmt.Errorf("invalid vacancy_id: %w", err)
	}

	teamID, err := uuid.Parse(payload.TeamID)
	if err != nil {
		return fmt.Errorf("invalid team_id: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, payload.CreatedAt)
	if err != nil {
		return fmt.Errorf("invalid created_at: %w", err)
	}

	desiredRoleIDs, err := parseUUIDs(payload.DesiredRoleIDs)
	if err != nil {
		return fmt.Errorf("invalid desired_role_ids: %w", err)
	}

	desiredSkillIDs, err := parseUUIDs(payload.DesiredSkillIDs)
	if err != nil {
		return fmt.Errorf("invalid desired_skill_ids: %w", err)
	}

	vacancy := &entity.Vacancy{
		VacancyID:       vacancyID,
		TeamID:          teamID,
		HackathonID:     hackathonID,
		Description:     payload.Description,
		DesiredRoleIDs:  desiredRoleIDs,
		DesiredSkillIDs: desiredSkillIDs,
		SlotsOpen:       payload.SlotsOpen,
		UpdatedAt:       createdAt,
	}

	if err := h.vacancyRepo.Upsert(ctx, vacancy); err != nil {
		return fmt.Errorf("failed to upsert vacancy: %w", err)
	}

	h.logger.InfoContext(ctx, "synced vacancy created",
		"vacancy_id", payload.VacancyID,
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type VacancyUpdatedHandler struct {
	vacancyRepo VacancyRepository
	logger      *slog.Logger
}

func NewVacancyUpdatedHandler(
	vacancyRepo VacancyRepository,
	logger *slog.Logger,
) *VacancyUpdatedHandler {
	return &VacancyUpdatedHandler{
		vacancyRepo: vacancyRepo,
		logger:      logger,
	}
}

func (h *VacancyUpdatedHandler) Subject() string {
	return "team.vacancy.updated.>"
}

func (h *VacancyUpdatedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload VacancyUpdatedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	vacancyID, err := uuid.Parse(payload.VacancyID)
	if err != nil {
		return fmt.Errorf("invalid vacancy_id: %w", err)
	}

	teamID, err := uuid.Parse(payload.TeamID)
	if err != nil {
		return fmt.Errorf("invalid team_id: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, payload.UpdatedAt)
	if err != nil {
		return fmt.Errorf("invalid updated_at: %w", err)
	}

	desiredRoleIDs, err := parseUUIDs(payload.DesiredRoleIDs)
	if err != nil {
		return fmt.Errorf("invalid desired_role_ids: %w", err)
	}

	desiredSkillIDs, err := parseUUIDs(payload.DesiredSkillIDs)
	if err != nil {
		return fmt.Errorf("invalid desired_skill_ids: %w", err)
	}

	vacancy := &entity.Vacancy{
		VacancyID:       vacancyID,
		TeamID:          teamID,
		HackathonID:     hackathonID,
		Description:     payload.Description,
		DesiredRoleIDs:  desiredRoleIDs,
		DesiredSkillIDs: desiredSkillIDs,
		SlotsOpen:       payload.SlotsOpen,
		UpdatedAt:       updatedAt,
	}

	if err := h.vacancyRepo.Upsert(ctx, vacancy); err != nil {
		return fmt.Errorf("failed to upsert vacancy: %w", err)
	}

	h.logger.InfoContext(ctx, "synced vacancy updated",
		"vacancy_id", payload.VacancyID,
		"team_id", payload.TeamID,
		"hackathon_id", payload.HackathonID,
	)

	return nil
}

type ParticipationRegisteredHandler struct {
	participationRepo ParticipationRepository
	userRepo          UserRepository
	logger            *slog.Logger
}

func NewParticipationRegisteredHandler(
	participationRepo ParticipationRepository,
	userRepo UserRepository,
	logger *slog.Logger,
) *ParticipationRegisteredHandler {
	return &ParticipationRegisteredHandler{
		participationRepo: participationRepo,
		userRepo:          userRepo,
		logger:            logger,
	}
}

func (h *ParticipationRegisteredHandler) Subject() string {
	return "participation.registered.>"
}

func (h *ParticipationRegisteredHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload ParticipationRegisteredPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	registeredAt, err := time.Parse(time.RFC3339, payload.RegisteredAt)
	if err != nil {
		return fmt.Errorf("invalid registered_at: %w", err)
	}

	wishedRoleIDs, err := parseUUIDs(payload.WishedRoleIDs)
	if err != nil {
		return fmt.Errorf("invalid wished_role_ids: %w", err)
	}

	// Ensure user exists in read-model (create stub if not)
	user := &entity.User{
		UserID:           userID,
		Username:         "",
		AvatarURL:        "",
		CatalogSkillIDs:  []uuid.UUID{},
		CustomSkillNames: []string{},
		UpdatedAt:        registeredAt,
	}
	if err := h.userRepo.Upsert(ctx, user); err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	participation := &entity.Participation{
		HackathonID:    hackathonID,
		UserID:         userID,
		Status:         payload.Status,
		WishedRoleIDs:  wishedRoleIDs,
		MotivationText: payload.MotivationText,
		TeamID:         nil,
		UpdatedAt:      registeredAt,
	}

	if err := h.participationRepo.Upsert(ctx, participation); err != nil {
		return fmt.Errorf("failed to upsert participation: %w", err)
	}

	h.logger.InfoContext(ctx, "synced participation registered",
		"hackathon_id", payload.HackathonID,
		"user_id", payload.UserID,
	)

	return nil
}

type ParticipationUpdatedHandler struct {
	participationRepo ParticipationRepository
	logger            *slog.Logger
}

func NewParticipationUpdatedHandler(
	participationRepo ParticipationRepository,
	logger *slog.Logger,
) *ParticipationUpdatedHandler {
	return &ParticipationUpdatedHandler{
		participationRepo: participationRepo,
		logger:            logger,
	}
}

func (h *ParticipationUpdatedHandler) Subject() string {
	return "participation.updated.>"
}

func (h *ParticipationUpdatedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload ParticipationUpdatedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, payload.UpdatedAt)
	if err != nil {
		return fmt.Errorf("invalid updated_at: %w", err)
	}

	wishedRoleIDs, err := parseUUIDs(payload.WishedRoleIDs)
	if err != nil {
		return fmt.Errorf("invalid wished_role_ids: %w", err)
	}

	participation := &entity.Participation{
		HackathonID:    hackathonID,
		UserID:         userID,
		Status:         payload.Status,
		WishedRoleIDs:  wishedRoleIDs,
		MotivationText: payload.MotivationText,
		TeamID:         nil,
		UpdatedAt:      updatedAt,
	}

	if err := h.participationRepo.Upsert(ctx, participation); err != nil {
		return fmt.Errorf("failed to upsert participation: %w", err)
	}

	h.logger.InfoContext(ctx, "synced participation updated",
		"hackathon_id", payload.HackathonID,
		"user_id", payload.UserID,
	)

	return nil
}

type ParticipationStatusChangedHandler struct {
	participationRepo ParticipationRepository
	logger            *slog.Logger
}

func NewParticipationStatusChangedHandler(
	participationRepo ParticipationRepository,
	logger *slog.Logger,
) *ParticipationStatusChangedHandler {
	return &ParticipationStatusChangedHandler{
		participationRepo: participationRepo,
		logger:            logger,
	}
}

func (h *ParticipationStatusChangedHandler) Subject() string {
	return "participation.status_changed.>"
}

func (h *ParticipationStatusChangedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload ParticipationStatusChangedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	changedAt, err := time.Parse(time.RFC3339, payload.ChangedAt)
	if err != nil {
		return fmt.Errorf("invalid changed_at: %w", err)
	}

	if payload.NewStatus == domain.ParticipationStatusNone {
		if err := h.participationRepo.Delete(ctx, hackathonID, userID); err != nil {
			return fmt.Errorf("failed to delete participation: %w", err)
		}

		h.logger.InfoContext(ctx, "deleted participation (status changed to not_participating)",
			"hackathon_id", payload.HackathonID,
			"user_id", payload.UserID,
		)

		return nil
	}

	participation := &entity.Participation{
		HackathonID:    hackathonID,
		UserID:         userID,
		Status:         payload.NewStatus,
		WishedRoleIDs:  []uuid.UUID{},
		MotivationText: "",
		TeamID:         nil,
		UpdatedAt:      changedAt,
	}

	if err := h.participationRepo.Upsert(ctx, participation); err != nil {
		return fmt.Errorf("failed to upsert participation: %w", err)
	}

	h.logger.InfoContext(ctx, "synced participation status changed",
		"hackathon_id", payload.HackathonID,
		"user_id", payload.UserID,
		"new_status", payload.NewStatus,
	)

	return nil
}

type ParticipationTeamAssignedHandler struct {
	participationRepo ParticipationRepository
	logger            *slog.Logger
}

func NewParticipationTeamAssignedHandler(
	participationRepo ParticipationRepository,
	logger *slog.Logger,
) *ParticipationTeamAssignedHandler {
	return &ParticipationTeamAssignedHandler{
		participationRepo: participationRepo,
		logger:            logger,
	}
}

func (h *ParticipationTeamAssignedHandler) Subject() string {
	return "participation.team_assigned.>"
}

func (h *ParticipationTeamAssignedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload ParticipationTeamAssignedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	teamID, err := uuid.Parse(payload.TeamID)
	if err != nil {
		return fmt.Errorf("invalid team_id: %w", err)
	}

	assignedAt, err := time.Parse(time.RFC3339, payload.AssignedAt)
	if err != nil {
		return fmt.Errorf("invalid assigned_at: %w", err)
	}

	status := domain.ParticipationStatusTeamMember
	if payload.IsCaptain {
		status = domain.ParticipationStatusTeamCaptain
	}

	participation := &entity.Participation{
		HackathonID:    hackathonID,
		UserID:         userID,
		Status:         status,
		WishedRoleIDs:  []uuid.UUID{},
		MotivationText: "",
		TeamID:         &teamID,
		UpdatedAt:      assignedAt,
	}

	if err := h.participationRepo.Upsert(ctx, participation); err != nil {
		return fmt.Errorf("failed to upsert participation: %w", err)
	}

	h.logger.InfoContext(ctx, "synced participation team assigned",
		"hackathon_id", payload.HackathonID,
		"user_id", payload.UserID,
		"team_id", payload.TeamID,
	)

	return nil
}

type ParticipationTeamRemovedHandler struct {
	participationRepo ParticipationRepository
	logger            *slog.Logger
}

func NewParticipationTeamRemovedHandler(
	participationRepo ParticipationRepository,
	logger *slog.Logger,
) *ParticipationTeamRemovedHandler {
	return &ParticipationTeamRemovedHandler{
		participationRepo: participationRepo,
		logger:            logger,
	}
}

func (h *ParticipationTeamRemovedHandler) Subject() string {
	return "participation.team_removed.>"
}

func (h *ParticipationTeamRemovedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload ParticipationTeamRemovedPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	hackathonID, err := uuid.Parse(payload.HackathonID)
	if err != nil {
		return fmt.Errorf("invalid hackathon_id: %w", err)
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	removedAt, err := time.Parse(time.RFC3339, payload.RemovedAt)
	if err != nil {
		return fmt.Errorf("invalid removed_at: %w", err)
	}

	participation := &entity.Participation{
		HackathonID:    hackathonID,
		UserID:         userID,
		Status:         domain.ParticipationStatusLookingForTeam,
		WishedRoleIDs:  []uuid.UUID{},
		MotivationText: "",
		TeamID:         nil,
		UpdatedAt:      removedAt,
	}

	if err := h.participationRepo.Upsert(ctx, participation); err != nil {
		return fmt.Errorf("failed to upsert participation: %w", err)
	}

	h.logger.InfoContext(ctx, "synced participation team removed",
		"hackathon_id", payload.HackathonID,
		"user_id", payload.UserID,
		"team_id", payload.TeamID,
	)

	return nil
}
