package config

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Config struct {
	Cloud Cloud
}

type Cloud struct {
	Plugin     string
	Properties CPIProperties
}

type CPIProperties struct {
	Vcenters []Vcenter
}

type Vcenter struct {
	Host        string
	User        string
	Password    string
	Datacenters []Datacenter
}

type Datacenter struct {
	Name              string
	Vm_Folder         string
	Template_Folder   string
	Disk_Path         string
	Datastore_Pattern string
}

//{"cloud":{"plugin":"vsphere","properties":{"vcenters":[{"host":172.1,"user":"root","password":"homelabnyc","datacenters":[{"name":"ha-datacenter","vm_folder":"BOSH_VMs","template_folder":"BOSH_Templates","disk_path":"bosh_disks","datastore_pattern":"datastore1"}]}],"agent":{"ntp":["time1.google.com","time2.google.com","time3.google.com","time4.google.com"],"blobstore":{"provider":"local","options":{"blobstore_path":"/var/vcap/micro_bosh/data/cache"}},"mbus":"https://mbus:p2an3m7idfm6vmqp3w74@0.0.0.0:6868"}}}}
func NewConfigFromPath(path string, fs boshsys.FileSystem) (Config, error) {
	var config Config

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return config, bosherr.WrapErrorf(err, "Reading config '%s'", path)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config")
	}

	err = config.Validate()
	if err != nil {
		return config, bosherr.WrapError(err, "Validating configuration")
	}

	return config, nil
}

func (c Config) Validate() error {
	return nil
}
