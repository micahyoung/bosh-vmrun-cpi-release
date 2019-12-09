package action

import (
	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/vm"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type SetVMMetadataMethod struct {
	driverClient driver.Client
	logger       boshlog.Logger
}

func NewSetVMMetadataMethod(driverClient driver.Client, logger boshlog.Logger) SetVMMetadataMethod {
	return SetVMMetadataMethod{driverClient: driverClient, logger: logger}
}

func (c SetVMMetadataMethod) SetVMMetadata(vmCid apiv1.VMCID, apiVMMeta apiv1.VMMeta) error {
	var err error
	vmId := "vm-" + vmCid.AsString()

	if !c.driverClient.NeedsVMNameChange(vmId) {
		return nil
	}

	c.logger.DebugWithDetails("SetVMMetadata", "metadata:", apiVMMeta)
	vmMeta, err := vm.NewVMMetadata(apiVMMeta)
	if err != nil {
		return err
	}

	vmDisplayName := vmMeta.DisplayName(vmId)

	c.logger.Debug("SetVMMetadata", "Setting VM display name %s", vmDisplayName)
	err = c.driverClient.SetVMDisplayName(vmId, vmDisplayName)
	if err != nil {
		return err
	}

	c.logger.Debug("SetVMMetadata", "Starting VM after setting vm metadata")
	err = c.driverClient.StartVM(vmId)
	if err != nil {
		return err
	}

	return nil
}
