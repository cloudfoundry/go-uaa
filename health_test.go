package uaa_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	uaa "github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testIsHealthy(t *testing.T, when spec.G, it spec.S) {
	var (
		s       *httptest.Server
		handler http.Handler
		called  int
		a       *uaa.API
	)

	it.Before(func() {
		RegisterTestingT(t)
		called = 0
		s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			called = called + 1
			Expect(handler).NotTo(BeNil())
			handler.ServeHTTP(w, req)
		}))
		a, _ = uaa.New(s.URL, uaa.WithNoAuthentication())
	})

	it.After(func() {
		if s != nil {
			s.Close()
		}
	})

	it("is healthy when a 200 response is received", func() {
		handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			Expect(req.URL.Path).To(Equal("/healthz"))
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("ok"))
			Expect(err).NotTo(HaveOccurred())
		})

		status, err := a.IsHealthy()
		Expect(status).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	it("is unhealthy when a non-200 response is received", func() {
		handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			Expect(req.URL.Path).To(Equal("/healthz"))
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("ok"))
			Expect(err).NotTo(HaveOccurred())
		})
		status, err := a.IsHealthy()
		Expect(status).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())
	})
}
