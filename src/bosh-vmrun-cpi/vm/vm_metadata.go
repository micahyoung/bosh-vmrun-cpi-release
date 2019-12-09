package vm

import (
	"encoding/json"
	"fmt"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type vmMetadata struct {
	Director      string
	Deployment    string
	InstanceGroup string `json:"instance_group"`
	Job           string
	Index         string
	Name          string
}

func NewVMMetadata(apiVMMeta apiv1.VMMeta) (*vmMetadata, error) {
	var err error
	data, err := apiVMMeta.MarshalJSON()
	if err != nil {
		return nil, err
	}

	vmMeta := &vmMetadata{}
	err = json.Unmarshal(data, vmMeta)
	if err != nil {
		return nil, err
	}

	return vmMeta, nil
}

func (v *vmMetadata) DisplayName(vmID string) string {
	var vmDisplayName string
	switch {
	case v.InstanceGroup != "" && v.Deployment != "" && v.Index != "":
		vmDisplayName = fmt.Sprintf("%s_%s_%s", v.InstanceGroup, v.Deployment, v.Index)
	case v.Name != "":
		vmDisplayName = v.Name
	default:
		vmDisplayName = vmID
	}

	return vmDisplayName
}
