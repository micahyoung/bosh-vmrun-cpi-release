package driver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakedriver "bosh-vmrun-cpi/driver/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"

	"bosh-vmrun-cpi/driver"
)

var _ = Describe("DriverClient", func() {
	var vmrunRunner *fakedriver.FakeVmrunRunner
	var ovftoolRunner *fakedriver.FakeOvftoolRunner
	var vdiskmanagerRunner *fakedriver.FakeVdiskmanagerRunner
	var vmxBuilder *fakedriver.FakeVmxBuilder
	var config *fakedriver.FakeConfig
	var logger *fakelogger.FakeLogger
	var client driver.Client

	BeforeEach(func() {
		vmrunRunner = &fakedriver.FakeVmrunRunner{}
		ovftoolRunner = &fakedriver.FakeOvftoolRunner{}
		vdiskmanagerRunner = &fakedriver.FakeVdiskmanagerRunner{}
		vmxBuilder = &fakedriver.FakeVmxBuilder{}
		config = &fakedriver.FakeConfig{}
		logger = &fakelogger.FakeLogger{}

		client = driver.NewClient(vmrunRunner, ovftoolRunner, vdiskmanagerRunner, vmxBuilder, config, logger)
	})

	Describe("ImportOvf", func() {
		It("runs the driver command", func() {
			config.VmPathReturns("vm-path")
			ovfPath := "ovf-path"
			stemcellId := "stemcell-uuid"

			ovftoolRunner.CliCommandReturns("", nil)

			result, err := client.ImportOvf(ovfPath, stemcellId)

			importArgs, importFlags := ovftoolRunner.CliCommandArgsForCall(0)
			Expect(importFlags).To(Equal(map[string]string{
				"sourceType":          "OVF",
				"allowAllExtraConfig": "true",
				"allowExtraConfig":    "true",
				"targetType":          "VMX",
				"name":                "stemcell-uuid",
			}))
			Expect(importArgs).To(Equal([]string{
				"ovf-path",
				"vm-path/stemcell-uuid/stemcell-uuid.vmx",
			}))

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
		})
	})

	//
	//	Describe("StartVM", func() {
	//		It("runs driver commands", func() {
	//			config.EsxUrlReturns("esx-url")
	//			client := driver.NewClient(vmrunRunner, config, logger)
	//			vmId := "vm-uuid"
	//
	//			vmrunRunner.CliCommandReturnsOnCall(0, "start-success", nil)
	//			vmrunRunner.CliCommandReturnsOnCall(1, `{"VirtualMachines":[{"Runtime":{"Question":null}}]}`, nil)
	//			vmrunRunner.CliCommandReturnsOnCall(2, `{"VirtualMachines":[{"Runtime":{"Question":true}}]}`, nil)
	//
	//			result, err := client.StartVM(vmId)
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(result).To(Equal("success"))
	//			Expect(vmrunRunner.CliCommandCallCount()).To(Equal(4))
	//
	//			powerBin, powerFlags, powerArgs := vmrunRunner.CliCommandArgsForCall(0)
	//			Expect(powerBin).To(Equal("vm.power"))
	//			Expect(powerFlags).To(Equal(map[string]string{
	//				"on": "true",
	//				"u":  "esx-url",
	//				"k":  "true",
	//			}))
	//			Expect(powerArgs).To(Equal([]string{"vm-uuid"}))
	//
	//			infoOffStateBin, infoOffStateFlags, infoOffStateArgs := vmrunRunner.CliCommandArgsForCall(1)
	//			Expect(infoOffStateBin).To(Equal("vm.info"))
	//			Expect(infoOffStateFlags).To(Equal(map[string]string{
	//				"u": "esx-url",
	//				"k": "true",
	//			}))
	//			Expect(infoOffStateArgs).To(Equal([]string{"vm-uuid"}))
	//
	//			infoOnStateBin, _, _ := vmrunRunner.CliCommandArgsForCall(2)
	//			Expect(infoOnStateBin).To(Equal("vm.info"))
	//
	//			questionBin, questionFlags, questionArgs := vmrunRunner.CliCommandArgsForCall(3)
	//			Expect(questionBin).To(Equal("vm.question"))
	//			Expect(questionFlags).To(Equal(map[string]string{
	//				"answer": "2",
	//				"vm":     "vm-uuid",
	//				"u":      "esx-url",
	//				"k":      "true",
	//			}))
	//			Expect(questionArgs).To(BeNil())
	//		})
	//	})

	//	Describe("SetVMNetworkAdapters", func() {
	//		It("runs driver commands", func() {
	//			config.EsxUrlReturns("esx-url")
	//			client := driver.NewClient(runner, config, logger)
	//			vmId := "vm-uuid"
	//
	//			runner.CliCommandReturnsOnCall(0, "network-success", nil)
	//
	//			err := client.SetVMNetworkAdapter(vmId, "VM Network", "00:11:22:33:44:55")
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(runner.CliCommandCallCount()).To(Equal(1))
	//
	//			networkBin, networkFlags, networkArgs := runner.CliCommandArgsForCall(0)
	//			Expect(networkBin).To(Equal("vm.network.add"))
	//			Expect(networkFlags).To(Equal(map[string]string{
	//				"vm":          "vm-uuid",
	//				"net":         "VM Network",
	//				"net.adapter": "vmxnet3",
	//				"net.address": "00:11:22:33:44:55",
	//				"u":           "esx-url",
	//				"k":           "true",
	//			}))
	//			Expect(networkArgs).To(BeNil())
	//		})
	//	})
	//
	//	Describe("SetVMResources", func() {
	//		It("runs driver commands", func() {
	//			config.EsxUrlReturns("esx-url")
	//			client := driver.NewClient(runner, config, logger)
	//			vmId := "vm-uuid"
	//
	//			runner.CliCommandReturnsOnCall(0, "change-success", nil)
	//
	//			err := client.SetVMResources(vmId, 2, 1024)
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(runner.CliCommandCallCount()).To(Equal(1))
	//
	//			changeBin, changeFlags, changeArgs := runner.CliCommandArgsForCall(0)
	//			Expect(changeBin).To(Equal("vm.change"))
	//			Expect(changeFlags).To(Equal(map[string]string{
	//				"vm": "vm-uuid",
	//				"c":  "2",
	//				"m":  "1024",
	//				"u":  "esx-url",
	//				"k":  "true",
	//			}))
	//			Expect(changeArgs).To(BeNil())
	//		})
	//	})
	//
	//	Describe("HasVM", func() {
	//		It("runs driver commands", func() {
	//			config.EsxUrlReturns("esx-url")
	//			client := driver.NewClient(runner, config, logger)
	//			vmId := "vm-uuid"
	//
	//			runner.CliCommandReturnsOnCall(0, `{"VirtualMachines":[{"Runtime":{"PowerState":null}}]}`, nil)
	//
	//			result, err := client.HasVM(vmId)
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(result).To(Equal(true))
	//			Expect(runner.CliCommandCallCount()).To(Equal(1))
	//
	//			infoBin, infoFlags, infoArgs := runner.CliCommandArgsForCall(0)
	//			Expect(infoBin).To(Equal("vm.info"))
	//			Expect(infoFlags).To(Equal(map[string]string{
	//				"u": "esx-url",
	//				"k": "true",
	//			}))
	//			Expect(infoArgs).To(Equal([]string{"vm-uuid"}))
	//		})
	//	})
	//
	//	Describe("CreateDisk", func() {
	//		It("runs driver commands", func() {
	//			config.EsxUrlReturns("esx-url")
	//			client := driver.NewClient(runner, config, logger)
	//			diskId := "disk-uuid"
	//			diskKB := 10240
	//
	//			runner.CliCommandReturnsOnCall(0, "success", nil)
	//
	//			err := client.CreateDisk(diskId, diskKB)
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(runner.CliCommandCallCount()).To(Equal(1))
	//
	//			diskCreateBin, diskCreateFlags, diskCreateArgs := runner.CliCommandArgsForCall(0)
	//			Expect(diskCreateBin).To(Equal("datastore.disk.create"))
	//			Expect(diskCreateFlags).To(Equal(map[string]string{
	//				"size": "10240MB",
	//				"u":    "esx-url",
	//				"k":    "true",
	//			}))
	//			Expect(diskCreateArgs).To(Equal([]string{diskId + ".vmdk"}))
	//		})
	//	})
	//
	//	Describe("DetachDisk", func() {
	//		It("runs driver commands", func() {
	//			config.EsxUrlReturns("esx-url")
	//			client := driver.NewClient(runner, config, logger)
	//			vmId := "vm-uuid"
	//			diskId := "disk-uuid"
	//
	//			runner.CliCommandReturnsOnCall(0, `{"devices":[{"name":"disk-1000-2","backing":{"parent":{"file_name":"[datastore] disk-uuid"}}}]}`, nil)
	//			runner.CliCommandReturnsOnCall(1, "success", nil)
	//
	//			err := client.DetachDisk(vmId, diskId)
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(runner.CliCommandCallCount()).To(Equal(2))
	//
	//			deviceInfoBin, deviceInfoFlags, _ := runner.CliCommandArgsForCall(0)
	//			Expect(deviceInfoBin).To(Equal("device.info"))
	//			Expect(deviceInfoFlags).To(Equal(map[string]string{
	//				"json": "true",
	//				"vm":   vmId,
	//				"u":    "esx-url",
	//				"k":    "true",
	//			}))
	//
	//			deviceRemoveBin, deviceRemoveFlags, _ := runner.CliCommandArgsForCall(1)
	//			Expect(deviceRemoveBin).To(Equal("device.remove"))
	//			Expect(deviceRemoveFlags).To(Equal(map[string]string{
	//				"keep": "true",
	//				"vm":   vmId,
	//				"u":    "esx-url",
	//				"k":    "true",
	//			}))
	//		})
	//	})
	//
	//	Describe("DestroyVM", func() {
	//		It("runs driver commands", func() {
	//			config.EsxUrlReturns("esx-url")
	//			client := driver.NewClient(runner, config, logger)
	//			vmId := "vm-uuid"
	//
	//			runner.CliCommandReturnsOnCall(0, `{"VirtualMachines":[{"Runtime":{"PowerState":"poweredOn"}}]}`, nil)
	//			runner.CliCommandReturnsOnCall(1, "stop-vm-success", nil)
	//			runner.CliCommandReturnsOnCall(2, "destroy-vm-success", nil)
	//			runner.CliCommandReturnsOnCall(3, fmt.Sprintf(`[{"Datastore":{"Type":"Datastore","Value":"5a83963c-9fd8a83a-c3b7-000c297e0932"},"FolderPath":"[datastore1]","File":[{"Path":"never-match","FriendlyName":"","FileSize":0,"Modification":null,"Owner":""},{"Path":"%s","FriendlyName":"","FileSize":0,"Modification":null,"Owner":""}]}]`, vmId), nil)
	//			runner.CliCommandReturnsOnCall(4, "delete-datastore-success", nil)
	//
	//			result, err := client.DestroyVM(vmId)
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(result).To(Equal("delete-datastore-success"))
	//			Expect(runner.CliCommandCallCount()).To(Equal(5))
	//
	//			infoBin, infoFlags, infoArgs := runner.CliCommandArgsForCall(0)
	//			Expect(infoBin).To(Equal("vm.info"))
	//			Expect(infoFlags).To(Equal(map[string]string{
	//				"u": "esx-url",
	//				"k": "true",
	//			}))
	//			Expect(infoArgs).To(Equal([]string{"vm-uuid"}))
	//
	//			powerBin, powerFlags, powerArgs := runner.CliCommandArgsForCall(1)
	//			Expect(powerBin).To(Equal("vm.power"))
	//			Expect(powerFlags).To(Equal(map[string]string{
	//				"off": "true",
	//				"u":   "esx-url",
	//				"k":   "true",
	//			}))
	//			Expect(powerArgs).To(Equal([]string{"vm-uuid"}))
	//
	//			destroyBin, destroyFlags, destroyArgs := runner.CliCommandArgsForCall(2)
	//			Expect(destroyBin).To(Equal("vm.destroy"))
	//			Expect(destroyFlags).To(Equal(map[string]string{
	//				"u": "esx-url",
	//				"k": "true",
	//			}))
	//			Expect(destroyArgs).To(Equal([]string{"vm-uuid"}))
	//
	//			listBin, listFlags, listArgs := runner.CliCommandArgsForCall(3)
	//			Expect(listBin).To(Equal("datastore.ls"))
	//			Expect(listFlags).To(Equal(map[string]string{
	//				"u": "esx-url",
	//				"k": "true",
	//			}))
	//			Expect(listArgs).To(BeNil())
	//
	//			deleteBin, deleteFlags, deleteArgs := runner.CliCommandArgsForCall(4)
	//			Expect(deleteBin).To(Equal("datastore.rm"))
	//			Expect(deleteFlags).To(Equal(map[string]string{
	//				"f": "true",
	//				"u": "esx-url",
	//				"k": "true",
	//			}))
	//			Expect(deleteArgs).To(Equal([]string{"vm-uuid"}))
	//		})
	//	})
})
