package uaa

import (
	"net/http"
)

// HealthStatus is either ok or an error.
type HealthStatus string

const (
	// OK is healthy.
	OK = HealthStatus("ok")
	// ERROR is unhealthy.
	ERROR = HealthStatus("health_error")
)

// Health gets the health of the UAA API.
func Health(target Target) (HealthStatus, error) {
	url, err := buildURL(target.BaseURL, "healthz")
	if err != nil {
		return "", err
	}

	resp, err := http.Get(url.String())
	if err != nil {
		return "", nil
	}

	if resp.StatusCode == 200 {
		return OK, nil
	}

	return ERROR, nil
}
