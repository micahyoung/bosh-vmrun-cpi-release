package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-vmrun-cpi/govc"
)

var _ = Describe("GovcRunner", func() {
	It("runs the govc command", func() {
		logger := boshlog.NewLogger(boshlog.LevelInfo)
		runner := govc.NewGovcRunner(logger)
		flags := map[string]string{"u": "foo"}
		result, err := runner.CliCommand("env", flags, nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainSubstring("GOVC_URL=foo"))
	})
})
