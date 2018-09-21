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

func NewDetachDiskMethod(driverClient driver.Client, agentSettings vm.AgentSettings) DetachDiskMethod {
	return DetachDiskMethod{
		driverClient:  driverClient,
		agentSettings: agentSettings,
	}
}

func (c DetachDiskMethod) DetachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	var err error
	var agentEnv apiv1.AgentEnv
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

	currentIsoPath := c.driverClient.GetVMIsoPath(vmId)
	agentEnv, err = c.agentSettings.GetIsoAgentEnv(currentIsoPath)
	if err != nil {
		return err
	}

	agentEnv.DetachPersistentDisk(diskCID)
	agentEnvBytes, err := agentEnv.AsBytes()
	if err != nil {
		return err
	}

	envIsoPath, err := c.agentSettings.GenerateAgentEnvIso(agentEnvBytes)
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
