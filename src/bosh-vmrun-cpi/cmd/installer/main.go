package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-vmrun-cpi/config"
)

var (
	configPathOpt       = flag.String("configPath", "", "Path to configuration file")
	configBase64JSONOpt = flag.String("configBase64JSON", "", "Base64-encoded JSON string of configuration")
	plaformOpt          = flag.String("platform", "", "Platform name (windows, linux, darwing) of the binary to copy")
	versionOpt          = flag.Bool("version", false, "Version")

	//set by X build flag
	version string
)

func cpiSourcePath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return filepath.Join(dir, fmt.Sprintf("cpi-%s", *plaformOpt))
}

func cpiDestPath(cpiConfig config.Config) (string, bool) {
	vmStorePath := cpiConfig.Cloud.Properties.Vmrun.Vm_Store_Path
	cpiPath := filepath.Join(vmStorePath, "cpi")

	installCpi := true
	if _, err := os.Stat(vmStorePath); os.IsNotExist(err) {
		installCpi = false
	}

	return cpiPath, installCpi
}

func main() {
	var err error
	var configJSON string
	var cpiConfig config.Config

	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)

	flag.Parse()

	if *versionOpt {
		fmt.Println(version)
		os.Exit(0)
	}

	if *configPathOpt != "" {
		configJSONBytes, err := ioutil.ReadFile(*configPathOpt)
		configJSON = string(configJSONBytes)
		if err != nil {
			logger.ErrorWithDetails("main", "loading cfg", err)
			os.Exit(1)
		}
	} else if *configBase64JSONOpt != "" {
		configJSONBytes, err := base64.StdEncoding.DecodeString(*configBase64JSONOpt)
		if err != nil {
			logger.ErrorWithDetails("main", "base64 decoding cfg", err)
			os.Exit(1)
		}
		configJSON = string(configJSONBytes)
	}

	cpiConfig, err = config.NewConfigFromJson(configJSON)
	if err != nil {
		logger.ErrorWithDetails("main", "config JSON is invalid", err, configJSON)
		os.Exit(1)
	}

	cpiDestPath, installCpi := cpiDestPath(cpiConfig)
	escapedCpiDestPath := strings.Trim(strconv.Quote(cpiDestPath), `"`)

	//output path
	fmt.Println(escapedCpiDestPath)

	if !installCpi {
		logger.Debug("main", "not possible to install CPI dest")
		os.Exit(0)
	}

	cpiSrcPath := cpiSourcePath()
	in, err := os.Open(cpiSrcPath)
	if err != nil {
		logger.ErrorWithDetails("main", "opening CPI config", err)
		os.Exit(1)
	}
	defer in.Close()

	out, err := os.Create(cpiDestPath)
	if err != nil {
		logger.ErrorWithDetails("main", "creating CPI destination file", err)
		os.Exit(1)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		logger.ErrorWithDetails("main", "copying from CPI source to destination file", err)
		os.Exit(1)
	}
	out.Close()
}
