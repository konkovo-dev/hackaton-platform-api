package queryutil

import (
	"encoding/base64"
	"encoding/json"
)

func EncodeCursor(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(jsonData), nil
}

func DecodeCursor(token string, dest interface{}) error {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}
