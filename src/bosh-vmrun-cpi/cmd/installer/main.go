package main

import (
	"bosh-vmrun-cpi/install"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-vmrun-cpi/config"
)

var (
	configPathOpt     = flag.String("configPath", "", "Path to configuration file")
	directorTmpDirOpt = flag.String("directorTmpDirPath", "", "Path to director's temp file containing extracted stemcell images")
	versionOpt        = flag.Bool("version", false, "Version")

	//set by X build flag
	version string
)

func main() {
	var err error
	var configJSON string
	var directorTmpDirPath string

	logLevel, err := boshlog.Levelify(os.Getenv("BOSH_LOG_LEVEL"))
	if err != nil {
		logLevel = boshlog.LevelDebug
	}
	logger := boshlog.NewWriterLogger(logLevel, os.Stderr)

	flag.Parse()

	if *configPathOpt != "" {
		configJSONBytes, err := ioutil.ReadFile(*configPathOpt)
		configJSON = string(configJSONBytes)
		if err != nil {
			logger.Error("main", "loading cfg", err)
			os.Exit(1)
		}
	}

	if *directorTmpDirOpt != "" {
		//optional, validated on use
		directorTmpDirPath = *directorTmpDirOpt
	}

	command := flag.Arg(0)

	if command == "version" {
		fmt.Println(version)
		os.Exit(0)
	}

	if command == "encoded-config" {
		configBase64 := base64.StdEncoding.EncodeToString([]byte(configJSON))
		fmt.Println(configBase64)
		os.Exit(0)
	}

	cpiConfig, err := config.NewConfigFromJson(configJSON)
	if err != nil {
		logger.Error("main", "parsing config JSON", err)
		os.Exit(1)
	}

	sshClient, err := install.NewSshClient(cpiConfig, logger)
	if err != nil {
		logger.Error("main", "initializing ssh client", err)
		os.Exit(1)
	}

	installer, err := install.NewInstaller(cpiConfig, sshClient, logger)
	if err != nil {
		logger.Error("main", "initializing installer", err)
		os.Exit(1)
	}

	switch command {
	case "install-cpi":
		err = installer.InstallCPI(version)
	case "sync-director-stemcells":
		err = installer.SyncDirectorStemcells(directorTmpDirPath)
	default:
		err = errors.New("command required")
	}

	if err != nil {
		logger.Error("main", "command failed", err)
		os.Exit(1)
	}
}
