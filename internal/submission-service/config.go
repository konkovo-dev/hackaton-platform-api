package submissionservice

import (
	"fmt"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	Limits LimitsConfig
}

type LimitsConfig struct {
	MaxFileSizeBytes       int64
	MaxTotalSizeBytes      int64
	MaxFilesPerSubmission  int
	MaxSubmissionsPerOwner int
	AllowedFileExtensions  []string
	AllowedContentTypes    []string
	PresignedURLExpiryMins int
}

func NewConfig() *Config {
	maxFileSizeBytes, err := env.GetEnvInt64("SUBMISSION_MAX_FILE_SIZE_BYTES", 50*1024*1024)
	if err != nil {
		panic(fmt.Errorf("invalid SUBMISSION_MAX_FILE_SIZE_BYTES: %w", err))
	}

	maxTotalSizeBytes, err := env.GetEnvInt64("SUBMISSION_MAX_TOTAL_SIZE_BYTES", 200*1024*1024)
	if err != nil {
		panic(fmt.Errorf("invalid SUBMISSION_MAX_TOTAL_SIZE_BYTES: %w", err))
	}

	maxFilesPerSubmission, err := env.GetEnvInt("SUBMISSION_MAX_FILES_PER_SUBMISSION", 20)
	if err != nil {
		panic(fmt.Errorf("invalid SUBMISSION_MAX_FILES_PER_SUBMISSION: %w", err))
	}

	maxSubmissionsPerOwner, err := env.GetEnvInt("SUBMISSION_MAX_SUBMISSIONS_PER_OWNER", 50)
	if err != nil {
		panic(fmt.Errorf("invalid SUBMISSION_MAX_SUBMISSIONS_PER_OWNER: %w", err))
	}

	presignedURLExpiryMins, err := env.GetEnvInt("SUBMISSION_PRESIGNED_URL_EXPIRY_MINS", 15)
	if err != nil {
		panic(fmt.Errorf("invalid SUBMISSION_PRESIGNED_URL_EXPIRY_MINS: %w", err))
	}

	allowedExtensions := parseCommaSeparated(
		env.GetEnv("SUBMISSION_ALLOWED_EXTENSIONS", ".pdf,.zip,.png,.jpg,.jpeg,.txt,.md,.csv"),
	)

	allowedContentTypes := parseCommaSeparated(
		env.GetEnv("SUBMISSION_ALLOWED_CONTENT_TYPES",
			"application/pdf,application/zip,image/png,image/jpeg,text/plain,text/markdown,text/csv"),
	)

	return &Config{
		Limits: LimitsConfig{
			MaxFileSizeBytes:       maxFileSizeBytes,
			MaxTotalSizeBytes:      maxTotalSizeBytes,
			MaxFilesPerSubmission:  maxFilesPerSubmission,
			MaxSubmissionsPerOwner: maxSubmissionsPerOwner,
			AllowedFileExtensions:  allowedExtensions,
			AllowedContentTypes:    allowedContentTypes,
			PresignedURLExpiryMins: presignedURLExpiryMins,
		},
	}
}

func MustNewConfig() *Config {
	return NewConfig()
}

func parseCommaSeparated(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
