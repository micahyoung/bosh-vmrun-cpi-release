package main

import (
	"encoding/base64"
	"flag"
	"fmt"
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
	configPathOpt       = flag.String("configPath", "", "Path to configuration file")
	configBase64JsonOpt = flag.String("configBase64JSON", "", "Base64-encoded JSON string of configuration")
	versionOpt          = flag.Bool("version", false, "Version")

	version string
)

func main() {
	var err error

	rand.Seed(time.Now().UTC().UnixNano()) // todo MAC generation

	logger, fs, compressor, uuidGen := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()

	if *versionOpt {
		fmt.Println(version)
		os.Exit(0)
	}

	var configJson string
	if *configPathOpt != "" {
		configJson, err = fs.ReadFileString(*configPathOpt)
		if err != nil {
			logger.ErrorWithDetails("main", "loading cfg", err)
			os.Exit(1)
		}
	} else if *configBase64JsonOpt != "" {
		configJsonBytes, err := base64.StdEncoding.DecodeString(*configBase64JsonOpt)
		if err != nil {
			logger.ErrorWithDetails("main", "base64 decoding cfg", err)
			os.Exit(1)
		}
		configJson = string(configJsonBytes)
	} else {
		logger.Error("main", "config option required")
		os.Exit(1)
	}

	cpiConfig, err := config.NewConfigFromJson(configJson)
	if err != nil {
		logger.ErrorWithDetails("main", "config JSON is invalid", err)
		os.Exit(1)
	}

	driverConfig := driver.NewConfig(cpiConfig)
	stemcellConfig := stemcell.NewConfig(cpiConfig)
	boshRunner := boshsys.NewExecCmdRunner(logger)
	retryFileLock := driver.NewRetryFileLock(logger)
	vmrunRunner := driver.NewVmrunRunner(driverConfig.VmrunPath(), retryFileLock, logger)
	ovftoolRunner := driver.NewOvftoolRunner(driverConfig.OvftoolPath(), boshRunner, logger)
	vdiskmanagerRunner := driver.NewVdiskmanagerRunner(driverConfig.VdiskmanagerPath(), boshRunner, logger)
	vmxBuilder := vmx.NewVmxBuilder(logger)
	driverClient := driver.NewClient(vmrunRunner, ovftoolRunner, vdiskmanagerRunner, vmxBuilder, driverConfig, logger)
	stemcellClient := stemcell.NewClient(compressor, fs, logger)
	stemcellStore := stemcell.NewStemcellStore(stemcellConfig, compressor, fs, logger)
	agentEnvFactory := apiv1.NewAgentEnvFactory()
	agentSettings := vm.NewAgentSettings(fs, logger, agentEnvFactory)
	cpiFactory := action.NewFactory(driverClient, stemcellClient, stemcellStore, agentSettings, agentEnvFactory, cpiConfig, fs, uuidGen, logger)

	cli := rpc.NewFactory(logger).NewCLI(cpiFactory)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error("main", "Serving once: %s", err)
		os.Exit(1)
	}
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem, boshcmd.Compressor, boshuuid.Generator) {
	logLevel, err := boshlog.Levelify(os.Getenv("BOSH_LOG_LEVEL"))
	if err != nil {
		logLevel = boshlog.LevelDebug
	}

	logger := boshlog.NewWriterLogger(logLevel, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	cmdRunner := boshsys.NewExecCmdRunner(logger)
	compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)
	uuidGen := boshuuid.NewGenerator()

	return logger, fs, compressor, uuidGen
}
