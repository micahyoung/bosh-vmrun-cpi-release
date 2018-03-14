package govc

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type GovcClientImpl struct {
	runner GovcRunner
	config GovcConfig
	logger boshlog.Logger
}

func NewClient(runner GovcRunner, config GovcConfig, logger boshlog.Logger) GovcClient {
	return GovcClientImpl{runner: runner, config: config, logger: logger}
}

func (c GovcClientImpl) ImportOvf(ovfPath string) (string, error) {
	flags := map[string]string{
		"u": c.config.EsxUrl,
		"k": "true",
	}
	args := []string{ovfPath}

	result, err := c.runner.CliCommand("import.ovf", flags, args)
	if err != nil {
		return "", err
	}

	return result, nil
}
