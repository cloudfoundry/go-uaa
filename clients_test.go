package uaa_test

import (
	"net/http"

	"github.com/cloudfoundry-community/go-uaa"
	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Clients", func() {
	var (
		server     *ghttp.Server
		config     uaa.Config
		httpClient *http.Client
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		httpClient = &http.Client{}
		config = uaa.NewConfigWithServerURL(server.URL())
		ctx := uaa.NewContextWithToken("access_token")
		config.AddContext(ctx)
	})

	Describe("UaaClient.PreCreateValidation()", func() {
		It("rejects empty grant types", func() {
			client := uaa.Client{}

			err := client.Validate()

			Expect(err.Error()).To(Equal(`grant type must be one of [authorization_code, implicit, password, client_credentials]`))
		})

		Describe("when authorization_code", func() {
			It("requires client_id", func() {
				client := uaa.Client{
					AuthorizedGrantTypes: []string{"authorization_code"},
					RedirectURI:          []string{"http://localhost:8080"},
					ClientSecret:         "secret",
				}

				err := client.Validate()

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("client_id must be specified in the client definition"))
			})

			It("requires redirect_uri", func() {
				client := uaa.Client{
					ClientID:             "myclient",
					AuthorizedGrantTypes: []string{"authorization_code"},
					ClientSecret:         "secret",
				}

				err := client.Validate()

				Expect(err.Error()).To(Equal("redirect_uri must be specified for authorization_code grant type"))
			})

			It("requires client_secret", func() {
				client := uaa.Client{
					ClientID:             "myclient",
					AuthorizedGrantTypes: []string{"authorization_code"},
					RedirectURI:          []string{"http://localhost:8080"},
				}

				err := client.Validate()

				Expect(err.Error()).To(Equal("client_secret must be specified for authorization_code grant type"))
			})
		})

		Describe("when implicit", func() {
			It("requires client_id", func() {
				client := uaa.Client{
					AuthorizedGrantTypes: []string{"implicit"},
					RedirectURI:          []string{"http://localhost:8080"},
				}

				err := client.Validate()

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("client_id must be specified in the client definition"))
			})

			It("requires redirect_uri", func() {
				client := uaa.Client{
					ClientID:             "myclient",
					AuthorizedGrantTypes: []string{"implicit"},
				}

				err := client.Validate()

				Expect(err.Error()).To(Equal("redirect_uri must be specified for implicit grant type"))
			})

			It("does not require client_secret", func() {
				client := uaa.Client{
					ClientID:             "someclient",
					AuthorizedGrantTypes: []string{"implicit"},
					RedirectURI:          []string{"http://localhost:8080"},
				}

				err := client.Validate()

				Expect(err).To(BeNil())
			})
		})

		Describe("when client_credentials", func() {
			It("requires client_id", func() {
				client := uaa.Client{
					AuthorizedGrantTypes: []string{"client_credentials"},
					ClientSecret:         "secret",
				}

				err := client.Validate()

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("client_id must be specified in the client definition"))
			})

			It("requires client_secret", func() {
				client := uaa.Client{
					ClientID:             "myclient",
					AuthorizedGrantTypes: []string{"client_credentials"},
				}

				err := client.Validate()

				Expect(err.Error()).To(Equal("client_secret must be specified for client_credentials grant type"))
			})
		})

		Describe("when password", func() {
			It("requires client_id", func() {
				client := uaa.Client{
					AuthorizedGrantTypes: []string{"password"},
					RedirectURI:          []string{"http://localhost:8080"},
					ClientSecret:         "secret",
				}

				err := client.Validate()

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("client_id must be specified in the client definition"))
			})

			It("requires client_secret", func() {
				client := uaa.Client{
					ClientID:             "myclient",
					AuthorizedGrantTypes: []string{"password"},
				}

				err := client.Validate()

				Expect(err.Error()).To(Equal("client_secret must be specified for password grant type"))
			})
		})
	})

	Describe("Get", func() {
		const GetClientResponseJSON string = `{
		  "scope" : [ "clients.read", "clients.write" ],
		  "client_id" : "clientid",
		  "resource_ids" : [ "none" ],
		  "authorized_grant_types" : [ "client_credentials" ],
		  "redirect_uri" : [ "http://ant.path.wildcard/**/passback/*", "http://test1.com" ],
		  "autoapprove" : [ "true" ],
		  "authorities" : [ "clients.read", "clients.write" ],
		  "token_salt" : "1SztLL",
		  "allowedproviders" : [ "uaa", "ldap", "my-saml-provider" ],
		  "name" : "My Client Name",
		  "lastModified" : 1502816030525,
		  "required_user_groups" : [ ]
		}`

		AfterEach(func() {
			server.Close()
		})

		It("calls the /oauth/clients endpoint", func() {
			server.RouteToHandler("GET", "/oauth/clients/clientid", ghttp.CombineHandlers(
				ghttp.RespondWith(200, GetClientResponseJSON),
				ghttp.VerifyRequest("GET", "/oauth/clients/clientid"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			clientResponse, _ := cm.Get("clientid")

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(clientResponse.Scope[0]).To(Equal("clients.read"))
			Expect(clientResponse.Scope[1]).To(Equal("clients.write"))
			Expect(clientResponse.ClientID).To(Equal("clientid"))
			Expect(clientResponse.ResourceIDs[0]).To(Equal("none"))
			Expect(clientResponse.AuthorizedGrantTypes[0]).To(Equal("client_credentials"))
			Expect(clientResponse.RedirectURI[0]).To(Equal("http://ant.path.wildcard/**/passback/*"))
			Expect(clientResponse.RedirectURI[1]).To(Equal("http://test1.com"))
			Expect(clientResponse.TokenSalt).To(Equal("1SztLL"))
			Expect(clientResponse.AllowedProviders[0]).To(Equal("uaa"))
			Expect(clientResponse.AllowedProviders[1]).To(Equal("ldap"))
			Expect(clientResponse.AllowedProviders[2]).To(Equal("my-saml-provider"))
			Expect(clientResponse.DisplayName).To(Equal("My Client Name"))
			Expect(clientResponse.LastModified).To(Equal(int64(1502816030525)))
		})

		It("returns helpful error when /oauth/clients request fails", func() {
			server.RouteToHandler("GET", "/oauth/clients/clientid", ghttp.CombineHandlers(
				ghttp.RespondWith(500, ""),
				ghttp.VerifyRequest("GET", "/oauth/clients/clientid"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			_, err := cm.Get("clientid")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
		})

		It("can handle boolean response in autoapprove field", func() {
			const ResponseWithBooleanAutoapprove string = `{
			  "scope" : [ "clients.read", "clients.write" ],
			  "client_id" : "clientid",
			  "resource_ids" : [ "none" ],
			  "authorized_grant_types" : [ "client_credentials" ],
			  "redirect_uri" : [ "http://ant.path.wildcard/**/passback/*", "http://test1.com" ],
			  "autoapprove" : true,
			  "authorities" : [ "clients.read", "clients.write" ],
			  "token_salt" : "1SztLL",
			  "allowedproviders" : [ "uaa", "ldap", "my-saml-provider" ],
			  "name" : "My Client Name",
			  "lastModified" : 1502816030525,
			  "required_user_groups" : [ ]
			}`

			server.RouteToHandler("GET", "/oauth/clients/clientid", ghttp.CombineHandlers(
				ghttp.RespondWith(200, ResponseWithBooleanAutoapprove),
				ghttp.VerifyRequest("GET", "/oauth/clients/clientid"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			clientResponse, err := cm.Get("clientid")

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(err).To(BeNil())
			Expect(clientResponse.Scope[0]).To(Equal("clients.read"))
			Expect(clientResponse.Scope[1]).To(Equal("clients.write"))
			Expect(clientResponse.ClientID).To(Equal("clientid"))
			Expect(clientResponse.ResourceIDs[0]).To(Equal("none"))
			Expect(clientResponse.AuthorizedGrantTypes[0]).To(Equal("client_credentials"))
			Expect(clientResponse.RedirectURI[0]).To(Equal("http://ant.path.wildcard/**/passback/*"))
			Expect(clientResponse.RedirectURI[1]).To(Equal("http://test1.com"))
			Expect(clientResponse.TokenSalt).To(Equal("1SztLL"))
			Expect(clientResponse.AllowedProviders[0]).To(Equal("uaa"))
			Expect(clientResponse.AllowedProviders[1]).To(Equal("ldap"))
			Expect(clientResponse.AllowedProviders[2]).To(Equal("my-saml-provider"))
			Expect(clientResponse.DisplayName).To(Equal("My Client Name"))
			Expect(clientResponse.LastModified).To(Equal(int64(1502816030525)))
		})

		It("returns helpful error when /oauth/clients/clientid response can't be parsed", func() {
			server.RouteToHandler("GET", "/oauth/clients/clientid", ghttp.CombineHandlers(
				ghttp.RespondWith(200, "{unparsable-json-response}"),
				ghttp.VerifyRequest("GET", "/oauth/clients/clientid"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			_, err := cm.Get("clientid")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("An unknown error occurred while parsing response from"))
			Expect(err.Error()).To(ContainSubstring("Response was {unparsable-json-response}"))
		})
	})

	Describe("Delete", func() {
		const DeleteClientResponseJSON string = `{
		  "scope" : [ "clients.read", "clients.write" ],
		  "client_id" : "clientid",
		  "resource_ids" : [ "none" ],
		  "authorized_grant_types" : [ "client_credentials" ],
		  "redirect_uri" : [ "http://ant.path.wildcard/**/passback/*", "http://test1.com" ],
		  "autoapprove" : [ "true" ],
		  "authorities" : [ "clients.read", "clients.write" ],
		  "token_salt" : "1SztLL",
		  "allowedproviders" : [ "uaa", "ldap", "my-saml-provider" ],
		  "name" : "My Client Name",
		  "lastModified" : 1502816030525,
		  "required_user_groups" : [ ]
		}`

		AfterEach(func() {
			server.Close()
		})

		It("calls the /oauth/clients endpoint", func() {
			server.RouteToHandler("DELETE", "/oauth/clients/clientid", ghttp.CombineHandlers(
				ghttp.RespondWith(200, DeleteClientResponseJSON),
				ghttp.VerifyRequest("DELETE", "/oauth/clients/clientid"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			clientResponse, _ := cm.Delete("clientid")

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(clientResponse.Scope[0]).To(Equal("clients.read"))
			Expect(clientResponse.Scope[1]).To(Equal("clients.write"))
			Expect(clientResponse.ClientID).To(Equal("clientid"))
			Expect(clientResponse.ResourceIDs[0]).To(Equal("none"))
			Expect(clientResponse.AuthorizedGrantTypes[0]).To(Equal("client_credentials"))
			Expect(clientResponse.RedirectURI[0]).To(Equal("http://ant.path.wildcard/**/passback/*"))
			Expect(clientResponse.RedirectURI[1]).To(Equal("http://test1.com"))
			Expect(clientResponse.TokenSalt).To(Equal("1SztLL"))
			Expect(clientResponse.AllowedProviders[0]).To(Equal("uaa"))
			Expect(clientResponse.AllowedProviders[1]).To(Equal("ldap"))
			Expect(clientResponse.AllowedProviders[2]).To(Equal("my-saml-provider"))
			Expect(clientResponse.DisplayName).To(Equal("My Client Name"))
			Expect(clientResponse.LastModified).To(Equal(int64(1502816030525)))
		})

		It("returns helpful error when /oauth/clients request fails", func() {
			server.RouteToHandler("DELETE", "/oauth/clients/clientid", ghttp.CombineHandlers(
				ghttp.RespondWith(500, ""),
				ghttp.VerifyRequest("DELETE", "/oauth/clients/clientid"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			_, err := cm.Delete("clientid")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
		})

		It("returns helpful error when /oauth/clients/clientid response can't be parsed", func() {
			server.RouteToHandler("DELETE", "/oauth/clients/clientid", ghttp.CombineHandlers(
				ghttp.RespondWith(200, "{unparsable-json-response}"),
				ghttp.VerifyRequest("DELETE", "/oauth/clients/clientid"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			_, err := cm.Delete("clientid")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("An unknown error occurred while parsing response from"))
			Expect(err.Error()).To(ContainSubstring("Response was {unparsable-json-response}"))
		})
	})

	Describe("Create", func() {
		const createdClientResponse string = `{
		  "scope" : [ "clients.read", "clients.write" ],
		  "client_id" : "peanuts_client",
		  "resource_ids" : [ "none" ],
		  "authorized_grant_types" : [ "client_credentials", "authorization_code" ],
		  "redirect_uri" : [ "http://snoopy.com/**", "http://woodstock.com" ],
		  "autoapprove" : ["true"],
		  "authorities" : [ "comics.read", "comics.write" ],
		  "token_salt" : "1SztLL",
		  "allowedproviders" : [ "uaa", "ldap", "my-saml-provider" ],
		  "name" : "The Peanuts Client",
		  "lastModified" : 1502816030525,
		  "required_user_groups" : [ ]
		}`

		It("calls the oauth/clients endpoint and returns response", func() {
			server.RouteToHandler("POST", "/oauth/clients", ghttp.CombineHandlers(
				ghttp.RespondWith(200, createdClientResponse),
				ghttp.VerifyRequest("POST", "/oauth/clients"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Content-Type", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			toCreate := uaa.Client{
				ClientID:             "peanuts_client",
				AuthorizedGrantTypes: []string{"client_credentials"},
				Scope:                []string{"clients.read", "clients.write"},
				ResourceIDs:          []string{"none"},
				RedirectURI:          []string{"http://snoopy.com/**", "http://woodstock.com"},
				Authorities:          []string{"comics.read", "comics.write"},
				DisplayName:          "The Peanuts Client",
			}

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			createdClient, _ := cm.Create(toCreate)

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(createdClient.Scope[0]).To(Equal("clients.read"))
			Expect(createdClient.Scope[1]).To(Equal("clients.write"))
			Expect(createdClient.ClientID).To(Equal("peanuts_client"))
			Expect(createdClient.ResourceIDs[0]).To(Equal("none"))
			Expect(createdClient.AuthorizedGrantTypes[0]).To(Equal("client_credentials"))
			Expect(createdClient.AuthorizedGrantTypes[1]).To(Equal("authorization_code"))
			Expect(createdClient.RedirectURI[0]).To(Equal("http://snoopy.com/**"))
			Expect(createdClient.RedirectURI[1]).To(Equal("http://woodstock.com"))
			Expect(createdClient.TokenSalt).To(Equal("1SztLL"))
			Expect(createdClient.AllowedProviders[0]).To(Equal("uaa"))
			Expect(createdClient.AllowedProviders[1]).To(Equal("ldap"))
			Expect(createdClient.AllowedProviders[2]).To(Equal("my-saml-provider"))
			Expect(createdClient.DisplayName).To(Equal("The Peanuts Client"))
			Expect(createdClient.LastModified).To(Equal(int64(1502816030525)))
		})
	})

	Describe("Update", func() {
		const updatedClientResponse string = `{
		  "scope" : [ "clients.read", "clients.write" ],
		  "client_id" : "peanuts_client",
		  "resource_ids" : [ "none" ],
		  "authorized_grant_types" : [ "client_credentials", "authorization_code" ],
		  "redirect_uri" : [ "http://snoopy.com/**", "http://woodstock.com" ],
		  "autoapprove" : ["true"],
		  "authorities" : [ "comics.read", "comics.write" ],
		  "token_salt" : "1SztLL",
		  "allowedproviders" : [ "uaa", "ldap", "my-saml-provider" ],
		  "name" : "The Peanuts Client",
		  "lastModified" : 1502816030525,
		  "required_user_groups" : [ ]
		}`

		It("calls the PUT oauth/clients/CLIENT_ID endpoint and returns response", func() {
			server.RouteToHandler("PUT", "/oauth/clients/peanuts_client", ghttp.CombineHandlers(
				ghttp.RespondWith(200, updatedClientResponse),
				ghttp.VerifyRequest("PUT", "/oauth/clients/peanuts_client"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Content-Type", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			toUpdate := uaa.Client{
				ClientID:             "peanuts_client",
				AuthorizedGrantTypes: []string{"client_credentials"},
				Scope:                []string{"clients.read", "clients.write"},
				ResourceIDs:          []string{"none"},
				RedirectURI:          []string{"http://snoopy.com/**", "http://woodstock.com"},
				Authorities:          []string{"comics.read", "comics.write"},
				DisplayName:          "The Peanuts Client",
			}

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			updatedClient, _ := cm.Update(toUpdate)

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(updatedClient.Scope[0]).To(Equal("clients.read"))
			Expect(updatedClient.Scope[1]).To(Equal("clients.write"))
			Expect(updatedClient.ClientID).To(Equal("peanuts_client"))
			Expect(updatedClient.ResourceIDs[0]).To(Equal("none"))
			Expect(updatedClient.AuthorizedGrantTypes[0]).To(Equal("client_credentials"))
			Expect(updatedClient.AuthorizedGrantTypes[1]).To(Equal("authorization_code"))
			Expect(updatedClient.RedirectURI[0]).To(Equal("http://snoopy.com/**"))
			Expect(updatedClient.RedirectURI[1]).To(Equal("http://woodstock.com"))
			Expect(updatedClient.TokenSalt).To(Equal("1SztLL"))
			Expect(updatedClient.AllowedProviders[0]).To(Equal("uaa"))
			Expect(updatedClient.AllowedProviders[1]).To(Equal("ldap"))
			Expect(updatedClient.AllowedProviders[2]).To(Equal("my-saml-provider"))
			Expect(updatedClient.DisplayName).To(Equal("The Peanuts Client"))
			Expect(updatedClient.LastModified).To(Equal(int64(1502816030525)))
		})
	})

	Describe("ChangeSecret", func() {
		It("calls the /oauth/clients/<clientid>/secret endpoint", func() {
			server.RouteToHandler("PUT", "/oauth/clients/peanuts_client/secret", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, ""),
				ghttp.VerifyRequest("PUT", "/oauth/clients/peanuts_client/secret"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Content-Type", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				ghttp.VerifyJSON(`{"clientId": "peanuts_client", "secret": "new_secret"}`),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			cm.ChangeSecret("peanuts_client", "new_secret")

			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("does not panic when error happens during network call", func() {
			server.RouteToHandler("PUT", "/oauth/clients/peanuts_client/secret", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusUnauthorized, ""),
				ghttp.VerifyRequest("PUT", "/oauth/clients/peanuts_client/secret"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Content-Type", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				ghttp.VerifyJSON(`{"clientId": "peanuts_client", "secret": "new_secret"}`),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			err := cm.ChangeSecret("peanuts_client", "new_secret")

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(err).NotTo(BeNil())
		})
	})

	It("returns error when response is unparsable", func() {
		server.RouteToHandler("PUT", "/oauth/clients/peanuts_client", ghttp.CombineHandlers(
			ghttp.RespondWith(200, "{unparsable}"),
		))

		cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
		_, err := cm.Update(uaa.Client{ClientID: "peanuts_client"})

		Expect(server.ReceivedRequests()).To(HaveLen(1))
		Expect(err).NotTo(BeNil())
	})

	Describe("List", func() {
		const ClientsListResponseJSONPage1 = `{
		  "resources" : [ {
			"client_id" : "client1"
		  },
		  {
			"client_id" : "client2"
		  }],
		  "startIndex" : 1,
		  "itemsPerPage" : 2,
		  "totalResults" : 6,
		  "schemas" : [ "http://cloudfoundry.org/schema/scim/oauth-clients-1.0" ]
		}`

		It("can fetch and display multiple pages of results", func() {
			// This test is fairly bogus. Even though the query params vary each
			// time the /oauth/clients endpoint is called, I couldn't figure out
			// a way to make ghttp return different responses for sequential
			// calls to the same endpoint. In reality, with totalResults=6 I would
			// expect the three calls to each get a different response.
			server.RouteToHandler("GET", "/oauth/clients", ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/oauth/clients"),
				ghttp.RespondWith(http.StatusOK, ClientsListResponseJSONPage1),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			clientList, err := cm.List()

			Expect(server.ReceivedRequests()).To(HaveLen(3))
			Expect(clientList).To(HaveLen(6))
			Expect(clientList[0].ClientID).To(Equal("client1"))
			Expect(err).To(BeNil())
		})

		It("returns an error if an error occurs while fetching clients", func() {
			server.RouteToHandler("GET", "/oauth/clients", ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/oauth/clients"),
				ghttp.RespondWith(http.StatusInternalServerError, ""),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			_, err := cm.List()

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(err).NotTo(BeNil())
		})

		It("returns an error when parsing fails", func() {
			server.RouteToHandler("GET", "/oauth/clients", ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/oauth/clients"),
				ghttp.RespondWith(http.StatusInternalServerError, "{garbage}"),
			))

			cm := &uaa.ClientManager{HTTPClient: httpClient, Config: config}
			_, err := cm.List()

			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(err).NotTo(BeNil())
		})
	})
})
