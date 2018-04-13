package action

import (
	"fmt"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-govmomi-cpi/config"
	"bosh-govmomi-cpi/govc"
	"bosh-govmomi-cpi/stemcell"
	"bosh-govmomi-cpi/vm"
)

type Factory struct {
	govcClient      govc.GovcClient
	stemcellClient  stemcell.StemcellClient
	agentSettings   vm.AgentSettings
	agentEnvFactory apiv1.AgentEnvFactory
	config          config.Config
	fs              boshsys.FileSystem
	uuidGen         boshuuid.Generator
	logger          boshlog.Logger
}

type CPI struct {
	CreateStemcellMethod
	DeleteStemcellMethod
	CreateVMMethod
	DeleteVMMethod
	HasVMMethod
	CreateDiskMethod
	AttachDiskMethod
	DetachDiskMethod
	DeleteDiskMethod
	MiscMethod
}

var _ apiv1.CPIFactory = Factory{}
var _ apiv1.CPI = CPI{}

func NewFactory(
	govcClient govc.GovcClient,
	stemcellClient stemcell.StemcellClient,
	agentSettings vm.AgentSettings,
	agentEnvFactory apiv1.AgentEnvFactory,
	config config.Config,
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) Factory {
	return Factory{
		govcClient,
		stemcellClient,
		agentSettings,
		agentEnvFactory,
		config,
		fs,
		uuidGen,
		logger,
	}
}

func (f Factory) New(_ apiv1.CallContext) (apiv1.CPI, error) {
	return CPI{
		NewCreateStemcellMethod(f.govcClient, f.stemcellClient, f.uuidGen, f.logger),
		NewDeleteStemcellMethod(f.govcClient, f.logger),
		NewCreateVMMethod(f.govcClient, f.agentSettings, f.config.GetAgentOptions(), f.agentEnvFactory, f.uuidGen, f.logger),
		NewDeleteVMMethod(f.govcClient),
		NewHasVMMethod(f.govcClient),
		NewCreateDiskMethod(f.govcClient, f.uuidGen),
		NewAttachDiskMethod(f.govcClient, f.agentSettings, f.agentEnvFactory),
		NewDetachDiskMethod(f.govcClient, f.agentSettings, f.agentEnvFactory),
		NewDeleteDiskMethod(f.govcClient, f.logger),
		NewMiscMethod(f.govcClient),
	}, nil
}

func (c CPI) CalculateVMCloudProperties(res apiv1.VMResources) (apiv1.VMCloudProps, error) {
	panic("CalculateVMCloudProperties")
	return apiv1.NewVMCloudPropsFromMap(map[string]interface{}{}), nil
}

func (c CPI) SetVMMetadata(cid apiv1.VMCID, metadata apiv1.VMMeta) error {
	//NOOP is sufficient for now
	fmt.Fprintf(os.Stderr, "metadata: %s\n", metadata)
	return nil
}

func (c CPI) SetDiskMetadata(cid apiv1.VMCID, metadata apiv1.VMMeta) error {
	//NOOP is sufficient for now
	fmt.Fprintf(os.Stderr, "metadata: %s\n", metadata)
	return nil
}

func (c CPI) RebootVM(cid apiv1.VMCID) error {
	panic("RebootVM")
	return nil
}

func (c CPI) GetDisks(cid apiv1.VMCID) ([]apiv1.DiskCID, error) {
	panic("GetDisks")
	return []apiv1.DiskCID{}, nil
}

func (c CPI) HasDisk(cid apiv1.DiskCID) (bool, error) {
	panic("HasDisk")
	return false, nil
}
