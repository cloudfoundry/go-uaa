package uaa

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"strings"
)

// CurlManager allows you to make arbitrary requests to the UAA API.
type CurlManager struct {
	HTTPClient *http.Client
	Config     Config
}

// Curl makes a request to the UAA API with the given path, method, data, and
// headers.
func (cm CurlManager) Curl(path, method, data string, headers []string) (resHeaders, resBody string, err error) {
	target := cm.Config.GetActiveTarget()
	context := target.GetActiveContext()

	url, err := buildURL(target.BaseURL, path)
	if err != nil {
		return
	}

	req, err := http.NewRequest(method, url.String(), strings.NewReader(data))
	if err != nil {
		return
	}
	err = mergeHeaders(req.Header, strings.Join(headers, "\n"))
	if err != nil {
		return
	}
	req, err = addAuthorization(req, context)
	if err != nil {
		return
	}

	if cm.Config.Verbose {
		logRequest(req)
	}

	resp, err := cm.HTTPClient.Do(req)
	if err != nil {
		if cm.Config.Verbose {
			fmt.Printf("%v\n\n", err)
		}
		return
	}
	defer resp.Body.Close()

	headerBytes, _ := httputil.DumpResponse(resp, false)
	resHeaders = string(headerBytes)

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil && cm.Config.Verbose {
		fmt.Printf("%v\n\n", err)
	}
	resBody = string(bytes)

	if cm.Config.Verbose {
		logResponse(resp)
	}

	return
}

func mergeHeaders(destination http.Header, headerString string) (err error) {
	headerString = strings.TrimSpace(headerString)
	headerString += "\n\n"
	headerReader := bufio.NewReader(strings.NewReader(headerString))
	headers, err := textproto.NewReader(headerReader).ReadMIMEHeader()
	if err != nil {
		return
	}

	for key, values := range headers {
		destination.Del(key)
		for _, value := range values {
			destination.Add(key, value)
		}
	}

	return
}
