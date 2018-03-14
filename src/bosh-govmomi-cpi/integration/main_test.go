package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	It("runs the cpi", func() {
		configFile, err := ioutil.TempFile("", "config")
		Expect(err).ToNot(HaveOccurred())

		configFile.WriteString(configContent)
		configPath, err := filepath.Abs(configFile.Name())
		defer os.Remove(configPath)

		cpiBin, err := gexec.Build("bosh-govmomi-cpi/main")
		Expect(err).ToNot(HaveOccurred())

		command := exec.Command(cpiBin, "-configPath", configPath)
		stdin, err := command.StdinPipe()
		Expect(err).ToNot(HaveOccurred())

		session, err := gexec.Start(command, GinkgoWriter, os.Stderr)
		Expect(err).ToNot(HaveOccurred())
		time.Sleep(2)

		imageTarballPath := filepath.Join(extractedStemcellTempDir, "image")

		request := fmt.Sprintf(`{
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

		fmt.Printf("CONFIG: %s\n", request)
		stdin.Write([]byte(request))

		err = stdin.Close()
		Expect(err).ShouldNot(HaveOccurred())

		Eventually(session.Out, "60s").Should(gbytes.Say(`"result"`))
	})
})
