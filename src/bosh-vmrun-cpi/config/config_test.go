package config_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	"bosh-vmrun-cpi/config"
)

var _ = Describe("Config", func() {
	It("runs the cpi", func() {
		fs := fakesys.NewFakeFileSystem()
		config_content := `{
			"cloud":{
				"plugin":"vsphere",
				"properties":{
					"vmrun":{
						"vm_store_path":"/store-dir",
						"vmrun_bin_path":"/vmrun-bin",
						"ovftool_bin_path":"/ovftool-bin",
						"vdiskmanager_bin_path":"/vdiskmanager-bin",
						"stemcell_store_path":"/stemcell-store-dir",
						"vm_soft_shutdown_max_wait_seconds":20,
						"vm_start_max_wait_seconds":10
					},
					"agent":{"ntp":["time1.google.com",
					"time2.google.com",
					"time3.google.com",
					"time4.google.com"],
					"blobstore":{
						"provider":"local",
						"options":{"blobstore_path":"/var/vcap/micro_bosh/data/cache"}
					 },
					"mbus":"https://mbus:mbuspassword@0.0.0.0:6868"}
				}
			}
		}`
		fs.WriteFileString("cpi_config.json", config_content)

		c, err := config.NewConfigFromPath("cpi_config.json", fs)
		Expect(err).ToNot(HaveOccurred())

		Expect(c).To(MatchAllFields(Fields{
			"Cloud": MatchAllFields(Fields{
				"Plugin": Equal("vsphere"),
				"Properties": MatchAllFields(Fields{
					"Vmrun": MatchAllFields(Fields{
						"Vm_Store_Path":                     Equal("/store-dir"),
						"Vmrun_Bin_Path":                    Equal("/vmrun-bin"),
						"Ovftool_Bin_Path":                  Equal("/ovftool-bin"),
						"Vdiskmanager_Bin_Path":             Equal("/vdiskmanager-bin"),
						"Stemcell_Store_Path":               Equal("/stemcell-store-dir"),
						"Vm_Soft_Shutdown_Max_Wait":         Equal(20 * time.Second),
						"Vm_Start_Max_Wait":                 Equal(10 * time.Second),
						"Vm_Soft_Shutdown_Max_Wait_Seconds": Equal(20),
						"Vm_Start_Max_Wait_Seconds":         Equal(10),
					}),
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
