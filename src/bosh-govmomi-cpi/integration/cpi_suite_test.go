package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BoshGovmomiCpi Suite")
}

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
