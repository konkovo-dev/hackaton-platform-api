package domain

type ValidationError struct {
	Code    string
	Field   string
	Message string
	Meta    map[string]string
}

const (
	ValidationCodeRequired   = "REQUIRED"
	ValidationCodeTimeRule   = "TIME_RULE"
	ValidationCodeTimeLocked = "TIME_LOCKED"
	ValidationCodeTypeRule   = "TYPE_RULE"
	ValidationCodePolicyRule = "POLICY_RULE"
	ValidationCodeForbidden  = "FORBIDDEN"
	ValidationCodeFormat     = "FORMAT"
)
