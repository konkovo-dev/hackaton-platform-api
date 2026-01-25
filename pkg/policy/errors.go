package policy

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PolicyError struct {
	Action     Action
	Violations []Violation
}

func (e *PolicyError) Error() string {
	if len(e.Violations) == 0 {
		return fmt.Sprintf("policy violation for action %s", e.Action)
	}

	messages := make([]string, len(e.Violations))
	for i, v := range e.Violations {
		if v.Field != "" {
			messages[i] = fmt.Sprintf("%s: %s", v.Field, v.Message)
		} else {
			messages[i] = v.Message
		}
	}

	return fmt.Sprintf("policy violation for action %s: %s", e.Action, strings.Join(messages, "; "))
}

func NewPolicyError(action Action, violations []Violation) *PolicyError {
	return &PolicyError{
		Action:     action,
		Violations: violations,
	}
}

func (e *PolicyError) ToGRPCError() error {
	if len(e.Violations) == 0 {
		return status.Error(codes.PermissionDenied, "access denied")
	}

	grpcCode := e.determineGRPCCode()

	return status.Error(grpcCode, e.Error())
}

func (e *PolicyError) determineGRPCCode() codes.Code {
	hasForbidden := false
	hasRequired := false
	hasFormat := false
	hasConflict := false
	hasStageRule := false
	hasNotFound := false

	for _, v := range e.Violations {
		switch v.Code {
		case ViolationCodeForbidden:
			hasForbidden = true
		case ViolationCodeRequired, ViolationCodeFormat:
			hasRequired = true
			if v.Code == ViolationCodeFormat {
				hasFormat = true
			}
		case ViolationCodeConflict:
			hasConflict = true
		case ViolationCodeStageRule, ViolationCodeTimeRule, ViolationCodeTimeLocked, ViolationCodePolicyRule, ViolationCodeLimitRule:
			hasStageRule = true
		case ViolationCodeNotFound:
			hasNotFound = true
		}
	}

	if hasNotFound {
		return codes.NotFound
	}
	if hasForbidden {
		return codes.PermissionDenied
	}
	if hasConflict {
		return codes.AlreadyExists
	}
	if hasStageRule {
		return codes.FailedPrecondition
	}
	if hasRequired || hasFormat {
		return codes.InvalidArgument
	}

	return codes.PermissionDenied
}

func IsPolicyError(err error) bool {
	_, ok := err.(*PolicyError)
	return ok
}
