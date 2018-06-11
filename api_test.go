package uaa_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	uaa "github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

func TestNew(t *testing.T) {
	spec.Run(t, "New", testNew, spec.Report(report.Terminal{}))
}

func testNew(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("NewWithClientCredentials()", func() {
		it("fails if the target url is invalid", func() {
			api, err := uaa.NewWithClientCredentials("(*#&^@%$&%)", "", "", "", uaa.OpaqueToken)
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("returns an API with a TargetURL", func() {
			api, err := uaa.NewWithClientCredentials("https://example.net", "", "", "", uaa.OpaqueToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal("https://example.net"))
		})

		it("returns an API with an HTTPClient", func() {
			api, err := uaa.NewWithClientCredentials("https://example.net", "", "", "", uaa.OpaqueToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.AuthenticatedClient).NotTo(BeNil())
		})
	})

	when("NewWithPasswordCredentials()", func() {
		it("fails if the target url is invalid", func() {
			api, err := uaa.NewWithPasswordCredentials("(*#&^@%$&%)", "", "", "", "", "", uaa.OpaqueToken)
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("returns an API with a TargetURL", func() {
			api, err := uaa.NewWithPasswordCredentials("https://example.net", "", "", "", "", "", uaa.OpaqueToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal("https://example.net"))
		})

		it("returns an API with an HTTPClient", func() {
			api, err := uaa.NewWithPasswordCredentials("https://example.net", "", "", "", "", "", uaa.OpaqueToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.AuthenticatedClient).NotTo(BeNil())
		})
	})

	when("NewWithAuthorizationCode", func() {
		var (
			s           *httptest.Server
			returnToken bool
		)

		it.Before(func() {
			returnToken = true
			s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				t := &oauth2.Token{
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
					TokenType:    "bearer",
					Expiry:       time.Now().Add(60 * time.Second),
				}
				if !returnToken {
					t = nil
				}
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(t)
				Expect(err).NotTo(HaveOccurred())
			}))
		})

		it.After(func() {
			if s != nil {
				s.Close()
			}
		})

		it("fails if the target url is invalid", func() {
			api, err := uaa.NewWithAuthorizationCode("(*#&^@%$&%)", "", "", "", "", false, uaa.OpaqueToken)
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("returns an API with a TargetURL", func() {
			api, err := uaa.NewWithAuthorizationCode(s.URL, "", "", "", "", false, uaa.OpaqueToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal(s.URL))
		})

		it("returns an API with an HTTPClient", func() {
			api, err := uaa.NewWithAuthorizationCode(s.URL, "", "", "", "", false, uaa.OpaqueToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.AuthenticatedClient).NotTo(BeNil())
		})

		it("returns an error if the token cannot be retrieved", func() {
			returnToken = false
			api, err := uaa.NewWithAuthorizationCode(s.URL, "", "", "", "", false, uaa.OpaqueToken)
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})
	})
}
