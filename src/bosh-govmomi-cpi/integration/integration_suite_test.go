package integration_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
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

func gexecCommandWithStdin(commandBin string, commandArgs ...string) (*gexec.Session, io.WriteCloser) {
	command := exec.Command(commandBin, commandArgs...)
	stdin, err := command.StdinPipe()
	Expect(err).ToNot(HaveOccurred())

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	return session, stdin
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
