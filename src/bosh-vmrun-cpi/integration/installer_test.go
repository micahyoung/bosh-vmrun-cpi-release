package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("installer integration", func() {
	It("runs the installer", func() {
		installerBin, err := gexec.Build("bosh-vmrun-cpi/cmd/installer")
		Expect(err).ToNot(HaveOccurred())

		cpiBin := filepath.Join(filepath.Dir(installerBin), "cpi-windows")
		f, err := os.Create(cpiBin)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		command := exec.Command(installerBin, "-platform", "windows", "-configPath", CpiConfigPath)

		session, err := gexec.Start(command, GinkgoWriter, os.Stderr)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		resultPath := strings.TrimSpace(string(session.Out.Contents()))
		Expect(resultPath).To(MatchRegexp(`vm-store-path.*cpi`))
		Expect(resultPath).To(BeAnExistingFile())
	})
})
