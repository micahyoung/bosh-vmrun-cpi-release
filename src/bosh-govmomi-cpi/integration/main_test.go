package integration_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

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
	var configPath string
	var stemcellCid string
	var vmCid string
	var diskId string
	var response map[string]interface{}

	BeforeEach(func() {
		configFile, err := ioutil.TempFile("", "config")
		Expect(err).ToNot(HaveOccurred())

		configFile.WriteString(configContent)
		configPath, _ = filepath.Abs(configFile.Name())
	})

	AfterEach(func() {
		os.Remove(configPath)
	})

	It("runs the cpi", func() {
		cpiBin, err := gexec.Build("bosh-govmomi-cpi/main")
		Expect(err).ToNot(HaveOccurred())

		request := fmt.Sprintf(`{ "method": "info", "arguments": [] }`)

		session, stdin := gexecCommandWithStdin(cpiBin, "-configPath", configPath)

		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "2s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).ToNot(BeNil())

		imageTarballPath := filepath.Join(extractedStemcellTempDir, "image")
		request = fmt.Sprintf(`{
			"method": "create_stemcell",
			"arguments": ["%s", {
				"architecture":"x86_64",
				"container_format":"bare",
				"disk":3072,
				"disk_format":"ovf",
				"hypervisor":"esxi",
				"infrastructure":"vsphere",
				"name":"bosh-vsphere-esxi-ubuntu-trusty-go_agent",
				"os_distro":"ubuntu",
				"os_type":"linux",
				"root_device_name":"/dev/sda1",
				"version":"3541.5"
			}]
		}`, imageTarballPath)

		session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "120s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		stemcellCid = response["result"].(string)

		request = fmt.Sprintf(`{
			"method":"create_vm",
			"arguments":[
			"0aa2b270-5da8-4596-728d-84df02143198",
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
			{"bosh":{"password":"$6$/UZ140gHNv$iZjNisgoF3DuQCfsmy6d8nXA5v7Agw34IjtuhqdthmMsXwIZaRJo2b4yFmXXVgXIU9mXjDXECa/eBu9z0YsPa0"}}
			]
		}`, stemcellCid)

		session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		vmCid = response["result"].(string)
		Expect(vmCid).ToNot(Equal(""))

		time.Sleep(30 * time.Second)

		diskKB := "10240"
		request = fmt.Sprintf(`{
			"method":"create_disk",
			"arguments":[%s,{},"%s"]
		}`, diskKB, vmCid)

		session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		diskId = response["result"].(string)
		Expect(diskId).ToNot(Equal(""))

		request = fmt.Sprintf(`{
		  "method":"attach_disk",
		  "arguments":["%s","%s"]
		}`, vmCid, diskId)

		session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())

		request = fmt.Sprintf(`{
			"method":"delete_vm",
			"arguments":["%s"]
		}`, vmCid)

		session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())

		request = fmt.Sprintf(`{
			"method":"delete_disk",
			"arguments":["%s"]
		}`, diskId)

		session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())

		request = fmt.Sprintf(`{
			"method":"delete_stemcell",
			"arguments":["%s"]
		}`, stemcellCid)

		session, stdin = gexecCommandWithStdin(cpiBin, "-configPath", configPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())
	})
})
