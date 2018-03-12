package action

import (
	//"fmt"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

func (c CPI) CreateStemcell(imagePath string, _ apiv1.StemcellCloudProps) (apiv1.StemcellCID, error) {
	ovfPath, err := c.stemcellClient.ExtractOvf(imagePath)
	if err != nil {
		return apiv1.StemcellCID{}, err
	}
	_, err = c.govcClient.ImportOvf(ovfPath)
	if err != nil {
		return apiv1.StemcellCID{}, err
	}

	return apiv1.NewStemcellCID("stemcell-cid"), nil
}
