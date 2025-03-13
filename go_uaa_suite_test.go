package uaa_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGoUaa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoUaa Suite")
}
