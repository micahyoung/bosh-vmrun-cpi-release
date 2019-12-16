package driver

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type ovftoolRunnerImpl struct {
	ovftoolBinPath string
	boshRunner     boshsys.CmdRunner
	logger         boshlog.Logger
}

func NewOvftoolRunner(ovftoolBinPath string, boshRunner boshsys.CmdRunner, logger boshlog.Logger) *ovftoolRunnerImpl {
	logger.Debug("ovftool-runner", "bin: %+s", ovftoolBinPath)

	return &ovftoolRunnerImpl{ovftoolBinPath: ovftoolBinPath, boshRunner: boshRunner, logger: logger}
}

func (r *ovftoolRunnerImpl) Configure() error {
	_, err := r.cliCommand([]string{"-v"}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *ovftoolRunnerImpl) ImportOvf(ovfPath, vmxPath, vmName string) error {
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
		r.logger.ErrorWithDetails("ovftool runner", "import ovf", err)
		return err
	}

	return nil
}

func (r *ovftoolRunnerImpl) Clone(sourceVmxPath, targetVmxPath, targetVmName string) error {
	var err error
	flags := map[string]string{
		"sourceType":          "VMX",
		"allowAllExtraConfig": "true",
		"allowExtraConfig":    "true",
		"targetType":          "VMX",
		"name":                targetVmName,
	}

	os.MkdirAll(filepath.Dir(targetVmxPath), 0700)

	args := []string{sourceVmxPath, targetVmxPath}

	_, err = r.cliCommand(args, flags)
	if err != nil {
		r.logger.ErrorWithDetails("ovftool runner", "clone", err)
		return err
	}

	return nil
}

func (r *ovftoolRunnerImpl) CreateDisk(diskPath string, diskMB int) error {
	var err error
	var outputDirPath string
	var inputOvfPath string
	var generatedVmxPath string
	var generatedDiskPath string
	var generatedDiskFile *os.File
	var outputDiskFile *os.File

	outputDirPath, err = ioutil.TempDir("", "ovf-disk-generate")
	if err != nil {
		return err
	}
	defer os.RemoveAll(outputDirPath)

	inputOvfPath = filepath.Join(outputDirPath, "makedisk.ovf")
	generatedVmxPath = filepath.Join(outputDirPath, "makedisk.vmx")
	generatedDiskPath = filepath.Join(outputDirPath, "makedisk-disk1.vmdk")

	var inputOvfContent bytes.Buffer
	err = tmpl.Execute(&inputOvfContent, struct{ SizeMB int }{diskMB})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(inputOvfPath, inputOvfContent.Bytes(), 0777)
	if err != nil {
		return err
	}

	flags := map[string]string{
		"sourceType": "OVF",
		"targetType": "VMX",
		"name":       "makedisk",
	}

	args := []string{inputOvfPath, generatedVmxPath}

	_, err = r.cliCommand(args, flags)
	if err != nil {
		r.logger.ErrorWithDetails("ovftool runner", "create disk", err)
		return err
	}

	generatedDiskFile, err = os.Open(generatedDiskPath)
	if err != nil {
		return err
	}

	outputDiskFile, err = os.Create(diskPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(outputDiskFile, generatedDiskFile)
	if err != nil {
		return err
	}

	return nil
}

func (r *ovftoolRunnerImpl) cliCommand(args []string, flagMap map[string]string) (string, error) {
	commandArgs := []string{}
	for option, value := range flagMap {
		commandArgs = append(commandArgs, fmt.Sprintf("--%s=%s", option, value))
	}
	commandArgs = append(commandArgs, args...)

	stdout, _, _, err := r.boshRunner.RunCommand(r.ovftoolBinPath, commandArgs...)

	return stdout, err
}

//TODO: investigate linking to parent disks https://www.vmware.com/pdf/ovf_spec_draft.pdf (Section 10.1)
var tmpl = template.Must(template.New("tmpl").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<Envelope vmw:buildId="build-1312298" xmlns="http://schemas.dmtf.org/ovf/envelope/1" xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vmw="http://www.vmware.com/schema/ovf" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <References>
  </References>
  <DiskSection>
    <Info>Virtual disk information</Info>
    <Disk ovf:capacity="${generatedDiskSizeMB}" ovf:capacityAllocationUnits="byte * 2^20" ovf:diskId="vmdisk1" ovf:format="http://www.vmware.com/specifications/vmdk.html#sparse" ovf:populatedSize="0"/>
  </DiskSection>
  <VirtualSystem ovf:id="vsan-esxi-1">
	<Info>A virtual machine</Info>
	<Name>vsan-esxi-1</Name>
	<ProductSection>
      <Info>Generated Disk Size</Info>
      <Property ovf:key="generatedDiskSizeMB"
		ovf:runtimeConfigurable="false"	
		ovf:type="int" 
		ovf:qualifiers="MinValue(1) MaxValue(10000000)"
		ovf:value="{{.SizeMB}}"
		ovf:userConfigurable="true">
        <Label>Generated Disk Size</Label>
        <Description>The size of the disk in gigabytes.</Description>
      </Property>
    </ProductSection>    
    <OperatingSystemSection ovf:id="104" ovf:version="5" vmw:osType="vmkernel5Guest">
      <Info>The kind of installed guest operating system</Info>
      <Description>VMware ESXi 5.x</Description>
    </OperatingSystemSection>
    <VirtualHardwareSection>
      <Info>Virtual hardware requirements</Info>
      <System>
        <vssd:ElementName>Virtual Hardware Family</vssd:ElementName>
        <vssd:InstanceID>0</vssd:InstanceID>
        <vssd:VirtualSystemIdentifier>vsan-esxi-1</vssd:VirtualSystemIdentifier>
        <vssd:VirtualSystemType>vmx-09</vssd:VirtualSystemType>
      </System>
      <Item>
        <rasd:AllocationUnits>hertz * 10^6</rasd:AllocationUnits>
        <rasd:Description>Number of Virtual CPUs</rasd:Description>
        <rasd:ElementName>2 virtual CPU(s)</rasd:ElementName>
        <rasd:InstanceID>1</rasd:InstanceID>
        <rasd:ResourceType>3</rasd:ResourceType>
        <rasd:VirtualQuantity>2</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:AllocationUnits>byte * 2^20</rasd:AllocationUnits>
        <rasd:Description>Memory Size</rasd:Description>
        <rasd:ElementName>5120MB of memory</rasd:ElementName>
        <rasd:InstanceID>2</rasd:InstanceID>
        <rasd:ResourceType>4</rasd:ResourceType>
        <rasd:VirtualQuantity>5120</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:Address>0</rasd:Address>
        <rasd:Description>SCSI Controller</rasd:Description>
        <rasd:ElementName>SCSI controller 0</rasd:ElementName>
        <rasd:InstanceID>3</rasd:InstanceID>
        <rasd:ResourceSubType>lsilogic</rasd:ResourceSubType>
        <rasd:ResourceType>6</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>0</rasd:AddressOnParent>
        <rasd:ElementName>Hard disk 1</rasd:ElementName>
        <rasd:HostResource>ovf:/disk/vmdisk1</rasd:HostResource>
        <rasd:InstanceID>9</rasd:InstanceID>
        <rasd:Parent>3</rasd:Parent>
        <rasd:ResourceType>17</rasd:ResourceType>
        <vmw:Config ovf:required="false" vmw:key="backing.writeThrough" vmw:value="false"/>
      </Item>
    </VirtualHardwareSection>
  </VirtualSystem>
</Envelope>
`))
