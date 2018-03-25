package govc_test

import (
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
			client := govc.NewClient(runner, config, logger)
			ovfPath := "ovf-path"
			stemcellId := "stemcell-uuid"

			config.EsxUrlReturns("esx-url")
			runner.CliCommandReturns("success", nil)

			result, err := client.ImportOvf(ovfPath, stemcellId)

			runnerCliCommandBin, runnerCliCommandFlags, runnerCliCommandArgs := runner.CliCommandArgsForCall(0)
			Expect(runnerCliCommandBin).To(Equal("import.ovf"))
			Expect(runnerCliCommandFlags).To(Equal(map[string]string{
				"name": "stemcell-uuid",
				"u":    "esx-url",
				"k":    "true",
			}))
			Expect(runnerCliCommandArgs).To(Equal([]string{"ovf-path"}))

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("success"))
		})
	})

	Describe("StartVM", func() {
		It("runs govc commands", func() {
			client := govc.NewClient(runner, config, logger)
			vmId := "vm-uuid"

			config.EsxUrlReturns("esx-url")
			runner.CliCommandReturnsOnCall(0, "start-success", nil)
			runner.CliCommandReturnsOnCall(1, "question-success", nil)

			result, err := client.StartVM(vmId)
			_ = result
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("start-success"))
			Expect(runner.CliCommandCallCount()).To(Equal(2))

			runnerPowerCliCommandBin,
				runnerPowerCliCommandFlags,
				runnerPowerCliCommandArgs :=
				runner.CliCommandArgsForCall(0)
			Expect(runnerPowerCliCommandBin).To(Equal("vm.power"))
			Expect(runnerPowerCliCommandFlags).To(Equal(map[string]string{
				"on": "true",
				"u":  "esx-url",
				"k":  "true",
			}))
			Expect(runnerPowerCliCommandArgs).To(Equal([]string{"vm-uuid"}))

			runnerQuestionCliCommandBin,
				runnerQuestionCliCommandFlags,
				runnerQuestionCliCommandArgs :=
				runner.CliCommandArgsForCall(1)
			Expect(runnerQuestionCliCommandBin).To(Equal("vm.question"))
			Expect(runnerQuestionCliCommandFlags).To(Equal(map[string]string{
				"answer": "2",
				"vm":     "vm-uuid",
				"u":      "esx-url",
				"k":      "true",
			}))
			Expect(runnerQuestionCliCommandArgs).To(BeNil())

		})
	})

	Describe("CloneVM", func() {
		It("runs govc commands", func() {
			client := govc.NewClient(runner, config, logger)
			stemcellId := "stemcell-uuid"
			vmId := "vm-uuid"

			config.EsxUrlReturns("esx-url")
			runner.CliCommandReturnsOnCall(0, "copy-success", nil)
			runner.CliCommandReturnsOnCall(1, "register-success", nil)
			runner.CliCommandReturnsOnCall(2, "change-success", nil)
			runner.CliCommandReturnsOnCall(3, "network-success", nil)

			result, err := client.CloneVM(stemcellId, vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("network-success"))
			Expect(runner.CliCommandCallCount()).To(Equal(4))

			runnerCopyCliCommandBin,
				runnerCopyCliCommandFlags,
				runnerCopyCliCommandArgs :=
				runner.CliCommandArgsForCall(0)
			Expect(runnerCopyCliCommandBin).To(Equal("datastore.cp"))
			Expect(runnerCopyCliCommandFlags).To(Equal(map[string]string{
				"u": "esx-url",
				"k": "true",
			}))
			Expect(runnerCopyCliCommandArgs).To(Equal([]string{"stemcell-uuid", "vm-uuid"}))

			runnerRegisterCliCommandBin,
				runnerRegisterCliCommandFlags,
				runnerRegisterCliCommandArgs :=
				runner.CliCommandArgsForCall(1)
			Expect(runnerRegisterCliCommandBin).To(Equal("vm.register"))
			Expect(runnerRegisterCliCommandFlags).To(Equal(map[string]string{
				"name": "vm-uuid",
				"u":    "esx-url",
				"k":    "true",
			}))
			Expect(runnerRegisterCliCommandArgs).To(Equal([]string{"vm-uuid/stemcell-uuid.vmx"}))

			runnerChangeCliCommandBin,
				runnerChangeCliCommandFlags,
				runnerChangeCliCommandArgs :=
				runner.CliCommandArgsForCall(2)
			Expect(runnerChangeCliCommandBin).To(Equal("vm.change"))
			Expect(runnerChangeCliCommandFlags).To(Equal(map[string]string{
				"vm":                  "vm-uuid",
				"nested-hv-enabled":   "true",
				"sync-time-with-host": "true",
				"u": "esx-url",
				"k": "true",
			}))
			Expect(runnerChangeCliCommandArgs).To(BeNil())

			runnerNetworkCliCommandBin,
				runnerNetworkCliCommandFlags,
				runnerNetworkCliCommandArgs :=
				runner.CliCommandArgsForCall(3)
			Expect(runnerNetworkCliCommandBin).To(Equal("vm.network.add"))
			Expect(runnerNetworkCliCommandFlags).To(Equal(map[string]string{
				"vm":          "vm-uuid",
				"net":         "VM Network",
				"net.adapter": "vmxnet3",
				"u":           "esx-url",
				"k":           "true",
			}))
			Expect(runnerNetworkCliCommandArgs).To(BeNil())
		})
	})

	Describe("DestroyVM", func() {
		It("runs govc commands", func() {
			client := govc.NewClient(runner, config, logger)
			vmId := "vm-uuid"

			config.EsxUrlReturns("esx-url")
			runner.CliCommandReturnsOnCall(0, "vm-exists", nil)
			runner.CliCommandReturnsOnCall(1, "stop-vm-success", nil)
			runner.CliCommandReturnsOnCall(2, "destroy-vm-success", nil)
			runner.CliCommandReturnsOnCall(3, "datastore-exists", nil)
			runner.CliCommandReturnsOnCall(4, "delete-datastore-success", nil)

			result, err := client.DestroyVM(vmId)
			_ = result
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("start-success"))
			Expect(runner.CliCommandCallCount()).To(Equal(2))

			runnerPowerCliCommandBin,
				runnerPowerCliCommandFlags,
				runnerPowerCliCommandArgs :=
				runner.CliCommandArgsForCall(0)
			Expect(runnerPowerCliCommandBin).To(Equal("vm.power"))
			Expect(runnerPowerCliCommandFlags).To(Equal(map[string]string{
				"on": "true",
				"u":  "esx-url",
				"k":  "true",
			}))
			Expect(runnerPowerCliCommandArgs).To(Equal([]string{"vm-uuid"}))

			runnerQuestionCliCommandBin,
				runnerQuestionCliCommandFlags,
				runnerQuestionCliCommandArgs :=
				runner.CliCommandArgsForCall(1)
			Expect(runnerQuestionCliCommandBin).To(Equal("vm.question"))
			Expect(runnerQuestionCliCommandFlags).To(Equal(map[string]string{
				"answer": "2",
				"vm":     "vm-uuid",
				"u":      "esx-url",
				"k":      "true",
			}))
			Expect(runnerQuestionCliCommandArgs).To(BeNil())

		})
	})
})
