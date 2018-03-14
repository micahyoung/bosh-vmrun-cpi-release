package registry_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"
	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"

	"bosh-govmomi-cpi/iso"
)

var _ = Describe("AgentSettings", func() {
	FIt("returns the registry config", func() {
		logger := &fakelogger.Logger{}

		vmId := "vm-cid"
		stemcellId := "d61a115a-f7ec-4ede-4392-c26da3293453"
		client := iso.NewIsoClient(runner, config, logger)
		result, err := client.Config(stemcellId, vmId)
		Expect(err).ToNot(HaveOccurred())

		expectedResult = `{
		  "vm": {
		  	"name":"vm-d70f1f0d-b9b8-4ad9-b9c0-9cd45853349f",
		    "id":"vm-54443"
		  },
		  "agent_id":"c292e270-d067-45a9-553d-645c9c7c4fd9",
		  "networks":{
		  	"private":{
		  		"cloud_properties":{"name":"scarlet"},
		      "default":["dns", "gateway"],
		      "dns":["8.8.8.8"],
		      "gateway":"10.85.57.1",
		      "ip":"10.85.57.200",
		      "netmask":"255.255.255.0",
		      "type":"manual",
		      "mac":"00:50:56:9a:20:b2"
		  	}
		  },
		  "disks":{
		  	"system":"0",
		    "ephemeral":"1",
		    "persistent":{ "disk-773d31f6-5245-468f-a324-8698b07330a8":"2" }
		  },
		  "ntp":["0.pool.ntp.org", "1.pool.ntp.org"],
		  "blobstore":{
		  	"provider":"local",
		    "options":{"blobstore_path":"/var/vcap/micro_bosh/data/cache"}
		  },
		  "mbus":"https://mbus:mbus-password@0.0.0.0:6868",
		  "env":{}
	  }`
		Expect(result).To(Equal("success"))
	})
})
