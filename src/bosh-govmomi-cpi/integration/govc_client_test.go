package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-govmomi-cpi/govc"
)

var _ = Describe("GovcClient", func() {
	FIt("runs the govc command", func() {
		logger := boshlog.NewLogger(boshlog.LevelInfo)
		runner := govc.NewGovcRunner(logger)
		config := &fakegovc.FakeGovcConfig{}
		config.EsxUrlReturns("https://root:homelabnyc@172.16.125.131")

		vmId := "vm-cid"
		stemcellId := "d61a115a-f7ec-4ede-4392-c26da3293453"
		client := govc.NewClient(runner, config, logger)
		result, err := client.CloneVM(stemcellId, vmId)
		Expect(err).ToNot(HaveOccurred())

		Expect(result).To(Equal("success"))
	})
})
