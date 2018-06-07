package uaa

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stringifier", func() {
	It("presents a string slice as a string", func() {
		Expect(stringSliceStringifier([]string{"foo", "bar", "baz"})).To(Equal("[foo, bar, baz]"))
		Expect(stringSliceStringifier([]string{"foo"})).To(Equal("[foo]"))
		Expect(stringSliceStringifier([]string{})).To(Equal("[]"))
		Expect(stringSliceStringifier([]string{" "})).To(Equal("[ ]"))
	})
})
