package driver

import "github.com/hooklift/govmx"

//go:generate counterfeiter -o fakes/fake_client.go $GOPATH/src/bosh-vmrun-cpi/driver/driver.go Client
type Client interface {
	ImportOvf(string, string) (bool, error)
	CloneVM(string, string) error
	UpdateVMIso(string, string) error
	StartVM(string) error
	StopVM(string) error
	HasVM(string) bool
	SetVMNetworkAdapter(string, string, string) error
	SetVMResources(string, int, int) error
	CreateEphemeralDisk(string, int) error
	CreateDisk(string, int) error
	AttachDisk(string, string) error
	DetachDisk(string, string) error
	DestroyDisk(string) error
	DestroyVM(string) error
	GetVMInfo(string) (VMInfo, error)
	BootstrapVM(string) error
}

//go:generate counterfeiter -o fakes/fake_config.go $GOPATH/src/bosh-vmrun-cpi/driver/driver.go Config
type Config interface {
	OvftoolPath() string
	VmrunPath() string
	VdiskmanagerPath() string
	VmPath() string
	BootstrapScriptPath() string
	BootstrapScriptContent() string
	BootstrapInterpreterPath() string
	BootstrapUsername() string
	BootstrapPassword() string
	NeedsBootstrap() bool
}

//go:generate counterfeiter -o fakes/fake_vmrun_runner.go $GOPATH/src/bosh-vmrun-cpi/driver/driver.go VmrunRunner
type VmrunRunner interface {
	CliCommand([]string, map[string]string) (string, error)
}

//go:generate counterfeiter -o fakes/fake_ovftool_runner.go $GOPATH/src/bosh-vmrun-cpi/driver/driver.go OvftoolRunner
type OvftoolRunner interface {
	CliCommand([]string, map[string]string) (string, error)
}

//go:generate counterfeiter -o fakes/fake_vdiskmanager_runner.go $GOPATH/src/bosh-vmrun-cpi/driver/driver.go VdiskmanagerRunner
type VdiskmanagerRunner interface {
	CreateDisk(string, int) error
}

//go:generate counterfeiter -o fakes/fake_vmx_builder.go $GOPATH/src/bosh-vmrun-cpi/driver/driver.go VmxBuilder
type VmxBuilder interface {
	InitHardware(string) error
	AddNetworkInterface(string, string, string) error
	SetVMResources(int, int, string) error
	AttachDisk(string, string) error
	DetachDisk(string, string) error
	AttachCdrom(string, string) error
	VMInfo(string) (VMInfo, error)
	GetVmx(string) (*vmx.VirtualMachine, error)
}

type VMInfo struct {
	Name string
	CPUs int
	RAM  int
	NICs []struct {
		Network string
		MAC     string
	}
	Disks []struct {
		ID   string
		Path string
	}
	CleanShutdown bool
}
