package config

import (
	"encoding/json"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
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
	Agent apiv1.AgentOptions
	Vmrun Vmrun
}

type Vmrun struct {
	Vm_Store_Path                     string
	Vmrun_Bin_Path                    string
	Ovftool_Bin_Path                  string
	Vm_Start_Max_Wait_Seconds         int
	Vm_Soft_Shutdown_Max_Wait_Seconds int
	Stemcell_Store_Path               string
	Enable_Human_Readable_Name        bool

	//calculated
	Vm_Start_Max_Wait         time.Duration
	Vm_Soft_Shutdown_Max_Wait time.Duration
	Ssh_Tunnel                struct {
		Host        string
		Port        string
		Username    string
		Private_Key string
		Platform    string
	}
}

func NewConfigFromJson(configJson string) (Config, error) {
	var config Config
	var err error

	err = json.Unmarshal([]byte(configJson), &config)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config")
	}

	err = config.Validate()
	if err != nil {
		return config, bosherr.WrapError(err, "Validating configuration")
	}

	config.Cloud.Properties.Vmrun.setDurations()
	config.Cloud.Properties.Vmrun.setDefaultStemcellStore()

	return config, nil
}

func (c Config) GetAgentOptions() apiv1.AgentOptions {
	return c.Cloud.Properties.Agent
}

func (c Config) Validate() error {
	return nil
}

func (v *Vmrun) PlatformPathSeparator() string {
	if v.Ssh_Tunnel.Platform == "windows" {
		return `\`
	}
	return `/`
}

func (v *Vmrun) setDurations() {
	v.Vm_Start_Max_Wait = secsIntToDuration(v.Vm_Start_Max_Wait_Seconds)
	v.Vm_Soft_Shutdown_Max_Wait = secsIntToDuration(v.Vm_Soft_Shutdown_Max_Wait_Seconds)
}

func (v *Vmrun) setDefaultStemcellStore() {
	if v.Stemcell_Store_Path == "" {
		v.Stemcell_Store_Path = strings.Join([]string{v.Vm_Store_Path, "stemcells"}, v.PlatformPathSeparator())
	}
}

func secsIntToDuration(secs int) time.Duration {
	return time.Duration(float64(secs) * float64(time.Second))
}
