package govc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestCpi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Govc Suite")
}

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
