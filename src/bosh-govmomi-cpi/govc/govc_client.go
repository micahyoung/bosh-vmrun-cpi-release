package govc

import (
	"context"
	"flag"
	"io/ioutil"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/vmware/govmomi/govc/cli"

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

type GovcClientImpl struct {
	config GovcConfig
	logger boshlog.Logger
}

func NewClient(config GovcConfig, logger boshlog.Logger) GovcClient {
	return GovcClientImpl{config: config, logger: logger}
}

func (c GovcClientImpl) ImportOvf(ovfPath string) (bool, error) {
	commands := cli.Commands()

	ctx := context.Background()
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)

	cmd := commands["import.ovf"]
	cmd.Register(ctx, fs)

	fs.Parse([]string{ovfPath})
	fs.Set("u", c.config.EsxUrl)
	fs.Set("k", "true")
	fs.Set("json", "true")

	var err error
	if err = cmd.Process(ctx); err != nil {
		return false, err
	}

	if err = cmd.Run(ctx, fs); err != nil {
		return false, err
	}

	return true, nil
}
