package uaa

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type uaaTransport struct {
	Transport      http.RoundTripper
	LoggingEnabled bool
}

func (t *uaaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.logRequest(req)

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.logResponse(resp)

	return resp, err
}

func newLoggingTransport(loggingEnabled bool) *uaaTransport {
	return &uaaTransport{LoggingEnabled: loggingEnabled}
}

func (t *uaaTransport) logRequest(req *http.Request) {
	if t.LoggingEnabled {
		bytes, _ := httputil.DumpRequest(req, false)
		fmt.Printf(string(bytes))
	}
}

func (t *uaaTransport) logResponse(resp *http.Response) {
	if t.LoggingEnabled {
		bytes, _ := httputil.DumpResponse(resp, true)
		fmt.Printf(string(bytes))
	}
}

func (t *uaaTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}
