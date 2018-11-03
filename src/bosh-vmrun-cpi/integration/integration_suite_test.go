package integration_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
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
var stemcellStoreDir string
var testStemcellUrl = "https://s3.amazonaws.com/bosh-core-stemcells/vsphere/bosh-stemcell-3586.42-vsphere-esxi-ubuntu-trusty-go_agent.tgz"

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
				"vm_soft_shutdown_max_wait_seconds": 20,
				"vm_start_max_wait_seconds": 20
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

func generateCPIConfig(configFile *os.File, vmStoreDir, stemcellStoreDir string) {
	var err error
	var configValues = struct {
		VmStorePath         string
		VmrunBinPath        string
		VdiskmanagerBinPath string
		OvftoolBinPath      string
		StemcellStorePath   string
	}{
		VmStorePath:         template.JSEscapeString(vmStoreDir),
		VmrunBinPath:        template.JSEscapeString(requirePath("vmrun")),
		VdiskmanagerBinPath: template.JSEscapeString(requirePath("vmware-vdiskmanager")),
		OvftoolBinPath:      template.JSEscapeString(requirePath("ovftool")),
		StemcellStorePath:   template.JSEscapeString(stemcellStoreDir),
	}

	var configContent bytes.Buffer
	err = configTemplate.Execute(&configContent, configValues)
	Expect(err).ToNot(HaveOccurred())

	_, err = configFile.WriteString(configContent.String())
	Expect(err).ToNot(HaveOccurred())
}

func requirePath(bin string) string {
	path, _ := exec.LookPath(bin)
	if path == "" {
		path, _ = exec.LookPath(bin + ".exe")
	}

	if path == "" {
		panic("test requires bin: " + bin)
	}
	return path
}

func getTestStemcell(testStemcellUrl, stemcellPath string) {
	var err error

	if _, err := os.Stat(stemcellPath); !os.IsNotExist(err) {
		return
	}

	output, err := os.Create(stemcellPath)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	response, err := http.Get(testStemcellUrl)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		panic(err)
	}
}

var _ = BeforeSuite(func() {
	var err error

	stemcellStoreDir := filepath.Join("..", "..", "..", "ci", "state", "stemcell-store")
	err = os.MkdirAll(stemcellStoreDir, 0777)
	Expect(err).ToNot(HaveOccurred())
	stemcellPath := filepath.Join(stemcellStoreDir, "stemcell.tgz")

	vmStoreDir, err := ioutil.TempDir("", "vm-store-path-")
	Expect(err).ToNot(HaveOccurred())

	configFile, err := ioutil.TempFile("", "config")
	Expect(err).ToNot(HaveOccurred())

	CpiConfigPath, err = filepath.Abs(configFile.Name())

	getTestStemcell(testStemcellUrl, stemcellPath)
	generateCPIConfig(configFile, vmStoreDir, stemcellStoreDir)
})

var _ = AfterSuite(func() {
	os.RemoveAll(CpiConfigPath)
	os.RemoveAll(vmStoreDir)
	gexec.CleanupBuildArtifacts()
})
