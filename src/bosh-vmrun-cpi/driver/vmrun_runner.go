package driver

import (
	"fmt"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type VmrunRunnerImpl struct {
	vmrunBinPath string
	logger       boshlog.Logger
	boshRunner   boshsys.CmdRunner
}

func NewVmrunRunner(vmrunBinPath string, boshRunner boshsys.CmdRunner, logger boshlog.Logger) VmrunRunner {
	logger.DebugWithDetails("vmrun-runner", "bin: %+s", vmrunBinPath)

	return &VmrunRunnerImpl{vmrunBinPath: vmrunBinPath, boshRunner: boshRunner, logger: logger}
}

func (c VmrunRunnerImpl) Clone(sourceVmxPath, targetVmxPath, targetVmName string) error {
	args := []string{"clone", sourceVmxPath, targetVmxPath, "linked"}
	flags := map[string]string{"cloneName": targetVmName}

	_, err := c.cliCommand(args, flags)
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
	commandArgs := []string{}
	commandArgs = append(commandArgs, args...)

	for option, value := range flagMap {
		commandArgs = append(commandArgs, fmt.Sprintf("-%s=%s", option, value))
	}

	stdout, _, _, err := c.boshRunner.RunCommand(c.vmrunBinPath, commandArgs...)

	return stdout, err
}
