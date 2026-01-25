package policy

import (
	"github.com/google/uuid"
)

type Action string

type PolicyContext interface {
	IsAuthenticated() bool
	ActorUserID() uuid.UUID
}

type Violation struct {
	Code    string                 `json:"code"`
	Field   string                 `json:"field,omitempty"`
	Message string                 `json:"message"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

type Decision struct {
	Allowed    bool        `json:"allowed"`
	Violations []Violation `json:"violations,omitempty"`
}

func NewDecision() *Decision {
	return &Decision{
		Allowed:    true,
		Violations: []Violation{},
	}
}

func (d *Decision) Allow() *Decision {
	d.Allowed = true
	return d
}

func (d *Decision) Deny(violations ...Violation) *Decision {
	d.Allowed = false
	d.Violations = append(d.Violations, violations...)
	return d
}

func (d *Decision) AddViolation(v Violation) {
	d.Violations = append(d.Violations, v)
	if !d.Allowed {
		d.Allowed = false
	}
}

func (d *Decision) HasViolations() bool {
	return len(d.Violations) > 0
}

const (
	ViolationCodeForbidden  = "FORBIDDEN"
	ViolationCodeRequired   = "REQUIRED"
	ViolationCodeFormat     = "FORMAT"
	ViolationCodeConflict   = "CONFLICT"
	ViolationCodeStageRule  = "STAGE_RULE"
	ViolationCodeTimeRule   = "TIME_RULE"
	ViolationCodeTimeLocked = "TIME_LOCKED"
	ViolationCodePolicyRule = "POLICY_RULE"
	ViolationCodeNotFound   = "NOT_FOUND"
	ViolationCodeLimitRule  = "LIMIT_RULE"
)
