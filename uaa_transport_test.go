package uaa_test

import (
	uaa "github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"net/http"
	"testing"
)

func testUaaTransport(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("adds X-CF-ENCODED-CREDENTIALS header when using basic auth", func() {
		transport := uaa.NewUaaTransport(false)
		request, _ := http.NewRequest("", "", nil)
		request.Header.Add("Authorization", "Basic ENCODEDCREDENTIALS")

		transport.RoundTrip(request)
		Expect(request.Header.Get("X-CF-ENCODED-CREDENTIALS")).To(Equal("true"))
	})

	it("does not add X-CF-ENCODED-CREDENTIALS header when not using basic auth", func() {
		transport := uaa.NewUaaTransport(false)
		request, _ := http.NewRequest("", "", nil)

		transport.RoundTrip(request)
		Expect(request.Header.Get("X-CF-ENCODED-CREDENTIALS")).To(BeEmpty())
	})
}
