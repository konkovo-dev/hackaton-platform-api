package participation

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/google/uuid"
)

type GetParticipationPermissionsIn struct {
	HackathonID uuid.UUID
}

type GetParticipationPermissionsOut struct {
	Register                   bool
	Unregister                 bool
	SwitchParticipationMode    bool
	UpdateParticipationProfile bool
	InviteStaff                bool
	ListParticipants           bool
}

type policyRepos struct {
	participRepo    ParticipationRepository
	roleRepo        StaffRoleRepository
	hackathonClient HackathonClient
}

func (r *policyRepos) GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error) {
	return r.roleRepo.GetRoleStrings(ctx, hackathonID, userID)
}

func (r *policyRepos) GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error) {
	return r.participRepo.GetStatus(ctx, hackathonID, userID)
}

func (r *policyRepos) GetHackathonInfo(ctx context.Context, hackathonID uuid.UUID) (*policy.HackathonInfo, error) {
	info, err := r.hackathonClient.GetHackathonInfo(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	return &policy.HackathonInfo{
		Stage:           info.Stage,
		AllowIndividual: info.AllowIndividual,
		AllowTeam:       info.AllowTeam,
	}, nil
}

func (s *Service) GetParticipationPermissions(ctx context.Context, in GetParticipationPermissionsIn) (*GetParticipationPermissionsOut, error) {
	out := &GetParticipationPermissionsOut{}

	repos := &policyRepos{
		participRepo:    s.participRepo,
		roleRepo:        s.roleRepo,
		hackathonClient: s.hackathonClient,
	}

	// Check register
	registerPolicy := policy.NewRegisterForHackathonPolicy(repos)
	if pctx, err := registerPolicy.LoadContext(ctx, policy.RegisterForHackathonParams{
		HackathonID:   in.HackathonID,
		DesiredStatus: "individual", // dummy value for permission check
	}); err == nil {
		decision := registerPolicy.Check(ctx, pctx)
		out.Register = decision.Allowed
	}

	// Check unregister
	unregisterPolicy := policy.NewUnregisterFromHackathonPolicy(repos)
	if pctx, err := unregisterPolicy.LoadContext(ctx, policy.UnregisterFromHackathonParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := unregisterPolicy.Check(ctx, pctx)
		out.Unregister = decision.Allowed
	}

	// Check switchParticipationMode.
	// We only need to know if the user CAN switch at all (CanSwitch), not whether a
	// specific target status differs from the current one (IsStatusDifferent).
	// Using the full Check() with a hardcoded dummy target causes a false-negative
	// when the user's current status equals the dummy value.
	switchModePolicy := policy.NewSwitchParticipationModePolicy(repos)
	if pctx, err := switchModePolicy.LoadContext(ctx, policy.SwitchParticipationModeParams{
		HackathonID: in.HackathonID,
		NewStatus:   "individual", // value doesn't matter here — we check CanSwitch directly
	}); err == nil {
		out.SwitchParticipationMode = pctx.IsAuthenticated() && pctx.CanSwitch()
	}

	// Check updateParticipationProfile
	updateMyPolicy := policy.NewUpdateMyParticipationPolicy(repos)
	if pctx, err := updateMyPolicy.LoadContext(ctx, policy.UpdateMyParticipationParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := updateMyPolicy.Check(ctx, pctx)
		out.UpdateParticipationProfile = decision.Allowed
	}

	// Check inviteStaff
	inviteStaffPolicy := policy.NewCreateStaffInvitationPolicy(repos)
	if pctx, err := inviteStaffPolicy.LoadContext(ctx, policy.CreateStaffInvitationParams{
		HackathonID:   in.HackathonID,
		TargetUserID:  uuid.New(), // dummy value
		RequestedRole: "organizer",
	}); err == nil {
		decision := inviteStaffPolicy.Check(ctx, pctx)
		out.InviteStaff = decision.Allowed
	}

	// Check listParticipants
	listParticipantsPolicy := policy.NewListParticipantsPolicy(repos)
	if pctx, err := listParticipantsPolicy.LoadContext(ctx, policy.ListParticipantsParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := listParticipantsPolicy.Check(ctx, pctx)
		out.ListParticipants = decision.Allowed
	}

	return out, nil
}
