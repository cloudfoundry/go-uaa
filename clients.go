package uaa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const clientsResource string = "/oauth/clients"

// ClientManager allows you to interact with the Clients resource.
type ClientManager struct {
	HTTPClient *http.Client
	Config     Config
}

// Client is a UAA client
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#clients.
type Client struct {
	ClientID             string   `json:"client_id,omitempty"`
	ClientSecret         string   `json:"client_secret,omitempty"`
	Scope                []string `json:"scope,omitempty"`
	ResourceIDs          []string `json:"resource_ids,omitempty"`
	AuthorizedGrantTypes []string `json:"authorized_grant_types,omitempty"`
	RedirectURI          []string `json:"redirect_uri,omitempty"`
	Authorities          []string `json:"authorities,omitempty"`
	TokenSalt            string   `json:"token_salt,omitempty"`
	AllowedProviders     []string `json:"allowedproviders,omitempty"`
	DisplayName          string   `json:"name,omitempty"`
	LastModified         int64    `json:"lastModified,omitempty"`
	RequiredUserGroups   []string `json:"required_user_groups,omitempty"`
	AccessTokenValidity  int64    `json:"access_token_validity,omitempty"`
	RefreshTokenValidity int64    `json:"refresh_token_validity,omitempty"`
}

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

	if err := requireClientSecretForGrantType(c, PASSWORD); err != nil {
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

// PaginatedClientList is the response from the API for a single page of clients.
type PaginatedClientList struct {
	Resources    []Client `json:"resources"`
	StartIndex   int      `json:"startIndex"`
	ItemsPerPage int      `json:"itemsPerPage"`
	TotalResults int      `json:"totalResults"`
	Schemas      []string `json:"schemas"`
}

// Get the client with the given ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#retrieve-3.
func (cm *ClientManager) Get(id string) (Client, error) {
	url := fmt.Sprintf("%s/%s", clientsResource, id)
	bytes, err := AuthenticatedRequestor{}.Get(cm.HTTPClient, cm.Config, url, "")
	if err != nil {
		return Client{}, err
	}

	c := Client{}
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return Client{}, parseError(url, bytes)
	}

	return c, err
}

// Delete the client with the given ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#delete-6.
func (cm *ClientManager) Delete(id string) (Client, error) {
	url := fmt.Sprintf("%s/%s", clientsResource, id)
	bytes, err := AuthenticatedRequestor{}.Delete(cm.HTTPClient, cm.Config, url, "")
	if err != nil {
		return Client{}, err
	}

	c := Client{}
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return Client{}, parseError(url, bytes)
	}

	return c, err
}

// Create the given client
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#create-6.
func (cm *ClientManager) Create(client Client) (Client, error) {
	bytes, err := AuthenticatedRequestor{}.PostJSON(cm.HTTPClient, cm.Config, clientsResource, "", client)
	if err != nil {
		return Client{}, err
	}

	c := Client{}
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return Client{}, parseError(clientsResource, bytes)
	}

	return c, err
}

// Update the given client
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#update-6.
func (cm *ClientManager) Update(client Client) (Client, error) {
	url := "/oauth/clients/" + client.ClientID
	bytes, err := AuthenticatedRequestor{}.PutJSON(cm.HTTPClient, cm.Config, url, "", client)
	if err != nil {
		return Client{}, err
	}

	c := Client{}
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return Client{}, parseError(url, bytes)
	}

	return c, err
}

// ChangeSecret updates the secret with the given value for the client
// with the given id
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#change-secret.
func (cm *ClientManager) ChangeSecret(id string, newSecret string) error {
	url := "/oauth/clients/" + id + "/secret"
	body := changeSecretBody{ClientID: id, ClientSecret: newSecret}
	_, err := AuthenticatedRequestor{}.PutJSON(cm.HTTPClient, cm.Config, url, "", body)
	return err
}

func getClientPage(cm *ClientManager, startIndex, count int) (PaginatedClientList, error) {
	query := fmt.Sprintf("startIndex=%v&count=%v", startIndex, count)
	if startIndex == 0 {
		query = ""
	}

	bytes, err := AuthenticatedRequestor{}.Get(cm.HTTPClient, cm.Config, "/oauth/clients", query)
	if err != nil {
		return PaginatedClientList{}, err
	}

	clientList := PaginatedClientList{}
	err = json.Unmarshal(bytes, &clientList)
	if err != nil {
		return PaginatedClientList{}, parseError("/oauth/clients", bytes)
	}
	return clientList, nil
}

// List all clients.
func (cm *ClientManager) List() ([]Client, error) {
	results, err := getClientPage(cm, 0, 0)
	if err != nil {
		return []Client{}, err
	}

	clientList := results.Resources
	startIndex, count := results.StartIndex, results.ItemsPerPage
	for results.TotalResults > len(clientList) {
		startIndex += count
		newResults, err := getClientPage(cm, startIndex, count)
		if err != nil {
			return []Client{}, err
		}
		clientList = append(clientList, newResults.Resources...)
	}

	return clientList, nil
}
