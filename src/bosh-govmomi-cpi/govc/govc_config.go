package govc

import (
	"net/url"

	cpiconfig "bosh-govmomi-cpi/config"
)

type GovcConfigImpl struct {
	cpiConfig cpiconfig.Config
}

func NewGovcConfig(cpiConfig cpiconfig.Config) GovcConfig {
	return GovcConfigImpl{cpiConfig: cpiConfig}
}

func (c GovcConfigImpl) EsxUrl() string {
	vcenter := c.cpiConfig.Cloud.Properties.Vcenters[0]

	url := &url.URL{
		Scheme: "https",
		User:   url.UserPassword(vcenter.User, vcenter.Password),
		Host:   vcenter.Host,
	}

	return url.String()
}
