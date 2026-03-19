package txrepo

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TeamRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewTeamRepository(tx pgx.Tx) *TeamRepository {
	return &TeamRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *TeamRepository) Create(ctx context.Context, team *entity.Team) error {
	row, err := r.Queries().CreateTeam(ctx, queries.CreateTeamParams{
		ID:          team.ID,
		HackathonID: team.HackathonID,
		Name:        team.Name,
		Description: team.Description,
		IsJoinable:  team.IsJoinable,
	})
	if err != nil {
		return fmt.Errorf("failed to create team: %w", pgxutil.MapDBError(err))
	}

	team.CreatedAt = row.CreatedAt
	team.UpdatedAt = row.UpdatedAt

	return nil
}

type MembershipRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewMembershipRepository(tx pgx.Tx) *MembershipRepository {
	return &MembershipRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *MembershipRepository) Create(ctx context.Context, membership *entity.Membership) error {
	var assignedVacancyID pgtype.UUID
	if membership.AssignedVacancyID != nil {
		assignedVacancyID = pgtype.UUID{
			Bytes: *membership.AssignedVacancyID,
			Valid: true,
		}
	}

	row, err := r.Queries().CreateMembership(ctx, queries.CreateMembershipParams{
		TeamID:            membership.TeamID,
		UserID:            membership.UserID,
		IsCaptain:         membership.IsCaptain,
		AssignedVacancyID: assignedVacancyID,
	})
	if err != nil {
		return fmt.Errorf("failed to create membership: %w", pgxutil.MapDBError(err))
	}

	membership.JoinedAt = row.JoinedAt

	return nil
}

func (r *MembershipRepository) Delete(ctx context.Context, teamID, userID uuid.UUID) error {
	err := r.Queries().DeleteMembership(ctx, queries.DeleteMembershipParams{
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *MembershipRepository) UpdateCaptainStatus(ctx context.Context, teamID, userID uuid.UUID, isCaptain bool) error {
	err := r.Queries().UpdateCaptainStatus(ctx, queries.UpdateCaptainStatusParams{
		TeamID:    teamID,
		UserID:    userID,
		IsCaptain: isCaptain,
	})
	if err != nil {
		return fmt.Errorf("failed to update captain status: %w", pgxutil.MapDBError(err))
	}

	return nil
}

type VacancyRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewVacancyRepository(tx pgx.Tx) *VacancyRepository {
	return &VacancyRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *VacancyRepository) Create(ctx context.Context, vacancy *entity.Vacancy) error {
	row, err := r.Queries().CreateVacancy(ctx, queries.CreateVacancyParams{
		ID:              vacancy.ID,
		TeamID:          vacancy.TeamID,
		Description:     vacancy.Description,
		DesiredRoleIds:  vacancy.DesiredRoleIDs,
		DesiredSkillIds: vacancy.DesiredSkillIDs,
		SlotsTotal:      vacancy.SlotsTotal,
		SlotsOpen:       vacancy.SlotsOpen,
		IsSystem:        vacancy.IsSystem,
		CreatedAt:       vacancy.CreatedAt,
		UpdatedAt:       vacancy.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to create vacancy: %w", pgxutil.MapDBError(err))
	}

	vacancy.CreatedAt = row.CreatedAt
	vacancy.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *VacancyRepository) IncrementSlotsOpen(ctx context.Context, vacancyID uuid.UUID) error {
	err := r.Queries().IncrementSlotsOpen(ctx, vacancyID)
	if err != nil {
		return fmt.Errorf("failed to increment slots open: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *VacancyRepository) DecrementSlotsOpen(ctx context.Context, vacancyID uuid.UUID) error {
	err := r.Queries().DecrementSlotsOpen(ctx, vacancyID)
	if err != nil {
		return fmt.Errorf("failed to decrement slots open: %w", pgxutil.MapDBError(err))
	}

	return nil
}

type TeamInvitationRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewTeamInvitationRepository(tx pgx.Tx) *TeamInvitationRepository {
	return &TeamInvitationRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *TeamInvitationRepository) UpdateStatus(ctx context.Context, invitationID uuid.UUID, status string) error {
	_, err := r.Queries().UpdateTeamInvitationStatus(ctx, queries.UpdateTeamInvitationStatusParams{
		ID:     invitationID,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to update invitation status: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *TeamInvitationRepository) CancelCompeting(ctx context.Context, userID, hackathonID uuid.UUID) error {
	err := r.Queries().CancelCompetingInvitations(ctx, queries.CancelCompetingInvitationsParams{
		TargetUserID: userID,
		HackathonID:  hackathonID,
	})
	if err != nil {
		return fmt.Errorf("failed to cancel competing invitations: %w", pgxutil.MapDBError(err))
	}

	return nil
}

type JoinRequestRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewJoinRequestRepository(tx pgx.Tx) *JoinRequestRepository {
	return &JoinRequestRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *JoinRequestRepository) UpdateStatus(ctx context.Context, requestID uuid.UUID, status string) error {
	_, err := r.Queries().UpdateJoinRequestStatus(ctx, queries.UpdateJoinRequestStatusParams{
		ID:     requestID,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to update join request status: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *JoinRequestRepository) CancelCompeting(ctx context.Context, userID, hackathonID uuid.UUID) error {
	err := r.Queries().CancelCompetingJoinRequests(ctx, queries.CancelCompetingJoinRequestsParams{
		RequesterUserID: userID,
		HackathonID:     hackathonID,
	})
	if err != nil {
		return fmt.Errorf("failed to cancel competing join requests: %w", pgxutil.MapDBError(err))
	}

	return nil
}

type OutboxRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewOutboxRepository(tx pgx.Tx) *OutboxRepository {
	return &OutboxRepository{
		BaseRepository: pgxutil.NewBaseRepository[*queries.Queries, queries.DBTX](tx, queries.New),
	}
}

func (r *OutboxRepository) Create(ctx context.Context, event *outbox.Event) error {
	err := r.Queries().CreateOutboxEvent(ctx, queries.CreateOutboxEventParams{
		ID:            event.ID,
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		EventType:     event.EventType,
		Payload:       event.Payload,
		Status:        string(event.Status),
		AttemptCount:  int32(event.AttemptCount),
		LastError:     event.LastError,
		CreatedAt:     event.CreatedAt,
		UpdatedAt:     event.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", pgxutil.MapDBError(err))
	}
	return nil
}
