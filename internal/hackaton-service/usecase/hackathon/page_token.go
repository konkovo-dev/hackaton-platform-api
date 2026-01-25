package hackathon

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

func encodePageToken(offset uint32) string {
	s := fmt.Sprintf("%d", offset)
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func parsePageToken(token string) (uint32, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}

	offset, err := strconv.ParseUint(string(decoded), 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(offset), nil
}
