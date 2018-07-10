package integration_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	cpiconfig "bosh-vmrun-cpi/config"
	"bosh-vmrun-cpi/driver"
)

var _ = Describe("driver integration", func() {
	var client driver.Client
	var esxNetworkName = os.Getenv("NETWORK_NAME")
	var vmId = "vm-virtualmachine"
	var stemcellId = "cs-stemcell"

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelDebug)
		boshRunner := boshsys.NewExecCmdRunner(logger)
		fs := boshsys.NewOsFileSystem(logger)
		cpiConfig, err := cpiconfig.NewConfigFromPath(CpiConfigPath, fs)
		Expect(err).ToNot(HaveOccurred())

		config := driver.NewConfig(cpiConfig)
		vmrunRunner := driver.NewVmrunRunner(config.VmrunPath(), boshRunner, logger)
		ovftoolRunner := driver.NewOvftoolRunner(config.OvftoolPath(), boshRunner, logger)
		vdiskmanagerRunner := driver.NewVdiskmanagerRunner(config.VdiskmanagerPath(), boshRunner, logger)
		vmxBuilder := driver.NewVmxBuilder(logger)
		client = driver.NewClient(vmrunRunner, ovftoolRunner, vdiskmanagerRunner, vmxBuilder, config, logger)
	})

	AfterEach(func() {
		if client.HasVM(vmId) {
			err := client.DestroyVM(vmId)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	Describe("full lifecycle", func() {
		It("runs the commands", func() {
			var success bool
			var found bool
			var err error
			var vmInfo driver.VMInfo

			ovfPath := "../test/fixtures/test.ovf"
			success, err = client.ImportOvf(ovfPath, stemcellId)
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(Equal(true))

			found = client.HasVM(vmId)
			Expect(found).To(Equal(false))

			err = client.CloneVM(stemcellId, vmId)
			Expect(err).ToNot(HaveOccurred())

			found = client.HasVM(vmId)
			Expect(found).To(Equal(true))

			vmInfo, err = client.GetVMInfo(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmInfo.Name).To(Equal(vmId))

			err = client.SetVMNetworkAdapter(vmId, esxNetworkName, "00:50:56:3F:00:00")
			Expect(err).ToNot(HaveOccurred())

			vmInfo, err = client.GetVMInfo(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmInfo.NICs[0].Network).To(Equal(esxNetworkName))
			Expect(vmInfo.NICs[0].MAC).To(Equal("00:50:56:3F:00:00"))

			err = client.SetVMResources(vmId, 2, 1024)
			Expect(err).ToNot(HaveOccurred())

			vmInfo, err = client.GetVMInfo(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmInfo.CPUs).To(Equal(2))
			Expect(vmInfo.RAM).To(Equal(1024))

			err = client.CreateEphemeralDisk(vmId, 2048)
			Expect(err).ToNot(HaveOccurred())

			vmInfo, err = client.GetVMInfo(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmInfo.Disks[2].Path).To(HaveSuffix("ephemeral-disks/vm-virtualmachine.vmdk"))

			err = client.CreateDisk("disk-1", 3096)
			Expect(err).ToNot(HaveOccurred())

			err = client.AttachDisk(vmId, "disk-1")
			Expect(err).ToNot(HaveOccurred())

			vmInfo, err = client.GetVMInfo(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmInfo.Disks[3].Path).To(HaveSuffix("persistent-disks/disk-1.vmdk"))

			envIsoPath := "../test/fixtures/env.iso"
			err = client.UpdateVMIso(vmId, envIsoPath)
			Expect(err).ToNot(HaveOccurred())

			err = client.StartVM(vmId)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(1 * time.Second)

			err = client.DetachDisk(vmId, "disk-1")
			Expect(err).ToNot(HaveOccurred())

			err = client.StopVM(vmId)
			Expect(err).ToNot(HaveOccurred())

			err = client.DestroyVM(vmId)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(1 * time.Second)

			found = client.HasVM(vmId)
			Expect(found).To(Equal(false))

			err = client.DestroyDisk("disk-1")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("partial state", func() {
		It("destroys unstarted vms", func() {
			vmId := "vm-virtualmachine"
			var success bool
			var err error

			err = client.DestroyVM(vmId)
			Expect(err).ToNot(HaveOccurred())

			ovfPath := "../test/fixtures/test.ovf"
			success, err = client.ImportOvf(ovfPath, vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(Equal(true))

			envIsoPath := "../test/fixtures/env.iso"
			err = client.UpdateVMIso(vmId, envIsoPath)
			Expect(err).ToNot(HaveOccurred())

			err = client.DestroyVM(vmId)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("empty state", func() {
		It("does not fail with nonexistant vms", func() {
			vmId := "doesnt-exist"
			err := client.DestroyVM(vmId)
			Expect(err).ToNot(HaveOccurred())

			found := client.HasVM(vmId)
			Expect(found).To(Equal(false))
		})
	})
})
