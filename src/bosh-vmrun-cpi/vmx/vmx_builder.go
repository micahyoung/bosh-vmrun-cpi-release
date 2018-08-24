package vmx

import (
	"fmt"
	"io/ioutil"
	"sort"

	govmx "github.com/hooklift/govmx"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type VmxBuilderImpl struct {
	logger boshlog.Logger
}

func NewVmxBuilder(logger boshlog.Logger) VmxBuilder {
	return VmxBuilderImpl{logger: logger}
}

func (p VmxBuilderImpl) InitHardware(vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *VM) *VM {
		vmxVM.VHVEnable = true
		vmxVM.Tools.SyncTime = true

		// Disable swap: https://kb.vmware.com/s/article/1008885
		vmxVM.MinVmMemPct = 100
		vmxVM.MemTrimRate = 0
		vmxVM.UseNamedFile = false
		vmxVM.Pshare = false
		vmxVM.UseRecommendedLockedMemSize = true
		vmxVM.MainmemBacking = "swap"

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) AddNetworkInterface(networkName, macAddress, vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *VM) *VM {
		vmxVM.Ethernet = append(vmxVM.Ethernet, govmx.Ethernet{
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
	err := p.replaceVmx(vmxPath, func(vmxVM *VM) *VM {
		vmxVM.NumvCPUs = uint(cpu)
		vmxVM.Memsize = uint(mem)

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) AttachDisk(diskPath, vmxPath string) error {
	err := p.replaceVmx(vmxPath, func(vmxVM *VM) *VM {
		newSCSIDevice := govmx.SCSIDevice{Device: govmx.Device{
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
	err := p.replaceVmx(vmxPath, func(vmxVM *VM) *VM {
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
	err := p.replaceVmx(vmxPath, func(vmxVM *VM) *VM {
		newCdromDevice := govmx.IDEDevice{Device: govmx.Device{
			Filename:       isoPath,
			Type:           govmx.CDROM_IMAGE,
			StartConnected: true,
			Present:        true,
		}}
		//assume all IDEDevices are this CDROM drive and clobber
		//TODO: detect existing one or add
		vmxVM.IDEDevices = []govmx.IDEDevice{newCdromDevice}

		return vmxVM
	})

	return err
}

func (p VmxBuilderImpl) GetVmx(vmxPath string) (*VM, error) {
	return p.getVmx(vmxPath)
}

func (p VmxBuilderImpl) replaceVmx(vmxPath string, vmUpdateFunc func(*VM) *VM) error {
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

func (p VmxBuilderImpl) getVmx(vmxPath string) (*VM, error) {
	var err error

	vmxBytes, err := ioutil.ReadFile(vmxPath)
	if err != nil {
		p.logger.ErrorWithDetails("vmx-builder", "reading file: %s", vmxPath)
		return nil, err
	}

	vmxVM := new(VM)
	err = govmx.Unmarshal(vmxBytes, vmxVM)
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

func (p VmxBuilderImpl) writeVmx(vmxVM *VM, vmxPath string) error {
	var err error

	vmxBytes, err := govmx.Marshal(vmxVM)
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
