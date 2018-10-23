package action

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/stemcell"
)

type CreateStemcellMethod struct {
	driverClient   driver.Client
	stemcellClient stemcell.StemcellClient
	stemcellStore  stemcell.StemcellStore
	uuidGen        boshuuid.Generator
	fs             boshsys.FileSystem
	logger         boshlog.Logger
}

func NewCreateStemcellMethod(driverClient driver.Client, stemcellClient stemcell.StemcellClient, stemcellStore stemcell.StemcellStore, uuidGen boshuuid.Generator, fs boshsys.FileSystem, logger boshlog.Logger) CreateStemcellMethod {
	return CreateStemcellMethod{driverClient: driverClient, stemcellClient: stemcellClient, stemcellStore: stemcellStore, uuidGen: uuidGen, fs: fs, logger: logger}
}

func (c CreateStemcellMethod) CreateStemcell(localImagePath string, stemcellCloudProps apiv1.StemcellCloudProps) (apiv1.StemcellCID, error) {
	stemcellUuid, _ := c.uuidGen.Generate()
	stemcellId := "cs-" + stemcellUuid
	stemcellCID := apiv1.NewStemcellCID(stemcellUuid)
	imagePath := ""

	c.logger.DebugWithDetails("cpi", "stemcellCloudProps:", stemcellCloudProps)
	stemcellProps, err := stemcell.NewStemcellProps(stemcellCloudProps)
	if err != nil {
		return stemcellCID, err
	}

	c.logger.DebugWithDetails("cpi", "LocalImagePath:", localImagePath)

        if c.fs.FileExists(localImagePath) {
		imagePath = localImagePath
	} else {
		storeImagePath, err := c.stemcellStore.GetImagePath(stemcellProps.Name, stemcellProps.Version)
		if err != nil {
			return stemcellCID, err
		}
		defer c.stemcellStore.Cleanup()

		c.logger.DebugWithDetails("cpi", "StoreImagePath:", storeImagePath)

		if c.fs.FileExists(storeImagePath) {
			imagePath = storeImagePath
		} else {
			return stemcellCID, errors.New("stemcell image not found locally or in image store")
		}
	}

	ovfPath, err := c.stemcellClient.ExtractOvf(imagePath)
	if err != nil {
		return stemcellCID, err
	}
	defer c.stemcellClient.Cleanup()

	_, err = c.driverClient.ImportOvf(ovfPath, stemcellId)
	if err != nil {
		return stemcellCID, err
	}

	return stemcellCID, nil
}
