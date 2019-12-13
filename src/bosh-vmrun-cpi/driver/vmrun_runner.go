package driver

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type VmrunRunnerImpl struct {
	vmrunBinPath     string
	vmrunBackendType string
	retryFileLock    RetryFileLock
	logger           boshlog.Logger
}

func NewVmrunRunner(vmrunBinPath string, vmrunBackendType string, retryFileLock RetryFileLock, logger boshlog.Logger) VmrunRunner {
	logger.Debug("vmrun-runner", "bin: %+s", vmrunBinPath)

	return &VmrunRunnerImpl{vmrunBinPath, vmrunBackendType, retryFileLock, logger}
}

func (c VmrunRunnerImpl) Clone(sourceVmxPath, targetVmxPath, targetVmName string) error {
	args := []string{"clone", sourceVmxPath, targetVmxPath, "linked"}
	flags := map[string]string{"cloneName": targetVmName}

	lockFilePath := filepath.Join(filepath.Dir(sourceVmxPath), "cpi-clone.lock")
	cloneMaxWait := 1 * time.Minute

	var err error

	c.retryFileLock.Try(lockFilePath, cloneMaxWait, func() error {
		_, err = c.cliCommand(args, flags)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (c VmrunRunnerImpl) List() (string, error) {
	args := []string{"list"}

	return c.cliCommand(args, nil)
}

func (c VmrunRunnerImpl) Start(vmxPath string) error {
	args := []string{"start", vmxPath, "nogui"}

	_, err := c.cliCommand(args, nil)
	return err
}

func (c VmrunRunnerImpl) SoftStop(vmxPath string) error {
	args := []string{"stop", vmxPath, "soft"}

	_, err := c.cliCommand(args, nil)
	return err
}

func (c VmrunRunnerImpl) HardStop(vmxPath string) error {
	args := []string{"stop", vmxPath, "hard"}

	_, err := c.cliCommand(args, nil)
	return err
}

func (c VmrunRunnerImpl) Delete(vmxPath string) error {
	args := []string{"deleteVM", vmxPath}

	_, err := c.cliCommand(args, nil)
	return err
}

func (c VmrunRunnerImpl) CopyFileFromHostToGuest(vmxPath, hostFilePath, guestFilePath, guestUsername, guestPassword string) error {
	args := []string{
		"-gu", guestUsername,
		"-gp", guestPassword,
		"copyFileFromHostToGuest",
		vmxPath,
		hostFilePath,
		guestFilePath,
	}

	_, err := c.cliCommand(args, nil)
	return err
}

func (c VmrunRunnerImpl) RunProgramInGuest(vmxPath, guestInterpreterPath, guestFilePath, guestUsername, guestPassword string) error {
	args := []string{
		"-gu", guestUsername,
		"-gp", guestPassword,
		"runProgramInGuest",
		vmxPath,
		guestInterpreterPath,
		guestFilePath,
	}

	_, err := c.cliCommand(args, nil)
	return err
}

func (c VmrunRunnerImpl) ListProcessesInGuest(vmxPath, guestUsername, guestPassword string) (string, error) {
	args := []string{
		"-gu", guestUsername,
		"-gp", guestPassword,
		"listProcessesInGuest",
		vmxPath,
	}

	return c.cliCommand(args, nil)
}

func (c VmrunRunnerImpl) cliCommand(args []string, flagMap map[string]string) (string, error) {
	var stdout string
	var err error

	commandArgs := []string{}

	if c.vmrunBackendType != "" {
		commandArgs = append(commandArgs, "-T", c.vmrunBackendType)
	}

	commandArgs = append(commandArgs, args...)

	for option, value := range flagMap {
		commandArgs = append(commandArgs, fmt.Sprintf("-%s=%s", option, value))
	}

	commandStr := fmt.Sprintf("%s %s", c.vmrunBinPath, strings.Join(commandArgs, " "))
	c.logger.DebugWithDetails("vmrun-runner", "Running command with args:", commandStr)

	for {
		retryableError := "Unable to connect to host"

		execCmd := newExecCmd(c.vmrunBinPath, commandArgs...)
		stdoutBytes, err := execCmd.Output()
		stdout = string(stdoutBytes)
		if err == nil {
			break
		}

		if strings.Contains(stdout, retryableError) {
			c.logger.Debug("vmrun-runner", "Retryable error: %s: %s (%s)", commandStr, stdout, err.Error())
			continue
		} else {
			return stdout, bosherr.WrapErrorf(err, "Running '%s: %s'", commandStr, stdout)
		}
	}

	c.logger.DebugWithDetails("vmrun-runner", "Command Succeeded:", stdout)

	return stdout, err
}
