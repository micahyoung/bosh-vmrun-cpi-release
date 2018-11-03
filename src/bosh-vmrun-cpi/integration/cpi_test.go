package integration_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("cpi integration", func() {
	var stemcellCid string
	var vmCid string
	var diskId string
	var response map[string]interface{}

	It("runs the cpi", func() {
		cpiBin, err := gexec.Build("bosh-vmrun-cpi/cmd/cpi")
		Expect(err).ToNot(HaveOccurred())

		request := fmt.Sprintf(`{ "method": "info", "arguments": [] }`)

		session, stdin := GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)

		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "2s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).ToNot(BeNil())

		//create_stemcell
		request = `{
			"method": "create_stemcell",
			"arguments": ["local-image-path-ignored--uses-store-instead", {
				"name":"bosh-vsphere-esxi-ubuntu-trusty-go_agent",
				"version":"3586.42",
				"architecture":"x86_64",
				"container_format":"bare",
				"disk":3072,
				"disk_format":"ovf",
				"hypervisor":"esxi",
				"infrastructure":"vsphere",
				"os_distro":"ubuntu",
				"os_type":"linux",
				"root_device_name":"/dev/sda1"
			}]
		}`

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "120s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		stemcellCid = response["result"].(string)

		//create_vm
		request = fmt.Sprintf(`{
		  "method":"create_vm",
		  "arguments":[
		    "0aa2b270-5da8-4596-728d-84df02143198",
		    "%s",
		    {
				  "cpu":2,
				  "ram":4096,
				  "disk":10240
			  },
		    {
		      "default":{
		        "cloud_properties":{ "name":"VMnet8"},
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

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "120s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		vmCid = response["result"].(string)
		Expect(vmCid).ToNot(Equal(""))

		//create_disk
		diskMB := "2048"
		request = fmt.Sprintf(`{
			"method":"create_disk",
			"arguments":[%s,{},"%s"]
		}`, diskMB, vmCid)

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		diskId = response["result"].(string)
		Expect(diskId).ToNot(Equal(""))

		//attach_disk
		request = fmt.Sprintf(`{
		  "method":"attach_disk",
		  "arguments":["%s","%s"]
		}`, vmCid, diskId)

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())

		//detach_disk
		request = fmt.Sprintf(`{
		  "method":"detach_disk",
		  "arguments":["%s","%s"]
		}`, vmCid, diskId)

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())

		//delete_vm
		request = fmt.Sprintf(`{
			"method":"delete_vm",
			"arguments":["%s"]
		}`, vmCid)

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())

		//delete_disk
		request = fmt.Sprintf(`{
			"method":"delete_disk",
			"arguments":["%s"]
		}`, diskId)

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())

		//create_vm with bootstrap
		request = fmt.Sprintf(`{
			"method":"create_vm",
			"arguments":[
				"0aa2b270-5da8-4596-728d-84df02143198",
				"%s",
				{
					"cpu":2,
					"ram":4096,
					"disk":1024,
					"bootstrap":{
						"script_content": "touch /home/vcap/bootstrapped.txt",
						"script_path": "/home/vcap/bootstrap.sh",
						"interpreter_path": "/bin/bash",
						"ready_process_name": "bosh-agent",
						"username": "vcap",
						"password": "c1oudc0w",
						"min_wait_seconds": 10,
						"max_wait_seconds": 30
					}
				},
				{
				"default":{
					"cloud_properties":{ "name":"VMnet8"},
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

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "120s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		vmCid = response["result"].(string)
		Expect(vmCid).ToNot(Equal(""))

		//delete_vm with bootstrap
		request = fmt.Sprintf(`{
			"method":"delete_vm",
			"arguments":["%s"]
		}`, vmCid)

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())

		//delete_stemcell
		request = fmt.Sprintf(`{
			"method":"delete_stemcell",
			"arguments":["%s"]
		}`, stemcellCid)

		session, stdin = GexecCommandWithStdin(cpiBin, "-configPath", CpiConfigPath)
		stdin.Write([]byte(request))
		stdin.Close()

		Eventually(session.Out, "60s").Should(gbytes.Say(`"error":null`))
		Expect(json.Unmarshal(session.Out.Contents(), &response)).To(Succeed())
		Expect(response["result"]).To(BeNil())
	})
})
