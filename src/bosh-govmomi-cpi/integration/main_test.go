package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	configContent = `{
		"cloud": {
			"plugin": "vsphere",
			"properties": {
				"vcenters": [
				{
					"host": "172.16.125.131",
					"user": "root",
					"password": "homelabnyc",
					"datacenters": [
					{
						"name": "ha-datacenter",
						"vm_folder": "BOSH_VMs",
						"template_folder": "BOSH_Templates",
						"disk_path": "bosh_disks",
						"datastore_pattern": "datastore1"
					}
					]
				}
				],
				"agent": {
					"ntp": [
					],
					"blobstore": {
						"provider": "local",
						"options": {
							"blobstore_path": "/var/vcap/micro_bosh/data/cache"
						}
					},
					"mbus": "https://mbus:p2an3m7idfm6vmqp3w74@0.0.0.0:6868"
				}
			}
		}
	}`
)

var _ = Describe("CPI", func() {
	It("runs the cpi", func() {
		configFile, err := ioutil.TempFile("", "config")
		Expect(err).ToNot(HaveOccurred())

		configFile.WriteString(configContent)
		configPath, err := filepath.Abs(configFile.Name())
		defer os.Remove(configPath)

		cpiBin, err := gexec.Build("bosh-govmomi-cpi/main")
		Expect(err).ToNot(HaveOccurred())

		request := fmt.Sprintf(`{ "method": "info", "arguments": [] }`)

		session, stdin := gexecCommandWithStdin(cpiBin, "-configPath", configPath)

		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "2s").Should(gbytes.Say(`"error":null`))

		//imageTarballPath := filepath.Join(extractedStemcellTempDir, "image")
		//request := fmt.Sprintf(`{
		//  "method": "create_stemcell",
		//  "arguments": ["%s", {
		//    "architecture":"x86_64",
		//    "container_format":"bare",
		//    "disk":3072,
		//    "disk_format":"ovf",
		//    "hypervisor":"esxi",
		//    "infrastructure":"vsphere",
		//    "name":"bosh-vsphere-esxi-ubuntu-trusty-go_agent",
		//    "os_distro":"ubuntu",
		//    "os_type":"linux",
		//    "root_device_name":"/dev/sda1",
		//    "version":"3541.5"
		//  }]
		//}`, imageTarballPath)

		//session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		//stdin.Write([]byte(request))
		//stdin.Close()

		//Eventually(session.Out, "120s").Should(gbytes.Say(`"error":null`))

		vmCid := "vm-cid"
		stemcellCid := "d61a115a-f7ec-4ede-4392-c26da3293453"

		request = fmt.Sprintf(`{
			"method":"create_vm",
			"arguments":[
			"%s",
			"%s",
			{"instance_type":"concourse"},
			{
				"default":{
					"cloud_properties":{
						"net_id":".",
						"security_groups":"."
					},
					"default":["dns", "gateway"],
					"dns":["10.0.0.2"],
					"gateway":"10.0.0.1",
					"ip":"10.0.0.3",
					"netmask":"255.255.255.0",
					"type":"manual"
				}
			},
			[],
			{"bosh":{"password":"*"}}
			]
		}`, vmCid, stemcellCid)

		session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
	})
})
