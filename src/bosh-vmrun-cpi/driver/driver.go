package driver

import (
	"time"
)

//go:generate counterfeiter -o fakes/fake_client.go driver.go Client
type Client interface {
	ImportOvf(string, string) (bool, error)
	CloneVM(string, string) error
	GetVMIsoPath(string) string
	UpdateVMIso(string, string) error
	StartVM(string) error
	StopVM(string) error
	NeedsVMNameChange(vmName string) bool
	HasVM(string) bool
	SetVMDisplayName(vmName string, displayName string) error
	SetVMNetworkAdapter(string, string, string) error
	SetVMResources(string, int, int) error
	CreateEphemeralDisk(string, int) error
	CreateDisk(string, int) error
	AttachDisk(string, string) error
	DetachDisk(string, string) error
	DestroyDisk(string) error
	HasDisk(string) bool
	DestroyVM(string) error
	GetVMInfo(string) (VMInfo, error)
	BootstrapVM(string, string, string, string, string, string, string, time.Duration, time.Duration) error
}

//go:generate counterfeiter -o fakes/fake_config.go driver.go Config
type Config interface {
	VmxPath(vmName string) string
	EphemeralDiskPath(vmName string) string
	EnvIsoPath(vmName string) string
	PersistentDiskPath(diskId string) string
	OvftoolPath() string
	VmrunPath() string
	VmStartMaxWait() time.Duration
	VmSoftShutdownMaxWait() time.Duration
	EnableHumanReadableName() bool
}

//go:generate counterfeiter -o fakes/fake_retry_file_lock.go driver.go RetryFileLock
type RetryFileLock interface {
	Try(string, time.Duration, func() error) error
}

//go:generate counterfeiter -o fakes/fake_vmrun_runner.go driver.go VmrunRunner
type VmrunRunner interface {
	Configure() error
	IsPlayer() bool
	Clone(sourceVmxPath, targetVmxPath, targetVmName string) error
	List() (string, error)
	Start(string) error
	SoftStop(string) error
	HardStop(string) error
	Delete(string) error
	CopyFileFromHostToGuest(string, string, string, string, string) error
	RunProgramInGuest(string, string, string, string, string) error
	ListProcessesInGuest(string, string, string) (string, error)
}

//go:generate counterfeiter -o fakes/fake_ovftool_runner.go driver.go OvftoolRunner
type OvftoolRunner interface {
	Configure() error
	ImportOvf(string, string, string) error
	Clone(sourceVmxPath, targetVmxPath, targetVmName string) error
	CreateDisk(string, int) error
}

//go:generate counterfeiter -o fakes/fake_clone_runner.go driver.go CloneRunner
type CloneRunner interface {
	Clone(sourceVmxPath, targetVmxPath, targetVmName string) error
}

//TODO: move to vm package
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
