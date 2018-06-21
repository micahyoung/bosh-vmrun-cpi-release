package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/govc"
)

type HasVMMethod struct {
	govcClient govc.GovcClient
}

func NewHasVMMethod(govcClient govc.GovcClient) HasVMMethod {
	return HasVMMethod{
		govcClient: govcClient,
	}
}

func (c HasVMMethod) HasVM(vmCid apiv1.VMCID) (bool, error) {
	vmId := "vm-" + vmCid.AsString()

	vmFound, err := c.govcClient.HasVM(vmId)
	if err != nil {
		return false, err
	}

	return vmFound, nil
}
