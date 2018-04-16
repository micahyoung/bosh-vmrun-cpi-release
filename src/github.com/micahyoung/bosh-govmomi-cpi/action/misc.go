package action

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"github.com/micahyoung/bosh-govmomi-cpi/govc"
)

type MiscMethod struct{}

func NewMiscMethod(govcClient govc.GovcClient) MiscMethod {
	return MiscMethod{}
}

func (c MiscMethod) Info() (apiv1.Info, error) {
	return apiv1.Info{
		StemcellFormats: []string{"general-ovf", "vsphere-ovf"},
	}, nil
}
