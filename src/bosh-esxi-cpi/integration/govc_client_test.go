package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-esxi-cpi/govc/fakes"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"time"

	"bosh-esxi-cpi/govc"
	"fmt"
	"os"
)

var _ = Describe("Govc Client", func() {
	var client govc.GovcClient
	var esxUrl = fmt.Sprintf(
		"https://%s:%s@%s",
		os.Getenv("VCENTER_USER"),
		os.Getenv("VCENTER_PASSWORD"),
		os.Getenv("VCENTER_HOST"),
	)
	var esxNetworkName = os.Getenv("VCENTER_NETWORK_NAME")

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelDebug)
		runner := govc.NewGovcRunner(logger)
		config := &fakegovc.FakeGovcConfig{}
		config.EsxUrlReturns(esxUrl)
		client = govc.NewClient(runner, config, logger)
	})

	Describe("full lifecycle", func() {
		It("runs the govc command", func() {
			vmId := "vm-virtualmachine"
			stemcellId := "cs-stemcell"
			var result string
			var found bool
			var err error

			ovfPath := "../test/fixtures/test.ovf"
			result, err = client.ImportOvf(ovfPath, stemcellId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(""))

			found, err = client.HasVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(Equal(false))

			result, err = client.CloneVM(stemcellId, vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(""))

			found, err = client.HasVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(Equal(true))

			err = client.SetVMNetworkAdapter(vmId, esxNetworkName, "00:50:56:3F:00:00")
			Expect(err).ToNot(HaveOccurred())

			err = client.SetVMResources(vmId, 2, 1024)
			Expect(err).ToNot(HaveOccurred())

			err = client.CreateEphemeralDisk(vmId, 2048)
			Expect(err).ToNot(HaveOccurred())

			err = client.CreateDisk("disk-1", 3096)
			Expect(err).ToNot(HaveOccurred())

			err = client.AttachDisk(vmId, "disk-1")
			Expect(err).ToNot(HaveOccurred())

			envIsoPath := "../test/fixtures/env.iso"
			result, err = client.UpdateVMIso(vmId, envIsoPath)
			Expect(err).ToNot(HaveOccurred())

			result, err = client.StartVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("success"))

			time.Sleep(1 * time.Second)

			err = client.DetachDisk(vmId, "disk-1")
			Expect(err).ToNot(HaveOccurred())

			result, err = client.DestroyVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(""))

			time.Sleep(1 * time.Second)

			found, err = client.HasVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(Equal(false))

			err = client.DestroyDisk("disk-1")
			Expect(err).ToNot(HaveOccurred())

			result, err = client.DestroyVM(stemcellId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(""))
		})
	})

	Describe("partial state", func() {
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
	})

	Describe("empty state", func() {
		It("does not fail with nonexistant vms", func() {
			vmId := "doesnt-exist"
			result, err := client.DestroyVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(""))

			found, err := client.HasVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(Equal(false))
		})
	})
})
