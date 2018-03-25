package action

import (
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-govmomi-cpi/govc"
	"bosh-govmomi-cpi/vm"
)

type CreateVMMethod struct {
	govcClient      govc.GovcClient
	agentSettings   vm.AgentSettings
	agentOptions    apiv1.AgentOptions
	agentEnvFactory apiv1.AgentEnvFactory
	uuidGen         boshuuid.Generator
}

func NewCreateVMMethod(govcClient govc.GovcClient, agentSettings vm.AgentSettings, agentOptions apiv1.AgentOptions, agentEnvFactory apiv1.AgentEnvFactory, uuidGen boshuuid.Generator) CreateVMMethod {
	return CreateVMMethod{
		govcClient:      govcClient,
		agentSettings:   agentSettings,
		agentOptions:    agentOptions,
		agentEnvFactory: agentEnvFactory,
		uuidGen:         uuidGen,
	}
}

func (c CreateVMMethod) CreateVM(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	associatedDiskCIDs []apiv1.DiskCID, vmEnv apiv1.VMEnv) (apiv1.VMCID, error) {

	stemcellId := stemcellCID.AsString()
	vmId, _ := c.uuidGen.Generate()
	newVMCID := apiv1.NewVMCID(vmId)

	_, err := c.govcClient.CloneVM(stemcellId, vmId)
	if err != nil {
		return newVMCID, err
	}

	agentEnv := c.agentEnvFactory.ForVM(agentID, newVMCID, networks, vmEnv, c.agentOptions)
	envIsoPath, err := c.agentSettings.GenerateAgentEnvIso(agentEnv)
	if err != nil {
		return newVMCID, err
	}

	_, err = c.govcClient.UpdateVMIso(vmId, envIsoPath)
	if err != nil {
		return newVMCID, err
	}
	c.agentSettings.Cleanup()

	_, err = c.govcClient.StartVM(vmId)
	if err != nil {
		return newVMCID, err
	}

	return newVMCID, nil
}
