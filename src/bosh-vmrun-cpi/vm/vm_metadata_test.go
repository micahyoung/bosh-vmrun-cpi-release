package vm_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bosh-vmrun-cpi/vm"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("VMMetadata", func() {
	Describe("NewVMMetadata", func() {
		It("sets the display name", func() {
			propsJson := []byte(`{"created_at":"2019-12-06T16:20:45-05:00","deployment":"bosh","director":"bosh-init","index":"0","instance_group":"director","job":"bosh","name":"bosh/0"}`)

			var apiMeta apiv1.VMMeta
			Expect(json.Unmarshal(propsJson, &apiMeta)).To(Succeed())

			vmMeta, err := vm.NewVMMetadata(apiMeta)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmMeta.DisplayName("vm-foo")).To(Equal("director_bosh_0"))
		})
	})
})
