package uaa

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Requestor makes requests with a client.
type Requestor interface {
	Get(client *http.Client, config Config, path string, query string) ([]byte, error)
	Delete(client *http.Client, config Config, path string, query string) ([]byte, error)
	PostForm(client *http.Client, config Config, path string, query string, body map[string]string) ([]byte, error)
	PostJSON(client *http.Client, config Config, path string, query string, body interface{}) ([]byte, error)
	PutJSON(client *http.Client, config Config, path string, query string, body interface{}) ([]byte, error)
}

// UnauthenticatedRequestor makes requests that are unauthenticated.
type UnauthenticatedRequestor struct{}

// AuthenticatedRequestor makes requests that are authenticated.
type AuthenticatedRequestor struct{}

func is2XX(status int) bool {
	if status >= 200 && status < 300 {
		return true
	}
	return false
}

func addZoneSwitchHeader(req *http.Request, config *Config) {
	req.Header.Add("X-Identity-Zone-Subdomain", config.ZoneSubdomain)
}

func mapToURLValues(body map[string]string) url.Values {
	data := url.Values{}
	for key, val := range body {
		data.Add(key, val)
	}
	return data
}

func doAndRead(req *http.Request, client *http.Client, config Config) ([]byte, error) {
	if config.Verbose {
		logRequest(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		if config.Verbose {
			fmt.Printf("%v\n\n", err)
		}

		return []byte{}, requestError(req.URL.String())
	}

	if config.Verbose {
		logResponse(resp)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.Verbose {
			fmt.Printf("%v\n\n", err)
		}

		return []byte{}, unknownError()
	}

	if !is2XX(resp.StatusCode) {
		return []byte{}, requestError(req.URL.String())
	}
	return bytes, nil
}

// Get makes a get request.
func (ug UnauthenticatedRequestor) Get(client *http.Client, config Config, path string, query string) ([]byte, error) {
	req, err := unauthenticatedRequestFactory{}.Get(config.GetActiveTarget(), path, query)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// Get makes a get request.
func (ag AuthenticatedRequestor) Get(client *http.Client, config Config, path string, query string) ([]byte, error) {
	req, err := authenticatedRequestFactory{}.Get(config.GetActiveTarget(), path, query)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// Delete makes a delete request.
func (ug UnauthenticatedRequestor) Delete(client *http.Client, config Config, path string, query string) ([]byte, error) {
	req, err := unauthenticatedRequestFactory{}.Delete(config.GetActiveTarget(), path, query)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// Delete makes a delete request.
func (ag AuthenticatedRequestor) Delete(client *http.Client, config Config, path string, query string) ([]byte, error) {
	req, err := authenticatedRequestFactory{}.Delete(config.GetActiveTarget(), path, query)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// PostForm makes a post request.
func (ug UnauthenticatedRequestor) PostForm(client *http.Client, config Config, path string, query string, body map[string]string) ([]byte, error) {
	data := mapToURLValues(body)

	req, err := unauthenticatedRequestFactory{}.PostForm(config.GetActiveTarget(), path, query, &data)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// PostForm makes a post request.
func (ag AuthenticatedRequestor) PostForm(client *http.Client, config Config, path string, query string, body map[string]string) ([]byte, error) {
	data := mapToURLValues(body)

	req, err := authenticatedRequestFactory{}.PostForm(config.GetActiveTarget(), path, query, &data)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// PostJSON makes a post request.
func (ug UnauthenticatedRequestor) PostJSON(client *http.Client, config Config, path string, query string, body interface{}) ([]byte, error) {
	req, err := unauthenticatedRequestFactory{}.PostJSON(config.GetActiveTarget(), path, query, body)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// PostJSON makes a post request.
func (ag AuthenticatedRequestor) PostJSON(client *http.Client, config Config, path string, query string, body interface{}) ([]byte, error) {
	req, err := authenticatedRequestFactory{}.PostJSON(config.GetActiveTarget(), path, query, body)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// PutJSON makes a put request.
func (ug UnauthenticatedRequestor) PutJSON(client *http.Client, config Config, path string, query string, body interface{}) ([]byte, error) {
	req, err := unauthenticatedRequestFactory{}.PutJSON(config.GetActiveTarget(), path, query, body)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// PutJSON makes a put request.
func (ag AuthenticatedRequestor) PutJSON(client *http.Client, config Config, path string, query string, body interface{}) ([]byte, error) {
	req, err := authenticatedRequestFactory{}.PutJSON(config.GetActiveTarget(), path, query, body)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// PatchJSON makes a patch request.
func (ug UnauthenticatedRequestor) PatchJSON(client *http.Client, config Config, path string, query string, body interface{}) ([]byte, error) {
	req, err := unauthenticatedRequestFactory{}.PatchJSON(config.GetActiveTarget(), path, query, body)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	return doAndRead(req, client, config)
}

// PatchJSON makes a patch request.
func (ag AuthenticatedRequestor) PatchJSON(client *http.Client, config Config, path string, query string, body interface{}, extraHeaders map[string]string) ([]byte, error) {
	req, err := authenticatedRequestFactory{}.PatchJSON(config.GetActiveTarget(), path, query, body)
	if err != nil {
		return []byte{}, err
	}
	addZoneSwitchHeader(req, &config)
	for k, v := range extraHeaders {
		req.Header.Add(k, v)
	}
	return doAndRead(req, client, config)
}
