package driver

import (
	"syscall"
)

func runnerOptions() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: 0x01000000}
}
