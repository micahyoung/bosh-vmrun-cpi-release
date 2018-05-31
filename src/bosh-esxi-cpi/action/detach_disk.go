package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-esxi-cpi/govc"
	"bosh-esxi-cpi/vm"
)

type DetachDiskMethod struct {
	govcClient      govc.GovcClient
	agentSettings   vm.AgentSettings
	agentEnvFactory apiv1.AgentEnvFactory
}

func NewDetachDiskMethod(govcClient govc.GovcClient, agentSettings vm.AgentSettings, agentEnvFactory apiv1.AgentEnvFactory) DetachDiskMethod {
	return DetachDiskMethod{
		govcClient:      govcClient,
		agentSettings:   agentSettings,
		agentEnvFactory: agentEnvFactory,
	}
}

func (c DetachDiskMethod) DetachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	vmId := "vm-" + vmCID.AsString()
	diskId := "disk-" + diskCID.AsString()

	err := c.govcClient.DetachDisk(vmId, diskId)
	if err != nil {
		return err
	}

	agentEnvBytes := c.agentSettings.AgentEnvBytesFromFile()
	agentEnv, err := c.agentEnvFactory.FromBytes(agentEnvBytes)
	if err != nil {
		return err
	}

	agentEnv.DetachPersistentDisk(diskCID)

	envIsoPath, err := c.agentSettings.GenerateAgentEnvIso(agentEnv)
	if err != nil {
		return err
	}

	_, err = c.govcClient.UpdateVMIso(vmId, envIsoPath)
	if err != nil {
		return err
	}
	c.agentSettings.Cleanup()

	return nil
}
