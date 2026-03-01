package policy

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

const (
	ActionCreateSubmission            policy.Action = "create_submission"
	ActionUpdateSubmission            policy.Action = "update_submission"
	ActionListSubmissions             policy.Action = "list_submissions"
	ActionSelectFinalSubmission       policy.Action = "select_final_submission"
	ActionGetSubmission               policy.Action = "get_submission"
	ActionGetFinalSubmission          policy.Action = "get_final_submission"
	ActionCreateSubmissionUpload      policy.Action = "create_submission_upload"
	ActionCompleteSubmissionUpload    policy.Action = "complete_submission_upload"
	ActionGetSubmissionFileDownloadURL policy.Action = "get_submission_file_download_url"
)
