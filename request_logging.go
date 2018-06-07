package uaa

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/cloudfoundry-community/uaa/internal/utils"
)

func logResponse(response *http.Response) {
	dumped, _ := httputil.DumpResponse(response, true)

	if is2XX(response.StatusCode) {
		fmt.Println(utils.Green(string(dumped)) + "\n")
	} else {
		fmt.Println(utils.Red(string(dumped)) + "\n")
	}
}

func logRequest(request *http.Request) {
	dumped, _ := httputil.DumpRequest(request, true)
	fmt.Println(string(dumped))
}
