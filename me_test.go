package uaa_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	uaa "github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testMe(t *testing.T, when spec.G, it spec.S) {
	var (
		s            *httptest.Server
		handler      http.Handler
		called       int
		a            *uaa.API
		userinfoJSON string
	)

	it.Before(func() {
		RegisterTestingT(t)
		called = 0
		userinfoJSON = `{
		  "user_id": "d6ef6c2e-02f6-477a-a7c6-18e27f9a6e87",
		  "sub": "d6ef6c2e-02f6-477a-a7c6-18e27f9a6e87",
		  "user_name": "charlieb",
		  "given_name": "Charlie",
		  "family_name": "Brown",
		  "email": "charlieb@peanuts.com",
		  "phone_number": null,
		  "previous_logon_time": 1503123277743,
		  "name": "Charlie Brown"
		}`
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

	it("calls the /userinfo endpoint", func() {
		handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			Expect(req.Header.Get("Accept")).To(Equal("application/json"))
			Expect(req.URL.Path).To(Equal("/userinfo"))
			Expect(req.URL.Query().Get("scheme")).To(Equal("openid"))
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(userinfoJSON))
			Expect(err).NotTo(HaveOccurred())
		})

		userinfo, err := a.GetMe()
		Expect(err).NotTo(HaveOccurred())
		Expect(userinfo).NotTo(BeNil())
		Expect(called).To(Equal(1))
		Expect(userinfo.UserID).To(Equal("d6ef6c2e-02f6-477a-a7c6-18e27f9a6e87"))
		Expect(userinfo.Sub).To(Equal("d6ef6c2e-02f6-477a-a7c6-18e27f9a6e87"))
		Expect(userinfo.Username).To(Equal("charlieb"))
		Expect(userinfo.GivenName).To(Equal("Charlie"))
		Expect(userinfo.FamilyName).To(Equal("Brown"))
		Expect(userinfo.Email).To(Equal("charlieb@peanuts.com"))
	})

	it("returns helpful error when /userinfo request fails", func() {
		handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			Expect(req.Header.Get("Accept")).To(Equal("application/json"))
			Expect(req.URL.Path).To(Equal("/userinfo"))
			Expect(req.URL.Query().Get("scheme")).To(Equal("openid"))
			w.WriteHeader(http.StatusInternalServerError)
		})
		u, err := a.GetMe()
		Expect(err).To(HaveOccurred())
		Expect(u).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("An error occurred while calling"))
	})

	it("returns helpful error when /userinfo response can't be parsed", func() {
		handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			Expect(req.Header.Get("Accept")).To(Equal("application/json"))
			Expect(req.URL.Path).To(Equal("/userinfo"))
			Expect(req.URL.Query().Get("scheme")).To(Equal("openid"))
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("{unparsable-json-response}"))
			Expect(err).NotTo(HaveOccurred())
		})
		u, err := a.GetMe()
		Expect(err).To(HaveOccurred())
		Expect(u).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("An unknown error occurred while parsing response from"))
		Expect(err.Error()).To(ContainSubstring("Response was {unparsable-json-response}"))
	})
}
