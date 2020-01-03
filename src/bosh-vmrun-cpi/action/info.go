package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type InfoMethod struct{}

func NewInfoMethod() InfoMethod {
	return InfoMethod{}
}

func (c InfoMethod) Info() (apiv1.Info, error) {
	return apiv1.Info{
		StemcellFormats: []string{"general-ovf", "vsphere-ovf"},
	}, nil
}
