package vmx_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestVmx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VMX Suite")
}

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
