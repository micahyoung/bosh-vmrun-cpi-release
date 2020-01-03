package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	cpiconfig "bosh-vmrun-cpi/config"
	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/vmx"
)

var _ = Describe("driver integration", func() {
	var client driver.Client
	var vmId = "vm-virtualmachine"
	var stemcellId = "cs-stemcell"
	var config driver.Config
	var vmrunRunner driver.VmrunRunner
	var ovftoolRunner driver.OvftoolRunner
	var vmxBuilder vmx.VmxBuilder
	var logger boshlog.Logger

	BeforeEach(func() {
		logLevel, err := boshlog.Levelify(os.Getenv("BOSH_LOG_LEVEL"))
		if err != nil {
			logLevel = boshlog.LevelInfo
		}
		logger = boshlog.NewLogger(logLevel)
		boshRunner := boshsys.NewExecCmdRunner(logger)
		fs := boshsys.NewOsFileSystem(logger)
		retryFileLock := driver.NewRetryFileLock(logger)
		vmxBuilder = vmx.NewVmxBuilder(logger)

		generateCPIConfig(CpiConfigPath, DirectCPIConfig)
		cpiConfigJson, err := fs.ReadFileString(CpiConfigPath)
		Expect(err).ToNot(HaveOccurred())
		cpiConfig, err := cpiconfig.NewConfigFromJson(cpiConfigJson)
		Expect(err).ToNot(HaveOccurred())

		config = driver.NewConfig(cpiConfig)

		vmrunRunner = driver.NewVmrunRunner(config.VmrunPath(), retryFileLock, logger)
		Expect(vmrunRunner.Configure()).To(Succeed())

		ovftoolRunner = driver.NewOvftoolRunner(config.OvftoolPath(), boshRunner, logger)
		Expect(ovftoolRunner.Configure()).To(Succeed())
	})

	AfterEach(func() {
		if client.HasVM(vmId) {
			err := client.DestroyVM(vmId)

			Expect(err).ToNot(HaveOccurred())
		}
		if client.HasVM(stemcellId) {
			err := client.DestroyVM(stemcellId)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	Describe("common client options", func() {
		BeforeEach(func() {
			client = driver.NewClient(vmrunRunner, ovftoolRunner, ovftoolRunner, vmxBuilder, config, logger)
		})

		Describe("full lifecycle", func() {
			It("runs the commands", func() {
				var success bool
				var found bool
				var err error
				var vmInfo driver.VMInfo

				ovfPath := filepath.Join("..", "test", "fixtures", "image.ovf")
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
				Expect(vmInfo.CPUs).To(Equal(1))
				Expect(vmInfo.RAM).To(Equal(512))
				Expect(len(vmInfo.NICs)).To(Equal(0))
				rootDiskPath := vmInfo.Disks[1].Path
				Expect(rootDiskPath).To(Equal("vm-virtualmachine-disk1.vmdk"))

				fileInfo, err := os.Stat(filepath.Join(filepath.Dir(config.VmxPath(vmId)), rootDiskPath))
				Expect(err).ToNot(HaveOccurred())
				Expect(fileInfo.Size()).To(Equal(int64(65536)))

				err = client.SetVMNetworkAdapter(vmId, "fake-network", "00:50:56:3F:00:00")
				Expect(err).ToNot(HaveOccurred())

				vmInfo, err = client.GetVMInfo(vmId)
				Expect(err).ToNot(HaveOccurred())
				Expect(vmInfo.NICs[0].Network).To(Equal("fake-network"))
				Expect(vmInfo.NICs[0].MAC).To(Equal("00:50:56:3F:00:00"))

				err = client.SetVMResources(vmId, 2, 1024)
				Expect(err).ToNot(HaveOccurred())

				Expect(client.NeedsVMNameChange(vmId)).To(BeTrue())

				err = client.SetVMDisplayName(vmId, "initial-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(client.NeedsVMNameChange(vmId)).To(BeFalse())

				vmInfo, err = client.GetVMInfo(vmId)
				Expect(err).ToNot(HaveOccurred())
				Expect(vmInfo.CPUs).To(Equal(2))
				Expect(vmInfo.RAM).To(Equal(1024))
				Expect(vmInfo.Name).To(Equal("initial-name"))

				err = client.CreateEphemeralDisk(vmId, 2048)
				Expect(err).ToNot(HaveOccurred())

				vmInfo, err = client.GetVMInfo(vmId)
				Expect(err).ToNot(HaveOccurred())
				Expect(vmInfo.Disks[2].Path).To(HaveSuffix(filepath.Join("ephemeral-disks", "vm-virtualmachine.vmdk")))

				fileInfo, err = os.Stat(vmInfo.Disks[2].Path)
				Expect(err).ToNot(HaveOccurred())
				Expect(fileInfo.Size()).To(Equal(int64(327680)))

				found = client.HasDisk("disk-1")
				Expect(found).To(Equal(false))

				err = client.CreateDisk("disk-1", 3096)
				Expect(err).ToNot(HaveOccurred())

				found = client.HasDisk("disk-1")
				Expect(found).To(Equal(true))

				err = client.AttachDisk(vmId, "disk-1")
				Expect(err).ToNot(HaveOccurred())

				vmInfo, err = client.GetVMInfo(vmId)
				Expect(err).ToNot(HaveOccurred())

				Expect(vmInfo.Disks[3].Path).To(HaveSuffix(filepath.Join("persistent-disks", "disk-1.vmdk")))

				fileInfo, err = os.Stat(vmInfo.Disks[3].Path)
				Expect(err).ToNot(HaveOccurred())
				Expect(fileInfo.Size()).To(Equal(int64(458752)))

				currentIsoPath := client.GetVMIsoPath(vmId)
				Expect(currentIsoPath).To(Equal(""))

				envIsoPath := filepath.Join("..", "test", "fixtures", "env.iso")
				err = client.UpdateVMIso(vmId, envIsoPath)
				Expect(err).ToNot(HaveOccurred())

				currentIsoPath = client.GetVMIsoPath(vmId)
				Expect(currentIsoPath).To(ContainSubstring(filepath.Join("env-isos", "vm-virtualmachine.iso")))

				err = client.StartVM(vmId)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(1 * time.Second)

				err = client.SetVMDisplayName(vmId, "ignored-name-when-vm-running")
				Expect(err).ToNot(HaveOccurred())

				err = client.DetachDisk(vmId, "disk-1")
				Expect(err).ToNot(HaveOccurred())

				err = client.StopVM(vmId)
				Expect(err).ToNot(HaveOccurred())

				vmInfo, err = client.GetVMInfo(vmId)
				Expect(err).ToNot(HaveOccurred())
				Expect(vmInfo.Name).To(Equal("initial-name"))

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

				ovfPath := filepath.Join("..", "test", "fixtures", "image.ovf")
				success, err = client.ImportOvf(ovfPath, vmId)
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(Equal(true))

				envIsoPath := filepath.Join("..", "test", "fixtures", "env.iso")
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

	Describe("with vmrun linked cloning", func() {
		BeforeEach(func() {
			if vmrunRunner.IsPlayer() {
				Skip("can't test linked cloning with player")
			}

			client = driver.NewClient(vmrunRunner, ovftoolRunner, vmrunRunner, vmxBuilder, config, logger)
		})

		It("clones with linked disks", func() {
			var success bool
			var found bool
			var err error
			var vmInfo driver.VMInfo

			ovfPath := filepath.Join("..", "test", "fixtures", "image.ovf")
			success, err = client.ImportOvf(ovfPath, stemcellId)
			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(Equal(true))

			err = client.CloneVM(stemcellId, vmId)
			Expect(err).ToNot(HaveOccurred())

			found = client.HasVM(vmId)
			Expect(found).To(Equal(true))

			vmInfo, err = client.GetVMInfo(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmInfo.Name).To(Equal(vmId))
			Expect(vmInfo.CPUs).To(Equal(1))
			Expect(vmInfo.RAM).To(Equal(512))
			Expect(len(vmInfo.NICs)).To(Equal(0))
			Expect(vmInfo.Disks[1].Path).To(HaveSuffix("cs-stemcell-disk1-cl1.vmdk"))
		})

		Describe("concurrent clone", func() {
			var iterations = 20

			It("can clone in parallel", func() {
				var success bool
				var err error

				ovfPath := filepath.Join("..", "test", "fixtures", "image.ovf")
				success, err = client.ImportOvf(ovfPath, stemcellId)
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(Equal(true))

				errorChannel := make(chan error)
				for i := 1; i <= iterations; i++ {
					go func(j int) {
						parallelVmId := fmt.Sprintf("vm-virtualmachine-%d", j)

						errorChannel <- client.CloneVM(stemcellId, parallelVmId)
					}(i)
				}

				for i := 1; i <= iterations; i++ {
					Expect(<-errorChannel).ToNot(HaveOccurred())
				}
			})

			AfterEach(func() {
				errorChannel := make(chan error)
				for i := 1; i <= iterations; i++ {
					go func(j int) {
						parallelVmId := fmt.Sprintf("vm-virtualmachine-%d", j)

						if client.HasVM(parallelVmId) {
							errorChannel <- client.DestroyVM(parallelVmId)
						}
					}(i)
				}

				for i := 1; i <= iterations; i++ {
					Expect(<-errorChannel).ToNot(HaveOccurred())
				}
			})
		})
	})
})
