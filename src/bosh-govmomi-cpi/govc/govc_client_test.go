package govc_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"

	"bosh-govmomi-cpi/govc"
)

var _ = Describe("GovcClient", func() {
	var runner *fakegovc.FakeGovcRunner
	var config *fakegovc.FakeGovcConfig
	var logger *fakelogger.FakeLogger

	BeforeEach(func() {
		runner = &fakegovc.FakeGovcRunner{}
		config = &fakegovc.FakeGovcConfig{}
		logger = &fakelogger.FakeLogger{}
	})

	Describe("ImportOvf", func() {
		It("runs the govc command", func() {
			config.EsxUrlReturns("esx-url")
			client := govc.NewClient(runner, config, logger)
			ovfPath := "ovf-path"
			stemcellId := "stemcell-uuid"

			runner.CliCommandReturns("success", nil)

			result, err := client.ImportOvf(ovfPath, stemcellId)

			importBin, importFlags, importArgs := runner.CliCommandArgsForCall(0)
			Expect(importBin).To(Equal("import.ovf"))
			Expect(importFlags).To(Equal(map[string]string{
				"name": "stemcell-uuid",
				"u":    "esx-url",
				"k":    "true",
			}))
			Expect(importArgs).To(Equal([]string{"ovf-path"}))

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("success"))
		})
	})

	Describe("StartVM", func() {
		It("runs govc commands", func() {
			config.EsxUrlReturns("esx-url")
			client := govc.NewClient(runner, config, logger)
			vmId := "vm-uuid"

			runner.CliCommandReturnsOnCall(0, "start-success", nil)
			runner.CliCommandReturnsOnCall(1, `{"VirtualMachines":[{"Runtime":{"Question":null}}]}`, nil)
			runner.CliCommandReturnsOnCall(2, `{"VirtualMachines":[{"Runtime":{"Question":true}}]}`, nil)

			result, err := client.StartVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("success"))
			Expect(runner.CliCommandCallCount()).To(Equal(4))

			powerBin, powerFlags, powerArgs := runner.CliCommandArgsForCall(0)
			Expect(powerBin).To(Equal("vm.power"))
			Expect(powerFlags).To(Equal(map[string]string{
				"on": "true",
				"u":  "esx-url",
				"k":  "true",
			}))
			Expect(powerArgs).To(Equal([]string{"vm-uuid"}))

			infoOffStateBin, infoOffStateFlags, infoOffStateArgs := runner.CliCommandArgsForCall(1)
			Expect(infoOffStateBin).To(Equal("vm.info"))
			Expect(infoOffStateFlags).To(Equal(map[string]string{
				"u": "esx-url",
				"k": "true",
			}))
			Expect(infoOffStateArgs).To(Equal([]string{"vm-uuid"}))

			infoOnStateBin, _, _ := runner.CliCommandArgsForCall(2)
			Expect(infoOnStateBin).To(Equal("vm.info"))

			questionBin, questionFlags, questionArgs := runner.CliCommandArgsForCall(3)
			Expect(questionBin).To(Equal("vm.question"))
			Expect(questionFlags).To(Equal(map[string]string{
				"answer": "2",
				"vm":     "vm-uuid",
				"u":      "esx-url",
				"k":      "true",
			}))
			Expect(questionArgs).To(BeNil())
		})
	})

	Describe("CloneVM", func() {
		It("runs govc commands", func() {
			config.EsxUrlReturns("esx-url")
			client := govc.NewClient(runner, config, logger)
			stemcellId := "stemcell-uuid"
			vmId := "vm-uuid"

			runner.CliCommandReturnsOnCall(0, "copy-success", nil)
			runner.CliCommandReturnsOnCall(1, "register-success", nil)
			runner.CliCommandReturnsOnCall(2, "change-success", nil)
			runner.CliCommandReturnsOnCall(3, "network-success", nil)

			result, err := client.CloneVM(stemcellId, vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("network-success"))
			Expect(runner.CliCommandCallCount()).To(Equal(4))

			copyBin, copyFlags, copyArgs := runner.CliCommandArgsForCall(0)
			Expect(copyBin).To(Equal("datastore.cp"))
			Expect(copyFlags).To(Equal(map[string]string{
				"u": "esx-url",
				"k": "true",
			}))
			Expect(copyArgs).To(Equal([]string{"stemcell-uuid", "vm-uuid"}))

			registerBin, registerFlags, registerArgs := runner.CliCommandArgsForCall(1)
			Expect(registerBin).To(Equal("vm.register"))
			Expect(registerFlags).To(Equal(map[string]string{
				"name": "vm-uuid",
				"u":    "esx-url",
				"k":    "true",
			}))
			Expect(registerArgs).To(Equal([]string{"vm-uuid/stemcell-uuid.vmx"}))

			changeBin, changeFlags, changeArgs := runner.CliCommandArgsForCall(2)
			Expect(changeBin).To(Equal("vm.change"))
			Expect(changeFlags).To(Equal(map[string]string{
				"vm":                  "vm-uuid",
				"nested-hv-enabled":   "true",
				"sync-time-with-host": "true",
				"u": "esx-url",
				"k": "true",
			}))
			Expect(changeArgs).To(BeNil())

			networkBin, networkFlags, networkArgs := runner.CliCommandArgsForCall(3)
			Expect(networkBin).To(Equal("vm.network.add"))
			Expect(networkFlags).To(Equal(map[string]string{
				"vm":          "vm-uuid",
				"net":         "VM Network",
				"net.adapter": "vmxnet3",
				"u":           "esx-url",
				"k":           "true",
			}))
			Expect(networkArgs).To(BeNil())
		})
	})

	Describe("HasVM", func() {
		It("runs govc commands", func() {
			config.EsxUrlReturns("esx-url")
			client := govc.NewClient(runner, config, logger)
			vmId := "vm-uuid"

			runner.CliCommandReturnsOnCall(0, `{"VirtualMachines":[{"Runtime":{"PowerState":null}}]}`, nil)

			result, err := client.HasVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
			Expect(runner.CliCommandCallCount()).To(Equal(1))

			infoBin, infoFlags, infoArgs := runner.CliCommandArgsForCall(0)
			Expect(infoBin).To(Equal("vm.info"))
			Expect(infoFlags).To(Equal(map[string]string{
				"u": "esx-url",
				"k": "true",
			}))
			Expect(infoArgs).To(Equal([]string{"vm-uuid"}))
		})
	})

	Describe("CreateDisk", func() {
		It("runs govc commands", func() {
			config.EsxUrlReturns("esx-url")
			client := govc.NewClient(runner, config, logger)
			diskId := "disk-uuid"
			diskKB := 10240

			runner.CliCommandReturnsOnCall(0, "success", nil)

			err := client.CreateDisk(diskId, diskKB)
			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CliCommandCallCount()).To(Equal(1))

			diskCreateBin, diskCreateFlags, diskCreateArgs := runner.CliCommandArgsForCall(0)
			Expect(diskCreateBin).To(Equal("datastore.disk.create"))
			Expect(diskCreateFlags).To(Equal(map[string]string{
				"size": "10240MB",
				"u":    "esx-url",
				"k":    "true",
			}))
			Expect(diskCreateArgs).To(Equal([]string{diskId + ".vmdk"}))
		})
	})

	Describe("DetachDisk", func() {
		It("runs govc commands", func() {
			config.EsxUrlReturns("esx-url")
			client := govc.NewClient(runner, config, logger)
			vmId := "vm-uuid"
			diskId := "disk-uuid"

			runner.CliCommandReturnsOnCall(0, `{"devices":[{"name":"disk-1000-2","backing":{"parent":{"file_name":"[datastore] disk-uuid"}}}]}`, nil)
			runner.CliCommandReturnsOnCall(1, "success", nil)

			err := client.DetachDisk(vmId, diskId)
			Expect(err).ToNot(HaveOccurred())
			Expect(runner.CliCommandCallCount()).To(Equal(2))

			deviceInfoBin, deviceInfoFlags, _ := runner.CliCommandArgsForCall(0)
			Expect(deviceInfoBin).To(Equal("device.info"))
			Expect(deviceInfoFlags).To(Equal(map[string]string{
				"json": "true",
				"vm":   vmId,
				"u":    "esx-url",
				"k":    "true",
			}))

			deviceRemoveBin, deviceRemoveFlags, _ := runner.CliCommandArgsForCall(1)
			Expect(deviceRemoveBin).To(Equal("device.remove"))
			Expect(deviceRemoveFlags).To(Equal(map[string]string{
				"keep": "true",
				"vm":   vmId,
				"u":    "esx-url",
				"k":    "true",
			}))
		})
	})

	Describe("DestroyVM", func() {
		It("runs govc commands", func() {
			config.EsxUrlReturns("esx-url")
			client := govc.NewClient(runner, config, logger)
			vmId := "vm-uuid"

			runner.CliCommandReturnsOnCall(0, `{"VirtualMachines":[{"Runtime":{"PowerState":"poweredOn"}}]}`, nil)
			runner.CliCommandReturnsOnCall(1, "stop-vm-success", nil)
			runner.CliCommandReturnsOnCall(2, "destroy-vm-success", nil)
			runner.CliCommandReturnsOnCall(3, fmt.Sprintf(`[{"Datastore":{"Type":"Datastore","Value":"5a83963c-9fd8a83a-c3b7-000c297e0932"},"FolderPath":"[datastore1]","File":[{"Path":"never-match","FriendlyName":"","FileSize":0,"Modification":null,"Owner":""},{"Path":"%s","FriendlyName":"","FileSize":0,"Modification":null,"Owner":""}]}]`, vmId), nil)
			runner.CliCommandReturnsOnCall(4, "delete-datastore-success", nil)

			result, err := client.DestroyVM(vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("delete-datastore-success"))
			Expect(runner.CliCommandCallCount()).To(Equal(5))

			infoBin, infoFlags, infoArgs := runner.CliCommandArgsForCall(0)
			Expect(infoBin).To(Equal("vm.info"))
			Expect(infoFlags).To(Equal(map[string]string{
				"u": "esx-url",
				"k": "true",
			}))
			Expect(infoArgs).To(Equal([]string{"vm-uuid"}))

			powerBin, powerFlags, powerArgs := runner.CliCommandArgsForCall(1)
			Expect(powerBin).To(Equal("vm.power"))
			Expect(powerFlags).To(Equal(map[string]string{
				"off": "true",
				"u":   "esx-url",
				"k":   "true",
			}))
			Expect(powerArgs).To(Equal([]string{"vm-uuid"}))

			destroyBin, destroyFlags, destroyArgs := runner.CliCommandArgsForCall(2)
			Expect(destroyBin).To(Equal("vm.destroy"))
			Expect(destroyFlags).To(Equal(map[string]string{
				"u": "esx-url",
				"k": "true",
			}))
			Expect(destroyArgs).To(Equal([]string{"vm-uuid"}))

			listBin, listFlags, listArgs := runner.CliCommandArgsForCall(3)
			Expect(listBin).To(Equal("datastore.ls"))
			Expect(listFlags).To(Equal(map[string]string{
				"u": "esx-url",
				"k": "true",
			}))
			Expect(listArgs).To(BeNil())

			deleteBin, deleteFlags, deleteArgs := runner.CliCommandArgsForCall(4)
			Expect(deleteBin).To(Equal("datastore.rm"))
			Expect(deleteFlags).To(Equal(map[string]string{
				"f": "true",
				"u": "esx-url",
				"k": "true",
			}))
			Expect(deleteArgs).To(Equal([]string{"vm-uuid"}))
		})
	})
})
