package action_test

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"
	fakevm "bosh-govmomi-cpi/vm/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"bosh-govmomi-cpi/action"
)

var _ = Describe("CreateVM", func() {
	It("runs the cpi", func() {
		govcClient := &fakegovc.FakeGovcClient{}
		agentSettings := &fakevm.FakeAgentSettings{}
		uuidGen := &fakeuuid.FakeGenerator{}
		agentOptions := apiv1.AgentOptions{}
		agentEnvFactory := apiv1.NewAgentEnvFactory()
		m := action.NewCreateVMMethod(govcClient, agentSettings, agentOptions, agentEnvFactory, uuidGen)

		agentId := apiv1.AgentID{}
		stemcellCid := apiv1.StemcellCID{}
		vmEnv := apiv1.VMEnv{}
		cid, err := m.CreateVM(agentId, stemcellCid, nil, nil, nil, vmEnv)

		Expect(err).ToNot(HaveOccurred())
		Expect(cid.AsString()).To(Equal("fake-uuid-0"))
	})
})
