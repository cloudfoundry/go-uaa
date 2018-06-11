package uaa_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	uaa "github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func newTrueP() *bool {
	b := true
	return &b
}

func newFalseP() *bool {
	b := false
	return &b
}

func TestUsers(t *testing.T) {
	spec.Run(t, "Users", testUsers, spec.Report(report.Terminal{}))
}

func testUsers(t *testing.T, when spec.G, it spec.S) {
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
		c := &http.Client{Transport: http.DefaultTransport}
		u, _ := url.Parse(s.URL)
		a = &uaa.API{
			TargetURL:             u,
			AuthenticatedClient:   c,
			UnauthenticatedClient: c,
		}
	})

	it.After(func() {
		if s != nil {
			s.Close()
		}
	})

	when("GetUser()", func() {
		when("the user is returned from the server", func() {
			it.Before(func() {
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(MarcusUserResponse))
				})
			})
			it("gets the user from the UAA by ID", func() {
				user, err := a.GetUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")
				Expect(err).NotTo(HaveOccurred())
				Expect(user.ID).To(Equal("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
				Expect(user.ExternalID).To(Equal("marcus-user"))
				Expect(user.Meta.Created).To(Equal("2017-01-15T16:54:15.677Z"))
				Expect(user.Meta.LastModified).To(Equal("2017-08-15T16:54:15.677Z"))
				Expect(user.Meta.Version).To(Equal(1))
				Expect(user.Username).To(Equal("marcus@stoicism.com"))
				Expect(user.Name.GivenName).To(Equal("Marcus"))
				Expect(user.Name.FamilyName).To(Equal("Aurelius"))
				Expect(*user.Emails[0].Primary).To(Equal(false))
				Expect(user.Emails[0].Value).To(Equal("marcus@stoicism.com"))
				Expect(user.Groups[0].Display).To(Equal("philosophy.read"))
				Expect(user.Groups[0].Type).To(Equal("DIRECT"))
				Expect(user.Groups[0].Value).To(Equal("ac2ab20e-0a2d-4b68-82e4-817ee6b258b4"))
				Expect(user.Groups[1].Display).To(Equal("philosophy.write"))
				Expect(user.Groups[1].Type).To(Equal("DIRECT"))
				Expect(user.Groups[1].Value).To(Equal("110b2434-4a30-439b-b5fc-f4cf47fc04f0"))
				Expect(user.Approvals[0].UserID).To(Equal("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
				Expect(user.Approvals[0].ClientID).To(Equal("shinyclient"))
				Expect(user.Approvals[0].ExpiresAt).To(Equal("2017-08-15T16:54:25.765Z"))
				Expect(user.Approvals[0].LastUpdatedAt).To(Equal("2017-08-15T16:54:15.765Z"))
				Expect(user.Approvals[0].Scope).To(Equal("philosophy.read"))
				Expect(user.Approvals[0].Status).To(Equal("APPROVED"))
				Expect(user.PhoneNumbers[0].Value).To(Equal("5555555555"))
				Expect(*user.Active).To(Equal(true))
				Expect(*user.Verified).To(Equal(true))
				Expect(user.Origin).To(Equal("uaa"))
				Expect(user.ZoneID).To(Equal("uaa"))
				Expect(user.PasswordLastModified).To(Equal("2017-08-15T16:54:15.000Z"))
				Expect(user.PreviousLogonTime).To(Equal(1502816055768))
				Expect(user.LastLogonTime).To(Equal(1502816055768))
				Expect(user.Schemas[0]).To(Equal("urn:scim:schemas:core:1.0"))
			})
		})

		when("the server errors", func() {
			it.Before(func() {
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"))
					w.WriteHeader(http.StatusInternalServerError)
				})
			})

			it("returns helpful error", func() {
				user, err := a.GetUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7")
				Expect(err).To(HaveOccurred())
				Expect(user).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			})
		})

		when("the server returns unparsable users", func() {
			it.Before(func() {
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("{unparsable-json-response}"))
				})
			})

			it("returns helpful error", func() {
				user, err := a.GetUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7")
				Expect(err).To(HaveOccurred())
				Expect(user).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("An unknown error occurred while parsing response from"))
				Expect(err.Error()).To(ContainSubstring("Response was {unparsable-json-response}"))
			})
		})
	})

	when("GetUserByUsername()", func() {
		when("no username is specified", func() {
			it("returns an error", func() {
				u, err := a.GetUserByUsername("", "", "")
				Expect(err).To(HaveOccurred())
				Expect(u).To(BeNil())
				Expect(err.Error()).To(Equal("username cannot be blank"))
			})
		})

		when("an origin is specified", func() {
			it("looks up a user with SCIM filter", func() {
				user := uaa.User{Username: "marcus", Origin: "uaa"}
				response := PaginatedResponse(user)
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users"))
					Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus" and origin eq "uaa"`))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(response))
				})

				u, err := a.GetUserByUsername("marcus", "uaa", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(u.Username).To(Equal("marcus"))
			})

			it("returns an error when request fails", func() {
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users"))
					Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus" and origin eq "uaa"`))
					w.WriteHeader(http.StatusInternalServerError)
				})

				_, err := a.GetUserByUsername("marcus", "uaa", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("An unknown error"))
			})

			it("returns an error if no results are found", func() {
				response := PaginatedResponse()
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users"))
					Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus" and origin eq "uaa"`))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(response))
				})
				_, err := a.GetUserByUsername("marcus", "uaa", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`user marcus not found in origin uaa`))
			})

			when("attributes are specified", func() {
				it("adds them to the GET request", func() {
					handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						Expect(req.Header.Get("Accept")).To(Equal("application/json"))
						Expect(req.URL.Path).To(Equal("/Users"))
						Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus" and origin eq "uaa"`))
						Expect(req.URL.Query().Get("attributes")).To(Equal(`userName,emails`))
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(PaginatedResponse(uaa.User{Username: "marcus", Origin: "uaa"})))
					})
					_, err := a.GetUserByUsername("marcus", "uaa", "userName,emails")
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		when("no origin is specified", func() {
			it("looks up a user with a SCIM filter", func() {
				user := uaa.User{Username: "marcus", Origin: "uaa"}
				response := PaginatedResponse(user)
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users"))
					Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus"`))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(response))
				})
				u, err := a.GetUserByUsername("marcus", "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(u.Username).To(Equal("marcus"))
			})

			it("returns an error when request fails", func() {
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users"))
					Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus"`))
					w.WriteHeader(http.StatusInternalServerError)
				})
				_, err := a.GetUserByUsername("marcus", "", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("An unknown error"))
			})

			it("returns an error when no users are found", func() {
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users"))
					Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus"`))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(PaginatedResponse()))
				})
				_, err := a.GetUserByUsername("marcus", "", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`user marcus not found`))
			})

			it("returns an error when username found in multiple origins", func() {
				user1 := uaa.User{Username: "marcus", Origin: "uaa"}
				user2 := uaa.User{Username: "marcus", Origin: "ldap"}
				user3 := uaa.User{Username: "marcus", Origin: "okta"}
				response := PaginatedResponse(user1, user2, user3)
				handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					Expect(req.URL.Path).To(Equal("/Users"))
					Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus"`))
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(response))
				})

				_, err := a.GetUserByUsername("marcus", "", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`Found users with username marcus in multiple origins [uaa, ldap, okta].`))
			})

			when("attributes are specified", func() {
				it("adds them to the GET request", func() {
					handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						Expect(req.Header.Get("Accept")).To(Equal("application/json"))
						Expect(req.URL.Path).To(Equal("/Users"))
						Expect(req.URL.Query().Get("filter")).To(Equal(`userName eq "marcus"`))
						Expect(req.URL.Query().Get("attributes")).To(Equal(`userName,emails`))
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(PaginatedResponse(uaa.User{Username: "marcus", Origin: "uaa"})))
					})
					_, err := a.GetUserByUsername("marcus", "", "userName,emails")
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	when("ListAllUsers()", func() {
		it("can return multiple pages", func() {
			page1 := MultiPaginatedResponse(1, 1, 2, uaa.User{Username: "marcus", Origin: "uaa"})
			page2 := MultiPaginatedResponse(2, 1, 2, uaa.User{Username: "drseuss", Origin: "uaa"})
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				w.WriteHeader(http.StatusOK)
				if called == 1 {
					Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
					Expect(req.URL.Query().Get("count")).To(Equal("100"))
					w.Write([]byte(page1))
				} else {
					Expect(req.URL.Query().Get("startIndex")).To(Equal("2"))
					Expect(req.URL.Query().Get("count")).To(Equal("1"))
					w.Write([]byte(page2))
				}
			})

			users, err := a.ListAllUsers("", "", "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(users[0].Username).To(Equal("marcus"))
			Expect(users[1].Username).To(Equal("drseuss"))
			Expect(called).To(Equal(2))
		})
	})

	when("ListUsers()", func() {
		var userListResponse = fmt.Sprintf(PaginatedResponseTmpl, MarcusUserResponse, DrSeussUserResponse)

		it("can accept a filter query to limit results", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				Expect(req.URL.Query().Get("count")).To(Equal("100"))
				Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
				Expect(req.URL.Query().Get("filter")).To(Equal(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(userListResponse))
			})
			userList, _, err := a.ListUsers(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`, "", "", "", 1, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(userList[0].Username).To(Equal("marcus@stoicism.com"))
			Expect(userList[1].Username).To(Equal("drseuss@whoville.com"))
		})

		it("does not include the filter param if no filter exists", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				Expect(req.URL.Query().Get("count")).To(Equal("100"))
				Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
				Expect(req.URL.Query().Get("filter")).To(Equal(""))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(userListResponse))
			})
			userList, _, err := a.ListUsers("", "", "", "", 1, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(userList[0].Username).To(Equal("marcus@stoicism.com"))
			Expect(userList[1].Username).To(Equal("drseuss@whoville.com"))
		})

		it("can accept an attributes list", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				Expect(req.URL.Query().Get("count")).To(Equal("100"))
				Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
				Expect(req.URL.Query().Get("filter")).To(Equal(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`))
				Expect(req.URL.Query().Get("attributes")).To(Equal(`userName,emails`))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(userListResponse))
			})
			userList, _, err := a.ListUsers(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`, "", "userName,emails", "", 1, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(userList[0].Username).To(Equal("marcus@stoicism.com"))
			Expect(userList[1].Username).To(Equal("drseuss@whoville.com"))
		})

		it("can accept sortBy", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				Expect(req.URL.Query().Get("count")).To(Equal("100"))
				Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
				Expect(req.URL.Query().Get("filter")).To(Equal(""))
				Expect(req.URL.Query().Get("attributes")).To(Equal(""))
				Expect(req.URL.Query().Get("sortBy")).To(Equal("userName"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(userListResponse))
			})
			userList, _, err := a.ListUsers("", "userName", "", "", 1, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(userList[0].Username).To(Equal("marcus@stoicism.com"))
			Expect(userList[1].Username).To(Equal("drseuss@whoville.com"))
		})

		it("can accept sort order ascending/descending", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				Expect(req.URL.Query().Get("count")).To(Equal("100"))
				Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
				Expect(req.URL.Query().Get("filter")).To(Equal(""))
				Expect(req.URL.Query().Get("attributes")).To(Equal(""))
				Expect(req.URL.Query().Get("sortBy")).To(Equal(""))
				Expect(req.URL.Query().Get("sortOrder")).To(Equal("ascending"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(userListResponse))
			})
			userList, _, err := a.ListUsers("", "", "", uaa.SortAscending, 1, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(userList[0].Username).To(Equal("marcus@stoicism.com"))
			Expect(userList[1].Username).To(Equal("drseuss@whoville.com"))
		})

		it("uses a startIndex of 1 if 0 is supplied", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				Expect(req.URL.Query().Get("count")).To(Equal("100"))
				Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(userListResponse))
			})
			userList, _, err := a.ListUsers("", "", "", "", 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(userList[0].Username).To(Equal("marcus@stoicism.com"))
			Expect(userList[1].Username).To(Equal("drseuss@whoville.com"))
		})

		it("returns an error when /Users doesn't respond", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				Expect(req.URL.Query().Get("count")).To(Equal("100"))
				Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
				Expect(req.URL.Query().Get("filter")).To(Equal(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`))
				w.WriteHeader(http.StatusInternalServerError)
			})

			userList, _, err := a.ListUsers(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`, "", "", "", 1, 100)
			Expect(err).To(HaveOccurred())
			Expect(userList).To(BeNil())
		})

		it("returns an error when response is unparseable", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.URL.Path).To(Equal("/Users"))
				Expect(req.URL.Query().Get("count")).To(Equal("100"))
				Expect(req.URL.Query().Get("startIndex")).To(Equal("1"))
				Expect(req.URL.Query().Get("filter")).To(Equal(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{unparsable}"))
			})
			userList, _, err := a.ListUsers(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`, "", "", "", 1, 100)
			Expect(err).To(HaveOccurred())
			Expect(userList).To(BeNil())
		})
	})

	when("CreateUser()", func() {
		var (
			u uaa.User
		)
		it.Before(func() {
			u = uaa.User{
				Username: "marcus@stoicism.com",
				Active:   newTrueP(),
			}
			u.Name = &uaa.UserName{GivenName: "Marcus", FamilyName: "Aurelius"}
		})

		it("performs a POST with the user data and returns the created user", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(req.Method).To(Equal(http.MethodPost))
				Expect(req.URL.Path).To(Equal("/Users"))
				defer req.Body.Close()
				body, _ := ioutil.ReadAll(req.Body)
				Expect(body).To(MatchJSON(`{ "userName": "marcus@stoicism.com", "active": true, "name" : { "familyName" : "Aurelius", "givenName" : "Marcus" }}`))
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(MarcusUserResponse))
			})

			created, err := a.CreateUser(u)
			Expect(called).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
			Expect(created).NotTo(BeNil())
			Expect(created.Username).To(Equal("marcus@stoicism.com"))
			Expect(created.ExternalID).To(Equal("marcus-user"))
		})

		it("returns error when response cannot be parsed", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal(http.MethodPost))
				Expect(req.URL.Path).To(Equal("/Users"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{unparseable}"))
			})
			created, err := a.CreateUser(u)
			Expect(err).To(HaveOccurred())
			Expect(created).To(BeNil())
		})

		it("returns error when response is not 200 OK", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal(http.MethodPost))
				Expect(req.URL.Path).To(Equal("/Users"))
				w.WriteHeader(http.StatusBadRequest)
			})
			created, err := a.CreateUser(u)
			Expect(err).To(HaveOccurred())
			Expect(created).To(BeNil())
		})
	})

	when("UpdateUser()", func() {
		var (
			u uaa.User
		)
		it.Before(func() {
			u = uaa.User{
				Username: "marcus@stoicism.com",
				Active:   newTrueP(),
			}
			u.Name = &uaa.UserName{GivenName: "Marcus", FamilyName: "Aurelius"}
		})

		it("performs a PUT with the user data and returns the updated user", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(req.Method).To(Equal(http.MethodPut))
				Expect(req.URL.Path).To(Equal("/Users"))
				defer req.Body.Close()
				body, _ := ioutil.ReadAll(req.Body)
				Expect(body).To(MatchJSON(body))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(MarcusUserResponse))
			})

			updated, err := a.UpdateUser(u)
			Expect(called).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
			Expect(updated).NotTo(BeNil())
			Expect(updated.Username).To(Equal("marcus@stoicism.com"))
			Expect(updated.ExternalID).To(Equal("marcus-user"))
		})

		it("returns error when response cannot be parsed", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal(http.MethodPut))
				Expect(req.URL.Path).To(Equal("/Users"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{unparseable}"))
			})
			updated, err := a.UpdateUser(u)
			Expect(err).To(HaveOccurred())
			Expect(updated).To(BeNil())
		})

		it("returns error when response is not 200 OK", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal(http.MethodPut))
				Expect(req.URL.Path).To(Equal("/Users"))
				w.WriteHeader(http.StatusBadRequest)
			})
			updated, err := a.UpdateUser(u)
			Expect(err).To(HaveOccurred())
			Expect(updated).To(BeNil())
		})
	})

	when("DeleteUser()", func() {
		it("performs DELETE with user data and bearer token", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.Method).To(Equal(http.MethodDelete))
				Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(MarcusUserResponse))
			})

			deleted, err := a.DeleteUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")
			Expect(called).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
			Expect(deleted).NotTo(BeNil())
			Expect(deleted.Username).To(Equal("marcus@stoicism.com"))
			Expect(deleted.ExternalID).To(Equal("marcus-user"))
		})

		it("returns error when response cannot be parsed", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal(http.MethodDelete))
				Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{unparseable}"))
			})
			deleted, err := a.DeleteUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")
			Expect(err).To(HaveOccurred())
			Expect(deleted).To(BeNil())
		})

		it("returns error when response is not 200 OK", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal(http.MethodDelete))
				Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
				w.WriteHeader(http.StatusBadRequest)
			})
			deleted, err := a.DeleteUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")
			Expect(err).To(HaveOccurred())
			Expect(deleted).To(BeNil())
		})
	})

	when("ActivateUser()", func() {
		it("returns an error when the userID is empty", func() {
			err := a.ActivateUser("", 10)
			Expect(err).To(HaveOccurred())
			Expect(called).To(Equal(0))
		})

		it("activates the user using the userID", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(req.Header.Get("If-Match")).To(Equal("10"))
				Expect(req.Method).To(Equal(http.MethodPatch))
				Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
				defer req.Body.Close()
				body, _ := ioutil.ReadAll(req.Body)
				Expect(body).To(MatchJSON(`{ "active": true }`))
				w.WriteHeader(http.StatusOK)
			})
			err := a.ActivateUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(called).To(Equal(1))
		})

		it("returns a helpful error the request fails", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(req.Header.Get("If-Match")).To(Equal("0"))
				Expect(req.Method).To(Equal(http.MethodPatch))
				Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"))
				defer req.Body.Close()
				body, _ := ioutil.ReadAll(req.Body)
				Expect(body).To(MatchJSON(`{ "active": true }`))
				w.WriteHeader(http.StatusInternalServerError)
			})
			err := a.ActivateUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7", 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			Expect(called).To(Equal(1))
		})
	})

	when("DeactivateUser()", func() {
		it("returns an error when the userID is empty", func() {
			err := a.DeactivateUser("", 10)
			Expect(err).To(HaveOccurred())
			Expect(called).To(Equal(0))
		})

		it("deactivates the user using the userID", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(req.Header.Get("If-Match")).To(Equal("10"))
				Expect(req.Method).To(Equal(http.MethodPatch))
				Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
				defer req.Body.Close()
				body, _ := ioutil.ReadAll(req.Body)
				Expect(body).To(MatchJSON(`{ "active": false }`))
				w.WriteHeader(http.StatusOK)
			})
			err := a.DeactivateUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70", 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(called).To(Equal(1))
		})

		it("returns a helpful error the request fails", func() {
			handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(req.Header.Get("If-Match")).To(Equal("0"))
				Expect(req.Method).To(Equal(http.MethodPatch))
				Expect(req.URL.Path).To(Equal("/Users/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"))
				defer req.Body.Close()
				body, _ := ioutil.ReadAll(req.Body)
				Expect(body).To(MatchJSON(`{ "active": false }`))
				w.WriteHeader(http.StatusInternalServerError)
			})
			err := a.DeactivateUser("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7", 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
			Expect(called).To(Equal(1))
		})
	})

	when("using user structs", func() {
		when("verified", func() {
			it("correctly shows false boolean values", func() {
				user := uaa.User{Verified: newFalseP()}
				userBytes, _ := json.Marshal(&user)
				Expect(string(userBytes)).To(MatchJSON(`{"verified": false}`))

				newUser := uaa.User{}
				json.Unmarshal([]byte(userBytes), &newUser)
				Expect(*newUser.Verified).To(BeFalse())
			})

			it("correctly shows true values", func() {
				user := uaa.User{Verified: newTrueP()}
				userBytes, _ := json.Marshal(&user)
				Expect(string(userBytes)).To(MatchJSON(`{"verified": true}`))

				newUser := uaa.User{}
				json.Unmarshal([]byte(userBytes), &newUser)
				Expect(*newUser.Verified).To(BeTrue())
			})

			it("correctly hides unset values", func() {
				user := uaa.User{}
				json.Unmarshal([]byte("{}"), &user)
				Expect(user.Verified).To(BeNil())

				userBytes, _ := json.Marshal(&user)
				Expect(string(userBytes)).To(MatchJSON(`{}`))
			})
		})

		when("emails", func() {
			it("correctly shows false boolean values", func() {
				user := uaa.User{}
				email := uaa.Email{Value: "foo@bar.com", Primary: newFalseP()}
				user.Emails = []uaa.Email{email}

				userBytes, _ := json.Marshal(&user)
				Expect(string(userBytes)).To(MatchJSON(`{"emails": [ { "value": "foo@bar.com", "primary": false } ]}`))

				newUser := uaa.User{}
				json.Unmarshal([]byte(userBytes), &newUser)
				Expect(*newUser.Emails[0].Primary).To(BeFalse())
			})

			it("correctly shows true values", func() {
				user := uaa.User{}
				email := uaa.Email{Value: "foo@bar.com", Primary: newTrueP()}
				user.Emails = []uaa.Email{email}

				userBytes, _ := json.Marshal(&user)
				Expect(string(userBytes)).To(MatchJSON(`{"emails": [ { "value": "foo@bar.com", "primary": true } ]}`))

				newUser := uaa.User{}
				json.Unmarshal([]byte(userBytes), &newUser)
				Expect(*newUser.Emails[0].Primary).To(BeTrue())
			})
		})

		when("active", func() {
			it("correctly shows false boolean values", func() {
				user := uaa.User{Active: newFalseP()}
				userBytes, _ := json.Marshal(&user)
				Expect(string(userBytes)).To(MatchJSON(`{"active": false}`))

				newUser := uaa.User{}
				json.Unmarshal([]byte(userBytes), &newUser)
				Expect(*newUser.Active).To(BeFalse())
			})

			it("correctly shows true values", func() {
				user := uaa.User{Active: newTrueP()}
				userBytes, _ := json.Marshal(&user)
				Expect(string(userBytes)).To(MatchJSON(`{"active": true}`))

				newUser := uaa.User{}
				json.Unmarshal([]byte(userBytes), &newUser)
				Expect(*newUser.Active).To(BeTrue())
			})
		})
	})
}
