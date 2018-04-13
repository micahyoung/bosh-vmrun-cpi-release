package action_test

import (
	"encoding/json"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"
	fakevm "bosh-govmomi-cpi/vm/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"bosh-govmomi-cpi/action"
)

var _ = Describe("CreateVM", func() {
	It("runs the cpi", func() {
		govcClient := &fakegovc.FakeGovcClient{}
		agentSettings := &fakevm.FakeAgentSettings{}
		uuidGen := &fakeuuid.FakeGenerator{}
		agentOptions := apiv1.AgentOptions{}
		networks := apiv1.Networks{}
		disks := []apiv1.DiskCID{}
		logger := &fakelogger.FakeLogger{}
		agentEnvFactory := apiv1.NewAgentEnvFactory()

		agentId := apiv1.AgentID{}
		stemcellCid := apiv1.NewStemcellCID("stemcell")
		vmEnv := apiv1.VMEnv{}

		var cloudProps apiv1.CloudPropsImpl
		json.Unmarshal([]byte(`{
			"cpu":  1,
			"ram":  1024,
			"disk": 2048
		}`), &cloudProps)

		govcClient.CloneVMReturns("", nil)
		govcClient.SetVMResourcesReturns(nil)
		govcClient.SetVMNetworkAdaptersReturns(nil)
		govcClient.CreateEphemeralDiskReturns(nil)
		govcClient.UpdateVMIsoReturns("", nil)
		govcClient.StartVMReturns("", nil)
		agentSettings.GenerateAgentEnvIsoReturns("iso-path", nil)

		m := action.NewCreateVMMethod(govcClient, agentSettings, agentOptions, agentEnvFactory, uuidGen, logger)
		cid, err := m.CreateVM(agentId, stemcellCid, cloudProps, networks, disks, vmEnv)

		Expect(err).ToNot(HaveOccurred())
		Expect(cid.AsString()).To(Equal("fake-uuid-0"))

		cloneVmStemcellId, cloneVmVmId := govcClient.CloneVMArgsForCall(0)
		Expect(cloneVmStemcellId).To(Equal("cs-stemcell"))
		Expect(cloneVmVmId).To(Equal("vm-fake-uuid-0"))

		setResourcesVmId, setResourcesCpu, setResourcesRam := govcClient.SetVMResourcesArgsForCall(0)
		Expect(setResourcesVmId).To(Equal("vm-fake-uuid-0"))
		Expect(setResourcesCpu).To(Equal(1))
		Expect(setResourcesRam).To(Equal(1024))

		setAdaptersVmId, setAdaptersCount := govcClient.SetVMNetworkAdaptersArgsForCall(0)
		Expect(setAdaptersVmId).To(Equal("vm-fake-uuid-0"))
		Expect(setAdaptersCount).To(Equal(0))

		ephemeralDiskVmId, ephemeralDiskSize := govcClient.CreateEphemeralDiskArgsForCall(0)
		Expect(ephemeralDiskVmId).To(Equal("vm-fake-uuid-0"))
		Expect(ephemeralDiskSize).To(Equal(2048))

		updateIsoVmId, updateIsoPath := govcClient.UpdateVMIsoArgsForCall(0)
		Expect(updateIsoVmId).To(Equal("vm-fake-uuid-0"))
		Expect(updateIsoPath).To(Equal("iso-path"))

		startVmVmId := govcClient.StartVMArgsForCall(0)
		Expect(startVmVmId).To(Equal("vm-fake-uuid-0"))
	})
})
