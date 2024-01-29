package pagination

import (
	"encoding/base64"
	"encoding/json"
)

func DecodeToken(token string, container interface{}) error {
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, container)
}

func EncodeToken(container interface{}) (string, error) {
	b, err := json.Marshal(container)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
