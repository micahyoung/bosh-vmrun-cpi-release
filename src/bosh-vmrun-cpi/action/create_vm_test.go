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

		agentSettings.GenerateAgentEnvIsoReturns("iso-path", nil)
		agentSettings.GenerateMacAddressReturnsOnCall(0, "00:11:22:33:44:55", nil)
		agentSettings.GenerateMacAddressReturnsOnCall(1, "55:44:33:22:11:00", nil)

		m := action.NewCreateVMMethod(driverClient, agentSettings, agentOptions, agentEnvFactory, uuidGen, logger)
		cid, err := m.CreateVM(agentId, stemcellCid, resourceCloudProps, networks, disks, vmEnv)

		Expect(err).ToNot(HaveOccurred())
		Expect(cid.AsString()).To(Equal("fake-uuid-0"))
	})
})
