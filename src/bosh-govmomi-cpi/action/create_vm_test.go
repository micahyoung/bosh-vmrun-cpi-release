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
		disks := []apiv1.DiskCID{}
		logger := &fakelogger.FakeLogger{}
		agentEnvFactory := apiv1.NewAgentEnvFactory()

		agentId := apiv1.AgentID{}
		stemcellCid := apiv1.NewStemcellCID("stemcell")
		vmEnv := apiv1.VMEnv{}

		networks := apiv1.Networks{}
		networks.UnmarshalJSON([]byte(`{
		  "first":{
		    "cloud_properties":{
		      "name":"VM Network",
					"type":"manual"
		    }
		  },
		  "second":{
		    "cloud_properties":{
		      "name":"BOSH Network",
					"type":"manual"
		    }
		  }
		}`))

		var resourceCloudProps apiv1.CloudPropsImpl
		json.Unmarshal([]byte(`{
			"cpu":  1,
			"ram":  1024,
			"disk": 2048
		}`), &resourceCloudProps)

		govcClient.CloneVMReturns("", nil)
		govcClient.SetVMResourcesReturns(nil)
		govcClient.SetVMNetworkAdapterReturns(nil)
		govcClient.CreateEphemeralDiskReturns(nil)
		govcClient.UpdateVMIsoReturns("", nil)
		govcClient.StartVMReturns("", nil)
		agentSettings.GenerateAgentEnvIsoReturns("iso-path", nil)
		agentSettings.GenerateMacAddressReturnsOnCall(0, "00:11:22:33:44:55", nil)
		agentSettings.GenerateMacAddressReturnsOnCall(1, "55:44:33:22:11:00", nil)

		m := action.NewCreateVMMethod(govcClient, agentSettings, agentOptions, agentEnvFactory, uuidGen, logger)
		cid, err := m.CreateVM(agentId, stemcellCid, resourceCloudProps, networks, disks, vmEnv)

		Expect(err).ToNot(HaveOccurred())
		Expect(cid.AsString()).To(Equal("fake-uuid-0"))

		cloneVmStemcellId, cloneVmVmId := govcClient.CloneVMArgsForCall(0)
		Expect(cloneVmStemcellId).To(Equal("cs-stemcell"))
		Expect(cloneVmVmId).To(Equal("vm-fake-uuid-0"))

		setResourcesVmId, setResourcesCpu, setResourcesRam := govcClient.SetVMResourcesArgsForCall(0)
		Expect(setResourcesVmId).To(Equal("vm-fake-uuid-0"))
		Expect(setResourcesCpu).To(Equal(1))
		Expect(setResourcesRam).To(Equal(1024))

		setAdapterVmId1, setAdapterNetName1, setAdapterMac1 := govcClient.SetVMNetworkAdapterArgsForCall(0)
		Expect(setAdapterVmId1).To(Equal("vm-fake-uuid-0"))
		Expect(setAdapterNetName1).To(Equal("VM Network"))
		Expect(setAdapterMac1).To(Equal("00:11:22:33:44:55"))

		setAdapterVmId2, setAdapterNetName2, setAdapterMac2 := govcClient.SetVMNetworkAdapterArgsForCall(1)
		Expect(setAdapterVmId2).To(Equal("vm-fake-uuid-0"))
		Expect(setAdapterNetName2).To(Equal("BOSH Network"))
		Expect(setAdapterMac2).To(Equal("55:44:33:22:11:00"))

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
