package driver

import (
	cpiconfig "bosh-vmrun-cpi/config"
)

type ConfigImpl struct {
	cpiConfig cpiconfig.Config
}

func NewConfig(cpiConfig cpiconfig.Config) Config {
	return &ConfigImpl{cpiConfig: cpiConfig}
}

func (c ConfigImpl) OvftoolPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Ovftool_Bin_Path
}

func (c ConfigImpl) VmrunPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Vmrun_Bin_Path
}

func (c ConfigImpl) VdiskmanagerPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Vdiskmanager_Bin_Path
}

func (c ConfigImpl) VmPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Vm_Store_Path
}
