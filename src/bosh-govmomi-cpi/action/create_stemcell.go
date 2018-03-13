package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-govmomi-cpi/govc"
	"bosh-govmomi-cpi/stemcell"
)

type CreateStemcellMethod struct {
	govcClient     govc.GovcClient
	stemcellClient stemcell.StemcellClient
	logger         boshlog.Logger
}

func NewCreateStemcellMethod(govcClient govc.GovcClient, stemcellClient stemcell.StemcellClient, logger boshlog.Logger) CreateStemcellMethod {
	return CreateStemcellMethod{govcClient: govcClient, stemcellClient: stemcellClient, logger: logger}
}

func (c CreateStemcellMethod) CreateStemcell(imagePath string, _ apiv1.StemcellCloudProps) (apiv1.StemcellCID, error) {
	c.logger.Debug("cpi", "ImagePath: %s", imagePath)
	ovfPath, err := c.stemcellClient.ExtractOvf(imagePath)
	if err != nil {
		return apiv1.StemcellCID{}, err
	}
	_, err = c.govcClient.ImportOvf(ovfPath)
	if err != nil {
		return apiv1.StemcellCID{}, err
	}
	err = c.stemcellClient.Cleanup()
	if err != nil {
		return apiv1.StemcellCID{}, err
	}

	return apiv1.NewStemcellCID("stemcell-cid"), nil
}
