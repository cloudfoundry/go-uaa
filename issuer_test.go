package uaa_test

import (
	"net/http"

	"github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Issuer", func() {
	var (
		server *ghttp.Server
		api    *uaa.API
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		server.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/.well-known/openid-configuration"),
			ghttp.RespondWithJSONEncoded(http.StatusOK, &uaa.OpenIDConfig{Issuer: "issuer"}),
		))

		target := server.URL()

		var err error
		api, err = uaa.New(target, uaa.WithNoAuthentication())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	It("return the issuer", func() {
		issuer, err := api.Issuer()
		Expect(err).NotTo(HaveOccurred())
		Expect(issuer).To(Equal("issuer"))
	})

	Context("when the server returns a non-200 status code", func() {
		BeforeEach(func() {
			server.Reset()
			server.AppendHandlers(
				ghttp.VerifyRequest("GET", "/.well-known/openid-configuration"),
				ghttp.RespondWith(http.StatusInternalServerError, nil),
			)
		})

		It("returns an error", func() {
			issuer, err := api.Issuer()
			Expect(err).To(HaveOccurred())
			Expect(issuer).To(Equal(""))
		})
	})

})
