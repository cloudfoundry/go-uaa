package uaa

import (
	"encoding/json"
	"net/http"
)

// JWK represents a JSON Web Key (https://tools.ietf.org/html/rfc7517).
type JWK struct {
	Kty   string `json:"kty"`
	E     string `json:"e,omitempty"`
	Use   string `json:"use"`
	Kid   string `json:"kid"`
	Alg   string `json:"alg"`
	Value string `json:"value"`
	N     string `json:"n,omitempty"`
}

// TokenKey retrieves a JWK from the token_key endpoint
// (http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#token-key-s).
func TokenKey(client *http.Client, config Config) (JWK, error) {
	body, err := UnauthenticatedRequestor{}.Get(client, config, "token_key", "")
	if err != nil {
		return JWK{}, err
	}

	key := JWK{}
	err = json.Unmarshal(body, &key)
	if err != nil {
		return JWK{}, parseError("/token_key", body)
	}

	return key, nil
}
