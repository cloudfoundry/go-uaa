package uaa

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("contains", func() {
	list := []string{"do", "re", "mi"}

	It("returns true if present", func() {
		Expect(contains(list, "re")).To(BeTrue())
	})

	It("returns false if not present", func() {
		Expect(contains(list, "fa")).To(BeFalse())
	})

	It("handles empty list", func() {
		Expect(contains([]string{}, "fa")).To(BeFalse())
	})
})
