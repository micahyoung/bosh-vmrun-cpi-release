package action

import (
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-govmomi-cpi/govc"
)

type CreateVMMethod struct {
	govcClient govc.GovcClient
	uuidGen    boshuuid.Generator
}

func NewCreateVMMethod(govcClient govc.GovcClient, uuidGen boshuuid.Generator) CreateVMMethod {
	return CreateVMMethod{govcClient: govcClient, uuidGen: uuidGen}
}

func (c CreateVMMethod) CreateVM(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	associatedDiskCIDs []apiv1.DiskCID, env apiv1.VMEnv) (apiv1.VMCID, error) {

	vmId, _ := c.uuidGen.Generate()
	stemcellId := stemcellCID.AsString()
	_, err := c.govcClient.CloneVM(stemcellId, vmId)
	if err != nil {
		return apiv1.VMCID{}, err
	}

	return apiv1.NewVMCID(vmId), nil
}
