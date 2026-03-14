package judging

import (
	"fmt"

	submissionv1 "github.com/belikoooova/hackaton-platform-api/api/submission/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

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

func ownerKindToString(kind submissionv1.OwnerKind) string {
	switch kind {
	case submissionv1.OwnerKind_OWNER_KIND_USER:
		return domain.OwnerKindUser
	case submissionv1.OwnerKind_OWNER_KIND_TEAM:
		return domain.OwnerKindTeam
	default:
		return ""
	}
}

func stringToOwnerKind(s string) submissionv1.OwnerKind {
	switch s {
	case domain.OwnerKindUser:
		return submissionv1.OwnerKind_OWNER_KIND_USER
	case domain.OwnerKindTeam:
		return submissionv1.OwnerKind_OWNER_KIND_TEAM
	default:
		return submissionv1.OwnerKind_OWNER_KIND_UNSPECIFIED
	}
}
