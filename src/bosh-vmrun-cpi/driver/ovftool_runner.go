package driver

import (
	"fmt"
	"os"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type OvftoolRunnerImpl struct {
	ovftoolBinPath string
	boshRunner     boshsys.CmdRunner
	logger         boshlog.Logger
}

func NewOvftoolRunner(ovftoolBinPath string, boshRunner boshsys.CmdRunner, logger boshlog.Logger) OvftoolRunner {
	logger.DebugWithDetails("ovftool-runner", "bin: %+s", ovftoolBinPath)

	return &OvftoolRunnerImpl{ovftoolBinPath: ovftoolBinPath, boshRunner: boshRunner, logger: logger}
}

func (r OvftoolRunnerImpl) ImportOvf(ovfPath, vmxPath, vmName string) error {
	var err error
	flags := map[string]string{
		"sourceType":          "OVF",
		"allowAllExtraConfig": "true",
		"allowExtraConfig":    "true",
		"targetType":          "VMX",
		"name":                vmName,
	}

	os.MkdirAll(filepath.Dir(vmxPath), 0700)

	args := []string{ovfPath, vmxPath}

	_, err = r.cliCommand(args, flags)
	if err != nil {
		r.logger.ErrorWithDetails("client", "import ovf: runner", err)
		return err
	}

	return nil
}

func (c OvftoolRunnerImpl) cliCommand(args []string, flagMap map[string]string) (string, error) {
	commandArgs := []string{}
	for option, value := range flagMap {
		commandArgs = append(commandArgs, fmt.Sprintf("--%s=%s", option, value))
	}
	commandArgs = append(commandArgs, args...)

	stdout, _, _, err := c.boshRunner.RunCommand(c.ovftoolBinPath, commandArgs...)

	return stdout, err
}
