package vm

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type VMProps struct {
	CPU                        int
	RAM                        int
	Disk                       int
	Bootstrap_Script_Content   string
	Bootstrap_Script_Path      string
	Bootstrap_Interpreter_Path string
	Bootstrap_Username         string
	Bootstrap_Password         string
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

func (p VMProps) NeedsBootstrap() bool {
	return p.Bootstrap_Script_Path != "" &&
		p.Bootstrap_Script_Content != "" &&
		p.Bootstrap_Interpreter_Path != "" &&
		p.Bootstrap_Username != "" &&
		p.Bootstrap_Password != ""
}
