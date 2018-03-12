package govc

import (
	"net/url"

	cpiConfig "bosh-govmomi-cpi/config"
)

type GovcConfig struct {
	EsxUrl string
}

func NewGovcConfig(cpi cpiConfig.Config) GovcConfig {
	return GovcConfig{EsxUrl: buildEsxUrl(cpi)}
}

func buildEsxUrl(cpi cpiConfig.Config) string {
	vcenter := cpi.Cloud.Properties.Vcenters[0]

	url := &url.URL{
		Scheme: "https",
		User:   url.UserPassword(vcenter.User, vcenter.Password),
		Host:   vcenter.Host,
	}

	return url.String()
}
