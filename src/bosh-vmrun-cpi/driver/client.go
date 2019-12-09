package driver

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-vmrun-cpi/vmx"
)

//TODO: use boshfs for fs operations
type ClientImpl struct {
	config             Config
	vmrunRunner        VmrunRunner
	ovftoolRunner      OvftoolRunner
	vmxBuilder         vmx.VmxBuilder
	vdiskmanagerRunner VdiskmanagerRunner
	logger             boshlog.Logger
}

var (
	STATE_NOT_FOUND = "state-not-found"
	STATE_POWER_ON  = "state-on"
	STATE_POWER_OFF = "state-off"
)

func NewClient(vmrunRunner VmrunRunner, ovftoolRunner OvftoolRunner, vdiskmanagerRunner VdiskmanagerRunner, vmxBuilder vmx.VmxBuilder, config Config, logger boshlog.Logger) Client {
	return ClientImpl{vmrunRunner: vmrunRunner, ovftoolRunner: ovftoolRunner, vdiskmanagerRunner: vdiskmanagerRunner, vmxBuilder: vmxBuilder, config: config, logger: logger}
}

func (c ClientImpl) ImportOvf(ovfPath string, vmName string) (bool, error) {
	err := c.ovftoolRunner.ImportOvf(ovfPath, c.config.VmxPath(vmName), vmName)
	if err != nil {
		c.logger.ErrorWithDetails("client", "import ovf: runner", err)
		return false, err
	}

	return true, nil
}

func (c ClientImpl) CloneVM(sourceVmName string, cloneVmName string) error {
	var err error

	err = c.vmrunRunner.Clone(c.config.VmxPath(sourceVmName), c.config.VmxPath(cloneVmName), cloneVmName)
	if err != nil {
		c.logger.ErrorWithDetails("client", "clone vm: clone stemcell", err)
		return err
	}

	err = c.vmxBuilder.InitHardware(c.config.VmxPath(cloneVmName))
	if err != nil {
		c.logger.ErrorWithDetails("client", "clone vm: configuring vm hardware", err)
		return err
	}

	return nil
}

func (c ClientImpl) SetVMNetworkAdapter(vmName string, networkName string, macAddress string) error {
	var err error

	err = c.vmxBuilder.AddNetworkInterface(networkName, macAddress, c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "adding network", err, vmName, networkName, macAddress)
		return err
	}

	return nil
}

func (c ClientImpl) SetVMResources(vmName string, cpuCount int, ramMB int) error {
	err := c.vmxBuilder.SetVMResources(cpuCount, ramMB, c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "setting vm cpu and ram", err)
		return err
	}

	return nil
}

func (c ClientImpl) GetVMIsoPath(vmName string) string {
	path := c.config.EnvIsoPath(vmName)
	if _, err := os.Stat(path); err != nil {
		return ""
	}

	return path
}

func (c ClientImpl) UpdateVMIso(vmName string, localIsoPath string) error {
	var err error

	isoBytes, err := ioutil.ReadFile(localIsoPath)
	if err != nil {
		c.logger.ErrorWithDetails("driver", "reading generated iso", err)
		return err
	}

	err = ioutil.WriteFile(c.config.EnvIsoPath(vmName), isoBytes, 0644)
	if err != nil {
		c.logger.ErrorWithDetails("driver", "writing vm iso contents", err)
		return err
	}

	err = c.vmxBuilder.AttachCdrom(c.config.EnvIsoPath(vmName), c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "connecting ENV cdrom", err)
		return err
	}

	return nil
}

func (c ClientImpl) StartVM(vmName string) error {
	var err error

	err = c.vmrunRunner.Start(c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "starting VM", err)
		return err
	}

	err = c.waitForVMStart(vmName)
	if err != nil {
		c.logger.ErrorWithDetails("driver", "waiting for VM to start", err)
		return err
	}

	return nil
}

func (c ClientImpl) waitForVMStart(vmName string) error {
	interval := time.Second
	for i := time.Duration(0); i < c.config.VmSoftShutdownMaxWait(); i += interval {
		var vmState string
		var err error

		if vmState, err = c.vmState(vmName); err != nil {
			return err
		}

		if vmState == STATE_POWER_ON {
			// vm not always ready as soon as state changes
			time.Sleep(interval)

			return nil
		}

		c.logger.DebugWithDetails("driver", "polling vm start state:", vmState)
		time.Sleep(interval)
	}

	return errors.New("timeout")
}

func (c ClientImpl) BootstrapVM(vmName, scriptContent, scriptPath, interpreterPath, readyProcessName, username, password string, vmReadyMinWait, vmReadyMaxWait time.Duration) error {
	var err error

	err = c.vmrunRunner.Start(c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "starting VM for bootstrapping", err)
		return err
	}

	c.logger.Debug("driver", "waiting for VM to be ready to bootstrap")
	err = c.waitForVMReady(vmName, readyProcessName, username, password, vmReadyMinWait, vmReadyMaxWait)
	if err != nil {
		c.logger.ErrorWithDetails("driver", "waiting for VM to be ready to bootstrap", err)
		return err
	}

	c.logger.Debug("driver", "copying bootstrap script")
	err = c.copyBootstrapScript(vmName, scriptContent, scriptPath, username, password)
	if err != nil {
		c.logger.ErrorWithDetails("driver", "copying bootstrap script for VM", err)
		return err
	}

	c.logger.Debug("driver", "running bootstrap script")
	err = c.vmrunRunner.RunProgramInGuest(c.config.VmxPath(vmName), interpreterPath, scriptPath, username, password)
	if err != nil {
		c.logger.ErrorWithDetails("driver", "running bootstrap script for VM", err)
		return err
	}

	err = c.vmrunRunner.SoftStop(c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "stopping VM after bootstrapping", err)
		return err
	}

	return nil
}

func (c ClientImpl) waitForVMReady(vmName, readyProcessName, username, password string, vmReadyMinWait, vmReadyMaxWait time.Duration) error {
	//allow VM to settle before polling
	time.Sleep(vmReadyMinWait)

	interval := time.Second
	for i := time.Duration(0); i < vmReadyMaxWait; i += interval {
		var processes string
		var err error

		processes, err = c.vmrunRunner.ListProcessesInGuest(c.config.VmxPath(vmName), username, password)

		c.logger.DebugWithDetails("driver", "polling vm processes for vm readiness", processes)

		if err != nil {
			return err
		}

		//continue if wait process has started
		if strings.Contains(processes, readyProcessName) {
			return nil
		}

		time.Sleep(interval)
	}

	return errors.New("timeout")
}

func (c ClientImpl) copyBootstrapScript(vmName, scriptContent, scriptPath, username, password string) error {
	var err error
	var file *os.File

	if file, err = ioutil.TempFile("", ""); err != nil {
		return err
	}
	defer file.Close()

	if file.WriteString(scriptContent); err != nil {
		return err
	}

	if err := c.vmrunRunner.CopyFileFromHostToGuest(c.config.VmxPath(vmName), file.Name(), scriptPath, username, password); err != nil {
		return err
	}

	return nil
}

func (c ClientImpl) runBootstrapScript(vmName, scriptPath, interpreterPath, username, password string) error {
	return nil
}

func (c ClientImpl) HasVM(vmName string) bool {
	vmxPath := c.config.VmxPath(vmName)
	if _, err := os.Stat(vmxPath); err != nil {
		return false
	} else {
		c.logger.Debug("driver", "vmx file exists %s", vmxPath)
		return true
	}
}

func (c ClientImpl) NeedsVMNameChange(vmName string) bool {
	vmInfo, err := c.GetVMInfo(vmName)
	if err != nil {
		return false
	}

	if vmInfo.Name == vmName && c.config.EnableHumanReadableName() {
		return true
	}

	return false
}

func (c ClientImpl) SetVMDisplayName(vmName, displayName string) error {
	var err error

	c.logger.Debug("driver", "Setting VM Display Name")
	err = c.vmxBuilder.SetVMDisplayName(displayName, c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "Setting VM Display Name", err)
		return err
	}

	return nil
}

func (c ClientImpl) CreateEphemeralDisk(vmName string, diskMB int) error {
	var err error

	err = c.vdiskmanagerRunner.CreateDisk(c.config.EphemeralDiskPath(vmName), diskMB)
	if err != nil {
		c.logger.ErrorWithDetails("driver", "CreateEphemeralDisk create", err)
		return err
	}

	err = c.vmxBuilder.AttachDisk(c.config.EphemeralDiskPath(vmName), c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "CreateEphemeralDisk attach", err)
		return err
	}

	return nil
}

func (c ClientImpl) CreateDisk(diskId string, diskMB int) error {
	var err error

	err = c.vdiskmanagerRunner.CreateDisk(c.config.PersistentDiskPath(diskId), diskMB)
	if err != nil {
		c.logger.ErrorWithDetails("driver", "CreateDisk", err)
		return err
	}

	return nil
}

func (c ClientImpl) AttachDisk(vmName string, diskId string) error {
	var err error

	err = c.vmxBuilder.AttachDisk(c.config.PersistentDiskPath(diskId), c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "AttachDisk", err)
		return err
	}
	return nil
}

func (c ClientImpl) DetachDisk(vmName string, diskId string) error {
	var err error

	err = c.vmxBuilder.DetachDisk(c.config.PersistentDiskPath(vmName), c.config.VmxPath(vmName))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "DetachDisk", err)
		return err
	}
	return nil
}

func (c ClientImpl) DestroyDisk(diskId string) error {
	var err error

	err = os.Remove(c.config.PersistentDiskPath(diskId))
	if err != nil {
		c.logger.ErrorWithDetails("driver", "DestroyDisk", err)
		return err
	}

	return nil
}

func (c ClientImpl) HasDisk(diskId string) bool {
	diskPath := c.config.PersistentDiskPath(diskId)
	if _, err := os.Stat(diskPath); err != nil {
		c.logger.Debug("driver", "persistent disk file does not exist %s", diskPath)
		return false
	} else {
		c.logger.Debug("driver", "persistent disk file exists %s", diskPath)
		return true
	}
}

func (c ClientImpl) StopVM(vmName string) error {
	var err error
	var vmState string

	vmState, err = c.vmState(vmName)
	if err != nil {
		return err
	}

	if vmState != STATE_POWER_ON {
		return nil
	}

	//run blocking soft-shutdown command in background
	go func() {
		err = c.vmrunRunner.SoftStop(c.config.VmxPath(vmName))
		if err != nil {
			c.logger.Error("driver", "soft stop")
		}
	}()

	interval := time.Second
	for i := time.Duration(0); i < c.config.VmSoftShutdownMaxWait(); i += interval {
		vmInfo, err := c.GetVMInfo(vmName)
		if err != nil {
			return err
		}

		if vmInfo.CleanShutdown {
			return nil
		}

		time.Sleep(interval)
	}

	err = c.vmrunRunner.HardStop(c.config.VmxPath(vmName))
	if err != nil {
		c.logger.Error("driver", "hard stop")
		return err
	}

	return nil
}

//TODO: add more graceful handling of locked vmx (when stopped but GUI has them open)
func (c ClientImpl) DestroyVM(vmName string) error {
	var err error
	var vmState string

	vmState, err = c.vmState(vmName)
	if err != nil {
		return err
	}

	if vmState == STATE_POWER_ON {
		err = c.vmrunRunner.HardStop(c.config.VmxPath(vmName))
		if err != nil {
			return err
		}
	}

	vmState, err = c.vmState(vmName)
	if err != nil {
		return err
	}

	if vmState == STATE_POWER_OFF {
		err = c.vmrunRunner.Delete(c.config.VmxPath(vmName))
		if err != nil {
			return err
		}
	}

	//attempt to cleanup ephemeral disk, ignore error
	_ = os.Remove(c.config.EphemeralDiskPath(vmName))

	return nil
}

func (c ClientImpl) GetVMInfo(vmName string) (VMInfo, error) {
	vmxVM, err := c.vmxBuilder.GetVmx(c.config.VmxPath(vmName))

	if err != nil {
		return VMInfo{}, err
	}
	vmInfo := VMInfo{
		Name:          vmxVM.DisplayName,
		CPUs:          int(vmxVM.NumvCPUs),
		RAM:           int(vmxVM.Memsize),
		CleanShutdown: vmxVM.CleanShutdown,
	}

	for _, vmxNic := range vmxVM.Ethernet {
		vmInfo.NICs = append(vmInfo.NICs, struct {
			Network string
			MAC     string
		}{
			Network: vmxNic.VNetwork,
			MAC:     vmxNic.Address,
		})
	}

	for _, scsiDevice := range vmxVM.SCSIDevices {
		vmInfo.Disks = append(vmInfo.Disks, struct {
			ID   string
			Path string
		}{
			ID:   scsiDevice.VMXID,
			Path: scsiDevice.Filename,
		})
	}

	return vmInfo, err
}

//TODO: should match on full VMX path instead of just name
//      failing due to vmxPath substring not matching with string.Contains (maybe unicode problem?)
func (c ClientImpl) vmState(vmName string) (string, error) {
	result, err := c.vmrunRunner.List()
	if err != nil {
		return result, err
	}

	if !c.HasVM(vmName) {
		return STATE_NOT_FOUND, nil
	}

	if strings.Contains(result, vmName) {
		return STATE_POWER_ON, nil
	}

	return STATE_POWER_OFF, nil
}
