package uaa_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoUaa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoUaa Suite")
}
