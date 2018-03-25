package config

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
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
	Agent    apiv1.AgentOptions
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

func (c Config) GetAgentOptions() apiv1.AgentOptions {
	return c.Cloud.Properties.Agent
}

func (c Config) Validate() error {
	return nil
}
