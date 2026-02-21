package teaminbox

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

func encodePageToken(offset int) string {
	if offset == 0 {
		return ""
	}
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", offset)))
}

func parsePageToken(token string) (int, error) {
	if token == "" {
		return 0, nil
	}

	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return 0, fmt.Errorf("failed to decode page token: %w", err)
	}

	offset, err := strconv.Atoi(string(decoded))
	if err != nil {
		return 0, fmt.Errorf("failed to parse offset: %w", err)
	}

	if offset < 0 {
		return 0, fmt.Errorf("invalid offset: %d", offset)
	}

	return offset, nil
}
