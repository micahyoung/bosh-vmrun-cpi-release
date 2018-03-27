package action

import (
	"fmt"

	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-govmomi-cpi/govc"
)

type DeleteVMMethod struct {
	govcClient govc.GovcClient
}

func NewDeleteVMMethod(govcClient govc.GovcClient) DeleteVMMethod {
	return DeleteVMMethod{
		govcClient: govcClient,
	}
}

func (c DeleteVMMethod) DeleteVM(vmCid apiv1.VMCID) error {
	vmId := "vm-" + vmCid.AsString()
	_, err := c.govcClient.DestroyVM(vmId)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return err
	}

	return nil
}
