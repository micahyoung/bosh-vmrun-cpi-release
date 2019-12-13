package integration_test

import (
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

		It("fails with a useful message", func() {
			command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")
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

		It("fails with a useful message", func() {
			command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Command not set correctly in authorized_key"))
			Eventually(session.Err).Should(gbytes.Say("ssh: handshake failed"))
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

	Context("when config has valid ssh settings", func() {
		Context("and vm store path is present", func() {
			BeforeEach(func() {
				generateCPIConfig(CpiConfigPath, SSHCPIConfig)
			})

			It("runs the installer to install the CPI locally", func() {
				command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))

				resultPath := strings.TrimSpace(string(session.Out.Contents()))
				Expect(resultPath).To(MatchRegexp(`vm-store-path.*cpi`))
				Expect(resultPath).To(BeAnExistingFile())
			})

			It("reuses already installed CPI over SSH", func() {
				command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Eventually(session.Err).Should(gbytes.Say("Installing local cpi"))

				// running again will reuse
				command = exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")
				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)

				Eventually(session).Should(gexec.Exit(0))
				Eventually(session.Err).Should(gbytes.Say("Using existing remote cpi"))
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

		Context("when vm store dir is not present locally, install over SSH", func() {
			var notPrexistingDir string
			BeforeEach(func() {
				notPrexistingDir, _ = ioutil.TempDir("", "not-preexisting-vm-store-path")
				Expect(os.RemoveAll(notPrexistingDir)).To(Succeed())

				nonExistantVmStorePathSSHConfig := SSHCPIConfig
				nonExistantVmStorePathSSHConfig.VmStorePath = template.JSEscapeString(notPrexistingDir)
				generateCPIConfig(CpiConfigPath, nonExistantVmStorePathSSHConfig)
			})

			AfterEach(func() {
				Expect(os.RemoveAll(notPrexistingDir)).To(Succeed())
			})

			It("runs the installer to install the CPI over SSH", func() {
				command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))

				resultPath := strings.TrimSpace(string(session.Out.Contents()))
				Expect(resultPath).To(MatchRegexp(`not-preexisting-vm-store-path.*cpi`))
				Expect(resultPath).To(BeAnExistingFile())
			})

			It("reuses already installed CPI over SSH", func() {
				command := exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Eventually(session.Err).Should(gbytes.Say("Installing remote cpi"))

				// running again will reuse
				command = exec.Command(installerBin, "-configPath", CpiConfigPath, "install-cpi")
				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)

				Eventually(session).Should(gexec.Exit(0))
				Eventually(session.Err).Should(gbytes.Say("Using existing remote cpi"))
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
	})
})
