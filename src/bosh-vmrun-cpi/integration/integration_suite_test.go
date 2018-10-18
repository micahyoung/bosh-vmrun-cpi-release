package integration_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var CpiConfigPath string
var vmStoreDir string

func GexecCommandWithStdin(commandBin string, commandArgs ...string) (*gexec.Session, io.WriteCloser) {
	command := exec.Command(commandBin, commandArgs...)
	stdin, err := command.StdinPipe()
	Expect(err).ToNot(HaveOccurred())

	session, err := gexec.Start(command, GinkgoWriter, os.Stderr)
	Expect(err).ToNot(HaveOccurred())

	return session, stdin
}

var configTemplate, _ = template.New("parse").Parse(`{
	"cloud": {
		"plugin": "vsphere",
		"properties": {
			"vmrun": {
				"vm_store_path": "{{.VmStorePath}}",
				"vmrun_bin_path": "{{.VmrunBinPath}}",
				"vdiskmanager_bin_path": "{{.VdiskmanagerBinPath}}",
				"ovftool_bin_path": "{{.OvftoolBinPath}}",
				"stemcell_store_path": "{{.StemcellStorePath}}",
				"vm_soft_shutdown_max_wait_seconds": 1,
				"vm_start_max_wait_seconds": 10
			},
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
}`)

func generateCPIConfig() (string, string) {
	stemcellStoreDir := "../../../ci/state/stemcell-store"
	vmStoreTempDir, err := ioutil.TempDir("", "vm-store-path-")
	Expect(err).ToNot(HaveOccurred())

	var configValues = struct {
		VmStorePath         string
		VmrunBinPath        string
		VdiskmanagerBinPath string
		OvftoolBinPath      string
		StemcellStorePath   string
	}{
		VmStorePath:         vmStoreTempDir,
		VmrunBinPath:        requirePath("vmrun"),
		VdiskmanagerBinPath: requirePath("vmware-vdiskmanager"),
		OvftoolBinPath:      requirePath("ovftool"),
		StemcellStorePath:   stemcellStoreDir,
	}

	configFile, err := ioutil.TempFile("", "config")
	Expect(err).ToNot(HaveOccurred())

	var configContent bytes.Buffer
	configTemplate.Execute(&configContent, configValues)

	configFile.WriteString(configContent.String())
	configPath, err := filepath.Abs(configFile.Name())
	Expect(err).ToNot(HaveOccurred())

	return configPath, vmStoreTempDir
}

func requirePath(bin string) string {
	path, err := exec.LookPath(bin)
	if err != nil {
		panic("test requires bin: " + bin)
	}
	return path
}

var _ = BeforeSuite(func() {
	CpiConfigPath, vmStoreDir = generateCPIConfig()
})

var _ = AfterSuite(func() {
	os.RemoveAll(CpiConfigPath)
	os.RemoveAll(vmStoreDir)
	gexec.CleanupBuildArtifacts()
})
