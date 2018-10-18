package stemcell

import (
	cpiconfig "bosh-vmrun-cpi/config"
)

type ConfigImpl struct {
	cpiConfig cpiconfig.Config
}

func NewConfig(cpiConfig cpiconfig.Config) Config {
	return &ConfigImpl{cpiConfig: cpiConfig}
}

func (c ConfigImpl) StemcellStorePath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Stemcell_Store_Path
}
