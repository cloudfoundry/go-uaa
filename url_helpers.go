package uaa

import (
	"net/url"
)

// buildURL validates that the baseURL is valid and then sets the given path on
// it.
func buildURL(baseURL, path string) (*url.URL, error) {
	newURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	newURL.Path = path
	return newURL, nil
}
