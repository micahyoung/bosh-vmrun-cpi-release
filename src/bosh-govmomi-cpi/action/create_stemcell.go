package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-govmomi-cpi/govc"
	"bosh-govmomi-cpi/stemcell"
)

type CreateStemcellMethod struct {
	govcClient     govc.GovcClient
	stemcellClient stemcell.StemcellClient
	uuidGen        boshuuid.Generator
	logger         boshlog.Logger
}

func NewCreateStemcellMethod(govcClient govc.GovcClient, stemcellClient stemcell.StemcellClient, uuidGen boshuuid.Generator, logger boshlog.Logger) CreateStemcellMethod {
	return CreateStemcellMethod{govcClient: govcClient, stemcellClient: stemcellClient, uuidGen: uuidGen, logger: logger}
}

func (c CreateStemcellMethod) CreateStemcell(imagePath string, _ apiv1.StemcellCloudProps) (apiv1.StemcellCID, error) {
	c.logger.Debug("cpi", "ImagePath: %s", imagePath)
	ovfPath, err := c.stemcellClient.ExtractOvf(imagePath)
	if err != nil {
		return apiv1.StemcellCID{}, err
	}

	id, _ := c.uuidGen.Generate()
	_, err = c.govcClient.ImportOvf(ovfPath, id)

	if err != nil {
		return apiv1.StemcellCID{}, err
	}
	c.stemcellClient.Cleanup()

	return apiv1.NewStemcellCID(id), nil
}
