package driver

import (
	"fmt"
	"io/ioutil"
	"sort"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/hooklift/govmx"
)

type VmxBuilderImpl struct {
	logger boshlog.Logger
}

func NewVmxBuilder(logger boshlog.Logger) VmxBuilder {
	return VmxBuilderImpl{logger: logger}
}

func (p VmxBuilderImpl) InitHardware(vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *vmx.VirtualMachine) *vmx.VirtualMachine {
		vmxVM.VHVEnable = true
		vmxVM.Tools.SyncTime = true

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) AddNetworkInterface(networkName, macAddress, vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *vmx.VirtualMachine) *vmx.VirtualMachine {
		vmxVM.Ethernet = append(vmxVM.Ethernet, vmx.Ethernet{
			VNetwork:       networkName,
			Address:        macAddress,
			AddressType:    "static",
			VirtualDev:     "vmxnet3",
			ConnectionType: "custom",
			Present:        true,
		})

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) SetVMResources(cpu int, mem int, vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *vmx.VirtualMachine) *vmx.VirtualMachine {
		vmxVM.NumvCPUs = uint(cpu)
		vmxVM.Memsize = uint(mem)

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) AttachDisk(diskPath, vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *vmx.VirtualMachine) *vmx.VirtualMachine {
		newSCSIDevice := vmx.SCSIDevice{Device: vmx.Device{
			Filename: diskPath,
			Present:  true,
			VMXID:    fmt.Sprintf("scsi0:%d", len(vmxVM.SCSIDevices)-1),
		}}
		vmxVM.SCSIDevices = append(vmxVM.SCSIDevices, newSCSIDevice)

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) DetachDisk(diskPath string, vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *vmx.VirtualMachine) *vmx.VirtualMachine {
		for i, device := range vmxVM.SCSIDevices {
			if device.Filename == diskPath {
				//remove i
				vmxVM.SCSIDevices = append(vmxVM.SCSIDevices[:i], vmxVM.SCSIDevices[i+1:]...)

				return vmxVM
			}
		}

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) AttachCdrom(isoPath, vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *vmx.VirtualMachine) *vmx.VirtualMachine {
		newCdromDevice := vmx.IDEDevice{Device: vmx.Device{
			Filename:       isoPath,
			Type:           vmx.CDROM_IMAGE,
			StartConnected: true,
			Present:        true,
		}}
		//assume all IDEDevices are this CDROM drive and clobber
		//TODO: detect existing one or add
		vmxVM.IDEDevices = []vmx.IDEDevice{newCdromDevice}

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) VMInfo(vmxPath string) (VMInfo, error) {
	vmxVM, err := p.getVmx(vmxPath)
	if err != nil {
		return VMInfo{}, err
	}

	//p.logger.DebugWithDetails("vmx-builder", "DEBUG: %+v", vmxVM)

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

	return vmInfo, nil
}

func (p VmxBuilderImpl) GetVmx(vmxPath string) (*vmx.VirtualMachine, error) {
	return p.getVmx(vmxPath)
}

func (p VmxBuilderImpl) replaceVmx(vmxPath string, vmUpdateFunc func(*vmx.VirtualMachine) *vmx.VirtualMachine) error {
	vmxVM, err := p.getVmx(vmxPath)
	if err != nil {
		return err
	}

	vmxVM = vmUpdateFunc(vmxVM)

	err = p.writeVmx(vmxVM, vmxPath)
	if err != nil {
		return err
	}

	return nil
}

func (p VmxBuilderImpl) getVmx(vmxPath string) (*vmx.VirtualMachine, error) {
	var err error

	vmxBytes, err := ioutil.ReadFile(vmxPath)
	if err != nil {
		p.logger.ErrorWithDetails("vmx-builder", "reading file: %s", vmxPath)
		return nil, err
	}

	vmxVM := new(vmx.VirtualMachine)
	err = vmx.Unmarshal(vmxBytes, vmxVM)
	if err != nil {
		p.logger.ErrorWithDetails("vmx-builder", "unmarshaling file: %s", vmxPath)
		return nil, err
	}

	//consistently sort disks by VMXID
	sort.SliceStable(vmxVM.SCSIDevices, func(i, j int) bool {
		return vmxVM.SCSIDevices[i].VMXID < vmxVM.SCSIDevices[j].VMXID
	})

	return vmxVM, nil
}

func (p VmxBuilderImpl) writeVmx(vmxVM *vmx.VirtualMachine, vmxPath string) error {
	var err error

	vmxBytes, err := vmx.Marshal(vmxVM)
	if err != nil {
		p.logger.ErrorWithDetails("vmx-builder", "marshaling content: %+v", vmxVM)
		return err
	}

	err = ioutil.WriteFile(vmxPath, vmxBytes, 0644)
	if err != nil {
		p.logger.ErrorWithDetails("vmx-builder", "writing file: %s", vmxPath)
		return err
	}

	return nil
}
