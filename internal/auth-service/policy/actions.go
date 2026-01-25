package policy

import "github.com/belikoooova/hackaton-platform-api/pkg/policy"

const (
	ActionRegister        policy.Action = "auth.register"
	ActionLogin           policy.Action = "auth.login"
	ActionRefresh         policy.Action = "auth.refresh"
	ActionLogout          policy.Action = "auth.logout"
	ActionIntrospectToken policy.Action = "auth.introspect_token"
)
