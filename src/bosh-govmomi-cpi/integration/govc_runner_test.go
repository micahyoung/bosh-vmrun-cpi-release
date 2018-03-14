package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-govmomi-cpi/govc"
)

var _ = Describe("GovcRunner", func() {
	It("runs the govc command", func() {
		logger := boshlog.NewLogger(boshlog.LevelInfo)
		runner := govc.NewGovcRunner(logger)
		result, err := runner.CliCommand("version", nil, nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("govc 0.17.0\n"))
	})
})
