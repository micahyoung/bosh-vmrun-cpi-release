package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
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
	logger          boshlog.Logger
}

func NewCreateVMMethod(govcClient govc.GovcClient, agentSettings vm.AgentSettings, agentOptions apiv1.AgentOptions, agentEnvFactory apiv1.AgentEnvFactory, uuidGen boshuuid.Generator, logger boshlog.Logger) CreateVMMethod {
	return CreateVMMethod{
		govcClient:      govcClient,
		agentSettings:   agentSettings,
		agentOptions:    agentOptions,
		agentEnvFactory: agentEnvFactory,
		uuidGen:         uuidGen,
		logger:          logger,
	}
}

func (c CreateVMMethod) CreateVM(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	associatedDiskCIDs []apiv1.DiskCID, vmEnv apiv1.VMEnv) (apiv1.VMCID, error) {

	c.logger.Debug("cpi", "CloudProps: %+v\n", cloudProps)
	c.logger.Debug("cpi", "Networks: %+v\n", networks)
	c.logger.Debug("cpi", "AssociatedDiskCIDs: %+v\n", associatedDiskCIDs)

	vmUuid, _ := c.uuidGen.Generate()
	newVMCID := apiv1.NewVMCID(vmUuid)

	stemcellId := "cs-" + stemcellCID.AsString()
	vmId := "vm-" + vmUuid

	_, err := c.govcClient.CloneVM(stemcellId, vmId)
	if err != nil {
		return newVMCID, err
	}

	agentEnv := c.agentEnvFactory.ForVM(agentID, newVMCID, networks, vmEnv, c.agentOptions)
	agentEnv.AttachSystemDisk("0")
	agentEnv.AttachEphemeralDisk("1")

	diskUuid, _ := c.uuidGen.Generate()
	diskId := "disk-" + diskUuid
	err = c.govcClient.CreateDisk(diskId, 16384000)
	if err != nil {
		return newVMCID, err
	}

	err = c.govcClient.AttachDisk(vmId, diskId)
	if err != nil {
		return newVMCID, err
	}

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
