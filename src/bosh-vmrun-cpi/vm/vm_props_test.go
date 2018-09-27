package vm_test

import (
	"bosh-vmrun-cpi/vm"
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("VMProps", func() {
	Describe("Initialize", func() {
		Describe("override", func() {
			It("sets the duration fields", func() {
				propsJson := []byte(`{
					"CPU": 2,
					"RAM": 2048,
					"Disk": 10000,
					"Bootstrap": {
						"Script_Content": "foo",
						"Script_Path": "bar",
						"Interpreter_Path": "baz",
						"Ready_Process_Name": "qux",
						"Username": "jane",
						"Password": "doe",
						"Min_Wait_Seconds": 10,
						"Max_Wait_Seconds": 20
					}
				}`)

				var cloudProps apiv1.CloudPropsImpl
				var vmProps vm.VMProps
				Expect(json.Unmarshal(propsJson, &cloudProps)).To(Succeed())
				Expect(cloudProps.As(&vmProps)).To(Succeed())
				vmProps.Initialize()

				Expect(vmProps.CPU).To(Equal(2))
				Expect(vmProps.RAM).To(Equal(2048))
				Expect(vmProps.Disk).To(Equal(10000))
				Expect(vmProps.Bootstrap.Script_Content).To(Equal("foo"))
				Expect(vmProps.Bootstrap.Script_Path).To(Equal("bar"))
				Expect(vmProps.Bootstrap.Interpreter_Path).To(Equal("baz"))
				Expect(vmProps.Bootstrap.Ready_Process_Name).To(Equal("qux"))
				Expect(vmProps.Bootstrap.Username).To(Equal("jane"))
				Expect(vmProps.Bootstrap.Password).To(Equal("doe"))
				Expect(vmProps.Bootstrap.Min_Wait).To(Equal(10 * time.Second))
				Expect(vmProps.Bootstrap.Max_Wait).To(Equal(20 * time.Second))
			})
		})
	})
})
