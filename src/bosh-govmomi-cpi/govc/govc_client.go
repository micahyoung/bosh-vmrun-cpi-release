package govc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type GovcClientImpl struct {
	runner GovcRunner
	config GovcConfig
	logger boshlog.Logger
}

var (
	STATE_NOT_FOUND         = "state-not-found"
	STATE_POWER_ON          = "state-on"
	STATE_POWER_OFF         = "state-off"
	STATE_BLOCKING_QUESTION = "state-blocking-question"
)

func NewClient(runner GovcRunner, config GovcConfig, logger boshlog.Logger) GovcClient {
	return GovcClientImpl{runner: runner, config: config, logger: logger}
}

func (c GovcClientImpl) ImportOvf(ovfPath string, vmName string) (string, error) {
	flags := map[string]string{
		"name": vmName,
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}
	args := []string{ovfPath}

	result, err := c.runner.CliCommand("import.ovf", flags, args)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "import ovf", err, result)
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) CloneVM(sourceVmName string, cloneVmName string) (string, error) {
	var result string
	var err error

	result, err = c.copyDatastoreStemcell(sourceVmName, cloneVmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "copying datastore", err, result)
		return result, err
	}

	result, err = c.registerDatastoreVm(sourceVmName, cloneVmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "registering VM", err, result)
		return result, err
	}

	result, err = c.configVMHardware(cloneVmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "configuring vm hardware", err, result)
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) SetVMNetworkAdapters(vmName string, adapterCount int) error {
	for i := 0; i < adapterCount; i++ {
		result, err := c.addNetwork(vmName)
		if err != nil {
			c.logger.ErrorWithDetails("govc", "adding network", err, result)
			return err
		}
	}

	return nil
}

func (c GovcClientImpl) SetVMResources(vmName string, cpus int, ram int) error {
	result, err := c.setVMResources(vmName, cpus, ram)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "setting vm cpu and ram", err, result)
		return err
	}
	return nil
}

func (c GovcClientImpl) UpdateVMIso(vmName string, localIsoPath string) (string, error) {
	result, err := c.disconnectCdrom(vmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "connecting ENV cdrom", err, result)
		return result, err
	}

	result, err = c.ejectCdrom(vmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "ejecting ENV cdrom", err, result)
		return result, err
	}

	datastoreIsoPath := fmt.Sprintf("/env/env-%s.iso", vmName)
	result, err = c.upload(vmName, localIsoPath, datastoreIsoPath)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "uploading ENV cdrom", err, result)
		return result, err
	}

	result, err = c.insertCdrom(vmName, datastoreIsoPath)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "inserting ENV cdrom", err, result)
		return result, err
	}

	result, err = c.connectCdrom(vmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "connecting ENV cdrom", err, result)
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) StartVM(vmName string) (string, error) {
	go func() {
		// blocks until question is answered
		result, err := c.powerOnVm(vmName)
		if err != nil {
			c.logger.ErrorWithDetails("govc", "powering on VM", err, result)
			return
		}
	}()

	// continually wait for then answer start-blocking question
	for {
		time.Sleep(300 * time.Millisecond)

		vmState, err := c.vmState(vmName)
		if err != nil {
			c.logger.ErrorWithDetails("govc", "fetching question state for VM", err, vmState)
			return "", err
		}

		if vmState == STATE_BLOCKING_QUESTION {
			break
		}
	}

	result, err := c.answerCopyQuestion(vmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "answering question state for VM", err, result)
		return "", err
	}

	return "success", err
}

func (c GovcClientImpl) HasVM(vmName string) (bool, error) {
	result, err := c.vmState(vmName)
	found := (result != STATE_NOT_FOUND)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "HasVM", err, result)
		return false, err
	}
	return found, nil
}

func (c GovcClientImpl) CreateEphemeralDisk(vmName string, diskMB int) error {
	result, err := c.createEphemeralDisk(vmName, diskMB)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "CreateEphemeralDisk", err, result)
		return err
	}
	return nil
}

func (c GovcClientImpl) CreateDisk(diskId string, diskMB int) error {
	result, err := c.createDisk(diskId, diskMB)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "CreateDisk", err, result)
		return err
	}
	return nil
}

func (c GovcClientImpl) AttachDisk(vmName string, diskId string) error {
	result, err := c.attachDisk(vmName, diskId)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "AttachDisk", err, result)
		return err
	}
	return nil
}

func (c GovcClientImpl) DetachDisk(vmName string, diskId string) error {
	diskDeviceName, err := c.getVMDiskName(vmName, diskId)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "getVMDiskName", err, diskDeviceName)
		return err
	}

	result, err := c.detachDisk(vmName, diskDeviceName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "DetachDisk", err, result)
		return err
	}
	return nil
}

func (c GovcClientImpl) DestroyDisk(diskName string) error {
	diskPath := fmt.Sprintf(`%s.vmdk`, diskName)
	pathFound, err := c.datastorePathExists(diskPath)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "finding Path", err, pathFound)
		return err
	}

	if pathFound {
		result, err := c.deleteDatastoreObject(diskPath)
		if err != nil {
			c.logger.ErrorWithDetails("govc", "delete VM files", err, result)
			return err
		}
	}

	return nil
}

func (c GovcClientImpl) DestroyVM(vmName string) (string, error) {
	var result string
	var err error

	vmState, err := c.vmState(vmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "getting state for VM", err, vmState)
		return result, err
	}

	if vmState == STATE_POWER_ON {
		result, err = c.stopVM(vmName)
		if err != nil {
			c.logger.ErrorWithDetails("govc", "stopping VM", err, result)
			return result, err
		}
	}

	if vmState != STATE_NOT_FOUND {
		result, err = c.destroyVm(vmName)
		if err != nil {
			c.logger.ErrorWithDetails("govc", "destroy VM", err, result)
			return result, err
		}
	}

	pathFound, err := c.datastorePathExists(vmName)
	if err != nil {
		c.logger.ErrorWithDetails("govc", "finding Path", err, pathFound)
		return result, err
	}

	if pathFound {
		result, err = c.deleteDatastoreObject(vmName)
		if err != nil {
			c.logger.ErrorWithDetails("govc", "delete VM files", err, result)
			return result, err
		}
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

func (c GovcClientImpl) configVMHardware(cloneVmName string) (string, error) {
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

func (c GovcClientImpl) setVMResources(vmName string, cpuCount int, ramMB int) (string, error) {
	cpu := strconv.Itoa(cpuCount)
	mem := strconv.Itoa(ramMB)
	flags := map[string]string{
		"vm": vmName,
		"c":  cpu,
		"m":  mem,
		"u":  c.config.EsxUrl(),
		"k":  "true",
	}

	return c.runner.CliCommand("vm.change", flags, nil)
}

func (c GovcClientImpl) ejectCdrom(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm": cloneVmName,
		"u":  c.config.EsxUrl(),
		"k":  "true",
	}

	return c.runner.CliCommand("device.cdrom.eject", flags, nil)
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

func (c GovcClientImpl) disconnectCdrom(cloneVmName string) (string, error) {
	flags := map[string]string{
		"vm": cloneVmName,
		"u":  c.config.EsxUrl(),
		"k":  "true",
	}
	args := []string{"cdrom-3000"}

	return c.runner.CliCommand("device.disconnect", flags, args)
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
		"off": "true",
		"u":   c.config.EsxUrl(),
		"k":   "true",
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

func (c GovcClientImpl) deleteDatastoreObject(datastorePath string) (string, error) {
	flags := map[string]string{
		"f": "true",
		"u": c.config.EsxUrl(),
		"k": "true",
	}
	args := []string{datastorePath}

	return c.runner.CliCommand("datastore.rm", flags, args)
}

func (c GovcClientImpl) vmState(vmName string) (string, error) {
	flags := map[string]string{
		"u": c.config.EsxUrl(),
		"k": "true",
	}
	args := []string{vmName}

	result, err := c.runner.CliCommand("vm.info", flags, args)
	if err != nil {
		return result, err
	}

	var response struct {
		VirtualMachines []struct {
			Runtime struct {
				PowerState string
				Question   interface{}
			}
		}
	}
	err = json.Unmarshal([]byte(result), &response)

	if err != nil {
		return result, err
	}

	if len(response.VirtualMachines) == 0 {
		return STATE_NOT_FOUND, nil
	}

	if response.VirtualMachines[0].Runtime.Question != nil {
		return STATE_BLOCKING_QUESTION, nil
	}

	if response.VirtualMachines[0].Runtime.PowerState == "poweredOn" {
		return STATE_POWER_ON, nil
	}

	return STATE_POWER_OFF, nil
}

func (c GovcClientImpl) datastorePathExists(datastorePath string) (bool, error) {
	flags := map[string]string{
		"u": c.config.EsxUrl(),
		"k": "true",
	}

	result, err := c.runner.CliCommand("datastore.ls", flags, nil)
	if err != nil {
		return false, err
	}

	var response []struct{ File []struct{ Path string } }
	err = json.Unmarshal([]byte(result), &response)
	if err != nil {
		return false, fmt.Errorf("error: %+v\nresult: %s\n", err, result)
	}

	files := response[0].File
	found := false
	for i := range files {
		file := files[i].Path
		if file == datastorePath {
			found = true
			break
		}
	}

	return found, nil
}

func (c GovcClientImpl) createDisk(diskId string, diskMB int) (string, error) {
	diskPath := fmt.Sprintf(`%s.vmdk`, diskId)
	diskSize := fmt.Sprintf(`%dMB`, diskMB)
	flags := map[string]string{
		"size": diskSize,
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}
	args := []string{diskPath}

	result, err := c.runner.CliCommand("datastore.disk.create", flags, args)
	if err != nil {
		return result, err
	}

	return result, err
}

func (c GovcClientImpl) createEphemeralDisk(vmName string, diskMB int) (string, error) {
	diskPath := fmt.Sprintf(`%s/ephemeral.vmdk`, vmName)
	diskSize := fmt.Sprintf(`%dMB`, diskMB)
	flags := map[string]string{
		"vm":   vmName,
		"name": diskPath,
		"size": diskSize,
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}

	result, err := c.runner.CliCommand("vm.disk.create", flags, nil)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) attachDisk(vmName string, diskId string) (string, error) {
	diskPath := fmt.Sprintf(`%s.vmdk`, diskId)
	flags := map[string]string{
		"vm":   vmName,
		"disk": diskPath,
		"link": "true",
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}

	result, err := c.runner.CliCommand("vm.disk.attach", flags, nil)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c GovcClientImpl) getVMDiskName(vmName string, diskId string) (string, error) {
	flags := map[string]string{
		"json": "true",
		"vm":   vmName,
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}

	result, err := c.runner.CliCommand("device.info", flags, nil)
	if err != nil {
		return result, err
	}

	var response struct {
		Devices []struct {
			Name    string
			Backing struct {
				Parent struct {
					FileName string
				}
			}
		}
	}
	err = json.Unmarshal([]byte(result), &response)
	if err != nil {
		return result, err
	}

	foundDevice := ""
	for _, device := range response.Devices {
		if strings.Contains(device.Backing.Parent.FileName, diskId) {
			foundDevice = device.Name
		}
	}

	return foundDevice, nil
}

func (c GovcClientImpl) detachDisk(vmName string, diskName string) (string, error) {
	flags := map[string]string{
		"vm":   vmName,
		"keep": "true",
		"u":    c.config.EsxUrl(),
		"k":    "true",
	}
	args := []string{diskName}

	result, err := c.runner.CliCommand("device.remove", flags, args)
	if err != nil {
		return result, err
	}

	return result, nil
}
