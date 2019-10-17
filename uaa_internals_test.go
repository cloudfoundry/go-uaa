package uaa

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var suite spec.Suite

func init() {
	suite = spec.New("uaa-internals", spec.Report(report.Terminal{}))
	suite("ensureTransport", testEnsureTransport)
	suite("contains", testContains)
	suite("URLWithPath", testURLWithPath)
	suite("api", testAPI)
	suite("uaaTransport", testUaaTransport)
}

func TestUAAInternals(t *testing.T) {
	suite.Run(t)
}
