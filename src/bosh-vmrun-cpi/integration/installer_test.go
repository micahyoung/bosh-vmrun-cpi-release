package integration_test

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("installer integration", func() {
	var installerBin string

	BeforeEach(func() {
		var err error

		installerBin, err = gexec.Build("bosh-vmrun-cpi/cmd/installer", "-ldflags", "-X main.version=1.2.3")
		Expect(err).ToNot(HaveOccurred())

		cpiTmpBinPath, err := gexec.Build("bosh-vmrun-cpi/cmd/cpi", "-ldflags", "-X main.version=1.2.3")
		Expect(err).ToNot(HaveOccurred())

		cpiBinContent, err := ioutil.ReadFile(cpiTmpBinPath)
		Expect(err).ToNot(HaveOccurred())

		cpiBinPath := filepath.Join(filepath.Dir(installerBin), fmt.Sprintf("cpi-%s", runtime.GOOS))
		err = ioutil.WriteFile(cpiBinPath, cpiBinContent, 0777)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when config is missing SSH settings", func() {
		BeforeEach(func() {
			missingSSHCPIConfig := SSHCPIConfig
			missingSSHCPIConfig.SshHostname = ""
			missingSSHCPIConfig.SshUsername = ""
			missingSSHCPIConfig.SshPrivateKey = ""
			missingSSHCPIConfig.SshPlatform = ""
			generateCPIConfig(CpiConfigPath, missingSSHCPIConfig)
		})

		It("fails install-cpi with a useful message", func() {
			command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("ssh: no key found"))
		})

		It("fails sync-director-stemcells with a useful message", func() {
			command := exec.Command(installerBin, "-configPath", CpiConfigPath, "sync-director-stemcells")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("ssh: no key found"))
		})

		It("returns base64-encoded config", func() {
			command := exec.Command(installerBin, "-configPath", CpiConfigPath, "encoded-config")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			config, err := base64.StdEncoding.DecodeString(string(session.Out.Contents()))
			Expect(err).ToNot(HaveOccurred())
			Expect(len(config)).To(BeNumerically(">", 0))
		})
	})

	Context("when ssh authorized key is missing SSH settings", func() {
		BeforeEach(func() {
			notAuthorizedSSHCPIConfig := SSHCPIConfig
			notAuthorizedSSHCPIConfig.SshUsername = "notexisting"
			generateCPIConfig(CpiConfigPath, notAuthorizedSSHCPIConfig)
		})

		It("fails install-cpi with a useful message", func() {
			command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Command not set correctly in authorized_key"))
			Eventually(session.Err).Should(gbytes.Say("ssh: handshake failed"))
		})
	})

	Context("when config has valid ssh settings", func() {
		Context("and directories exist", func() {
			BeforeEach(func() {
				generateCPIConfig(CpiConfigPath, SSHCPIConfig)
			})

			Context("install-cpi", func() {
				Context("when vm store path is present", func() {
					It("installs successfully", func() {
						command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")

						session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Expect(err).ToNot(HaveOccurred())

						Eventually(session).Should(gexec.Exit(0))
						Eventually(session.Err).Should(gbytes.Say("Installing remote cpi"))

						resultPath := strings.TrimSpace(string(session.Out.Contents()))
						Expect(resultPath).To(MatchRegexp(`vm-store-path.*cpi`))
						Expect(resultPath).To(BeAnExistingFile())

						// running again will reuse
						command = exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)

						Eventually(session).Should(gexec.Exit(0))
						Eventually(session.Err).Should(gbytes.Say("Using existing remote cpi"))
					})
				})
			})
		})

		Context("and vm store and stemcell store dir do not exist", func() {
			var notPrexistingVmStoreDir string
			var notPrexistingStemcellStoreDir string
			BeforeEach(func() {
				notPrexistingVmStoreDir, _ = ioutil.TempDir("", "vm-store-path-not-preexisting")
				notPrexistingStemcellStoreDir, _ = ioutil.TempDir("", "stemcell-store-path-not-preexisting")
				Expect(os.RemoveAll(notPrexistingVmStoreDir)).To(Succeed())
				Expect(os.RemoveAll(notPrexistingStemcellStoreDir)).To(Succeed())

				nonExistantDirsSSHConfig := SSHCPIConfig
				nonExistantDirsSSHConfig.VmStorePath = template.JSEscapeString(notPrexistingVmStoreDir)
				nonExistantDirsSSHConfig.StemcellStorePath = template.JSEscapeString(notPrexistingStemcellStoreDir)
				generateCPIConfig(CpiConfigPath, nonExistantDirsSSHConfig)
			})

			AfterEach(func() {
				Expect(os.RemoveAll(notPrexistingVmStoreDir)).To(Succeed())
				Expect(os.RemoveAll(notPrexistingStemcellStoreDir)).To(Succeed())
			})

			Context("install-cpi", func() {
				It("runs the installer to install the CPI over SSH", func() {
					command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")

					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					Eventually(session).Should(gexec.Exit(0))
					Eventually(session.Err).Should(gbytes.Say("Installing remote cpi"))

					resultPath := strings.TrimSpace(string(session.Out.Contents()))
					Expect(resultPath).To(MatchRegexp(`vm-store-path-not-preexisting.*cpi`))
					Expect(resultPath).To(BeAnExistingFile())

					// running again will reuse
					command = exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)

					Eventually(session).Should(gexec.Exit(0))
					Eventually(session.Err).Should(gbytes.Say("Using existing remote cpi"))
				})
			})

			Context("sync-director-stemcells", func() {
				It("copy stemcell temp files", func() {
					var err error
					directorStemcellTempPath, err := ioutil.TempDir("", "director-data-tmp-")
					Expect(err).ToNot(HaveOccurred())
					defer os.RemoveAll(directorStemcellTempPath)

					command := exec.Command(installerBin, "-configPath", CpiConfigPath, "-directorTmpDirPath", directorStemcellTempPath, "sync-director-stemcells")

					stemcellTempPath := filepath.Join(directorStemcellTempPath, "stemcell-abc123")
					Expect(os.MkdirAll(stemcellTempPath, 0777)).To(Succeed())

					stemcellImagePath := filepath.Join(stemcellTempPath, "image")
					stemcellManifestPath := filepath.Join(stemcellTempPath, "stemcell.MF")
					stemcellManifestContent := strings.TrimSpace(`
name: bosh-vsphere-esxi-ubuntu-xenial-go_agent
version: '97.16'
`)
					err = ioutil.WriteFile(stemcellManifestPath, []byte(stemcellManifestContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = ioutil.WriteFile(stemcellImagePath, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					Eventually(session).Should(gexec.Exit(0))

					expectedStemcellPath := filepath.Join(notPrexistingStemcellStoreDir, "bosh-stemcell-97.16-vsphere-esxi-ubuntu-xenial-go_agent.tgz")
					Expect(expectedStemcellPath).To(BeAnExistingFile())
					expectedMappingPath := filepath.Join(notPrexistingStemcellStoreDir, "mappings", fmt.Sprintf("%x.mapping", sha1.Sum([]byte(stemcellImagePath))))
					Expect(expectedMappingPath).To(BeAnExistingFile())

					// running again will reuse stemcell but regenerate mapping
					Expect(os.RemoveAll(expectedMappingPath)).To(Succeed())

					command = exec.Command(installerBin, "-configPath", CpiConfigPath, "-directorTmpDirPath", directorStemcellTempPath, "sync-director-stemcells")
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)

					Eventually(session).Should(gexec.Exit(0))
					Eventually(session.Err).Should(gbytes.Say("remote stemcell already exists: bosh-stemcell-97.16-vsphere-esxi-ubuntu-xenial-go_agent.tgz"))

					Expect(expectedMappingPath).To(BeAnExistingFile())

					//verify rebuilt stemcell is valid
					command = exec.Command("tar", "-t", "-z", "-f", expectedStemcellPath)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Eventually(session).Should(gexec.Exit(0))
					Expect(string(session.Out.Contents())).To(ContainSubstring("stemcell.MF"))
					Expect(string(session.Out.Contents())).To(ContainSubstring("image"))
				})
			})
		})
	})
})
