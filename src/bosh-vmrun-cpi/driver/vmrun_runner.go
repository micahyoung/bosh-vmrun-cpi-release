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

func (c VmrunRunnerImpl) CliCommand(args []string, flagMap map[string]string) (string, error) {
	commandArgs := []string{}
	commandArgs = append(commandArgs, args...)

	for option, value := range flagMap {
		commandArgs = append(commandArgs, fmt.Sprintf("-%s=%s", option, value))
	}

	stdout, _, _, err := c.boshRunner.RunCommand(c.vmrunBinPath, commandArgs...)

	return stdout, err
}
