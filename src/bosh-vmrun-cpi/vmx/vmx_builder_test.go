package vmx_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"

	"bosh-vmrun-cpi/vmx"

	govmx "github.com/hooklift/govmx"
)

var _ = Describe("VmxBuilder", func() {
	var logger *fakelogger.FakeLogger
	var builder vmx.VmxBuilder
	var vmxPath string

	BeforeEach(func() {
		var err error

		vmxBytes, err := ioutil.ReadFile(filepath.Join("..", "test", "fixtures", "image.vmx"))
		Expect(err).ToNot(HaveOccurred())

		vmxFile, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(vmxFile.Name(), vmxBytes, 0644)
		Expect(err).ToNot(HaveOccurred())

		vmxPath = vmxFile.Name()
		logger = &fakelogger.FakeLogger{}
		builder = vmx.NewVmxBuilder(logger)
	})

	AfterEach(func() {
		os.Remove(vmxPath)
	})

	Describe("VMInfo", func() {
		It("reads the fixture", func() {
			vmxVM, err := builder.GetVmx(vmxPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(vmxVM.DisplayName).To(Equal("vm-virtualmachine"))
		})
	})

	Describe("InitHardware", func() {
		It("configs basic hardware settings", func() {
			err := builder.InitHardware(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmxVM.VHVEnable).To(Equal(true))
			Expect(vmxVM.Tools.SyncTime).To(Equal(true))
		})

		It("configs swap optimizations", func() {
			err := builder.InitHardware(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmxVM.MinVmMemPct).To(Equal(100))
			Expect(vmxVM.MemTrimRate).To(Equal(0))
			Expect(vmxVM.UseNamedFile).To(Equal(false))
			Expect(vmxVM.Pshare).To(Equal(false))
			Expect(vmxVM.UseRecommendedLockedMemSize).To(Equal(true))
			Expect(vmxVM.MainmemBacking).To(Equal("swap"))
		})
	})

	Describe("AddNetworkInterface", func() {
		It("add a NIC", func() {
			err := builder.AddNetworkInterface("fooNetwork", "00:11:22:33:44:55", vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmxVM.Ethernet[0].VNetwork).To(Equal("fooNetwork"))
			Expect(vmxVM.Ethernet[0].Address).To(Equal("00:11:22:33:44:55"))
			Expect(vmxVM.Ethernet[0].AddressType).To(Equal(govmx.EthernetAddressType("static")))
			Expect(vmxVM.Ethernet[0].VirtualDev).To(Equal("vmxnet3"))
			Expect(vmxVM.Ethernet[0].Present).To(BeTrue())
		})
	})

	Describe("SetVMResources", func() {
		It("sets cpu and mem", func() {
			err := builder.SetVMResources(2, 4096, vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmxVM.NumvCPUs).To(Equal(uint(2)))
			Expect(vmxVM.Memsize).To(Equal(uint(4096)))
		})
	})

	Describe("AttachDisk", func() {
		It("adds a disk entry", func() {
			err := builder.AttachDisk(filepath.Join("disk", "path.vmdk"), vmxPath)
			Expect(err).ToNot(HaveOccurred())

			err = builder.AttachDisk(filepath.Join("disk", "path.vmdk"), vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			disks := vmxVM.SCSIDevices

			Expect(disks[1].Filename).To(Equal("image.vmdk"))
			Expect(disks[1].Present).To(BeTrue())
			Expect(disks[2].Filename).To(Equal(filepath.Join("disk", "path.vmdk")))
			Expect(disks[2].Present).To(BeTrue())
			Expect(disks[3].Filename).To(Equal(filepath.Join("disk", "path.vmdk")))
			Expect(disks[3].Present).To(BeTrue())
		})
	})

	Describe("DetachDisk", func() {
		Context("when disk is attached", func() {
			It("adds a disk entry", func() {
				var err error
				var vmxVM *vmx.VM

				vmxVM, err = builder.GetVmx(vmxPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(vmxVM.SCSIDevices)).To(Equal(2))

				err = builder.AttachDisk(filepath.Join("disk", "path.vmdk"), vmxPath)
				Expect(err).ToNot(HaveOccurred())

				err = builder.DetachDisk("image.vmdk", vmxPath)
				Expect(err).ToNot(HaveOccurred())

				vmxVM, err = builder.GetVmx(vmxPath)
				Expect(err).ToNot(HaveOccurred())

				disks := vmxVM.SCSIDevices

				Expect(len(disks)).To(Equal(2))

				Expect(disks[1].Filename).To(Equal(filepath.Join("disk", "path.vmdk")))
				Expect(disks[1].Present).To(BeTrue())
			})
		})
	})

	Describe("AttachCdrom", func() {
		It("overwrites the cdrom entry", func() {
			err := builder.AttachCdrom(filepath.Join("disk", "path.iso"), vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmxVM.IDEDevices[0].Filename).To(Equal(filepath.Join("disk", "path.iso")))
			Expect(vmxVM.IDEDevices[0].Type).To(Equal(govmx.CDROM_IMAGE))
			Expect(vmxVM.IDEDevices[0].Present).To(BeTrue())
		})
	})
})
