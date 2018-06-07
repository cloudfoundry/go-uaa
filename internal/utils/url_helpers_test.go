package utils_test

import (
	"github.com/cloudfoundry-community/uaa/internal/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UrlHelpers", func() {
	Describe("BuildUrl", func() {
		It("adds path to base url", func() {
			url, _ := utils.BuildURL("http://localhost:8080", "foo")
			Expect(url.String()).To(Equal("http://localhost:8080/foo"))

			url, _ = utils.BuildURL("http://localhost:8080/", "foo")
			Expect(url.String()).To(Equal("http://localhost:8080/foo"))

			url, _ = utils.BuildURL("http://localhost:8080/", "/foo")
			Expect(url.String()).To(Equal("http://localhost:8080/foo"))

			url, _ = utils.BuildURL("http://localhost:8080", "/foo")
			Expect(url.String()).To(Equal("http://localhost:8080/foo"))
		})
	})
})
