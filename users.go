package uaa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	UserID        string `json:"userId,omitempty"`
	ClientID      string `json:"clientId,omitempty"`
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
	ExternalID           string        `json:"externalId,omitempty"`
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
	ZoneID               string        `json:"zoneId,omitempty"`
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

// GetUser with the given userID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#get-3.
func (a *API) GetUser(userID string) (*User, error) {
	u := urlWithPath(*a.TargetURL, fmt.Sprintf("%s/%s", usersEndpoint, userID))
	user := &User{}
	err := a.doJSON(http.MethodGet, &u, nil, user, true)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByUsername gets the user with the given username
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#list-with-attribute-filtering.
func (a *API) GetUserByUsername(username, origin, attributes string) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be blank")
	}

	filter := fmt.Sprintf(`userName eq "%v"`, username)
	help := fmt.Sprintf("user %v not found", username)

	if origin != "" {
		filter = fmt.Sprintf(`%s and origin eq "%v"`, filter, origin)
		help = fmt.Sprintf(`%s in origin %v`, help, origin)
	}

	users, err := a.ListAllUsers(filter, "", attributes, "")
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New(help)
	}
	if len(users) > 1 && origin == "" {
		var foundOrigins []string
		for _, user := range users {
			foundOrigins = append(foundOrigins, user.Origin)
		}

		msgTmpl := "Found users with username %v in multiple origins %v."
		msg := fmt.Sprintf(msgTmpl, username, stringSliceStringifier(foundOrigins))
		return nil, errors.New(msg)
	}
	return &users[0], nil
}

func stringSliceStringifier(stringsList []string) string {
	return "[" + strings.Join(stringsList, ", ") + "]"
}

// SortOrder defines the sort order when listing users or groups.
type SortOrder string

const (
	// SortAscending sorts in ascending order.
	SortAscending = SortOrder("ascending")
	// SortDescending sorts in descending order.
	SortDescending = SortOrder("descending")
)

// ListAllUsers retrieves UAA users
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#list-with-attribute-filtering.
func (a *API) ListAllUsers(filter, sortBy, attributes string, sortOrder SortOrder) ([]User, error) {
	page := Page{
		StartIndex:   1,
		ItemsPerPage: 100,
	}
	var (
		results     []User
		currentPage []User
		err         error
	)

	for {
		currentPage, page, err = a.ListUsers(filter, sortBy, attributes, sortOrder, page.StartIndex, page.ItemsPerPage)
		if err != nil {
			return nil, err
		}
		results = append(results, currentPage...)

		if (page.StartIndex + page.ItemsPerPage) > page.TotalResults {
			break
		}
		page.StartIndex = page.StartIndex + page.ItemsPerPage
	}
	return results, nil
}

// Page represents a page of information returned from the UAA API.
type Page struct {
	StartIndex   int
	ItemsPerPage int
	TotalResults int
}

// ListUsers with the given filter, sortBy, attributes, sortOrder, startIndex
// (1-based), and count (default 100).
// If successful, ListUsers returns the users and the total itemsPerPage of users for
// all pages. If unsuccessful, ListUsers returns the error.
func (a *API) ListUsers(filter string, sortBy string, attributes string, sortOrder SortOrder, startIndex int, itemsPerPage int) ([]User, Page, error) {
	u := urlWithPath(*a.TargetURL, usersEndpoint)
	query := url.Values{}
	if filter != "" {
		query.Set("filter", filter)
	}
	if attributes != "" {
		query.Set("attributes", attributes)
	}
	if sortBy != "" {
		query.Set("sortBy", sortBy)
	}
	if sortOrder != "" {
		query.Set("sortOrder", string(sortOrder))
	}
	if startIndex <= 0 {
		startIndex = 1
	}
	query.Set("startIndex", strconv.Itoa(startIndex))
	if itemsPerPage <= 0 {
		itemsPerPage = 100
	}
	query.Set("count", strconv.Itoa(itemsPerPage))
	u.RawQuery = query.Encode()

	users := &PaginatedUserList{}
	err := a.doJSON(http.MethodGet, &u, nil, users, true)
	if err != nil {
		return nil, Page{}, err
	}
	page := Page{
		StartIndex:   users.StartIndex,
		ItemsPerPage: users.ItemsPerPage,
		TotalResults: users.TotalResults,
	}
	return users.Resources, page, err
}

// CreateUser creates the given user
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#create-4.
func (a *API) CreateUser(user User) (*User, error) {
	u := urlWithPath(*a.TargetURL, usersEndpoint)
	created := &User{}
	j, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	err = a.doJSON(http.MethodPost, &u, bytes.NewBuffer([]byte(j)), created, true)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateUser updates the given user
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#update-4.
func (a *API) UpdateUser(user User) (*User, error) {
	u := urlWithPath(*a.TargetURL, usersEndpoint)
	created := &User{}
	j, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	err = a.doJSON(http.MethodPut, &u, bytes.NewBuffer([]byte(j)), created, true)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// DeleteUser deletes the user with the given user ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#delete-4.
func (a *API) DeleteUser(userID string) (*User, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be blank")
	}
	u := urlWithPath(*a.TargetURL, fmt.Sprintf("%s/%s", usersEndpoint, userID))
	deleted := &User{}
	err := a.doJSON(http.MethodDelete, &u, nil, deleted, true)
	if err != nil {
		return nil, err
	}
	return deleted, nil
}

// DeactivateUser deactivates the user with the given user ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#patch.
func (a *API) DeactivateUser(userID string, userMetaVersion int) error {
	return a.setActive(false, userID, userMetaVersion)
}

// ActivateUser activates the user with the given user ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#patch.
func (a *API) ActivateUser(userID string, userMetaVersion int) error {
	return a.setActive(true, userID, userMetaVersion)
}

func (a *API) setActive(active bool, userID string, userMetaVersion int) error {
	if userID == "" {
		return errors.New("userID cannot be blank")
	}
	u := urlWithPath(*a.TargetURL, fmt.Sprintf("%s/%s", usersEndpoint, userID))
	user := &User{}
	user.Active = &active

	extraHeaders := map[string]string{"If-Match": strconv.Itoa(userMetaVersion)}
	j, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return a.doJSONWithHeaders(http.MethodPatch, &u, extraHeaders, bytes.NewBuffer([]byte(j)), nil, true)
}
