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
	defer c.stemcellStore.Cleanup()

	stemcellUuid, _ := c.uuidGen.Generate()
	stemcellId := "cs-" + stemcellUuid
	stemcellCID := apiv1.NewStemcellCID(stemcellUuid)
	storeImagePath := ""

	c.logger.Debug("cpi", "stemcellCloudProps: %#+v", stemcellCloudProps)
	stemcellProps, err := stemcell.NewStemcellProps(stemcellCloudProps)
	if err != nil {
		return stemcellCID, err
	}

	c.logger.Debug("cpi", "LocalImagePath: %s", localImagePath)

	storeImagePath, err = c.stemcellStore.GetByImagePathMapping(localImagePath)
	if err != nil {
		c.logger.Debug("cpi", "Stemcell image found at path:", localImagePath)

		return stemcellCID, err
	}

	if storeImagePath == "" {
		storeImagePath, err = c.stemcellStore.GetByMetadata(stemcellProps.Name, stemcellProps.Version)
		if err != nil {
			c.logger.Debug("cpi", "Stemcell image found by metadata")
			return stemcellCID, err
		}
	}

	c.logger.DebugWithDetails("cpi", "StoreImagePath:", storeImagePath)

	if storeImagePath == "" {
		return stemcellCID, errors.New("stemcell image not found locally or in image store")
	}

	ovfPath, err := c.stemcellClient.ExtractOvf(storeImagePath)
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
