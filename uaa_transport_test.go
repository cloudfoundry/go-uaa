package uaa

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

type fakeTransport struct {
	roundtripper func(req *http.Request)
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if f.roundtripper != nil {
		f.roundtripper(req)
	}
	return nil, nil
}

func testUaaTransport(t *testing.T, when spec.G, it spec.S) {
	var (
		request   *http.Request
		transport *uaaTransport
	)
	it.Before(func() {
		RegisterTestingT(t)
		transport = &uaaTransport{
			base: &fakeTransport{roundtripper: func(req *http.Request) {
				request = req
			}},
			LoggingEnabled: false,
		}
	})

	it("can identify a nil baseTransport", func() {
		a := API{}
		Expect(a.baseTransportIsNil()).To(BeTrue())
		a.baseTransport = nil
		Expect(a.baseTransportIsNil()).To(BeTrue())
	})

	it("can identify a non-nil baseTransport", func() {
		a := API{baseTransport: &fakeTransport{}}
		Expect(a.baseTransportIsNil()).To(BeFalse())
	})

	it("adds X-CF-ENCODED-CREDENTIALS header when using basic auth", func() {
		req, _ := http.NewRequest("", "", nil)
		req.Header.Add("Authorization", "Basic ENCODEDCREDENTIALS")
		_, err := transport.RoundTrip(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(request).NotTo(BeNil())
		Expect(request.Header.Get("X-CF-ENCODED-CREDENTIALS")).To(Equal("true"))
	})

	it("does not add X-CF-ENCODED-CREDENTIALS header when not using basic auth", func() {
		req, _ := http.NewRequest("", "", nil)
		_, err := transport.RoundTrip(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(request).NotTo(BeNil())
		Expect(request.Header.Get("X-CF-ENCODED-CREDENTIALS")).To(BeEmpty())
	})
}
