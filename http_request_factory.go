package uaa

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cloudfoundry-community/uaa/internal/utils"
)

// httpRequestFactory is a request builder.
type httpRequestFactory interface {
	Get(Target, string, string) (*http.Request, error)
	Delete(Target, string, string) (*http.Request, error)
	PostForm(Target, string, string, *url.Values) (*http.Request, error)
	PostJSON(Target, string, string, interface{}) (*http.Request, error)
	PutJSON(Target, string, string, interface{}) (*http.Request, error)
	PatchJSON(Target, string, string, interface{}) (*http.Request, error)
}

// unauthenticatedRequestFactory builds requests that are unauthenticated.
type unauthenticatedRequestFactory struct{}

// authenticatedRequestFactory builds requests that are authenticated.
type authenticatedRequestFactory struct{}

// Get creates a get request with the given target, path and query.
func (urf unauthenticatedRequestFactory) Get(target Target, path string, query string) (*http.Request, error) {
	targetURL, err := utils.BuildURL(target.BaseURL, path)
	if err != nil {
		return nil, err
	}
	targetURL.RawQuery = query

	req, err := http.NewRequest("GET", targetURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	return req, nil
}

// Delete creates a delete request with the given target, path, and query.
func (urf unauthenticatedRequestFactory) Delete(target Target, path string, query string) (*http.Request, error) {
	targetURL, err := utils.BuildURL(target.BaseURL, path)
	if err != nil {
		return nil, err
	}
	targetURL.RawQuery = query

	req, err := http.NewRequest("DELETE", targetURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	return req, nil
}

// PostForm creates a post request with the given target, path, query, and data.
func (urf unauthenticatedRequestFactory) PostForm(target Target, path string, query string, data *url.Values) (*http.Request, error) {
	targetURL, err := utils.BuildURL(target.BaseURL, path)
	if err != nil {
		return nil, err
	}
	targetURL.RawQuery = query

	bodyBytes := []byte(data.Encode())
	req, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(bodyBytes)))

	return req, nil
}

// PostJSON creates a post request with the given target, path, query, and body.
func (urf unauthenticatedRequestFactory) PostJSON(target Target, path string, query string, body interface{}) (*http.Request, error) {
	targetURL, err := utils.BuildURL(target.BaseURL, path)
	if err != nil {
		return nil, err
	}
	targetURL.RawQuery = query

	j, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyBytes := []byte(j)
	req, err := http.NewRequest("POST", targetURL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(bodyBytes)))

	return req, nil
}

// PutJSON creates a put request with the given target, path, query, and body.
func (urf unauthenticatedRequestFactory) PutJSON(target Target, path string, query string, body interface{}) (*http.Request, error) {
	targetURL, err := utils.BuildURL(target.BaseURL, path)
	if err != nil {
		return nil, err
	}
	targetURL.RawQuery = query

	j, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyBytes := []byte(j)
	req, err := http.NewRequest("PUT", targetURL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(bodyBytes)))

	return req, nil
}

// PatchJSON creates a patch request with the given body
func (urf unauthenticatedRequestFactory) PatchJSON(target Target, path string, query string, body interface{}) (*http.Request, error) {
	targetURL, err := utils.BuildURL(target.BaseURL, path)
	if err != nil {
		return nil, err
	}
	targetURL.RawQuery = query

	j, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyBytes := []byte(j)
	req, err := http.NewRequest("PATCH", targetURL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(bodyBytes)))

	return req, nil
}

func addAuthorization(req *http.Request, ctx AuthContext) (*http.Request, error) {
	accessToken := ctx.AccessToken
	req.Header.Add("Authorization", "bearer "+accessToken)
	if accessToken == "" {
		return nil, errors.New("An access token is required to call " + req.URL.String())
	}

	return req, nil
}

// Get creates a get request with the given target, path and query.
func (arf authenticatedRequestFactory) Get(target Target, path string, query string) (*http.Request, error) {
	req, err := unauthenticatedRequestFactory{}.Get(target, path, query)

	if err != nil {
		return nil, err
	}

	return addAuthorization(req, target.GetActiveContext())
}

// Delete creates a delete request with the given target, path, and query.
func (arf authenticatedRequestFactory) Delete(target Target, path string, query string) (*http.Request, error) {
	req, err := unauthenticatedRequestFactory{}.Delete(target, path, query)

	if err != nil {
		return nil, err
	}

	return addAuthorization(req, target.GetActiveContext())
}

// PostForm creates a post request with the given target, path, query, and data.
func (arf authenticatedRequestFactory) PostForm(target Target, path string, query string, data *url.Values) (*http.Request, error) {
	req, err := unauthenticatedRequestFactory{}.PostForm(target, path, query, data)
	if err != nil {
		return nil, err
	}

	return addAuthorization(req, target.GetActiveContext())
}

// PostJSON creates a post request with the given target, path, query, and body.
func (arf authenticatedRequestFactory) PostJSON(target Target, path string, query string, body interface{}) (*http.Request, error) {
	req, err := unauthenticatedRequestFactory{}.PostJSON(target, path, query, body)
	if err != nil {
		return nil, err
	}

	return addAuthorization(req, target.GetActiveContext())
}

// PutJSON creates a put request with the given target, path, query, and body.
func (arf authenticatedRequestFactory) PutJSON(target Target, path string, query string, body interface{}) (*http.Request, error) {
	req, err := unauthenticatedRequestFactory{}.PutJSON(target, path, query, body)
	if err != nil {
		return nil, err
	}

	return addAuthorization(req, target.GetActiveContext())
}

// PatchJSON creates a patch request with the given body
func (arf authenticatedRequestFactory) PatchJSON(target Target, path string, query string, body interface{}) (*http.Request, error) {
	req, err := unauthenticatedRequestFactory{}.PatchJSON(target, path, query, body)
	if err != nil {
		return nil, err
	}

	return addAuthorization(req, target.GetActiveContext())
}
