package participation

import (
	"context"
	"strings"

	hackathonclient "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/client/hackathon"
	"github.com/google/uuid"
)

type hackathonClientAdapter struct {
	client *hackathonclient.Client
}

func NewHackathonClientAdapter(client *hackathonclient.Client) HackathonClient {
	return &hackathonClientAdapter{client: client}
}

func (a *hackathonClientAdapter) GetHackathonInfo(ctx context.Context, hackathonID uuid.UUID) (*HackathonInfo, error) {
	info, err := a.client.GetHackathonInfo(ctx, hackathonID.String())
	if err != nil {
		return nil, err
	}

	// Normalize stage: "HACKATHON_STAGE_REGISTRATION" -> "registration"
	stage := strings.ToLower(info.Stage)
	stage = strings.TrimPrefix(stage, "hackathon_stage_")

	return &HackathonInfo{
		Stage:           stage,
		AllowIndividual: info.AllowIndividual,
		AllowTeam:       info.AllowTeam,
	}, nil
}
