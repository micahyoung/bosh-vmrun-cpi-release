package vmx

import (
	govmx "github.com/hooklift/govmx"
)

type VM struct {
	// Disable swap: https://kb.vmware.com/s/article/1008885
	MinVmMemPct                 int    `vmx:"prefvmx.minVmMemPct"`
	MemTrimRate                 int    `vmx:"MemTrimRate"`
	UseNamedFile                bool   `vmx:"mainMem.useNamedFile"`
	Pshare                      bool   `vmx:"sched.mem.pshare.enable"`
	UseRecommendedLockedMemSize bool   `vmx:"prefvmx.useRecommendedLockedMemSize"`
	MainmemBacking              string `vmx:"mainmem.backing"`

	govmx.VirtualMachine
}
