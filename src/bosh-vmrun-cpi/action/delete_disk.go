package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/driver"
)

type DeleteDiskMethod struct {
	driverClient driver.Client
	logger       boshlog.Logger
}

func NewDeleteDiskMethod(driverClient driver.Client, logger boshlog.Logger) DeleteDiskMethod {
	return DeleteDiskMethod{
		driverClient: driverClient,
		logger:       logger,
	}
}

func (c DeleteDiskMethod) DeleteDisk(cid apiv1.DiskCID) error {
	diskId := "disk-" + cid.AsString()

	err := c.driverClient.DestroyDisk(diskId)
	if err != nil {
		c.logger.Error("cpi", "deleting disk: %s\n", diskId)
		return err
	}
	return nil
}
