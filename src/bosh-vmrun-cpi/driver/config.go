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

func (c ConfigImpl) BootstrapScriptPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Bootstrap_Script_Path
}

func (c ConfigImpl) BootstrapScriptContent() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Bootstrap_Script_Content
}

func (c ConfigImpl) BootstrapInterpreterPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Bootstrap_Interpreter_Path
}

func (c ConfigImpl) BootstrapUsername() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Bootstrap_Username
}

func (c ConfigImpl) BootstrapPassword() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Bootstrap_Password
}

func (c ConfigImpl) NeedsBootstrap() bool {
	return c.BootstrapScriptPath() != "" &&
		c.BootstrapScriptContent() != "" &&
		c.BootstrapInterpreterPath() != "" &&
		c.BootstrapUsername() != "" &&
		c.BootstrapPassword() != ""
}
