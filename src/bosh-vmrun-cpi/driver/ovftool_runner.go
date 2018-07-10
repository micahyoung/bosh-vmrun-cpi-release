package driver

import (
	"fmt"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type OvftoolRunnerImpl struct {
	ovftoolBinPath string
	logger         boshlog.Logger
	boshRunner     boshsys.CmdRunner
}

func NewOvftoolRunner(ovftoolBinPath string, boshRunner boshsys.CmdRunner, logger boshlog.Logger) OvftoolRunner {
	logger.DebugWithDetails("ovftool-runner", "bin: %+s", ovftoolBinPath)

	return &OvftoolRunnerImpl{ovftoolBinPath: ovftoolBinPath, boshRunner: boshRunner, logger: logger}
}

func (c OvftoolRunnerImpl) CliCommand(args []string, flagMap map[string]string) (string, error) {
	commandArgs := []string{}
	for option, value := range flagMap {
		commandArgs = append(commandArgs, fmt.Sprintf("--%s=%s", option, value))
	}
	commandArgs = append(commandArgs, args...)

	stdout, _, _, err := c.boshRunner.RunCommand(c.ovftoolBinPath, commandArgs...)

	return stdout, err
}
