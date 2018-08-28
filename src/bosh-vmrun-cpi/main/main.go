package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	"github.com/cppforlife/bosh-cpi-go/rpc"

	"bosh-vmrun-cpi/action"
	"bosh-vmrun-cpi/config"
	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/stemcell"
	"bosh-vmrun-cpi/vm"
	"bosh-vmrun-cpi/vmx"
)

var (
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // todo MAC generation

	logger, fs, compressor, uuidGen := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()
	cpiConfig, err := config.NewConfigFromPath(*configPathOpt, fs)
	if err != nil {
		logger.ErrorWithDetails("main", "loading cfg", err)
		os.Exit(1)
	}

	driverConfig := driver.NewConfig(cpiConfig)
	boshRunner := boshsys.NewExecCmdRunner(logger)
	vmrunRunner := driver.NewVmrunRunner(driverConfig.VmrunPath(), boshRunner, logger)
	ovftoolRunner := driver.NewOvftoolRunner(driverConfig.OvftoolPath(), boshRunner, logger)
	vdiskmanagerRunner := driver.NewVdiskmanagerRunner(driverConfig.VdiskmanagerPath(), boshRunner, logger)
	vmxBuilder := vmx.NewVmxBuilder(logger)
	driverClient := driver.NewClient(vmrunRunner, ovftoolRunner, vdiskmanagerRunner, vmxBuilder, driverConfig, logger)
	stemcellClient := stemcell.NewClient(compressor, fs, logger)
	agentEnvFactory := apiv1.NewAgentEnvFactory()
	agentSettings := vm.NewAgentSettings(fs, logger, agentEnvFactory)
	cpiFactory := action.NewFactory(driverClient, stemcellClient, agentSettings, agentEnvFactory, cpiConfig, fs, uuidGen, logger)

	cli := rpc.NewFactory(logger).NewCLI(cpiFactory)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error("main", "Serving once: %s", err)
		os.Exit(1)
	}
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem, boshcmd.Compressor, boshuuid.Generator) {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	cmdRunner := boshsys.NewExecCmdRunner(logger)
	compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)
	uuidGen := boshuuid.NewGenerator()

	return logger, fs, compressor, uuidGen
}
