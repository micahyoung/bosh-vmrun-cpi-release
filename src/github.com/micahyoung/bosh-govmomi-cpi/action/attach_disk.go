package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"github.com/micahyoung/bosh-govmomi-cpi/govc"
	"github.com/micahyoung/bosh-govmomi-cpi/vm"
)

type AttachDiskMethod struct {
	govcClient      govc.GovcClient
	agentSettings   vm.AgentSettings
	agentEnvFactory apiv1.AgentEnvFactory
}

func NewAttachDiskMethod(govcClient govc.GovcClient, agentSettings vm.AgentSettings, agentEnvFactory apiv1.AgentEnvFactory) AttachDiskMethod {
	return AttachDiskMethod{
		govcClient:      govcClient,
		agentSettings:   agentSettings,
		agentEnvFactory: agentEnvFactory,
	}
}

func (c AttachDiskMethod) AttachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	vmId := "vm-" + vmCID.AsString()
	diskId := "disk-" + diskCID.AsString()

	err := c.govcClient.AttachDisk(vmId, diskId)
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

	_, err = c.govcClient.UpdateVMIso(vmId, envIsoPath)
	if err != nil {
		return err
	}
	c.agentSettings.Cleanup()

	return nil
}
