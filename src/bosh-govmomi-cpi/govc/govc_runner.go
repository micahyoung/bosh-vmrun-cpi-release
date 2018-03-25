package govc

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/vmware/govmomi/govc/cli"

	//TOOD: disable all but needed
	_ "github.com/vmware/govmomi/govc/about"
	_ "github.com/vmware/govmomi/govc/cluster"
	_ "github.com/vmware/govmomi/govc/cluster/group"
	_ "github.com/vmware/govmomi/govc/cluster/override"
	_ "github.com/vmware/govmomi/govc/cluster/rule"
	_ "github.com/vmware/govmomi/govc/datacenter"
	_ "github.com/vmware/govmomi/govc/datastore"
	_ "github.com/vmware/govmomi/govc/datastore/disk"
	_ "github.com/vmware/govmomi/govc/datastore/vsan"
	_ "github.com/vmware/govmomi/govc/device"
	_ "github.com/vmware/govmomi/govc/device/cdrom"
	_ "github.com/vmware/govmomi/govc/device/floppy"
	_ "github.com/vmware/govmomi/govc/device/scsi"
	_ "github.com/vmware/govmomi/govc/device/serial"
	_ "github.com/vmware/govmomi/govc/device/usb"
	_ "github.com/vmware/govmomi/govc/dvs"
	_ "github.com/vmware/govmomi/govc/dvs/portgroup"
	_ "github.com/vmware/govmomi/govc/env"
	_ "github.com/vmware/govmomi/govc/events"
	_ "github.com/vmware/govmomi/govc/export"
	_ "github.com/vmware/govmomi/govc/extension"
	_ "github.com/vmware/govmomi/govc/fields"
	_ "github.com/vmware/govmomi/govc/folder"
	_ "github.com/vmware/govmomi/govc/host"
	_ "github.com/vmware/govmomi/govc/host/account"
	_ "github.com/vmware/govmomi/govc/host/autostart"
	_ "github.com/vmware/govmomi/govc/host/cert"
	_ "github.com/vmware/govmomi/govc/host/date"
	_ "github.com/vmware/govmomi/govc/host/esxcli"
	_ "github.com/vmware/govmomi/govc/host/firewall"
	_ "github.com/vmware/govmomi/govc/host/maintenance"
	_ "github.com/vmware/govmomi/govc/host/option"
	_ "github.com/vmware/govmomi/govc/host/portgroup"
	_ "github.com/vmware/govmomi/govc/host/service"
	_ "github.com/vmware/govmomi/govc/host/storage"
	_ "github.com/vmware/govmomi/govc/host/vnic"
	_ "github.com/vmware/govmomi/govc/host/vswitch"
	_ "github.com/vmware/govmomi/govc/importx"
	_ "github.com/vmware/govmomi/govc/license"
	_ "github.com/vmware/govmomi/govc/logs"
	_ "github.com/vmware/govmomi/govc/ls"
	_ "github.com/vmware/govmomi/govc/metric"
	_ "github.com/vmware/govmomi/govc/metric/interval"
	_ "github.com/vmware/govmomi/govc/object"
	_ "github.com/vmware/govmomi/govc/option"
	_ "github.com/vmware/govmomi/govc/permissions"
	_ "github.com/vmware/govmomi/govc/pool"
	_ "github.com/vmware/govmomi/govc/role"
	_ "github.com/vmware/govmomi/govc/session"
	_ "github.com/vmware/govmomi/govc/task"
	_ "github.com/vmware/govmomi/govc/vapp"
	_ "github.com/vmware/govmomi/govc/version"
	_ "github.com/vmware/govmomi/govc/vm"
	_ "github.com/vmware/govmomi/govc/vm/disk"
	_ "github.com/vmware/govmomi/govc/vm/guest"
	_ "github.com/vmware/govmomi/govc/vm/network"
	_ "github.com/vmware/govmomi/govc/vm/rdm"
	_ "github.com/vmware/govmomi/govc/vm/snapshot"
)

type GovcRunnerImpl struct {
	logger      boshlog.Logger
	cliCommands map[string]cli.Command
}

func NewGovcRunner(logger boshlog.Logger) GovcRunner {
	cliCommands := cli.Commands()
	return &GovcRunnerImpl{logger: logger, cliCommands: cliCommands}
}

func (c GovcRunnerImpl) CliCommand(command string, flagMap map[string]string, args []string) (string, error) {
	c.logger.Debug("govc-runner", fmt.Sprintf("command: %s, flags: %+v, args: %s", command, flagMap, args))
	ctx := context.Background()

	cliCommand := c.cliCommands[command]
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	stdoutReader, stdoutWriter, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = stdoutWriter
	defer func() {
		os.Stdout = oldStdout
	}()

	cliCommand.Register(ctx, flagSet)

	flagSet.Set("json", "true")
	flagSet.Set("persist-session", "false")

	if args != nil {
		flagSet.Parse(args)
	}

	if flagMap != nil {
		for k, v := range flagMap {
			flagSet.Set(k, v)
		}
	}

	var err error
	if err = cliCommand.Process(ctx); err != nil {
		return "", err
	}

	if err = cliCommand.Run(ctx, flagSet); err != nil {
		return "", err
	}

	stdoutWriter.Close()
	var output bytes.Buffer
	io.Copy(&output, stdoutReader)

	return output.String(), nil
}
