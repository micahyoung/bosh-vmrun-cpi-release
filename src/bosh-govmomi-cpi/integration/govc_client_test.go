package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-govmomi-cpi/govc"
)

var _ = Describe("GovcClient", func() {
	It("runs the govc command", func() {
		logger := boshlog.NewLogger(boshlog.LevelDebug)
		runner := govc.NewGovcRunner(logger)
		config := &fakegovc.FakeGovcConfig{}
		config.EsxUrlReturns("https://root:homelabnyc@172.16.125.131")

		vmId := "virtualmachine"
		stemcellId := "stemcell"
		client := govc.NewClient(runner, config, logger)

		ovfPath := "../test/fixtures/test.ovf"
		result, err := client.ImportOvf(ovfPath, stemcellId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))

		result, err = client.CloneVM(stemcellId, vmId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))

		envIsoPath := "../test/fixtures/env.iso"
		result, err = client.UpdateVMIso(vmId, envIsoPath)
		Expect(err).ToNot(HaveOccurred())

		result, err = client.StartVM(vmId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))

		result, err = client.DestroyVM(vmId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))

		result, err = client.DestroyStemcell(stemcellId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))
	})
})
