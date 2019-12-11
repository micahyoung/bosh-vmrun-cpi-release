package action_test

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakedriver "bosh-vmrun-cpi/driver/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"

	"bosh-vmrun-cpi/action"
)

var _ = Describe("SetVMMetadata", func() {
	It("runs the cpi", func() {
		driverClient := &fakedriver.FakeClient{}
		logger := &fakelogger.FakeLogger{}

		vmMeta := apiv1.VMMeta{}
		vmMeta.UnmarshalJSON([]byte(`{
			"director": "director-784430",
			"deployment": "redis",
			"name": "redis/ce7d2040-212e-4d5a-a62d-952a12c50741",
			"instance_group": "cache",
			"job": "redis",
			"id": "ce7d2040-212e-4d5a-a62d-952a12c50741",
			"index": "1"
  		}`))

		vmCid := apiv1.NewVMCID("foo")

		driverClient.NeedsVMNameChangeReturns(true)
		driverClient.SetVMDisplayNameReturns(nil)
		driverClient.StartVMReturns(nil)

		m := action.NewSetVMMetadataMethod(driverClient, logger)

		err := m.SetVMMetadata(vmCid, vmMeta)
		Expect(err).ToNot(HaveOccurred())

		driverVMID, driverVMDisplayName := driverClient.SetVMDisplayNameArgsForCall(0)

		Expect(driverVMID).To(Equal("vm-foo"))
		Expect(driverVMDisplayName).To(Equal("cache_redis_1"))

		driverVMID = driverClient.StartVMArgsForCall(0)

		Expect(driverVMID).To(Equal("vm-foo"))

		Expect(err).ToNot(HaveOccurred())
	})
})
