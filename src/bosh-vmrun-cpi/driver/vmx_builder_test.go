package driver_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"

	"bosh-vmrun-cpi/driver"
	vmx "bosh-vmrun-cpi/vmx"
	govmx "github.com/hooklift/govmx"
)

var _ = Describe("VmxBuilder", func() {
	var logger *fakelogger.FakeLogger
	var builder driver.VmxBuilder
	var vmxPath string

	BeforeEach(func() {
		var err error

		vmxBytes, err := ioutil.ReadFile("../test/fixtures/test.vmx")
		Expect(err).ToNot(HaveOccurred())

		vmxFile, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())

		err = ioutil.WriteFile(vmxFile.Name(), vmxBytes, 0644)
		Expect(err).ToNot(HaveOccurred())

		vmxPath = vmxFile.Name()
		logger = &fakelogger.FakeLogger{}
		builder = driver.NewVmxBuilder(logger)
	})

	AfterEach(func() {
		os.Remove(vmxPath)
	})

	Describe("VMInfo", func() {
		It("reads the fixture", func() {
			vmInfo, err := builder.VMInfo(vmxPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(vmInfo.Name).To(Equal("vm-virtualmachine"))
			Expect(vmInfo.NICs).To(BeEmpty())
		})
	})

	Describe("InitHardware", func() {
		It("configs basic hardware settings", func() {
			err := builder.InitHardware(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmxVM.VHVEnable).To(BeTrue())
			Expect(vmxVM.Tools.SyncTime).To(BeTrue())
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
			err := builder.AttachDisk("/disk/path.vmdk", vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			disks := vmxVM.SCSIDevices

			Expect(disks[1].Filename).To(Equal("vm-virtualmachine-disk1.vmdk"))
			Expect(disks[1].Present).To(BeTrue())
			Expect(disks[2].Filename).To(Equal("/disk/path.vmdk"))
			Expect(disks[2].Present).To(BeTrue())
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

				err = builder.DetachDisk("vm-virtualmachine-disk1.vmdk", vmxPath)
				Expect(err).ToNot(HaveOccurred())

				vmxVM, err = builder.GetVmx(vmxPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(vmxVM.SCSIDevices)).To(Equal(1))
			})
		})
	})

	Describe("AttachCdrom", func() {
		It("overwrites the cdrom entry", func() {
			err := builder.AttachCdrom("/disk/path.iso", vmxPath)
			Expect(err).ToNot(HaveOccurred())

			vmxVM, err := builder.GetVmx(vmxPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmxVM.IDEDevices[0].Filename).To(Equal("/disk/path.iso"))
			Expect(vmxVM.IDEDevices[0].Type).To(Equal(govmx.CDROM_IMAGE))
			Expect(vmxVM.IDEDevices[0].Present).To(BeTrue())
		})
	})
})
