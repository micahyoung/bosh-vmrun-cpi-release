package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakedriver "bosh-vmrun-cpi/driver/fakes"
	fakestemcell "bosh-vmrun-cpi/stemcell/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"bosh-vmrun-cpi/action"
)

var _ = Describe("CreateStemcell", func() {
	It("runs the cpi", func() {
		stemcellClient := &fakestemcell.FakeStemcellClient{}
		driverClient := &fakedriver.FakeClient{}
		logger := &fakelogger.FakeLogger{}
		uuidGen := &fakeuuid.FakeGenerator{}

		stemcellClient.ExtractOvfReturns("extracted-path", nil)

		m := action.NewCreateStemcellMethod(driverClient, stemcellClient, uuidGen, logger)
		var cid, err = m.CreateStemcell("image-path", nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(cid.AsString()).To(Equal("fake-uuid-0"))

		clientImportOvfPath, clientImportOvfVmId := driverClient.ImportOvfArgsForCall(0)
		Expect(clientImportOvfPath).To(Equal("extracted-path"))
		Expect(clientImportOvfVmId).To(Equal("cs-fake-uuid-0"))

		Expect(stemcellClient.ExtractOvfArgsForCall(0)).To(Equal("image-path"))
		Expect(stemcellClient.CleanupCallCount()).To(Equal(1))
	})
})
