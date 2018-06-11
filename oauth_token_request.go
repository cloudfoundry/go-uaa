package uaa

import (
	"encoding/json"
	"net/http"
)

func postToOAuthToken(httpClient *http.Client, config Config, body map[string]string) (TokenResponse, error) {
	bytes, err := UnauthenticatedRequestor{}.PostForm(httpClient, config, "/oauth/token", "", body)
	if err != nil {
		return TokenResponse{}, err
	}

	tokenResponse := TokenResponse{}
	err = json.Unmarshal(bytes, &tokenResponse)
	if err != nil {
		return TokenResponse{}, parseError("/oauth/token", bytes)
	}

	return tokenResponse, nil
}

// ClientCredentialsClient is used to authenticate with the authorization server.
type ClientCredentialsClient struct {
	ClientID     string
	ClientSecret string
}

// RequestToken gets a token from the token endpoint.
func (cc ClientCredentialsClient) RequestToken(httpClient *http.Client, config Config, format TokenFormat) (TokenResponse, error) {
	body := map[string]string{
		"grant_type":    string(CLIENTCREDENTIALS),
		"client_id":     cc.ClientID,
		"client_secret": cc.ClientSecret,
		"token_format":  format.String(),
		"response_type": "token",
	}

	return postToOAuthToken(httpClient, config, body)
}

// ResourceOwnerPasswordClient is used to authenticate with the authorization server.
type ResourceOwnerPasswordClient struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

// RequestToken gets a token from the token endpoint.
func (rop ResourceOwnerPasswordClient) RequestToken(httpClient *http.Client, config Config, format TokenFormat) (TokenResponse, error) {
	body := map[string]string{
		"grant_type":    string(PASSWORD),
		"client_id":     rop.ClientID,
		"client_secret": rop.ClientSecret,
		"username":      rop.Username,
		"password":      rop.Password,
		"token_format":  format.String(),
		"response_type": "token",
	}

	return postToOAuthToken(httpClient, config, body)
}

// AuthorizationCodeClient is used to authenticate with the authorization server.
type AuthorizationCodeClient struct {
	ClientID     string
	ClientSecret string
}

// RequestToken gets a token from the token endpoint.
func (acc AuthorizationCodeClient) RequestToken(httpClient *http.Client, config Config, format TokenFormat, code string, redirectURI string) (TokenResponse, error) {
	body := map[string]string{
		"grant_type":    string(AUTHCODE),
		"client_id":     acc.ClientID,
		"client_secret": acc.ClientSecret,
		"token_format":  format.String(),
		"response_type": "token",
		"redirect_uri":  redirectURI,
		"code":          code,
	}

	return postToOAuthToken(httpClient, config, body)
}

// RefreshTokenClient is used to authenticate with the authorization server.
type RefreshTokenClient struct {
	ClientID     string
	ClientSecret string
}

// RequestToken gets a token from the token endpoint.
func (rc RefreshTokenClient) RequestToken(httpClient *http.Client, config Config, format TokenFormat, refreshToken string) (TokenResponse, error) {
	body := map[string]string{
		"grant_type":    string(REFRESHTOKEN),
		"refresh_token": refreshToken,
		"client_id":     rc.ClientID,
		"client_secret": rc.ClientSecret,
		"token_format":  format.String(),
		"response_type": "token",
	}

	return postToOAuthToken(httpClient, config, body)
}

// // TokenFormat is the format of a token.
// type TokenFormat string
//
// // Valid TokenFormat values.
// const (
// 	OPAQUE = TokenFormat("opaque")
// 	JWT    = TokenFormat("jwt")
// )

// GrantType is a type of oauth2 grant.
type GrantType string

// Valid GrantType values.
const (
	REFRESHTOKEN      = GrantType("refresh_token")
	AUTHCODE          = GrantType("authorization_code")
	IMPLICIT          = GrantType("implicit")
	PASSWORD          = GrantType("password")
	CLIENTCREDENTIALS = GrantType("client_credentials")
)

// TokenResponse is a token.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int32  `json:"expires_in"`
	Scope        string `json:"scope"`
	JTI          string `json:"jti"`
}
