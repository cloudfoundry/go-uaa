package uaa

import (
	"encoding/json"
	"net/http"
)

// Keys is a slice of JSON Web Keys.
type Keys struct {
	Keys []JWK `json:"keys"`
}

// TokenKeys gets the JSON Web Token signing keys with the given client and
// config.
func TokenKeys(client *http.Client, config Config) ([]JWK, error) {
	body, err := UnauthenticatedRequestor{}.Get(client, config, "/token_keys", "")
	if err != nil {
		key, e := TokenKey(client, config)
		return []JWK{key}, e
	}

	keys := Keys{}
	err = json.Unmarshal(body, &keys)
	if err != nil {
		return []JWK{}, parseError("/token_keys", body)
	}

	return keys.Keys, nil
}
