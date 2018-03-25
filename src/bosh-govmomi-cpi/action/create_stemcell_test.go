package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"
	fakestemcell "bosh-govmomi-cpi/stemcell/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"bosh-govmomi-cpi/action"
)

var _ = Describe("CreateStemcell", func() {
	It("runs the cpi", func() {
		stemcellClient := &fakestemcell.FakeStemcellClient{}
		govcClient := &fakegovc.FakeGovcClient{}
		logger := &fakelogger.FakeLogger{}
		uuidGen := &fakeuuid.FakeGenerator{}

		stemcellClient.ExtractOvfReturns("extracted-path", nil)

		m := action.NewCreateStemcellMethod(govcClient, stemcellClient, uuidGen, logger)
		var cid, err = m.CreateStemcell("image-path", nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(cid.AsString()).To(Equal("fake-uuid-0"))

		govcClientImportOvfPath, govcClientImportOvfVmId := govcClient.ImportOvfArgsForCall(0)
		Expect(govcClientImportOvfPath).To(Equal("extracted-path"))
		Expect(govcClientImportOvfVmId).To(Equal("cs-fake-uuid-0"))

		Expect(stemcellClient.ExtractOvfArgsForCall(0)).To(Equal("image-path"))
		Expect(stemcellClient.CleanupCallCount()).To(Equal(1))
	})
})
