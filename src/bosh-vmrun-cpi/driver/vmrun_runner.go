package driver

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type vmrunRunnerImpl struct {
	vmrunBinPath     string
	vmrunBackendType string
	retryFileLock    RetryFileLock
	logger           boshlog.Logger
}

func NewVmrunRunner(vmrunBinPath string, retryFileLock RetryFileLock, logger boshlog.Logger) *vmrunRunnerImpl {
	logger.Debug("vmrun-runner", "bin: %+s", vmrunBinPath)

	return &vmrunRunnerImpl{vmrunBinPath: vmrunBinPath, retryFileLock: retryFileLock, logger: logger}
}

func (r *vmrunRunnerImpl) Configure() error {
	stdout, err := r.cliCommand([]string{"list"}, nil)
	if err != nil {
		if strings.Contains(stdout, "VIX_SERVICEPROVIDER_VMWARE_WORKSTATION") {
			r.logger.Debug("vmrun-runner", "Setting runner to use backend type 'player'")
			r.vmrunBackendType = "player"
		} else {
			return err
		}
	}

	_, err = r.cliCommand([]string{"list"}, nil)
	if err != nil {
		return err
	}

	return nil
}

func (r *vmrunRunnerImpl) IsPlayer() bool {
	return r.vmrunBackendType == "player"
}

func (r *vmrunRunnerImpl) Clone(sourceVmxPath, targetVmxPath, targetVmName string) error {
	args := []string{"clone", sourceVmxPath, targetVmxPath, "linked"}
	flags := map[string]string{"cloneName": targetVmName}

	lockFilePath := filepath.Join(filepath.Dir(sourceVmxPath), "cpi-clone.lock")
	cloneMaxWait := 1 * time.Minute

	var err error

	r.retryFileLock.Try(lockFilePath, cloneMaxWait, func() error {
		_, err = r.cliCommand(args, flags)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r *vmrunRunnerImpl) List() (string, error) {
	args := []string{"list"}

	return r.cliCommand(args, nil)
}

func (r *vmrunRunnerImpl) Start(vmxPath string) error {
	args := []string{"start", vmxPath, "nogui"}

	_, err := r.cliCommand(args, nil)
	return err
}

func (r *vmrunRunnerImpl) SoftStop(vmxPath string) error {
	args := []string{"stop", vmxPath, "soft"}

	_, err := r.cliCommand(args, nil)
	return err
}

func (r *vmrunRunnerImpl) HardStop(vmxPath string) error {
	args := []string{"stop", vmxPath, "hard"}

	_, err := r.cliCommand(args, nil)
	return err
}

func (r *vmrunRunnerImpl) Delete(vmxPath string) error {
	args := []string{"deleteVM", vmxPath}

	_, err := r.cliCommand(args, nil)
	return err
}

func (r *vmrunRunnerImpl) CopyFileFromHostToGuest(vmxPath, hostFilePath, guestFilePath, guestUsername, guestPassword string) error {
	args := []string{
		"-gu", guestUsername,
		"-gp", guestPassword,
		"copyFileFromHostToGuest",
		vmxPath,
		hostFilePath,
		guestFilePath,
	}

	_, err := r.cliCommand(args, nil)
	return err
}

func (r *vmrunRunnerImpl) RunProgramInGuest(vmxPath, guestInterpreterPath, guestFilePath, guestUsername, guestPassword string) error {
	args := []string{
		"-gu", guestUsername,
		"-gp", guestPassword,
		"runProgramInGuest",
		vmxPath,
		guestInterpreterPath,
		guestFilePath,
	}

	_, err := r.cliCommand(args, nil)
	return err
}

func (r *vmrunRunnerImpl) ListProcessesInGuest(vmxPath, guestUsername, guestPassword string) (string, error) {
	args := []string{
		"-gu", guestUsername,
		"-gp", guestPassword,
		"listProcessesInGuest",
		vmxPath,
	}

	return r.cliCommand(args, nil)
}

func (r *vmrunRunnerImpl) cliCommand(args []string, flagMap map[string]string) (string, error) {
	var stdout string
	var err error

	commandArgs := []string{}

	if r.vmrunBackendType != "" {
		commandArgs = append(commandArgs, "-T", r.vmrunBackendType)
	}

	commandArgs = append(commandArgs, args...)

	for option, value := range flagMap {
		commandArgs = append(commandArgs, fmt.Sprintf("-%s=%s", option, value))
	}

	commandStr := fmt.Sprintf("%s %s", r.vmrunBinPath, strings.Join(commandArgs, " "))
	r.logger.DebugWithDetails("vmrun-runner", "Running command with args:", commandStr)

	for {
		retryableError := "The operation is not supported for the specified parameters"

		execCmd := newExecCmd(r.vmrunBinPath, commandArgs...)
		stdoutBytes, err := execCmd.Output()
		stdout = string(stdoutBytes)
		if err == nil {
			break
		}

		if strings.Contains(stdout, retryableError) {
			r.logger.Debug("vmrun-runner", "Retryable error: %s: %s (%s)", commandStr, stdout, err.Error())
			continue
		} else {
			return stdout, bosherr.WrapErrorf(err, "Running '%s: %s'", commandStr, stdout)
		}
	}

	r.logger.DebugWithDetails("vmrun-runner", "Command Succeeded:", stdout)

	return stdout, err
}
