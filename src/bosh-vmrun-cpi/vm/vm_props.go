package vm

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type boostrapProps struct {
	Script_Content   string
	Script_Path      string
	Interpreter_Path string
	Username         string
	Password         string
}

type VMProps struct {
	CPU       int
	RAM       int
	Disk      int
	Bootstrap boostrapProps
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
	return p.Bootstrap.Script_Path != "" &&
		p.Bootstrap.Script_Content != "" &&
		p.Bootstrap.Interpreter_Path != "" &&
		p.Bootstrap.Username != "" &&
		p.Bootstrap.Password != ""
}
