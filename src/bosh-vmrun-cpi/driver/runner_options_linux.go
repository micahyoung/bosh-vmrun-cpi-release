package driver

import (
	"syscall"
)

func runnerOptions() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
