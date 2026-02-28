package mentors

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) getTicketRecipients(ctx context.Context, ticketID uuid.UUID, ownerKind string, ownerID uuid.UUID, hackathonID uuid.UUID) ([]string, error) {
	recipients := make([]string, 0)

	switch ownerKind {
	case domain.OwnerKindTeam:
		teamMembers, err := s.teamClient.ListTeamMembers(ctx, hackathonID.String(), ownerID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to list team members: %w", err)
		}
		recipients = append(recipients, teamMembers...)
	case domain.OwnerKindUser:
		recipients = append(recipients, ownerID.String())
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	if ticket != nil && ticket.AssignedMentorUserID != nil {
		mentorID := ticket.AssignedMentorUserID.String()
		alreadyIncluded := slices.Contains(recipients, mentorID)
		if !alreadyIncluded {
			recipients = append(recipients, mentorID)
		}
	}

	return recipients, nil
}

func mapPolicyError(decision *policy.Decision) error {
	if len(decision.Violations) == 0 {
		return ErrUnauthorized
	}

	v := decision.Violations[0]
	switch v.Code {
	case policy.ViolationCodeForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	case policy.ViolationCodeNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, v.Message)
	case policy.ViolationCodeConflict:
		return fmt.Errorf("%w: %s", ErrConflict, v.Message)
	case policy.ViolationCodeStageRule, policy.ViolationCodePolicyRule:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	default:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	}
}
