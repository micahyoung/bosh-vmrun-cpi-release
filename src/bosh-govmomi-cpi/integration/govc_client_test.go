package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-govmomi-cpi/govc"
)

var _ = Describe("GovcClient", func() {
	var client govc.GovcClient

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelDebug)
		runner := govc.NewGovcRunner(logger)
		config := &fakegovc.FakeGovcConfig{}
		config.EsxUrlReturns("https://root:homelabnyc@172.16.125.131")
		client = govc.NewClient(runner, config, logger)
	})

	It("runs the govc command", func() {
		vmId := "vm-virtualmachine"
		stemcellId := "cs-stemcell"
		var result string
		var err error

		ovfPath := "../test/fixtures/test.ovf"
		result, err = client.ImportOvf(ovfPath, stemcellId)
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

		result, err = client.DestroyVM(stemcellId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))
	})

	It("destroys unstarted vms", func() {
		vmId := "vm-virtualmachine"
		var result string
		var err error

		result, err = client.DestroyVM(vmId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))

		ovfPath := "../test/fixtures/test.ovf"
		result, err = client.ImportOvf(ovfPath, vmId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))

		envIsoPath := "../test/fixtures/env.iso"
		result, err = client.UpdateVMIso(vmId, envIsoPath)
		Expect(err).ToNot(HaveOccurred())

		result, err = client.DestroyVM(vmId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))
	})

	It("destroys nonexistant vms", func() {
		vmId := "doesnt-exist"
		result, err := client.DestroyVM(vmId)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(""))
	})

})
