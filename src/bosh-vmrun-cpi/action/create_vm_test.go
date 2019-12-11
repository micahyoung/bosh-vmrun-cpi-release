package action_test

import (
	"encoding/json"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakedriver "bosh-vmrun-cpi/driver/fakes"
	fakevm "bosh-vmrun-cpi/vm/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"bosh-vmrun-cpi/action"
)

var _ = Describe("CreateVM", func() {
	It("runs the cpi", func() {
		driverClient := &fakedriver.FakeClient{}
		agentSettings := &fakevm.FakeAgentSettings{}
		uuidGen := &fakeuuid.FakeGenerator{}
		agentOptions := apiv1.AgentOptions{}
		disks := []apiv1.DiskCID{}
		logger := &fakelogger.FakeLogger{}
		agentEnvFactory := apiv1.NewAgentEnvFactory()

		agentId := apiv1.NewAgentID("agent-0")
		stemcellCid := apiv1.NewStemcellCID("stemcell")
		vmEnv := apiv1.NewVMEnv(map[string]interface{}{"A": "B"})

		var resourceCloudProps apiv1.CloudPropsImpl
		json.Unmarshal([]byte(`{
			"cpu":  1,
			"ram":  1024,
			"disk": 2048,
			"bootstrap": {
				"script_path": "script path",
				"script_content": "script content",
				"interpreter_path": "interpreter path",
				"ready_process_name": "ready process name",
				"username": "me",
				"password": "secret"
			}
		}`), &resourceCloudProps)

		networks := apiv1.Networks{}
		networks.UnmarshalJSON([]byte(`{
		  "first":{
		    "cloud_properties":{
		      "name":"VM Network",
			  "type":"manual"
		    }
		  }
		}`))

		driverClient.HasVMReturns(true)
		driverClient.NeedsVMNameChangeReturns(false)

		agentSettings.GenerateAgentEnvIsoReturns("iso-path", nil)
		agentSettings.GetNetworkSettingsReturnsOnCall(0, "VM Network", "00:11:22:33:44:55", nil)
		agentSettings.GetNetworkSettingsReturnsOnCall(1, "BOSH Network", "55:44:33:22:11:00", nil)

		m := action.NewCreateVMMethod(driverClient, agentSettings, agentOptions, agentEnvFactory, uuidGen, logger)
		cid, err := m.CreateVM(agentId, stemcellCid, resourceCloudProps, networks, disks, vmEnv)

		Expect(err).ToNot(HaveOccurred())
		Expect(cid.AsString()).To(Equal("fake-uuid-0"))

		driverStemcellId := driverClient.HasVMArgsForCall(0)
		Expect(driverStemcellId).To(Equal("cs-stemcell"))

		driverStemcellId, driverVMID := driverClient.CloneVMArgsForCall(0)
		Expect(driverStemcellId).To(Equal("cs-stemcell"))
		Expect(driverVMID).To(Equal("vm-fake-uuid-0"))

		driverVMID, vmPropsCPU, vmPropsRAM := driverClient.SetVMResourcesArgsForCall(0)
		Expect(driverVMID).To(Equal("vm-fake-uuid-0"))
		Expect(vmPropsCPU).To(Equal(1))
		Expect(vmPropsRAM).To(Equal(1024))

		driverVMID, adapterNetworkName, macAddress := driverClient.SetVMNetworkAdapterArgsForCall(0)
		Expect(driverVMID).To(Equal("vm-fake-uuid-0"))
		Expect(adapterNetworkName).To(Equal("VM Network"))
		Expect(macAddress).To(Equal("00:11:22:33:44:55"))

		driverVMID, a, _, _, _, _, _, _, _ := driverClient.BootstrapVMArgsForCall(0)
		Expect(driverVMID).To(Equal("vm-fake-uuid-0"))
		Expect(a).To(Equal("script content"))

		driverVMID, vmPropsDisk := driverClient.CreateEphemeralDiskArgsForCall(0)
		Expect(driverVMID).To(Equal("vm-fake-uuid-0"))
		Expect(vmPropsDisk).To(Equal(2048))

		actualAgentEnv := agentSettings.GenerateAgentEnvIsoArgsForCall(0)
		expectedAgentEnv, _ := agentEnvFactory.FromBytes([]byte(`
		{	
		    "agent_id":"agent-0",
		    "vm":{"name":"fake-uuid-0","id":"fake-uuid-0"},
			"networks":{
				"first": {"mac":"00:11:22:33:44:55"}
			},
			"disks":{"system":"0","ephemeral":"1"},
			"env":{"A":"B"}
		}
		`))
		Expect(actualAgentEnv).To(Equal(expectedAgentEnv))

		driverVMID, isoPath := driverClient.UpdateVMIsoArgsForCall(0)
		Expect(driverVMID).To(Equal("vm-fake-uuid-0"))
		Expect(isoPath).To(Equal("iso-path"))

		driverVMID = driverClient.StartVMArgsForCall(0)
		Expect(driverVMID).To(Equal("vm-fake-uuid-0"))
	})
})
