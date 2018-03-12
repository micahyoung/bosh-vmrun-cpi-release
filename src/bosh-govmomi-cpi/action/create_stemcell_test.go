package action_test

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bosh-govmomi-cpi/action"
)

var _ = Describe("CreateStemcell", func() {
	It("runs the cpi", func() {
		imagePath := "foo"
		cpi := &action.CPI{}
		var cid, err = cpi.CreateStemcell(imagePath, nil)
		Expect(err).ToNot(HaveOccurred())

		var expectedCID = apiv1.NewStemcellCID("stemcell-cid")
		Expect(cid).To(Equal(expectedCID))
	})
})
