package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/vm"
)

type DetachDiskMethod struct {
	driverClient    driver.Client
	agentSettings   vm.AgentSettings
	agentEnvFactory apiv1.AgentEnvFactory
}

func NewDetachDiskMethod(driverClient driver.Client, agentSettings vm.AgentSettings, agentEnvFactory apiv1.AgentEnvFactory) DetachDiskMethod {
	return DetachDiskMethod{
		driverClient:    driverClient,
		agentSettings:   agentSettings,
		agentEnvFactory: agentEnvFactory,
	}
}

func (c DetachDiskMethod) DetachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	var err error
	vmId := "vm-" + vmCID.AsString()
	diskId := "disk-" + diskCID.AsString()

	err = c.driverClient.StopVM(vmId)
	if err != nil {
		return err
	}

	err = c.driverClient.DetachDisk(vmId, diskId)
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

	err = c.driverClient.UpdateVMIso(vmId, envIsoPath)
	if err != nil {
		return err
	}
	c.agentSettings.Cleanup()

	err = c.driverClient.StartVM(vmId)
	if err != nil {
		return err
	}

	return nil
}
