package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type MiscMethod struct{}

func NewMiscMethod() MiscMethod {
	return MiscMethod{}
}

func (c MiscMethod) Info() (apiv1.Info, error) {
	return apiv1.Info{
		StemcellFormats: []string{"general-ovf", "vsphere-ovf"},
	}, nil
}
