package main_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
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
	FIt("runs the cpi", func() {
		configFile, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
		defer configFile.Close()

		cpiBin, err := gexec.Build("bosh-govmomi-cpi/main")
		Expect(err).ToNot(HaveOccurred())

		configFile.WriteString(configContent)
		configPath, err := filepath.Abs(configFile.Name())

		command := exec.Command(cpiBin, "-configPath", configPath)
		stdin := &bytes.Buffer{}
		command.Stdin = stdin

		session, err := gexec.Start(command, GinkgoWriter, os.Stderr)
		Expect(err).ToNot(HaveOccurred())

		var stemcellFile = "../../../ci/deploy-test/state/stemcell.tgz"
		var stemcellVersion = `bar`
		request := fmt.Sprintf(`{
			"method": "create_stemcell",
			"arguments": ["%s", {
				"name": "bosh-vsphere-kvm-ubuntu-trusty",
				"version": "%s",
				"infrastructure": "vsphere"
			}]
		}`, stemcellFile, stemcellVersion)

		io.WriteString(stdin, request)

		Eventually(session.Out, "30s").Should(gbytes.Say("prefix"))

		err = os.Remove(configFile.Name())
		Expect(err).NotTo(HaveOccurred())

	})
})
