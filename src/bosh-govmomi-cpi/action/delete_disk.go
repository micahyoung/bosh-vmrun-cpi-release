package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-govmomi-cpi/govc"
)

type DeleteDiskMethod struct {
	govcClient govc.GovcClient
	logger     boshlog.Logger
}

func NewDeleteDiskMethod(govcClient govc.GovcClient, logger boshlog.Logger) DeleteDiskMethod {
	return DeleteDiskMethod{
		govcClient: govcClient,
		logger:     logger,
	}
}

func (c DeleteDiskMethod) DeleteDisk(cid apiv1.DiskCID) error {
	diskId := "disk-" + cid.AsString()

	err := c.govcClient.DestroyDisk(diskId)
	if err != nil {
		c.logger.Error("cpi", "deleting disk: %s\n", diskId)
		return err
	}
	return nil
}
