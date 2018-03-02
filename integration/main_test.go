package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("CPI", func() {
	It("runs the cpi", func() {
		cpiBin, err := gexec.Build("github.com/micahyoung/bosh-govmomi-cpi/cmd")
		Expect(err).ToNot(HaveOccurred())

		command := exec.Command(cpiBin)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("{}"))
	})
})
