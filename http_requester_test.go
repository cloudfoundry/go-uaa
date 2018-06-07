package uaa_test

import (
	"net/http"

	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HttpGetter", func() {
	var (
		server       *ghttp.Server
		client       *http.Client
		config       Config
		responseJSON string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		client = &http.Client{}
		config = NewConfigWithServerURL(server.URL())
		responseJSON = `{"foo": "bar"}`
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("UnauthenticatedRequester", func() {
		Describe("Get", func() {
			It("calls an endpoint with Accept application/json header", func() {
				server.RouteToHandler("GET", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("GET", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
				))

				UnauthenticatedRequestor{}.Get(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("supports zone switching", func() {
				server.RouteToHandler("GET", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))
				config.ZoneSubdomain = "twilight-zone"
				UnauthenticatedRequestor{}.Get(client, config, "/testPath", "someQueryParam=true")
			})

			It("returns helpful error when GET request fails", func() {
				server.RouteToHandler("GET", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(500, ""),
					ghttp.VerifyRequest("GET", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
				))

				_, err := UnauthenticatedRequestor{}.Get(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})
		})

		Describe("Delete", func() {
			It("calls an endpoint with Accept application/json header", func() {
				server.RouteToHandler("DELETE", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("DELETE", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
				))

				UnauthenticatedRequestor{}.Delete(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("supports zone switching", func() {
				server.RouteToHandler("DELETE", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))
				config.ZoneSubdomain = "twilight-zone"
				UnauthenticatedRequestor{}.Delete(client, config, "/testPath", "someQueryParam=true")
			})

			It("returns helpful error when DELETE request fails", func() {
				server.RouteToHandler("DELETE", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(500, ""),
					ghttp.VerifyRequest("DELETE", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
				))

				_, err := UnauthenticatedRequestor{}.Delete(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})
		})

		Describe("PostForm", func() {
			It("calls an endpoint with correct body and headers", func() {
				responseJSON = `{
				  "access_token" : "bc4885d950854fed9a938e96b13ca519",
				  "token_type" : "bearer",
				  "expires_in" : 43199,
				  "scope" : "clients.read emails.write scim.userids password.write idps.write notifications.write oauth.login scim.write critical_notifications.write",
				  "jti" : "bc4885d950854fed9a938e96b13ca519"
				}`

				server.RouteToHandler("POST", "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("POST", "/oauth/token", ""),
					ghttp.VerifyBody([]byte("hello=world")),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
				))

				body := map[string]string{"hello": "world"}
				returnedBytes, _ := UnauthenticatedRequestor{}.PostForm(client, config, "/oauth/token", "", body)
				parsedResponse := string(returnedBytes)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(parsedResponse).To(ContainSubstring("expires_in"))
			})

			It("treats 201 as success", func() {
				server.RouteToHandler("POST", "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(201, responseJSON),
					ghttp.VerifyRequest("POST", "/oauth/token", ""),
				))

				_, err := UnauthenticatedRequestor{}.PostForm(client, config, "/oauth/token", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).To(BeNil())
			})

			It("treats 405 as error", func() {
				server.RouteToHandler("PUT", "/oauth/token/foo/secret", ghttp.CombineHandlers(
					ghttp.RespondWith(405, responseJSON),
					ghttp.VerifyRequest("PUT", "/oauth/token/foo/secret", ""),
				))

				_, err := UnauthenticatedRequestor{}.PutJSON(client, config, "/oauth/token/foo/secret", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
			})

			It("returns an error when request fails", func() {
				server.RouteToHandler("POST", "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(500, "garbage"),
					ghttp.VerifyRequest("POST", "/oauth/token", ""),
				))

				_, err := UnauthenticatedRequestor{}.PostForm(client, config, "/oauth/token", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})

			It("supports zone switching", func() {
				server.RouteToHandler("POST", "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(201, responseJSON),
					ghttp.VerifyRequest("POST", "/oauth/token", ""),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))

				config.ZoneSubdomain = "twilight-zone"
				_, err := UnauthenticatedRequestor{}.PostForm(client, config, "/oauth/token", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).To(BeNil())
			})
		})

		Describe("PostJSON", func() {
			It("calls an endpoint with correct body and headers", func() {
				responseJSON = `{ "status" : "great successs" }`

				server.RouteToHandler("POST", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("POST", "/foo", ""),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Content-Type", "application/json"),
					ghttp.VerifyJSON(`{"Field1": "hello", "Field2": "world"}`),
				))

				bodyObj := TestData{Field1: "hello", Field2: "world"}

				returnedBytes, _ := UnauthenticatedRequestor{}.PostJSON(client, config, "/foo", "", bodyObj)
				parsedResponse := string(returnedBytes)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(parsedResponse).To(ContainSubstring("great success"))
			})

			It("returns an error when request fails", func() {
				server.RouteToHandler("POST", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(500, "garbage"),
					ghttp.VerifyRequest("POST", "/foo", ""),
				))

				bodyObj := TestData{Field1: "hello", Field2: "world"}
				_, err := UnauthenticatedRequestor{}.PostJSON(client, config, "/foo", "", bodyObj)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})

			It("supports zone switching", func() {
				server.RouteToHandler("POST", "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(201, responseJSON),
					ghttp.VerifyRequest("POST", "/oauth/token", ""),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))

				config.ZoneSubdomain = "twilight-zone"
				_, err := UnauthenticatedRequestor{}.PostJSON(client, config, "/oauth/token", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).To(BeNil())
			})
		})

		Describe("PutJSON", func() {
			It("calls an endpoint with correct body and headers", func() {
				responseJSON = `{ "status" : "great successs" }`

				server.RouteToHandler("PUT", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("PUT", "/foo", ""),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Content-Type", "application/json"),
					ghttp.VerifyJSON(`{"Field1": "hello", "Field2": "world"}`),
				))

				bodyObj := TestData{Field1: "hello", Field2: "world"}

				returnedBytes, _ := UnauthenticatedRequestor{}.PutJSON(client, config, "/foo", "", bodyObj)
				parsedResponse := string(returnedBytes)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(parsedResponse).To(ContainSubstring("great success"))
			})

			It("returns an error when request fails", func() {
				server.RouteToHandler("PUT", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(500, "garbage"),
					ghttp.VerifyRequest("PUT", "/foo", ""),
				))

				bodyObj := TestData{Field1: "hello", Field2: "world"}
				_, err := UnauthenticatedRequestor{}.PutJSON(client, config, "/foo", "", bodyObj)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})

			It("supports zone switching", func() {
				responseJSON = `{ "status" : "great successs" }`

				server.RouteToHandler("PUT", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("PUT", "/foo", ""),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))

				config.ZoneSubdomain = "twilight-zone"
				UnauthenticatedRequestor{}.PutJSON(client, config, "/foo", "", TestData{Field1: "hello", Field2: "world"})
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})
	})

	Describe("AuthenticatedRequester", func() {
		Describe("Get", func() {
			It("calls an endpoint with Accept and Authorization headers", func() {
				server.RouteToHandler("GET", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("GET", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				))

				config.AddContext(NewContextWithToken("access_token"))
				AuthenticatedRequestor{}.Get(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("supports zone switching", func() {
				server.RouteToHandler("GET", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("GET", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))

				config.AddContext(NewContextWithToken("access_token"))
				config.ZoneSubdomain = "twilight-zone"
				AuthenticatedRequestor{}.Get(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns a helpful error when GET request fails", func() {
				server.RouteToHandler("GET", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(500, ""),
					ghttp.VerifyRequest("GET", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
				))

				config.AddContext(NewContextWithToken("access_token"))
				_, err := AuthenticatedRequestor{}.Get(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})

			It("returns a helpful error when no token in context", func() {
				config.AddContext(NewContextWithToken(""))
				_, err := AuthenticatedRequestor{}.Get(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(0))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An access token is required to call"))
			})
		})

		Describe("Delete", func() {
			It("calls an endpoint with Accept and Authorization headers", func() {
				server.RouteToHandler("DELETE", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("DELETE", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				))

				config.AddContext(NewContextWithToken("access_token"))
				AuthenticatedRequestor{}.Delete(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("supports zone switching", func() {
				server.RouteToHandler("DELETE", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("DELETE", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))

				config.AddContext(NewContextWithToken("access_token"))
				config.ZoneSubdomain = "twilight-zone"
				AuthenticatedRequestor{}.Delete(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns a helpful error when DELETE request fails", func() {
				server.RouteToHandler("DELETE", "/testPath", ghttp.CombineHandlers(
					ghttp.RespondWith(500, ""),
					ghttp.VerifyRequest("DELETE", "/testPath", "someQueryParam=true"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
				))

				config.AddContext(NewContextWithToken("access_token"))
				_, err := AuthenticatedRequestor{}.Delete(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})

			It("returns a helpful error when no token in context", func() {
				config.AddContext(NewContextWithToken(""))
				_, err := AuthenticatedRequestor{}.Delete(client, config, "/testPath", "someQueryParam=true")

				Expect(server.ReceivedRequests()).To(HaveLen(0))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An access token is required to call"))
			})
		})

		Describe("PostForm", func() {
			It("calls an endpoint with correct body and headers", func() {
				responseJSON = `{
				  "access_token" : "bc4885d950854fed9a938e96b13ca519",
				  "token_type" : "bearer",
				  "expires_in" : 43199,
				  "scope" : "clients.read emails.write scim.userids password.write idps.write notifications.write oauth.login scim.write critical_notifications.write",
				  "jti" : "bc4885d950854fed9a938e96b13ca519"
				}`

				server.RouteToHandler("POST", "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("POST", "/oauth/token", ""),
					ghttp.VerifyBody([]byte("hello=world")),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Content-Type", "application/x-www-form-urlencoded"),
				))

				body := map[string]string{"hello": "world"}
				config.AddContext(NewContextWithToken("access_token"))

				returnedBytes, _ := AuthenticatedRequestor{}.PostForm(client, config, "/oauth/token", "", body)
				parsedResponse := string(returnedBytes)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(parsedResponse).To(ContainSubstring("expires_in"))
			})

			It("supports zone switching", func() {
				server.RouteToHandler("POST", "/oauth/token", ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token", ""),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))

				config.AddContext(NewContextWithToken("access_token"))
				config.ZoneSubdomain = "twilight-zone"

				AuthenticatedRequestor{}.PostForm(client, config, "/oauth/token", "", map[string]string{})
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error when request fails", func() {
				server.RouteToHandler("POST", "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(500, "garbage"),
					ghttp.VerifyRequest("POST", "/oauth/token", ""),
				))

				config.AddContext(NewContextWithToken("access_token"))
				_, err := AuthenticatedRequestor{}.PostForm(client, config, "/oauth/token", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})

			It("returns a helpful error when no token in context", func() {
				config.AddContext(NewContextWithToken(""))
				_, err := AuthenticatedRequestor{}.PostForm(client, config, "/oauth/token", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(0))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An access token is required to call"))
			})
		})

		Describe("PostJSON", func() {
			It("calls an endpoint with correct body and headers", func() {
				responseJSON = `{ "status" : "great successs" }`

				server.RouteToHandler("POST", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("POST", "/foo", ""),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Content-Type", "application/json"),
					ghttp.VerifyJSON(`{"Field1": "hello", "Field2": "world"}`),
				))

				bodyObj := TestData{Field1: "hello", Field2: "world"}
				config.AddContext(NewContextWithToken("access_token"))

				returnedBytes, _ := AuthenticatedRequestor{}.PostJSON(client, config, "/foo", "", bodyObj)
				parsedResponse := string(returnedBytes)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(parsedResponse).To(ContainSubstring("great success"))
			})

			It("returns an error when request fails", func() {
				server.RouteToHandler("POST", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(500, "garbage"),
					ghttp.VerifyRequest("POST", "/foo", ""),
				))

				config.AddContext(NewContextWithToken("access_token"))
				bodyObj := TestData{Field1: "hello", Field2: "world"}
				_, err := AuthenticatedRequestor{}.PostJSON(client, config, "/foo", "", bodyObj)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})

			It("returns a helpful error when no token in context", func() {
				config.AddContext(NewContextWithToken(""))
				_, err := AuthenticatedRequestor{}.PostJSON(client, config, "/foo", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(0))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An access token is required to call"))
			})
		})

		Describe("PutJSON", func() {
			It("calls an endpoint with correct body and headers", func() {
				responseJSON = `{ "status" : "great successs" }`

				server.RouteToHandler("PUT", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(200, responseJSON),
					ghttp.VerifyRequest("PUT", "/foo", ""),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Content-Type", "application/json"),
					ghttp.VerifyJSON(`{"Field1": "hello", "Field2": "world"}`),
				))

				bodyObj := TestData{Field1: "hello", Field2: "world"}
				config.AddContext(NewContextWithToken("access_token"))

				returnedBytes, _ := AuthenticatedRequestor{}.PutJSON(client, config, "/foo", "", bodyObj)
				parsedResponse := string(returnedBytes)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(parsedResponse).To(ContainSubstring("great success"))
			})

			It("returns an error when request fails", func() {
				server.RouteToHandler("PUT", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(500, "garbage"),
					ghttp.VerifyRequest("PUT", "/foo", ""),
				))

				config.AddContext(NewContextWithToken("access_token"))
				bodyObj := TestData{Field1: "hello", Field2: "world"}
				_, err := AuthenticatedRequestor{}.PutJSON(client, config, "/foo", "", bodyObj)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})

			It("supports zone switching", func() {
				server.RouteToHandler("PUT", "/foo", ghttp.CombineHandlers(
					ghttp.RespondWith(200, `{ "status" : "great successs" }`),
					ghttp.VerifyRequest("PUT", "/foo", ""),
					ghttp.VerifyHeaderKV("X-Identity-Zone-Subdomain", "twilight-zone"),
				))

				config.AddContext(NewContextWithToken("access_token"))
				config.ZoneSubdomain = "twilight-zone"
				_, err := AuthenticatedRequestor{}.PutJSON(client, config, "/foo", "", TestData{Field1: "hello", Field2: "world"})
				Expect(err).To(BeNil())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns a helpful error when no token in context", func() {
				config.AddContext(NewContextWithToken(""))
				_, err := AuthenticatedRequestor{}.PutJSON(client, config, "/foo", "", map[string]string{})

				Expect(server.ReceivedRequests()).To(HaveLen(0))
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("An access token is required to call"))
			})
		})

	})
})
