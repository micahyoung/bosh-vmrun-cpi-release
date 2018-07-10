package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/driver"
)

type HasVMMethod struct {
	driverClient driver.Client
}

func NewHasVMMethod(driverClient driver.Client) HasVMMethod {
	return HasVMMethod{
		driverClient: driverClient,
	}
}

func (c HasVMMethod) HasVM(vmCid apiv1.VMCID) (bool, error) {
	vmId := "vm-" + vmCid.AsString()

	vmFound := c.driverClient.HasVM(vmId)

	return vmFound, nil
}
