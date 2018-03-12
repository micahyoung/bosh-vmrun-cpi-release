package stemcell_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestCpi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CPI Suite")
}

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
