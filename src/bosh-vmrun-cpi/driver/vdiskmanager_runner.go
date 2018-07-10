package driver

import (
	"fmt"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type VdiskmanagerRunnerImpl struct {
	vmdiskmanagerBinPath string
	logger               boshlog.Logger
	boshRunner           boshsys.CmdRunner
}

func NewVdiskmanagerRunner(vmdiskmanagerBinPath string, boshRunner boshsys.CmdRunner, logger boshlog.Logger) VdiskmanagerRunner {
	logger.DebugWithDetails("vdiskmanager-runner", "bin: %+s", vmdiskmanagerBinPath)

	return VdiskmanagerRunnerImpl{vmdiskmanagerBinPath: vmdiskmanagerBinPath, boshRunner: boshRunner, logger: logger}
}

func (p VdiskmanagerRunnerImpl) CreateDisk(diskPath string, diskMB int) error {
	var err error

	_, err = p.run([]string{"-c", diskPath}, map[string]string{
		"s": fmt.Sprintf("%dMB", diskMB),
		"t": "0", //single growable virtual disk
	})
	if err != nil {
		return err
	}

	return nil
}

func (c VdiskmanagerRunnerImpl) run(args []string, flagMap map[string]string) (string, error) {
	commandArgs := []string{}
	for option, value := range flagMap {
		commandArgs = append(commandArgs, fmt.Sprintf("-%s %s", option, value))
	}
	commandArgs = append(commandArgs, args...)

	stdout, _, _, err := c.boshRunner.RunCommand(c.vmdiskmanagerBinPath, commandArgs...)

	return stdout, err
}
