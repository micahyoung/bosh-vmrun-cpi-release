package main_test

import (
	"io/ioutil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("CPI", func() {
	It("runs the cpi", func() {
		configFile, err = ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
		defer configFile.Close()

		cpiBin, err := gexec.Build("bosh-govmomi-cpi/cmd")
		Expect(err).ToNot(HaveOccurred())

		command := exec.Command(cpiBin, "-configPath", "/tmp/cpi.conf")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("{}"))

		err := os.Remove(configFile.Name())
		Expect(err).NotTo(HaveOccurred())

	})
})
