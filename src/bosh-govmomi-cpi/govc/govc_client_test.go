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
				"name": "cs-stemcell-uuid",
				"u":    "esx-url",
				"k":    "true",
			}))
			Expect(runnerCliCommandArgs).To(Equal([]string{"ovf-path"}))

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("success"))
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
			runner.CliCommandReturnsOnCall(4, "start-success", nil)
			runner.CliCommandReturnsOnCall(5, "question-success", nil)

			result, err := client.CloneVM(stemcellId, vmId)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("start-success"))
			Expect(runner.CliCommandCallCount()).To(Equal(6))

			runnerCopyCliCommandBin,
				runnerCopyCliCommandFlags,
				runnerCopyCliCommandArgs :=
				runner.CliCommandArgsForCall(0)
			Expect(runnerCopyCliCommandBin).To(Equal("datastore.cp"))
			Expect(runnerCopyCliCommandFlags).To(Equal(map[string]string{
				"u": "esx-url",
				"k": "true",
			}))
			Expect(runnerCopyCliCommandArgs).To(Equal([]string{"cs-stemcell-uuid", "vm-vm-uuid"}))

			runnerRegisterCliCommandBin,
				runnerRegisterCliCommandFlags,
				runnerRegisterCliCommandArgs :=
				runner.CliCommandArgsForCall(1)
			Expect(runnerRegisterCliCommandBin).To(Equal("vm.register"))
			Expect(runnerRegisterCliCommandFlags).To(Equal(map[string]string{
				"name": "vm-vm-uuid",
				"u":    "esx-url",
				"k":    "true",
			}))
			Expect(runnerRegisterCliCommandArgs).To(Equal([]string{"vm-vm-uuid/cs-stemcell-uuid.vmx"}))

			runnerChangeCliCommandBin,
				runnerChangeCliCommandFlags,
				runnerChangeCliCommandArgs :=
				runner.CliCommandArgsForCall(2)
			Expect(runnerChangeCliCommandBin).To(Equal("vm.change"))
			Expect(runnerChangeCliCommandFlags).To(Equal(map[string]string{
				"vm":                  "vm-vm-uuid",
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
				"vm":          "vm-vm-uuid",
				"net":         "VM Network",
				"net.adapter": "vmxnet3",
				"u":           "esx-url",
				"k":           "true",
			}))
			Expect(runnerNetworkCliCommandArgs).To(BeNil())

			runnerPowerCliCommandBin,
				runnerPowerCliCommandFlags,
				runnerPowerCliCommandArgs :=
				runner.CliCommandArgsForCall(4)
			Expect(runnerPowerCliCommandBin).To(Equal("vm.power"))
			Expect(runnerPowerCliCommandFlags).To(Equal(map[string]string{
				"on": "true",
				"u":  "esx-url",
				"k":  "true",
			}))
			Expect(runnerPowerCliCommandArgs).To(Equal([]string{"vm-vm-uuid"}))

			runnerQuestionCliCommandBin,
				runnerQuestionCliCommandFlags,
				runnerQuestionCliCommandArgs :=
				runner.CliCommandArgsForCall(5)
			Expect(runnerQuestionCliCommandBin).To(Equal("vm.question"))
			Expect(runnerQuestionCliCommandFlags).To(Equal(map[string]string{
				"answer": "2",
				"vm":     "vm-vm-uuid",
				"u":      "esx-url",
				"k":      "true",
			}))
			Expect(runnerQuestionCliCommandArgs).To(BeNil())
		})
	})
})
