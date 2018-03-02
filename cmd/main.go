package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/rpc"

	"github.com/micahyoung/bosh-govmomi-cpi/cpi"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // todo MAC generation

	logger := basicDeps()
	defer logger.HandlePanic("Main")

	flag.Parse()

	cpiFactory := cpi.NewFactory(logger)

	cli := rpc.NewFactory(logger).NewCLI(cpiFactory)

	err := cli.ServeOnce()
	if err != nil {
		logger.Error("main", "Serving once: %s", err)
		os.Exit(1)
	}
}

func basicDeps() boshlog.Logger {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	return logger
}
