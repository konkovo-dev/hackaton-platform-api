package team

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetTeamPermissionsIn struct {
	HackathonID uuid.UUID
}

type GetTeamPermissionsOut struct {
	CreateTeam bool
	CanInMyTeam struct {
		EditTeam           bool
		DeleteTeam         bool
		ManageVacancies    bool
		InviteMember       bool
		ManageJoinRequests bool
		KickMember         bool
		TransferCaptain    bool
		LeaveTeam          bool
	}
}

func (s *Service) GetTeamPermissions(ctx context.Context, in GetTeamPermissionsIn) (*GetTeamPermissionsOut, error) {
	out := &GetTeamPermissionsOut{}

	// Get user ID
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return out, nil
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return out, nil
	}

	// Get hackathon context
	stage, allowTeam, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return out, nil
	}

	// Get participation context
	parUserID, participationStatus, roles, err := s.parClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return out, nil
	}

	// Check createTeam
	createTeamPolicy := policy.NewCreateTeamPolicy()
	createTeamCtx := policy.NewCreateTeamContext()
	createTeamCtx.SetAuthenticated(parUserID != "")
	if parUserID != "" {
		uid, _ := uuid.Parse(parUserID)
		createTeamCtx.SetActorUserID(uid)
	}
	createTeamCtx.SetHackathonStage(stage)
	createTeamCtx.SetAllowTeam(allowTeam)
	createTeamCtx.SetActorRoles(roles)
	createTeamCtx.SetParticipationStatus(participationStatus)
	decision := createTeamPolicy.Check(ctx, createTeamCtx)
	out.CreateTeam = decision.Allowed

	// If user is in a team (team_member or team_captain), check team management permissions
	isInTeam := participationStatus == "team_member" || participationStatus == "team_captain"
	if isInTeam {
		// Determine if user is captain
		isCaptain := participationStatus == "team_captain"
		
		// For delete team check, we need members count
		// We'll assume 1 member for captain-only check
		membersCount := int64(1)
		if !isCaptain {
			membersCount = 2 // If not captain, assume there are other members
		}

		// Check editTeam
		updateTeamPolicy := policy.NewUpdateTeamPolicy()
		updateTeamCtx := policy.NewUpdateTeamContext()
		updateTeamCtx.SetAuthenticated(true)
		updateTeamCtx.SetActorUserID(userUUID)
		updateTeamCtx.SetHackathonStage(stage)
		updateTeamCtx.SetAllowTeam(allowTeam)
		updateTeamCtx.SetIsCaptain(isCaptain)
		decision = updateTeamPolicy.Check(ctx, updateTeamCtx)
		out.CanInMyTeam.EditTeam = decision.Allowed

		// Check deleteTeam
		deleteTeamPolicy := policy.NewDeleteTeamPolicy()
		deleteTeamCtx := policy.NewDeleteTeamContext()
		deleteTeamCtx.SetAuthenticated(true)
		deleteTeamCtx.SetActorUserID(userUUID)
		deleteTeamCtx.SetHackathonStage(stage)
		deleteTeamCtx.SetAllowTeam(allowTeam)
		deleteTeamCtx.SetIsCaptain(isCaptain)
		deleteTeamCtx.SetMembersCount(membersCount)
		decision = deleteTeamPolicy.Check(ctx, deleteTeamCtx)
		out.CanInMyTeam.DeleteTeam = decision.Allowed

		// Check manageVacancies
		upsertVacancyPolicy := policy.NewUpsertVacancyPolicy()
		upsertVacancyCtx := policy.NewUpsertVacancyContext()
		upsertVacancyCtx.SetAuthenticated(true)
		upsertVacancyCtx.SetActorUserID(userUUID)
		upsertVacancyCtx.SetHackathonStage(stage)
		upsertVacancyCtx.SetAllowTeam(allowTeam)
		upsertVacancyCtx.SetIsCaptain(isCaptain)
		decision = upsertVacancyPolicy.Check(ctx, upsertVacancyCtx)
		out.CanInMyTeam.ManageVacancies = decision.Allowed

		// Check inviteMember
		invitePolicy := policy.NewCreateTeamInvitationPolicy()
		inviteCtx := policy.NewCreateTeamInvitationContext()
		inviteCtx.SetAuthenticated(true)
		inviteCtx.SetActorUserID(userUUID)
		inviteCtx.SetHackathonStage(stage)
		inviteCtx.SetAllowTeam(allowTeam)
		inviteCtx.SetIsCaptain(isCaptain)
		inviteCtx.SetTargetIsStaff(false)
		inviteCtx.SetIsTargetTeamMember(false)
		decision = invitePolicy.Check(ctx, inviteCtx)
		out.CanInMyTeam.InviteMember = decision.Allowed

		// Check manageJoinRequests
		listJoinRequestsPolicy := policy.NewListJoinRequestsPolicy()
		listJoinRequestsCtx := policy.NewListJoinRequestsContext()
		listJoinRequestsCtx.SetAuthenticated(true)
		listJoinRequestsCtx.SetActorUserID(userUUID)
		listJoinRequestsCtx.SetHackathonStage(stage)
		listJoinRequestsCtx.SetIsCaptain(isCaptain)
		decision = listJoinRequestsPolicy.Check(ctx, listJoinRequestsCtx)
		out.CanInMyTeam.ManageJoinRequests = decision.Allowed

		// Check kickMember
		kickPolicy := policy.NewKickTeamMemberPolicy()
		kickCtx := policy.NewKickTeamMemberContext()
		kickCtx.SetAuthenticated(true)
		kickCtx.SetActorUserID(userUUID)
		kickCtx.SetHackathonStage(stage)
		kickCtx.SetAllowTeam(allowTeam)
		kickCtx.SetIsCaptain(isCaptain)
		kickCtx.SetTargetIsMember(true)
		kickCtx.SetTargetIsCaptain(false)
		kickCtx.SetTargetUserID(uuid.New()) // dummy
		decision = kickPolicy.Check(ctx, kickCtx)
		out.CanInMyTeam.KickMember = decision.Allowed

		// Check transferCaptain
		transferPolicy := policy.NewTransferCaptainPolicy()
		transferCtx := policy.NewTransferCaptainContext()
		transferCtx.SetAuthenticated(true)
		transferCtx.SetActorUserID(userUUID)
		transferCtx.SetHackathonStage(stage)
		transferCtx.SetAllowTeam(allowTeam)
		transferCtx.SetIsCaptain(isCaptain)
		transferCtx.SetTargetIsMember(true)
		transferCtx.SetNewCaptainID(uuid.New()) // dummy
		decision = transferPolicy.Check(ctx, transferCtx)
		out.CanInMyTeam.TransferCaptain = decision.Allowed

		// Check leaveTeam
		leavePolicy := policy.NewLeaveTeamPolicy()
		leaveCtx := policy.NewLeaveTeamContext()
		leaveCtx.SetAuthenticated(true)
		leaveCtx.SetActorUserID(userUUID)
		leaveCtx.SetHackathonStage(stage)
		leaveCtx.SetAllowTeam(allowTeam)
		leaveCtx.SetIsMember(true)
		leaveCtx.SetIsCaptain(isCaptain)
		decision = leavePolicy.Check(ctx, leaveCtx)
		out.CanInMyTeam.LeaveTeam = decision.Allowed
	}

	return out, nil
}
