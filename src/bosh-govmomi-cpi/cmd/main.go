package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cppforlife/bosh-cpi-go/rpc"

	"github.com/micahyoung/bosh-govmomi-cpi/cpi"
)

var (
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // todo MAC generation

	logger, fs := basicDeps()
	defer logger.HandlePanic("Main")

	flag.Parse()
	config, err := NewConfigFromPath(*configPathOpt, fs)
	if err != nil {
		logger.Error("main", "Loading config %s", err.Error())
		os.Exit(1)
	}

	cpiFactory := cpi.NewFactory(fs, cpi.FactoryOpts(config), logger)

	cli := rpc.NewFactory(logger).NewCLI(cpiFactory)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error("main", "Serving once: %s", err)
		os.Exit(1)
	}
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem) {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	return logger, fs
}
