package integration_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mholt/archiver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var extractedStemcellTempDir string

func extractStemcell() string {
	stemcellFile := "../../../ci/deploy-test/state/stemcell.tgz"

	stemcellTempDir, err := ioutil.TempDir("", "stemcell-")
	Expect(err).ToNot(HaveOccurred())

	err = archiver.TarGz.Open(stemcellFile, stemcellTempDir)
	Expect(err).ToNot(HaveOccurred())

	return stemcellTempDir
}

var _ = BeforeSuite(func() {
	extractedStemcellTempDir = extractStemcell()
})

var _ = AfterSuite(func() {
	os.RemoveAll(extractedStemcellTempDir)
	gexec.CleanupBuildArtifacts()
})
