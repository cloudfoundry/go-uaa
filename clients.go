package uaa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ClientsEndpoint is the path to the clients resource.
const ClientsEndpoint string = "/oauth/clients"

// paginatedClientList is the response from the API for a single page of clients.
type paginatedClientList struct {
	Page
	Resources []Client `json:"resources"`
	Schemas   []string `json:"schemas"`
}

// Client is a UAA client
// http://docs.cloudfoundry.org/api/uaa/version/4.19.0/index.html#clients.
type Client struct {
	ClientID              string          `json:"client_id,omitempty" generator:"id"`
	AuthorizedGrantTypes  []string        `json:"authorized_grant_types,omitempty"`
	RedirectURI           []string        `json:"redirect_uri,omitempty"`
	Scope                 []string        `json:"scope,omitempty"`
	ResourceIDs           []string        `json:"resource_ids,omitempty"`
	Authorities           []string        `json:"authorities,omitempty"`
	AutoApproveRaw        interface{}     `json:"autoapprove,omitempty"`
	AccessTokenValidity   int64           `json:"access_token_validity,omitempty"`
	RefreshTokenValidity  int64           `json:"refresh_token_validity,omitempty"`
	AllowedProvidersRaw   interface{}     `json:"allowedproviders,omitempty"`
	DisplayName           string          `json:"name,omitempty"`
	TokenSalt             string          `json:"token_salt,omitempty"`
	CreatedWith           string          `json:"createdwith,omitempty"`
	ApprovalsDeletedRaw   interface{}     `json:"approvals_deleted,omitempty"`
	RequiredUserGroupsRaw interface{}     `json:"required_user_groups,omitempty"`
	ClientSecret          string          `json:"client_secret,omitempty"`
	LastModifiedRaw       interface{}     `json:"lastModified,omitempty"`
	AllowPublicRaw        interface{}     `json:"allowpublic,omitempty"`
	JwksURI               string          `json:"jwks_uri,omitempty"`
	Jwks                  json.RawMessage `json:"jwks,omitempty"`
}

// Identifier returns the field used to uniquely identify a Client.
func (c Client) Identifier() string {
	return c.ClientID
}

func (c Client) AutoApprove() []string {
	switch t := c.AutoApproveRaw.(type) {
	case bool:
		return []string{strconv.FormatBool(t)}
	case string:
		return []string{t}
	case []string:
		return t
	}
	return []string{}
}

// AllowPublic returns whether the client allows public access, tolerating
// UAA responses that encode allowpublic as either a JSON boolean or string.
func (c Client) AllowPublic() bool {
	switch t := c.AllowPublicRaw.(type) {
	case bool:
		return t
	case string:
		b, err := strconv.ParseBool(t)
		if err != nil {
			return false
		}
		return b
	}
	return false
}

// ApprovalsDeleted returns whether the client's approvals were deleted,
// tolerating UAA responses that encode approvals_deleted as either a JSON
// boolean or string.
func (c Client) ApprovalsDeleted() bool {
	switch t := c.ApprovalsDeletedRaw.(type) {
	case bool:
		return t
	case string:
		b, err := strconv.ParseBool(t)
		if err != nil {
			return false
		}
		return b
	}
	return false
}

// AllowedProviders returns the client's allowed identity providers,
// tolerating UAA responses that encode allowedproviders as either a JSON
// array or a single string.
func (c Client) AllowedProviders() []string {
	switch t := c.AllowedProvidersRaw.(type) {
	case string:
		return []string{t}
	case []string:
		return t
	}
	return []string{}
}

// RequiredUserGroups returns the client's required user groups, tolerating
// UAA responses that encode required_user_groups as either a JSON array or a
// single string.
func (c Client) RequiredUserGroups() []string {
	switch t := c.RequiredUserGroupsRaw.(type) {
	case string:
		return []string{t}
	case []string:
		return t
	}
	return []string{}
}

// LastModified returns the client's last-modified timestamp in epoch
// milliseconds, tolerating UAA responses that encode lastModified as either a
// JSON number or string.
func (c Client) LastModified() int64 {
	switch t := c.LastModifiedRaw.(type) {
	case float64:
		return int64(t)
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0
		}
		return i
	}
	return 0
}

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

func errorMissingValueForGrantType(value string, grantType GrantType) error {
	return fmt.Errorf("%v must be specified for %v grant type", value, grantType)
}

func errorMissingValue(value string) error {
	return fmt.Errorf("%v must be specified in the client definition", value)
}

func requireRedirectURIForGrantType(c *Client, grantType GrantType) error {
	if contains(c.AuthorizedGrantTypes, string(grantType)) {
		if len(c.RedirectURI) == 0 {
			return errorMissingValueForGrantType("redirect_uri", grantType)
		}
	}
	return nil
}

func requireClientSecretForGrantType(c *Client, grantType GrantType) error {
	if contains(c.AuthorizedGrantTypes, string(grantType)) {
		if c.ClientSecret == "" {
			return errorMissingValueForGrantType("client_secret", grantType)
		}
	}
	return nil
}

func knownGrantTypesStr() string {
	grantTypeStrings := []string{}
	knownGrantTypes := []GrantType{AUTHCODE, IMPLICIT, PASSWORD, CLIENTCREDENTIALS}
	for _, grant := range knownGrantTypes {
		grantTypeStrings = append(grantTypeStrings, string(grant))
	}

	return "[" + strings.Join(grantTypeStrings, ", ") + "]"
}

// Validate returns nil if the client is valid, or an error if it is invalid.
func (c *Client) Validate() error {
	if len(c.AuthorizedGrantTypes) == 0 {
		return fmt.Errorf("grant type must be one of %v", knownGrantTypesStr())
	}

	if c.ClientID == "" {
		return errorMissingValue("client_id")
	}

	if err := requireRedirectURIForGrantType(c, AUTHCODE); err != nil {
		return err
	}
	if err := requireClientSecretForGrantType(c, AUTHCODE); err != nil {
		return err
	}

	if err := requireClientSecretForGrantType(c, CLIENTCREDENTIALS); err != nil {
		return err
	}

	if err := requireRedirectURIForGrantType(c, IMPLICIT); err != nil {
		return err
	}

	return nil
}

type changeSecretBody struct {
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"secret,omitempty"`
}

// ChangeClientSecret updates the secret with the given value for the client
// with the given id
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#change-secret.
func (a *API) ChangeClientSecret(id string, newSecret string) error {
	u := urlWithPath(*a.TargetURL, fmt.Sprintf("%s/%s/secret", ClientsEndpoint, id))
	change := &changeSecretBody{ClientID: id, ClientSecret: newSecret}
	j, err := json.Marshal(change)
	if err != nil {
		return err
	}
	err = a.doJSON(http.MethodPut, &u, bytes.NewBuffer([]byte(j)), nil, true)
	if err != nil {
		return err
	}
	return nil
}

// ChangeClientJWTMode is the operation mode for ChangeClientJWT.
type ChangeClientJWTMode string

const (
	ChangeClientJWTModeAdd    = ChangeClientJWTMode("ADD")
	ChangeClientJWTModeUpdate = ChangeClientJWTMode("UPDATE")
	ChangeClientJWTModeDelete = ChangeClientJWTMode("DELETE")
)

// ClientJWTChangeRequest is the request body for the PUT /oauth/clients/{id}/clientjwt endpoint.
type ClientJWTChangeRequest struct {
	ClientID   string              `json:"client_id,omitempty"`
	ChangeMode ChangeClientJWTMode `json:"changeMode,omitempty"`
	JwksURI    string              `json:"jwks_uri,omitempty"`
	Jwks       json.RawMessage     `json:"jwks,omitempty"`
	Kid        string              `json:"kid,omitempty"`
	Issuer     string              `json:"iss,omitempty"`
	Subject    string              `json:"sub,omitempty"`
	Audience   string              `json:"aud,omitempty"`
}

// ChangeClientJWT configures the JWT trust for the client with the given id.
// Use jwks_uri or jwks to specify public keys; use iss/sub/aud for federation JWT trust.
// changeMode controls whether the key is ADDed, UPDATEd, or DELETEd (kid required for DELETE).
func (a *API) ChangeClientJWT(req ClientJWTChangeRequest) error {
	if req.ClientID == "" {
		return errorMissingValue("client_id")
	}
	if req.ChangeMode == ChangeClientJWTModeDelete && req.Kid == "" {
		return fmt.Errorf("kid must be specified when changeMode is %v", ChangeClientJWTModeDelete)
	}
	u := urlWithPath(*a.TargetURL, fmt.Sprintf("%s/%s/clientjwt", ClientsEndpoint, req.ClientID))
	j, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return a.doJSON(http.MethodPut, &u, bytes.NewBuffer(j), nil, true)
}
