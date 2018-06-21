package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/govc"
	"bosh-vmrun-cpi/vm"
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

	c.logger.DebugWithDetails("create-vm", "networks", networks)

	vmUuid, _ := c.uuidGen.Generate()
	newVMCID := apiv1.NewVMCID(vmUuid)

	stemcellId := "cs-" + stemcellCID.AsString()
	vmId := "vm-" + vmUuid

	var vmProps vm.VMProps
	err := cloudProps.As(&vmProps)
	if err != nil {
		return newVMCID, err
	}

	_, err = c.govcClient.CloneVM(stemcellId, vmId)
	if err != nil {
		return newVMCID, err
	}

	err = c.govcClient.SetVMResources(vmId, vmProps.CPU, vmProps.RAM)
	if err != nil {
		return newVMCID, err
	}

	updatedNetworks := apiv1.Networks{}
	for networkName, network := range networks {
		var networkCloudProps struct {
			Name string
			Type string //remove?
		}
		network.CloudProps().As(&networkCloudProps)
		adapterNetworkName := networkCloudProps.Name

		macAddress, err := c.agentSettings.GenerateMacAddress()
		if err != nil {
			return newVMCID, err
		}

		err = c.govcClient.SetVMNetworkAdapter(vmId, adapterNetworkName, macAddress)
		if err != nil {
			return newVMCID, err
		}

		network.SetMAC(macAddress)
		updatedNetworks[networkName] = network
	}

	agentEnv := c.agentEnvFactory.ForVM(agentID, newVMCID, updatedNetworks, vmEnv, c.agentOptions)
	agentEnv.AttachSystemDisk("0")

	err = c.govcClient.CreateEphemeralDisk(vmId, vmProps.Disk)
	if err != nil {
		return newVMCID, err
	}

	agentEnv.AttachEphemeralDisk("1")

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
