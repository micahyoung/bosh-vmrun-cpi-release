package driver

import (
	"os/exec"
	"syscall"
)

const (
	// missing from syscall/types_windows.go
	// https://docs.microsoft.com/en-us/windows/desktop/ProcThread/process-creation-flags
	CREATE_BREAKAWAY_FROM_JOB = 0x01000000
)

func newExecCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)

	// required when running through Win32-OpenSSH
	// - child process of vmrun (vmware-vmx) were put in SSH JobObject and would be shut down when session was closed
	// - ref: https://github.com/PowerShell/Win32-OpenSSH/issues/1032
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: CREATE_BREAKAWAY_FROM_JOB}

	return cmd
}
