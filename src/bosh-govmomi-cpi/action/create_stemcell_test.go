package action_test

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakegovc "bosh-govmomi-cpi/govc/fakes"
	fakestemcell "bosh-govmomi-cpi/stemcell/fakes"
	"github.com/cloudfoundry/bosh-utils/logger/loggerfakes"

	"bosh-govmomi-cpi/action"
)

var _ = Describe("CreateStemcell", func() {
	It("runs the cpi", func() {
		stemcellClient := &fakestemcell.FakeStemcellClient{}
		govcClient := &fakegovc.FakeGovcClient{}
		logger := &loggerfakes.FakeLogger{}

		stemcellClient.ExtractOvfReturns("bar", nil)

		m := action.NewCreateStemcellMethod(govcClient, stemcellClient, logger)
		var cid, err = m.CreateStemcell("foo", nil)
		Expect(err).ToNot(HaveOccurred())

		var expectedCID = apiv1.NewStemcellCID("stemcell-cid")
		Expect(cid).To(Equal(expectedCID))

		Expect(stemcellClient.ExtractOvfArgsForCall(0)).To(Equal("foo"))
		Expect(govcClient.ImportOvfArgsForCall(0)).To(Equal("bar"))
		Expect(stemcellClient.CleanupCallCount()).To(Equal(1))
	})
})
