package action

import (
	"fmt"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"github.com/micahyoung/bosh-govmomi-cpi/govc"
)

type DeleteStemcellMethod struct {
	govcClient govc.GovcClient
	logger     boshlog.Logger
}

func NewDeleteStemcellMethod(govcClient govc.GovcClient, logger boshlog.Logger) DeleteStemcellMethod {
	return DeleteStemcellMethod{
		govcClient: govcClient,
		logger:     logger,
	}
}

func (c DeleteStemcellMethod) DeleteStemcell(stemcellCid apiv1.StemcellCID) error {
	stemcellId := "cs-" + stemcellCid.AsString()
	_, err := c.govcClient.DestroyVM(stemcellId)
	if err != nil {
		c.logger.Error("delete-stemcell", fmt.Sprintf("failed to delete stemcell. cid: %s", stemcellCid))
		return err
	}

	return nil
}
