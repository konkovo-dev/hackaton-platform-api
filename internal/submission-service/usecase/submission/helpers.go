package submission

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

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

func validateFileExtension(filename string, allowed []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return slices.Contains(allowed, ext)
}

func validateContentType(contentType string, allowed []string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	return slices.Contains(allowed, ct)
}
