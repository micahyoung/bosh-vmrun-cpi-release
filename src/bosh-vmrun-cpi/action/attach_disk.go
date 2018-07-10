package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/vm"
)

type AttachDiskMethod struct {
	driverClient    driver.Client
	agentSettings   vm.AgentSettings
	agentEnvFactory apiv1.AgentEnvFactory
}

func NewAttachDiskMethod(driverClient driver.Client, agentSettings vm.AgentSettings, agentEnvFactory apiv1.AgentEnvFactory) AttachDiskMethod {
	return AttachDiskMethod{
		driverClient:    driverClient,
		agentSettings:   agentSettings,
		agentEnvFactory: agentEnvFactory,
	}
}

func (c AttachDiskMethod) AttachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	var err error
	vmId := "vm-" + vmCID.AsString()
	diskId := "disk-" + diskCID.AsString()

	err = c.driverClient.StopVM(vmId)
	if err != nil {
		return err
	}

	err = c.driverClient.AttachDisk(vmId, diskId)
	if err != nil {
		return err
	}

	agentEnvBytes := c.agentSettings.AgentEnvBytesFromFile()
	agentEnv, err := c.agentEnvFactory.FromBytes(agentEnvBytes)
	if err != nil {
		return err
	}
	agentEnv.AttachPersistentDisk(diskCID, struct {
		Path     string `json:"path"`      //can be removed?
		VolumeID string `json:"volume_id"` //should be 3?
		Lun      string `json:"lun"`
	}{"/dev/sdc", "2", "0"})

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
