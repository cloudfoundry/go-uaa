package uaa

import (
	"encoding/json"
	"net/http"
)

// UserInfo is a protected resource required for OpenID Connect compatibility.
// The response format is defined here: https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse.
type UserInfo struct {
	UserID            string   `json:"user_id"`
	Sub               string   `json:"sub"`
	Username          string   `json:"user_name"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	Email             string   `json:"email"`
	PhoneNumber       []string `json:"phone_number"`
	PreviousLoginTime int64    `json:"previous_logon_time"`
	Name              string   `json:"name"`
}

// Me retrieves the UserInfo for the current user.
func Me(client *http.Client, config Config) (UserInfo, error) {
	body, err := AuthenticatedRequestor{}.Get(client, config, "/userinfo", "scheme=openid")
	if err != nil {
		return UserInfo{}, err
	}

	info := UserInfo{}
	err = json.Unmarshal(body, &info)
	if err != nil {
		return UserInfo{}, parseError("/userinfo", body)
	}

	return info, nil
}
