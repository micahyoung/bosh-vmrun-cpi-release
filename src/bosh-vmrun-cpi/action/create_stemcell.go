package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/stemcell"
)

type CreateStemcellMethod struct {
	driverClient   driver.Client
	stemcellClient stemcell.StemcellClient
	uuidGen        boshuuid.Generator
	logger         boshlog.Logger
}

func NewCreateStemcellMethod(driverClient driver.Client, stemcellClient stemcell.StemcellClient, uuidGen boshuuid.Generator, logger boshlog.Logger) CreateStemcellMethod {
	return CreateStemcellMethod{driverClient: driverClient, stemcellClient: stemcellClient, uuidGen: uuidGen, logger: logger}
}

func (c CreateStemcellMethod) CreateStemcell(imagePath string, _ apiv1.StemcellCloudProps) (apiv1.StemcellCID, error) {
	stemcellUuid, _ := c.uuidGen.Generate()
	stemcellId := "cs-" + stemcellUuid
	stemcellCID := apiv1.NewStemcellCID(stemcellUuid)

	c.logger.Debug("cpi", "ImagePath: %s", imagePath)
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
