package uaa

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const groupResource string = "/Groups"

// GroupMember is a user or a group.
type GroupMember struct {
	Origin string `json:"origin,omitempty"`
	Type   string `json:"type,omitempty"`
	Value  string `json:"value,omitempty"`
}

// Group is a container for users and groups.
type Group struct {
	ID          string        `json:"id,omitempty"`
	Meta        *Meta         `json:"meta,omitempty"`
	DisplayName string        `json:"displayName,omitempty"`
	ZoneID      string        `json:"zoneID,omitempty"`
	Description string        `json:"description,omitempty"`
	Members     []GroupMember `json:"members,omitempty"`
	Schemas     []string      `json:"schemas,omitempty"`
}

// PaginatedGroupList is the response from the API for a single page of groups.
type PaginatedGroupList struct {
	Resources    []Group  `json:"resources"`
	StartIndex   int32    `json:"startIndex"`
	ItemsPerPage int32    `json:"itemsPerPage"`
	TotalResults int32    `json:"totalResults"`
	Schemas      []string `json:"schemas"`
}

// GroupManager allows you to interact with the Groups resource.
type GroupManager struct {
	HTTPClient *http.Client
	Config     Config
}

// AddMember adds the user with the given ID to the group with the given ID.
func (gm GroupManager) AddMember(groupID, userID string) error {
	url := fmt.Sprintf("%s/%s/members", groupResource, groupID)
	membership := GroupMember{Origin: "uaa", Type: "USER", Value: userID}
	_, err := AuthenticatedRequestor{}.PostJSON(gm.HTTPClient, gm.Config, url, "", membership)
	if err != nil {
		return err
	}

	return nil
}

// Get the group with the given ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#retrieve-2.
func (gm GroupManager) Get(id string) (Group, error) {
	url := "/Groups/" + id
	bytes, err := AuthenticatedRequestor{}.Get(gm.HTTPClient, gm.Config, url, "")
	if err != nil {
		return Group{}, err
	}

	group := Group{}
	err = json.Unmarshal(bytes, &group)
	if err != nil {
		return Group{}, parseError(url, bytes)
	}

	return group, err
}

// GetByName gets the group with the given name
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#list-4.
func (gm GroupManager) GetByName(name, attributes string) (Group, error) {
	if name == "" {
		return Group{}, errors.New("group name may not be blank")
	}

	filter := fmt.Sprintf(`displayName eq "%v"`, name)
	groups, err := gm.List(filter, "", attributes, "", 0, 0)
	if err != nil {
		return Group{}, err
	}
	if len(groups.Resources) == 0 {
		return Group{}, fmt.Errorf("group %v not found", name)
	}
	return groups.Resources[0], nil
}

// List groups
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#list-4.
func (gm GroupManager) List(filter, sortBy, attributes string, sortOrder SortOrder, startIdx, count int) (PaginatedGroupList, error) {
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
	if count != 0 {
		query.Add("count", strconv.Itoa(count))
	}
	if startIdx != 0 {
		query.Add("startIndex", strconv.Itoa(startIdx))
	}
	if sortOrder != "" {
		query.Add("sortOrder", string(sortOrder))
	}

	bytes, err := AuthenticatedRequestor{}.Get(gm.HTTPClient, gm.Config, groupResource, query.Encode())
	if err != nil {
		return PaginatedGroupList{}, err
	}

	groupsResp := PaginatedGroupList{}
	err = json.Unmarshal(bytes, &groupsResp)
	if err != nil {
		return PaginatedGroupList{}, parseError(groupResource, bytes)
	}

	return groupsResp, err
}

// Create the given group
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#create-5.
func (gm GroupManager) Create(group Group) (Group, error) {
	bytes, err := AuthenticatedRequestor{}.PostJSON(gm.HTTPClient, gm.Config, groupResource, "", group)
	if err != nil {
		return Group{}, err
	}

	created := Group{}
	err = json.Unmarshal(bytes, &created)
	if err != nil {
		return Group{}, parseError(groupResource, bytes)
	}

	return created, err
}

// Update the given group
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#update-5.
func (gm GroupManager) Update(group Group) (Group, error) {
	url := "/Groups"
	bytes, err := AuthenticatedRequestor{}.PutJSON(gm.HTTPClient, gm.Config, url, "", group)
	if err != nil {
		return Group{}, err
	}

	updated := Group{}
	err = json.Unmarshal(bytes, &updated)
	if err != nil {
		return Group{}, parseError(url, bytes)
	}

	return updated, err
}

// Delete the group with the given ID
// http://docs.cloudfoundry.org/api/uaa/version/4.14.0/index.html#delete-5.
func (gm GroupManager) Delete(groupID string) (Group, error) {
	url := "/Groups/" + groupID
	bytes, err := AuthenticatedRequestor{}.Delete(gm.HTTPClient, gm.Config, url, "")
	if err != nil {
		return Group{}, err
	}

	deleted := Group{}
	err = json.Unmarshal(bytes, &deleted)
	if err != nil {
		return Group{}, parseError(url, bytes)
	}

	return deleted, err
}
