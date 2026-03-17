package hackathon

import (
	"context"

	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type GetHackathonPermissionsIn struct {
	HackathonID uuid.UUID
}

type GetHackathonPermissionsOut struct {
	ManageHackathon    bool
	ReadDraft          bool
	PublishHackathon   bool
	ViewAnnouncements  bool
	CreateAnnouncement bool
	ReadTask           bool
	ViewResultPublic   bool
	ReadResultDraft    bool
	PublishResult      bool
	UpdateResultDraft  bool
}

func (s *Service) GetHackathonPermissions(ctx context.Context, in GetHackathonPermissionsIn) (*GetHackathonPermissionsOut, error) {
	out := &GetHackathonPermissionsOut{}

	// Check manageHackathon (update hackathon)
	updatePolicy := hackathonpolicy.NewUpdateHackathonPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := updatePolicy.LoadContext(ctx, hackathonpolicy.UpdateHackathonParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := updatePolicy.Check(ctx, pctx)
		out.ManageHackathon = decision.Allowed
	}

	// Check readDraft (get hackathon when state=DRAFT)
	getPolicy := hackathonpolicy.NewGetHackathonPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := getPolicy.LoadContext(ctx, hackathonpolicy.GetHackathonParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := getPolicy.Check(ctx, pctx)
		out.ReadDraft = decision.Allowed
	}

	// Check publishHackathon
	publishPolicy := hackathonpolicy.NewPublishHackathonPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := publishPolicy.LoadContext(ctx, hackathonpolicy.PublishHackathonParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := publishPolicy.Check(ctx, pctx)
		out.PublishHackathon = decision.Allowed
	}

	// Check viewAnnouncements
	readAnnouncementsPolicy := hackathonpolicy.NewReadAnnouncementsPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := readAnnouncementsPolicy.LoadContext(ctx, hackathonpolicy.AnnouncementPolicyParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := readAnnouncementsPolicy.Check(ctx, pctx)
		out.ViewAnnouncements = decision.Allowed
	}

	// Check createAnnouncement
	createAnnouncementPolicy := hackathonpolicy.NewCreateAnnouncementPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := createAnnouncementPolicy.LoadContext(ctx, hackathonpolicy.AnnouncementPolicyParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := createAnnouncementPolicy.Check(ctx, pctx)
		out.CreateAnnouncement = decision.Allowed
	}

	// Check readTask
	readTaskPolicy := hackathonpolicy.NewReadTaskPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := readTaskPolicy.LoadContext(ctx, hackathonpolicy.ReadTaskParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := readTaskPolicy.Check(ctx, pctx)
		out.ReadTask = decision.Allowed
	}

	// Check viewResultPublic and readResultDraft (same policy, different conditions)
	readResultPolicy := hackathonpolicy.NewReadResultPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := readResultPolicy.LoadContext(ctx, hackathonpolicy.ReadResultParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := readResultPolicy.Check(ctx, pctx)
		// If allowed, determine if it's public or draft access
		if decision.Allowed {
			hctx := pctx.(*hackathonpolicy.HackathonPolicyContext)
			if hctx.Stage() == "FINISHED" {
				out.ViewResultPublic = true
			} else {
				out.ReadResultDraft = true
			}
		}
	}

	// Check publishResult
	publishResultPolicy := hackathonpolicy.NewPublishResultPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := publishResultPolicy.LoadContext(ctx, hackathonpolicy.PublishResultParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := publishResultPolicy.Check(ctx, pctx)
		out.PublishResult = decision.Allowed
	}

	// Check updateResultDraft
	updateResultDraftPolicy := hackathonpolicy.NewUpdateResultDraftPolicy(s.hackathonRepo, s.parClient)
	if pctx, err := updateResultDraftPolicy.LoadContext(ctx, hackathonpolicy.UpdateResultDraftParams{
		HackathonID: in.HackathonID,
	}); err == nil {
		decision := updateResultDraftPolicy.Check(ctx, pctx)
		out.UpdateResultDraft = decision.Allowed
	}

	return out, nil
}
