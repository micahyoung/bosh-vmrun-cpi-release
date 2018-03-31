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
		stemcellCid := apiv1.StemcellCID{}
		vmEnv := apiv1.VMEnv{}

		var cloudProps apiv1.CloudPropsImpl
		json.Unmarshal([]byte(`{
			"cpu":  1,
			"ram":  1024,
			"disk": 2048
		}`), &cloudProps)

		m := action.NewCreateVMMethod(govcClient, agentSettings, agentOptions, agentEnvFactory, uuidGen, logger)
		cid, err := m.CreateVM(agentId, stemcellCid, cloudProps, networks, disks, vmEnv)

		Expect(err).ToNot(HaveOccurred())
		Expect(cid.AsString()).To(Equal("fake-uuid-0"))
	})
})
