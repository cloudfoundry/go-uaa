package utils

import (
	"net/url"
)

// BuildURL validates that the baseURL is valid and then sets the given path on
// it.
func BuildURL(baseURL, path string) (*url.URL, error) {
	newURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	newURL.Path = path
	return newURL, nil
}
