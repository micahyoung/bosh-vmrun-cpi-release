package driver

import (
	cpiconfig "bosh-vmrun-cpi/config"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ConfigImpl struct {
	cpiConfig cpiconfig.Config
}

func NewConfig(cpiConfig cpiconfig.Config) Config {
	return &ConfigImpl{cpiConfig: cpiConfig}
}

func (c ConfigImpl) VmxPath(vmName string) string {
	return filepath.Join(c.vmPath(), fmt.Sprintf("%s", vmName), fmt.Sprintf("%s.vmx", vmName))
}

func (c ConfigImpl) EphemeralDiskPath(vmName string) string {
	baseDir := filepath.Join(c.vmPath(), "ephemeral-disks")
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0755)
	}

	return filepath.Join(baseDir, fmt.Sprintf("%s.vmdk", vmName))
}

func (c ConfigImpl) PersistentDiskPath(diskId string) string {
	baseDir := filepath.Join(c.vmPath(), "persistent-disks")
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0755)
	}

	return filepath.Join(baseDir, fmt.Sprintf("%s.vmdk", diskId))
}

func (c ConfigImpl) EnvIsoPath(vmName string) string {
	baseDir := filepath.Join(c.vmPath(), "env-isos")
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0755)
	}

	return filepath.Join(baseDir, fmt.Sprintf("%s.iso", vmName))
}

func (c ConfigImpl) VmrunPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Vmrun_Bin_Path
}

func (c ConfigImpl) OvftoolPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Ovftool_Bin_Path
}

func (c ConfigImpl) VmStartMaxWait() time.Duration {
	return c.cpiConfig.Cloud.Properties.Vmrun.Vm_Start_Max_Wait
}

func (c ConfigImpl) VmSoftShutdownMaxWait() time.Duration {
	return c.cpiConfig.Cloud.Properties.Vmrun.Vm_Soft_Shutdown_Max_Wait
}

func (c ConfigImpl) EnableHumanReadableName() bool {
	return c.cpiConfig.Cloud.Properties.Vmrun.Enable_Human_Readable_Name
}

func (c ConfigImpl) UseLinkedCloning() bool {
	return c.cpiConfig.Cloud.Properties.Vmrun.Use_Linked_Cloning
}

func (c ConfigImpl) vmPath() string {
	return c.cpiConfig.Cloud.Properties.Vmrun.Vm_Store_Path
}
