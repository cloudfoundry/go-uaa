package uaa_test

import (
	"fmt"

	"net/http"

	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Groups", func() {
	var (
		gm        GroupManager
		uaaServer *ghttp.Server
	)

	BeforeEach(func() {
		uaaServer = ghttp.NewServer()
		config := NewConfigWithServerURL(uaaServer.URL())
		config.AddContext(NewContextWithToken("access_token"))
		gm = GroupManager{HTTPClient: &http.Client{}, Config: config}
	})

	CloudControllerReadGroupResponse := `{
		"id" : "ea777017-883e-48ba-800a-637c71409b5e",
		"meta" : {
			"version" : 1,
			"created" : "2017-01-15T16:54:15.677Z",
			"lastModified" : "2017-08-15T16:54:15.677Z"
		},
		"displayName" : "cloud_controller.read",
		"description" : "View details of your applications and services",
		"members" : [ {
			"origin" : "uaa",
			"type" : "USER",
			"value" : "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"
		} ],
		"zoneID" : "uaa",
		"schemas" : [ "urn:scim:schemas:core:1.0" ]
	}`

	UaaAdminGroupResponse := `{
		"id" : "05a0c169-3592-4a45-b109-a16d9246e0ab",
		"meta" : {
			"version" : 1,
			"created" : "2017-01-15T16:54:15.677Z",
			"lastModified" : "2017-08-15T16:54:15.677Z"
		},
		"displayName" : "uaa.admin",
		"description" : "Act as an administrator throughout the UAA",
		"members" : [ {
			"origin" : "uaa",
			"type" : "USER",
			"value" : "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"
		} ],
		"zoneID" : "uaa",
		"schemas" : [ "urn:scim:schemas:core:1.0" ]
	}`

	var groupListResponse = fmt.Sprintf(PaginatedResponseTmpl, UaaAdminGroupResponse, CloudControllerReadGroupResponse)

	Describe("GroupManager#Get", func() {
		It("gets a group from the UAA by id", func() {
			uaaServer.RouteToHandler("GET", "/Groups/05a0c169-3592-4a45-b109-a16d9246e0ab", ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/Groups/05a0c169-3592-4a45-b109-a16d9246e0ab"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.RespondWith(http.StatusOK, UaaAdminGroupResponse),
			))

			group, _ := gm.Get("05a0c169-3592-4a45-b109-a16d9246e0ab")

			Expect(group.ID).To(Equal("05a0c169-3592-4a45-b109-a16d9246e0ab"))
			Expect(group.Meta.Created).To(Equal("2017-01-15T16:54:15.677Z"))
			Expect(group.Meta.LastModified).To(Equal("2017-08-15T16:54:15.677Z"))
			Expect(group.Meta.Version).To(Equal(1))
			Expect(group.DisplayName).To(Equal("uaa.admin"))
			Expect(group.ZoneID).To(Equal("uaa"))
			Expect(group.Description).To(Equal("Act as an administrator throughout the UAA"))
			Expect(group.Members[0].Origin).To(Equal("uaa"))
			Expect(group.Members[0].Type).To(Equal("USER"))
			Expect(group.Members[0].Value).To(Equal("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"))
			Expect(group.Schemas[0]).To(Equal("urn:scim:schemas:core:1.0"))
		})

		It("returns helpful error when /Groups/groupid request fails", func() {
			uaaServer.RouteToHandler("GET", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusInternalServerError, ""),
				ghttp.VerifyRequest("GET", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			_, err := gm.Get("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("An unknown error occurred while calling"))
		})

		It("returns helpful error when /Groups/groupid response can't be parsed", func() {
			uaaServer.RouteToHandler("GET", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, "{unparsable-json-response}"),
				ghttp.VerifyRequest("GET", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			_, err := gm.Get("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("An unknown error occurred while parsing response from"))
			Expect(err.Error()).To(ContainSubstring("Response was {unparsable-json-response}"))
		})
	})

	Describe("GroupManager#GetByName", func() {
		Context("when no group name is specified", func() {
			It("returns an error", func() {
				_, err := gm.GetByName("", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("group name may not be blank"))
			})
		})

		Context("when no origin is specified", func() {
			It("looks up a group with a SCIM filter", func() {
				group := Group{DisplayName: "uaa.admin"}
				response := PaginatedResponse(group)

				uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusOK, response),
					ghttp.VerifyRequest("GET", "/Groups", "filter=displayName+eq+%22uaa.admin%22"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				))

				group, err := gm.GetByName("uaa.admin", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(group.DisplayName).To(Equal("uaa.admin"))
			})

			It("returns an error when request fails", func() {
				uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusInternalServerError, ""),
					ghttp.VerifyRequest("GET", "/Groups", "filter=displayName+eq+%22uaa.admin%22"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				))

				_, err := gm.GetByName("uaa.admin", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("An unknown error"))
			})

			It("returns an error when no groups are found", func() {
				uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusOK, PaginatedResponse()),
					ghttp.VerifyRequest("GET", "/Groups", "filter=displayName+eq+%22uaa.admin%22"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				))

				_, err := gm.GetByName("uaa.admin", "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`group uaa.admin not found`))
			})
		})

		Context("when attributes are specified", func() {
			It("adds them to the GET request", func() {
				uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusOK, PaginatedResponse(Group{DisplayName: "uaa.admin"})),
					ghttp.VerifyRequest("GET", "/Groups", "filter=displayName+eq+%22uaa.admin%22&attributes=displayName"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				))

				_, err := gm.GetByName("uaa.admin", "displayName")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("GroupManager#List", func() {
		It("can accept a filter query to limit results", func() {
			uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, groupListResponse),
				ghttp.VerifyRequest("GET", "/Groups", "filter=id+eq+%22fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7%22"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			groupList, err := gm.List(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`, "", "", "")

			Expect(err).NotTo(HaveOccurred())
			Expect(groupList[0].DisplayName).To(Equal("uaa.admin"))
			Expect(groupList[1].DisplayName).To(Equal("cloud_controller.read"))
		})

		It("gets all groups when no filter is passed", func() {
			uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, groupListResponse),
				ghttp.VerifyRequest("GET", "/Groups", ""),
			))

			groupList, err := gm.List("", "", "", "")

			Expect(err).NotTo(HaveOccurred())
			Expect(groupList[0].DisplayName).To(Equal("uaa.admin"))
			Expect(groupList[1].DisplayName).To(Equal("cloud_controller.read"))
		})

		It("can accept an attributes list", func() {
			uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, groupListResponse),
				ghttp.VerifyRequest("GET", "/Groups", "filter=id+eq+%22fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7%22&attributes=displayName"),
			))

			groupList, err := gm.List(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`, "", "displayName", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(groupList[0].DisplayName).To(Equal("uaa.admin"))
			Expect(groupList[1].DisplayName).To(Equal("cloud_controller.read"))
		})

		It("can accept sortBy", func() {
			uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, groupListResponse),
				ghttp.VerifyRequest("GET", "/Groups", "sortBy=displayName"),
			))

			_, err := gm.List("", "displayName", "", "")
			Expect(err).NotTo(HaveOccurred())
		})

		It("can accept sort order ascending/descending", func() {
			uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, groupListResponse),
				ghttp.VerifyRequest("GET", "/Groups", "sortOrder=ascending"),
			))

			_, err := gm.List("", "", "", SortAscending)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can retrieve multiple pages", func() {
			page1 := MultiPaginatedResponse(1, 1, 2, Group{DisplayName: "uaa.admin"})
			page2 := MultiPaginatedResponse(2, 1, 2, Group{DisplayName: "cloud_controller.read"})
			uaaServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusOK, page1),
					ghttp.VerifyRequest("GET", "/Groups", ""),
				),
				ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusOK, page2),
					ghttp.VerifyRequest("GET", "/Groups", "count=1&startIndex=2"),
				),
			)

			groupList, err := gm.List("", "", "", "")

			Expect(err).NotTo(HaveOccurred())
			Expect(groupList[0].DisplayName).To(Equal("uaa.admin"))
			Expect(groupList[1].DisplayName).To(Equal("cloud_controller.read"))
		})

		It("returns an error when /Groups doesn't respond", func() {
			uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusInternalServerError, ""),
				ghttp.VerifyRequest("GET", "/Groups", "filter=id+eq+%22fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7%22"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			_, err := gm.List(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`, "", "", "")

			Expect(err).To(HaveOccurred())
		})

		It("returns an error when response is unparseable", func() {
			uaaServer.RouteToHandler("GET", "/Groups", ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, "{unparsable}"),
				ghttp.VerifyRequest("GET", "/Groups", "filter=id+eq+%22fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7%22"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
			))

			_, err := gm.List(`id eq "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd7"`, "", "", "")

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GroupManager#Create", func() {
		var group Group

		BeforeEach(func() {
			group = Group{
				DisplayName: "uaa.admin",
			}
		})

		It("performs POST with group data and bearer token", func() {
			uaaServer.RouteToHandler("POST", "/Groups", ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/Groups"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Content-Type", "application/json"),
				ghttp.VerifyJSON(`{ "displayName": "uaa.admin" }`),
				ghttp.RespondWith(http.StatusOK, UaaAdminGroupResponse),
			))

			gm.Create(group)

			Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns the created group", func() {
			uaaServer.RouteToHandler("POST", "/Groups", ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/Groups"),
				ghttp.RespondWith(http.StatusOK, UaaAdminGroupResponse),
			))

			group, _ = gm.Create(group)

			Expect(group.DisplayName).To(Equal("uaa.admin"))
		})

		It("returns error when response cannot be parsed", func() {
			uaaServer.RouteToHandler("POST", "/Groups", ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/Groups"),
				ghttp.RespondWith(http.StatusOK, "{unparseable}"),
			))

			_, err := gm.Create(group)

			Expect(err).To(HaveOccurred())
		})

		It("returns error when response is not 200 OK", func() {
			uaaServer.RouteToHandler("POST", "/Groups", ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/Groups"),
				ghttp.RespondWith(http.StatusBadRequest, ""),
			))

			_, err := gm.Create(group)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GroupManager#Update", func() {
		var group Group

		BeforeEach(func() {
			group = Group{
				DisplayName: "uaa.admin",
			}
		})

		It("performs PUT with group data and bearer token", func() {
			uaaServer.RouteToHandler("PUT", "/Groups", ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/Groups"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyHeaderKV("Content-Type", "application/json"),
				ghttp.VerifyJSON(`{ "displayName": "uaa.admin" }`),
				ghttp.RespondWith(http.StatusOK, UaaAdminGroupResponse),
			))

			gm.Update(group)

			Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns the updated group", func() {
			uaaServer.RouteToHandler("PUT", "/Groups", ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/Groups"),
				ghttp.RespondWith(http.StatusOK, UaaAdminGroupResponse),
			))

			group, _ = gm.Update(group)

			Expect(group.DisplayName).To(Equal("uaa.admin"))
		})

		It("returns error when response cannot be parsed", func() {
			uaaServer.RouteToHandler("PUT", "/Groups", ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/Groups"),
				ghttp.RespondWith(http.StatusOK, "{unparseable}"),
			))

			_, err := gm.Update(group)

			Expect(err).To(HaveOccurred())
		})

		It("returns error when response is not 200 OK", func() {
			uaaServer.RouteToHandler("PUT", "/Groups", ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/Groups"),
				ghttp.RespondWith(http.StatusBadRequest, ""),
			))

			_, err := gm.Update(group)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GroupManager#Delete", func() {
		It("performs DELETE with group data and bearer token", func() {
			uaaServer.RouteToHandler("DELETE", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70", ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.RespondWith(http.StatusOK, UaaAdminGroupResponse),
			))

			gm.Delete("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")

			Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns the deleted group", func() {
			uaaServer.RouteToHandler("DELETE", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70", ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"),
				ghttp.RespondWith(http.StatusOK, UaaAdminGroupResponse),
			))

			group, _ := gm.Delete("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")

			Expect(group.DisplayName).To(Equal("uaa.admin"))
		})

		It("returns error when response cannot be parsed", func() {
			uaaServer.RouteToHandler("DELETE", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70", ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"),
				ghttp.RespondWith(http.StatusOK, "{unparseable}"),
			))

			_, err := gm.Delete("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")

			Expect(err).To(HaveOccurred())
		})

		It("returns error when response is not 200 OK", func() {
			uaaServer.RouteToHandler("DELETE", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70", ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", "/Groups/fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"),
				ghttp.RespondWith(http.StatusBadRequest, ""),
			))

			_, err := gm.Delete("fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GroupManager#AddMember", func() {
		It("adds a membership", func() {
			membershipJSON := `{"origin":"uaa","type":"USER","value":"fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70"}`
			uaaServer.RouteToHandler("POST", "/Groups/05a0c169-3592-4a45-b109-a16d9246e0ab/members", ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/Groups/05a0c169-3592-4a45-b109-a16d9246e0ab/members"),
				ghttp.VerifyHeaderKV("Authorization", "bearer access_token"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.VerifyJSON(membershipJSON),
				ghttp.RespondWith(http.StatusOK, membershipJSON),
			))

			gm.AddMember("05a0c169-3592-4a45-b109-a16d9246e0ab", "fb5f32e1-5cb3-49e6-93df-6df9c8c8bd70")

			Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
