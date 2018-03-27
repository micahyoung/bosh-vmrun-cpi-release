package action

import (
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-govmomi-cpi/govc"
)

type CreateDiskMethod struct {
	govcClient govc.GovcClient
	uuidGen    boshuuid.Generator
}

func NewCreateDiskMethod(govcClient govc.GovcClient, uuidGen boshuuid.Generator) CreateDiskMethod {
	return CreateDiskMethod{
		govcClient: govcClient,
		uuidGen:    uuidGen,
	}
}

func (c CreateDiskMethod) CreateDisk(sizeMB int,
	cloudProps apiv1.DiskCloudProps, associatedVMCID *apiv1.VMCID) (apiv1.DiskCID, error) {

	diskUuid, _ := c.uuidGen.Generate()
	diskId := "disk-" + diskUuid
	newDiskCID := apiv1.NewDiskCID(diskUuid)
	sizeKB := sizeMB * 1000

	err := c.govcClient.CreateDisk(diskId, sizeKB)
	if err != nil {
		return newDiskCID, err
	}
	return newDiskCID, nil
}
