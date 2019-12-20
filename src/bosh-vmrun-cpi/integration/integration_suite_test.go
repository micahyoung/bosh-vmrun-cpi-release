package integration_test

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

var (
	CpiConfigPath    string
	VmStoreDir       string
	StemcellStoreDir string
	SSHCPIConfig     cpiConfig
	DirectCPIConfig  cpiConfig
	// must match cpi test metadata
	TestStemcellUrl  = "https://s3.amazonaws.com/bosh-core-stemcells/vsphere/bosh-stemcell-3586.42-vsphere-esxi-ubuntu-trusty-go_agent.tgz"
	TestStemcellSha1 = "72212fc00b10f162cfc23c42f3ab20d393970418"
)

var configTemplate, _ = template.New("parse").Parse(`{
	"cloud": {
		"plugin": "vsphere",
		"properties": {
			"vmrun": {
				"vm_store_path": "{{.VmStorePath}}",
				"vmrun_bin_path": "{{.VmrunBinPath}}",
				"ovftool_bin_path": "{{.OvftoolBinPath}}",
				"stemcell_store_path": "{{.StemcellStorePath}}",
				"vm_soft_shutdown_max_wait_seconds": 20,
				"vm_start_max_wait_seconds": 20,
				"use_linked_cloning": true,
				"enable_human_readable_name": true,
				"ssh_tunnel":{
					"host":"{{.SshHostname}}",
					"port":"{{.SshPort}}",
					"username":"{{.SshUsername}}",
					"private_key":"{{.SshPrivateKey}}",
					"platform":"{{.SshPlatform}}"
				}
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

type cpiConfig struct {
	VmStorePath       string
	VmrunBinPath      string
	OvftoolBinPath    string
	StemcellStorePath string
	SshHostname       string
	SshPort           string
	SshUsername       string
	SshPrivateKey     string
	SshPlatform       string
}

func sshCPIConfig(vmStoreDir, stemcellStoreDir string) cpiConfig {
	var config = cpiConfig{
		VmStorePath:       template.JSEscapeString(vmStoreDir),
		VmrunBinPath:      template.JSEscapeString(requirePath("vmrun")),
		OvftoolBinPath:    template.JSEscapeString(requirePath("ovftool")),
		StemcellStorePath: template.JSEscapeString(stemcellStoreDir),
		SshHostname:       template.JSEscapeString(requireEnv("SSH_HOSTNAME")),
		SshPort:           template.JSEscapeString(requireEnv("SSH_PORT")),
		SshUsername:       template.JSEscapeString(requireEnv("SSH_USERNAME")),
		SshPrivateKey:     template.JSEscapeString(requireEnv("SSH_PRIVATE_KEY")),
		SshPlatform:       template.JSEscapeString(requireEnv("SSH_PLATFORM")),
	}

	return config
}

func directCPIConfig(vmStoreDir, stemcellStoreDir string) cpiConfig {
	var config = cpiConfig{
		VmStorePath:       template.JSEscapeString(vmStoreDir),
		VmrunBinPath:      template.JSEscapeString(requirePath("vmrun")),
		OvftoolBinPath:    template.JSEscapeString(requirePath("ovftool")),
		StemcellStorePath: template.JSEscapeString(stemcellStoreDir),
		SshHostname:       "",
		SshPort:           "",
		SshUsername:       "",
		SshPrivateKey:     "",
		SshPlatform:       "",
	}

	return config
}

func generateCPIConfig(configFilePath string, config cpiConfig) {
	var err error

	var configContent bytes.Buffer
	err = configTemplate.Execute(&configContent, &config)
	Expect(err).ToNot(HaveOccurred())

	var configFile *os.File
	configFile, err = os.OpenFile(configFilePath, os.O_TRUNC|os.O_WRONLY, 0666)
	Expect(err).ToNot(HaveOccurred())
	defer configFile.Close()

	_, err = configFile.WriteString(configContent.String())
	Expect(err).ToNot(HaveOccurred())
}

func requireEnv(key string) string {
	val := os.Getenv(key)

	if val == "" {
		panic("test environment variable: " + key)
	}
	return val
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

func getTestStemcell(testStemcellUrl, testStemcellSha1, stemcellPath string) {
	var err error

	if _, err := os.Stat(stemcellPath); os.IsNotExist(err) {
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

	fileReader, err := os.Open(stemcellPath)
	if err != nil {
		panic(err)
	}
	defer fileReader.Close()

	shaWriter := sha1.New()
	if _, err := io.Copy(shaWriter, fileReader); err != nil {
		log.Fatal(err)
	}

	actualSha1 := fmt.Sprintf("%x", shaWriter.Sum(nil))
	if actualSha1 != testStemcellSha1 {
		panic(fmt.Sprintf("Test stemcell shasum mismatch %s != %s", actualSha1, testStemcellSha1))
	}
}

func GexecCommandWithStdin(commandBin string, commandArgs ...string) (*gexec.Session, io.WriteCloser) {
	command := exec.Command(commandBin, commandArgs...)
	stdin, err := command.StdinPipe()
	Expect(err).ToNot(HaveOccurred())

	session, err := gexec.Start(command, GinkgoWriter, os.Stderr)
	Expect(err).ToNot(HaveOccurred())

	return session, stdin
}

var _ = BeforeSuite(func() {
	var err error

	relativeStemcellStoreDir := filepath.Join("..", "..", "..", "ci", "state", "stemcell-store")
	StemcellStoreDir, err = filepath.Abs(relativeStemcellStoreDir)
	Expect(err).ToNot(HaveOccurred())

	err = os.MkdirAll(StemcellStoreDir, 0777)
	Expect(err).ToNot(HaveOccurred())
	stemcellPath := filepath.Join(StemcellStoreDir, "stemcell.tgz")

	getTestStemcell(TestStemcellUrl, TestStemcellSha1, stemcellPath)

	//write mapping for test stemcell
	mappingFileName := fmt.Sprintf("%x.mapping", sha1.Sum([]byte("/test-stemcell-tmp-image-path")))
	mappingFilePath := filepath.Join(StemcellStoreDir, "mappings", mappingFileName)

	err = os.MkdirAll(filepath.Dir(mappingFilePath), 0777)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(mappingFilePath, []byte(stemcellPath), 0666)
	Expect(err).ToNot(HaveOccurred())
})

var _ = BeforeEach(func() {
	var err error
	VmStoreDir, err = ioutil.TempDir("", "vm-store-path-")
	Expect(err).ToNot(HaveOccurred())

	SSHCPIConfig = sshCPIConfig(VmStoreDir, StemcellStoreDir)
	DirectCPIConfig = directCPIConfig(VmStoreDir, StemcellStoreDir)
	_, _ = SSHCPIConfig, DirectCPIConfig

	configFile, err := ioutil.TempFile("", "config")
	Expect(err).ToNot(HaveOccurred())
	configFile.Close()

	CpiConfigPath, err = filepath.Abs(configFile.Name())
	generateCPIConfig(CpiConfigPath, DirectCPIConfig)
})

var _ = AfterEach(func() {
	SSHCPIConfig, DirectCPIConfig = cpiConfig{}, cpiConfig{}
	os.RemoveAll(CpiConfigPath)
	os.RemoveAll(VmStoreDir)
	CpiConfigPath = ""
	VmStoreDir = ""
	//do not clean up StemcellStoreDir as it contains large stemcells that are re-used
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
