package stemcell

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type stemcellProps struct {
	Name    string
	Version string
}

func NewStemcellProps(cloudProps apiv1.StemcellCloudProps) (*stemcellProps, error) {
	stemcellProps := &stemcellProps{}

	err := cloudProps.As(&stemcellProps)
	if err != nil {
		return nil, err
	}

	return stemcellProps, nil
}
