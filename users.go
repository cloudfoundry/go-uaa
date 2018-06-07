package uaa

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cloudfoundry-community/uaa/internal/utils"
)

const usersEndpoint string = "/Users"

// Meta describes the version and timestamps for a resource.
type Meta struct {
	Version      int    `json:"version,omitempty"`
	Created      string `json:"created,omitempty"`
	LastModified string `json:"lastModified,omitempty"`
}

// UserName is a person's name.
type UserName struct {
	FamilyName string `json:"familyName,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
}

// Email is an email address.
type Email struct {
	Value   string `json:"value,omitempty"`
	Primary *bool  `json:"primary,omitempty"`
}

// UserGroup is a group that a user belongs to.
type UserGroup struct {
	Value   string `json:"value,omitempty"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
}

// Approval is a record of the user's explicit approval or rejection for an
// application's request for delegated permissions.
type Approval struct {
	UserID        string `json:"userID,omitempty"`
	ClientID      string `json:"clientID,omitempty"`
	Scope         string `json:"scope,omitempty"`
	Status        string `json:"status,omitempty"`
	LastUpdatedAt string `json:"lastUpdatedAt,omitempty"`
	ExpiresAt     string `json:"expiresAt,omitempty"`
}

// PhoneNumber is a phone number for a user.
type PhoneNumber struct {
	Value string `json:"value"`
}

// User is a UAA user
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#get-3.
type User struct {
	ID                   string        `json:"id,omitempty"`
	Password             string        `json:"password,omitempty"`
	ExternalID           string        `json:"externalID,omitempty"`
	Meta                 *Meta         `json:"meta,omitempty"`
	Username             string        `json:"userName,omitempty"`
	Name                 *UserName     `json:"name,omitempty"`
	Emails               []Email       `json:"emails,omitempty"`
	Groups               []UserGroup   `json:"groups,omitempty"`
	Approvals            []Approval    `json:"approvals,omitempty"`
	PhoneNumbers         []PhoneNumber `json:"phoneNumbers,omitempty"`
	Active               *bool         `json:"active,omitempty"`
	Verified             *bool         `json:"verified,omitempty"`
	Origin               string        `json:"origin,omitempty"`
	ZoneID               string        `json:"zoneID,omitempty"`
	PasswordLastModified string        `json:"passwordLastModified,omitempty"`
	PreviousLogonTime    int           `json:"previousLogonTime,omitempty"`
	LastLogonTime        int           `json:"lastLogonTime,omitempty"`
	Schemas              []string      `json:"schemas,omitempty"`
}

// UserManager allows you to interact with the Users resource.
type UserManager struct {
	HTTPClient *http.Client
	Config     Config
}

// PaginatedUserList is the response from the API for a single page of users.
type PaginatedUserList struct {
	Resources    []User   `json:"resources"`
	StartIndex   int      `json:"startIndex"`
	ItemsPerPage int      `json:"itemsPerPage"`
	TotalResults int      `json:"totalResults"`
	Schemas      []string `json:"schemas"`
}

// Get the user with the given userID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#get-3.
func (um UserManager) Get(userID string) (User, error) {
	url := fmt.Sprintf("%s/%s", usersEndpoint, userID)
	bytes, err := AuthenticatedRequestor{}.Get(um.HTTPClient, um.Config, url, "")
	if err != nil {
		return User{}, err
	}

	user := User{}
	err = json.Unmarshal(bytes, &user)
	if err != nil {
		return User{}, parseError(url, bytes)
	}

	return user, err
}

// GetByUsername gets the user with the given username and origin
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#list-with-attribute-filtering.
func (um UserManager) GetByUsername(username, origin, attributes string) (User, error) {
	if username == "" {
		return User{}, errors.New("username may not be blank")
	}

	filter := fmt.Sprintf(`userName eq "%v"`, username)
	help := fmt.Sprintf("user %v not found", username)

	if origin != "" {
		filter = fmt.Sprintf(`%s and origin eq "%v"`, filter, origin)
		help = fmt.Sprintf(`%s in origin %v`, help, origin)
	}

	users, err := um.List(filter, "", attributes, "")
	if err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, errors.New(help)
	}
	if len(users) > 1 && origin == "" {
		var foundOrigins []string
		for _, user := range users {
			foundOrigins = append(foundOrigins, user.Origin)
		}

		msgTmpl := "Found users with username %v in multiple origins %v."
		msg := fmt.Sprintf(msgTmpl, username, utils.StringSliceStringifier(foundOrigins))
		return User{}, errors.New(msg)
	}
	return users[0], nil
}

// SortOrder defines the sort order when listing users or groups.
type SortOrder string

const (
	// SortAscending sorts in ascending order.
	SortAscending = SortOrder("ascending")
	// SortDescending sorts in descending order.
	SortDescending = SortOrder("descending")
)

func getUserPage(um UserManager, query url.Values, startIndex, count int) (PaginatedUserList, error) {
	if startIndex != 0 {
		query.Add("startIndex", strconv.Itoa(startIndex))
	}
	if count != 0 {
		query.Add("count", strconv.Itoa(count))
	}

	bytes, err := AuthenticatedRequestor{}.Get(um.HTTPClient, um.Config, usersEndpoint, query.Encode())
	if err != nil {
		return PaginatedUserList{}, err
	}

	userList := PaginatedUserList{}
	err = json.Unmarshal(bytes, &userList)
	if err != nil {
		return PaginatedUserList{}, parseError(usersEndpoint, bytes)
	}
	return userList, nil
}

// List users
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#list-with-attribute-filtering.
func (um UserManager) List(filter, sortBy, attributes string, sortOrder SortOrder) ([]User, error) {
	query := url.Values{}
	if filter != "" {
		query.Add("filter", filter)
	}
	if attributes != "" {
		query.Add("attributes", attributes)
	}
	if sortBy != "" {
		query.Add("sortBy", sortBy)
	}
	if sortOrder != "" {
		query.Add("sortOrder", string(sortOrder))
	}

	results, err := getUserPage(um, query, 0, 0)
	if err != nil {
		return []User{}, err
	}

	userList := results.Resources
	startIndex, count := results.StartIndex, results.ItemsPerPage
	for results.TotalResults > len(userList) {
		startIndex += count
		newResults, err := getUserPage(um, query, startIndex, count)
		if err != nil {
			return []User{}, err
		}
		userList = append(userList, newResults.Resources...)
	}

	return userList, nil
}

// Create the given user
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#create-4.
func (um UserManager) Create(user User) (User, error) {
	bytes, err := AuthenticatedRequestor{}.PostJSON(um.HTTPClient, um.Config, usersEndpoint, "", user)
	if err != nil {
		return User{}, err
	}

	created := User{}
	err = json.Unmarshal(bytes, &created)
	if err != nil {
		return User{}, parseError(usersEndpoint, bytes)
	}

	return created, err
}

// Update the given user
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#update-4.
func (um UserManager) Update(user User) (User, error) {
	bytes, err := AuthenticatedRequestor{}.PutJSON(um.HTTPClient, um.Config, usersEndpoint, "", user)
	if err != nil {
		return User{}, err
	}

	updated := User{}
	err = json.Unmarshal(bytes, &updated)
	if err != nil {
		return User{}, parseError(usersEndpoint, bytes)
	}

	return updated, err
}

// Delete the user with the given user ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#delete-4.
func (um UserManager) Delete(userID string) (User, error) {
	url := fmt.Sprintf("%s/%s", usersEndpoint, userID)
	bytes, err := AuthenticatedRequestor{}.Delete(um.HTTPClient, um.Config, url, "")
	if err != nil {
		return User{}, err
	}

	deleted := User{}
	err = json.Unmarshal(bytes, &deleted)
	if err != nil {
		return User{}, parseError(url, bytes)
	}

	return deleted, err
}

// Deactivate the user with the given user ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#patch.
func (um UserManager) Deactivate(userID string, userMetaVersion int) error {
	return um.setActive(false, userID, userMetaVersion)
}

// Activate the user with the given user ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#patch.
func (um UserManager) Activate(userID string, userMetaVersion int) error {
	return um.setActive(true, userID, userMetaVersion)
}

func (um UserManager) setActive(active bool, userID string, userMetaVersion int) error {
	url := fmt.Sprintf("%s/%s", usersEndpoint, userID)
	user := User{}
	user.Active = &active

	extraHeaders := map[string]string{"If-Match": strconv.Itoa(userMetaVersion)}
	_, err := AuthenticatedRequestor{}.PatchJSON(um.HTTPClient, um.Config, url, "", user, extraHeaders)

	return err
}
