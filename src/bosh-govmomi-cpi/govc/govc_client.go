package govc

import (
	"fmt"
	"strings"
	"sync"

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

func (c GovcClientImpl) ImportOvf(ovfPath string, stemcellId string) (string, error) {
	stemcellVmName := "cs-" + stemcellId
	flags := map[string]string{
		"name": stemcellVmName,
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}
	args := []string{ovfPath}

	result, err := c.runner.CliCommand("import.ovf", flags, args)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) CloneVM(stemcellId string, vmId string) (string, error) {
	var result string
	var err error
	var wg sync.WaitGroup

	stemcellVmName := "cs-" + stemcellId
	cloneVmName := "vm-" + vmId

	result, err = c.copyDatastoreStemcell(stemcellVmName, cloneVmName)
	if err != nil {
		return result, err
	}

	result, err = c.registerDatastoreVm(stemcellVmName, cloneVmName)
	if err != nil {
		return result, err
	}

	result, err = c.changeVm(cloneVmName)
	if err != nil {
		return result, err
	}

	result, err = c.addNetwork(cloneVmName)
	if err != nil {
		return result, err
	}

	result, err = c.addEnvCdrom(cloneVmName)
	if err != nil {
		return result, err
	}

	// continually try to answer start-blocking question
	wg.Add(1)
	go func() {
		for {
			fResult, fErr := c.answerCopyQuestion(cloneVmName)
			if fErr != nil || !strings.Contains(fResult, "No pending question") {
				wg.Done()
				break
			}
		}
	}()

	// blocks until question is answered
	result, err = c.powerOnVm(cloneVmName)
	if err != nil {
		return result, err
	}

	wg.Wait()

	return result, nil
}

func (c GovcClientImpl) copyDatastoreStemcell(stemcellVmName string, cloneVmName string) (string, error) {
	flags := map[string]string{
		"u": c.config.EsxUrl(),
		"k": "true",
	}
	args := []string{stemcellVmName, cloneVmName}

	result, err := c.runner.CliCommand("datastore.cp", flags, args)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) registerDatastoreVm(stemcellVmName string, cloneVmName string) (string, error) {
	vmxPath := fmt.Sprintf("%s/%s.vmx", cloneVmName, stemcellVmName)
	flags := map[string]string{
		"name": cloneVmName,
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}
	args := []string{vmxPath}

	result, err := c.runner.CliCommand("vm.register", flags, args)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) changeVm(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm":                  cloneVmName,
		"nested-hv-enabled":   "true",
		"sync-time-with-host": "true",
		"u": c.config.EsxUrl(),
		"k": "true",
	}

	result, err := c.runner.CliCommand("vm.change", flags, nil)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) addNetwork(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm":          cloneVmName,
		"net":         "VM Network",
		"net.adapter": "vmxnet3",
		"u":           c.config.EsxUrl(),
		"k":           "true",
	}

	result, err := c.runner.CliCommand("vm.network.add", flags, nil)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) addCdrom(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm":          cloneVmName,
		"net":         "VM Network",
		"net.adapter": "vmxnet3",
		"u":           c.config.EsxUrl(),
		"k":           "true",
	}

	result, err := c.runner.CliCommand("vm.network.add", flags, nil)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) powerOnVm(cloneVmName string) (string, error) {
	flags := map[string]string{
		"on": "true",
		"u":  c.config.EsxUrl(),
		"k":  "true",
	}
	args := []string{cloneVmName}

	result, err := c.runner.CliCommand("vm.power", flags, args)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) answerCopyQuestion(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm":     cloneVmName,
		"answer": "2",
		"u":      c.config.EsxUrl(),
		"k":      "true",
	}

	result, err := c.runner.CliCommand("vm.question", flags, nil)
	if err != nil {
		return result, err
	}

	return result, nil
}
