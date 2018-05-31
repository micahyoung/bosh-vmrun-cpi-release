package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-esxi-cpi/govc"
	"bosh-esxi-cpi/stemcell"
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
	stemcellUuid, _ := c.uuidGen.Generate()
	stemcellId := "cs-" + stemcellUuid
	stemcellCID := apiv1.NewStemcellCID(stemcellUuid)

	c.logger.Debug("cpi", "ImagePath: %s", imagePath)
	ovfPath, err := c.stemcellClient.ExtractOvf(imagePath)
	if err != nil {
		return stemcellCID, err
	}

	_, err = c.govcClient.ImportOvf(ovfPath, stemcellId)
	if err != nil {
		return stemcellCID, err
	}
	c.stemcellClient.Cleanup()

	return stemcellCID, nil
}
