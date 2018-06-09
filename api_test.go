package uaa_test

import (
	"testing"

	uaa "github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
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
			api, err := uaa.NewWithClientCredentials("(*#&^@%$&%)", "", "", "")
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("returns an API with a TargetURL", func() {
			api, err := uaa.NewWithClientCredentials("https://example.net", "", "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal("https://example.net"))
		})

		it("returns an API with an HTTPClient", func() {
			api, err := uaa.NewWithClientCredentials("https://example.net", "", "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.AuthenticatedClient).NotTo(BeNil())
		})
	})

	when("NewWithPasswordCredentials()", func() {
		it("fails if the target url is invalid", func() {
			api, err := uaa.NewWithPasswordCredentials("(*#&^@%$&%)", "", "", "", "", "")
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("returns an API with a TargetURL", func() {
			api, err := uaa.NewWithPasswordCredentials("https://example.net", "", "", "", "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal("https://example.net"))
		})

		it("returns an API with an HTTPClient", func() {
			api, err := uaa.NewWithPasswordCredentials("https://example.net", "", "", "", "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.AuthenticatedClient).NotTo(BeNil())
		})
	})
}
