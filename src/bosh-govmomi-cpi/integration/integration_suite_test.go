package integration_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	session, err := gexec.Start(command, GinkgoWriter, os.Stderr)
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

var configContent = `{
	"cloud": {
		"plugin": "vsphere",
		"properties": {
			"vcenters": [
			{
				"host": "10.10.1.3",
				"user": "root",
				"password": "homelabnyc",
				"datacenters": [
				{
					"name": "ha-datacenter",
					"vm_folder": "BOSH_VMs",
					"template_folder": "BOSH_Templates",
					"disk_path": "bosh_disks",
					"datastore_pattern": "datastore1"
				}
				]
			}
			],
			"agent": {
				"ntp": [
				],
				"blobstore": {
					"provider": "local",
					"options": {
						"blobstore_path": "/var/vcap/micro_bosh/data/cache"
					}
				},
				"mbus": "https://mbus:p2an3m7idfm6vmqp3w74@0.0.0.0:6868"
			}
		}
	}
}`

func GenerateCPIConfig() string {
	configFile, err := ioutil.TempFile("", "config")
	Expect(err).ToNot(HaveOccurred())

	configFile.WriteString(configContent)
	configPath, err := filepath.Abs(configFile.Name())
	Expect(err).ToNot(HaveOccurred())

	return configPath
}

var _ = BeforeSuite(func() {
	extractedStemcellTempDir = extractStemcell()
})

var _ = AfterSuite(func() {
	os.RemoveAll(extractedStemcellTempDir)
	gexec.CleanupBuildArtifacts()
})
