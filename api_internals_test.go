package uaa

import (
	"log"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"golang.org/x/oauth2"
)

func testAPI(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
		log.SetFlags(log.Lshortfile)
	})

	when("New", func() {
		it("sets the zoneID", func() {
			api, err := New("https://example.net", WithToken(&oauth2.Token{
				AccessToken: "blergh",
				Expiry:      time.Now().Add(60 * time.Second),
			}), WithZoneID("zone-1"))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.zoneID).To(Equal("zone-1"))
		})
	})
}
