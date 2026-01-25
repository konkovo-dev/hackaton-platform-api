package policy

import "github.com/belikoooova/hackaton-platform-api/pkg/policy"

const (
	ActionReadMe           policy.Action = "identity.read_me"
	ActionUpdateMe         policy.Action = "identity.update_me"
	ActionUpdateMySkills   policy.Action = "identity.update_my_skills"
	ActionUpdateMyContacts policy.Action = "identity.update_my_contacts"

	ActionReadUser      policy.Action = "identity.read_user"
	ActionBatchGetUsers policy.Action = "identity.batch_get_users"
	ActionListUsers     policy.Action = "identity.list_users"
)
