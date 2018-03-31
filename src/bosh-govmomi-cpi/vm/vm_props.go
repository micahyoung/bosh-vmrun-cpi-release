package vm

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type VMProps struct {
	CPU  int
	RAM  int
	Disk int
}

func NewVMProps(cloudProps apiv1.VMCloudProps) (VMProps, error) {
	vmProps := VMProps{
		CPU:  1,
		RAM:  512,
		Disk: 5000,
	}

	err := cloudProps.As(&vmProps)
	if err != nil {
		return VMProps{}, err
	}

	return vmProps, nil
}
