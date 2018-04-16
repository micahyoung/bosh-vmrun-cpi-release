package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	"github.com/micahyoung/bosh-govmomi-cpi/config"
)

var _ = Describe("CreateStemcell", func() {
	It("runs the cpi", func() {
		fs := fakesys.NewFakeFileSystem()
		config_content := `{"cloud":{"plugin":"vsphere","properties":{"vcenters":[{"host":"1.2.3.4","user":"root","password":"password","datacenters":[{"name":"ha-datacenter","vm_folder":"BOSH_VMs","template_folder":"BOSH_Templates","disk_path":"bosh_disks","datastore_pattern":"datastore1"}]}],"agent":{"ntp":["time1.google.com","time2.google.com","time3.google.com","time4.google.com"],"blobstore":{"provider":"local","options":{"blobstore_path":"/var/vcap/micro_bosh/data/cache"}},"mbus":"https://mbus:mbuspassword@0.0.0.0:6868"}}}}`
		fs.WriteFileString("cpi_config.json", config_content)

		c, err := config.NewConfigFromPath("cpi_config.json", fs)
		Expect(err).ToNot(HaveOccurred())

		Expect(c).To(MatchAllFields(Fields{
			"Cloud": MatchAllFields(Fields{
				"Plugin": Equal("vsphere"),
				"Properties": MatchAllFields(Fields{
					"Vcenters": ContainElement(MatchAllFields(Fields{
						"Host":     Equal("1.2.3.4"),
						"User":     Equal("root"),
						"Password": Equal("password"),
						"Datacenters": ContainElement(MatchAllFields(Fields{
							"Name":              Equal("ha-datacenter"),
							"Vm_Folder":         Equal("BOSH_VMs"),
							"Template_Folder":   Equal("BOSH_Templates"),
							"Disk_Path":         Equal("bosh_disks"),
							"Datastore_Pattern": Equal("datastore1"),
						})),
					})),
					"Agent": MatchAllFields(Fields{
						"Mbus": Equal("https://mbus:mbuspassword@0.0.0.0:6868"),
						"NTP":  ConsistOf("time1.google.com", "time2.google.com", "time3.google.com", "time4.google.com"),
						"Blobstore": MatchAllFields(Fields{
							"Type":    Equal("local"),
							"Options": HaveKeyWithValue("blobstore_path", "/var/vcap/micro_bosh/data/cache"),
						}),
					}),
				}),
			}),
		}))
	})
})
