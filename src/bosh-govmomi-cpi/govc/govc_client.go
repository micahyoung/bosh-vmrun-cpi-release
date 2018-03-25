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
	stemcellVmName := stemcellName(stemcellId)
	flags := map[string]string{
		"name": stemcellVmName,
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}
	args := []string{ovfPath}

	result, err := c.runner.CliCommand("import.ovf", flags, args)
	if err != nil {
		c.logger.Error("govc", "import ovf")
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) CloneVM(stemcellId string, vmId string) (string, error) {
	var result string
	var err error

	stemcellVmName := stemcellName(stemcellId)
	cloneVmName := vmName(vmId)

	result, err = c.copyDatastoreStemcell(stemcellVmName, cloneVmName)
	if err != nil {
		c.logger.Error("govc", "copying datastore")
		return result, err
	}

	result, err = c.registerDatastoreVm(stemcellVmName, cloneVmName)
	if err != nil {
		c.logger.Error("govc", "registering VM")
		return result, err
	}

	result, err = c.changeVm(cloneVmName)
	if err != nil {
		c.logger.Error("govc", "changing VM")
		return result, err
	}

	result, err = c.addNetwork(cloneVmName)
	if err != nil {
		c.logger.Error("govc", "adding network")
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) UpdateVMIso(vmId string, localIsoPath string) (string, error) {
	vmName := vmName(vmId)

	datastoreIsoPath := fmt.Sprintf("%s/env.iso", vmName)
	result, err := c.upload(vmName, localIsoPath, datastoreIsoPath)
	if err != nil {
		c.logger.Error("govc", "uploading ENV cdrom")
		return result, err
	}

	result, err = c.insertCdrom(vmName, datastoreIsoPath)
	if err != nil {
		c.logger.Error("govc", "inserting ENV cdrom")
		return result, err
	}

	result, err = c.connectCdrom(vmName)
	if err != nil {
		c.logger.Error("govc", "inserting ENV cdrom")
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) StartVM(vmId string) (string, error) {
	vmName := vmName(vmId)
	var wg sync.WaitGroup

	// continually try to answer start-blocking question
	wg.Add(1)
	go func() {
		for {
			fResult, fErr := c.answerCopyQuestion(vmName)
			if fErr != nil || !strings.Contains(fResult, "No pending question") {
				wg.Done()
				break
			}
		}
	}()

	// blocks until question is answered
	result, err := c.powerOnVm(vmName)
	if err != nil {
		c.logger.Error("govc", "powering on VM")
		return result, err
	}

	wg.Wait()

	return result, nil
}

func (c GovcClientImpl) DestroyVM(vmId string) (string, error) {
	vmName := vmName(vmId)
	result, err := c.stopVM(vmName)
	if err != nil {
		c.logger.Error("govc", "stopping VM")
		return result, err
	}

	result, err = c.destroyVm(vmName)
	if err != nil {
		c.logger.Error("govc", "destroy VM")
		return result, err
	}

	result, err = c.deleteDatastoreDir(vmName)
	if err != nil {
		c.logger.Error("govc", "delete VM files")
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) DestroyStemcell(stemcellId string) (string, error) {
	stemcellVmName := stemcellName(stemcellId)

	result, err := c.destroyVm(stemcellVmName)
	if err != nil {
		c.logger.Error("govc", "destroy Stemcell VM")
		return result, err
	}

	result, err = c.deleteDatastoreDir(stemcellVmName)
	if err != nil {
		c.logger.Error("govc", "delete Stemcell VM files")
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) copyDatastoreStemcell(stemcellVmName string, cloneVmName string) (string, error) {
	flags := map[string]string{
		"u": c.config.EsxUrl(),
		"k": "true",
	}
	args := []string{stemcellVmName, cloneVmName}

	return c.runner.CliCommand("datastore.cp", flags, args)
}

func (c GovcClientImpl) registerDatastoreVm(stemcellVmName string, cloneVmName string) (string, error) {
	vmxPath := fmt.Sprintf("%s/%s.vmx", cloneVmName, stemcellVmName)
	flags := map[string]string{
		"name": cloneVmName,
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}
	args := []string{vmxPath}

	return c.runner.CliCommand("vm.register", flags, args)
}

func (c GovcClientImpl) changeVm(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm":                  cloneVmName,
		"nested-hv-enabled":   "true",
		"sync-time-with-host": "true",
		"u": c.config.EsxUrl(),
		"k": "true",
	}

	return c.runner.CliCommand("vm.change", flags, nil)
}

func (c GovcClientImpl) addNetwork(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm":          cloneVmName,
		"net":         "VM Network",
		"net.adapter": "vmxnet3",
		"u":           c.config.EsxUrl(),
		"k":           "true",
	}

	return c.runner.CliCommand("vm.network.add", flags, nil)
}

func (c GovcClientImpl) upload(cloneVmName string, localPath string, datastorePath string) (string, error) {
	flags := map[string]string{
		"u": c.config.EsxUrl(),
		"k": "true",
	}
	args := []string{localPath, datastorePath}

	return c.runner.CliCommand("datastore.upload", flags, args)
}

func (c GovcClientImpl) insertCdrom(cloneVmName string, datastorePath string) (string, error) {
	flags := map[string]string{
		"vm": cloneVmName,
		"u":  c.config.EsxUrl(),
		"k":  "true",
	}
	args := []string{datastorePath}

	return c.runner.CliCommand("device.cdrom.insert", flags, args)
}

func (c GovcClientImpl) connectCdrom(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm": cloneVmName,
		"u":  c.config.EsxUrl(),
		"k":  "true",
	}
	args := []string{"cdrom-3000"}

	return c.runner.CliCommand("device.connect", flags, args)
}

func (c GovcClientImpl) powerOnVm(cloneVmName string) (string, error) {
	flags := map[string]string{
		"on": "true",
		"u":  c.config.EsxUrl(),
		"k":  "true",
	}
	args := []string{cloneVmName}

	return c.runner.CliCommand("vm.power", flags, args)
}

func (c GovcClientImpl) answerCopyQuestion(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm":     cloneVmName,
		"answer": "2",
		"u":      c.config.EsxUrl(),
		"k":      "true",
	}

	return c.runner.CliCommand("vm.question", flags, nil)
}

func (c GovcClientImpl) stopVM(cloneVmName string) (string, error) {
	flags := map[string]string{
		"off":   "true",
		"force": "true",
		"u":     c.config.EsxUrl(),
		"k":     "true",
	}
	args := []string{cloneVmName}

	return c.runner.CliCommand("vm.power", flags, args)
}

func (c GovcClientImpl) destroyVm(vmName string) (string, error) {
	flags := map[string]string{
		"u": c.config.EsxUrl(),
		"k": "true",
	}
	args := []string{vmName}

	return c.runner.CliCommand("vm.destroy", flags, args)
}

func (c GovcClientImpl) deleteDatastoreDir(datastorePath string) (string, error) {
	flags := map[string]string{
		"u": c.config.EsxUrl(),
		"k": "true",
	}
	args := []string{datastorePath}

	return c.runner.CliCommand("datastore.rm", flags, args)
}

func vmName(vmId string) string {
	return "vm-" + vmId
}

func stemcellName(stemcellId string) string {
	return "cs-" + stemcellId
}
