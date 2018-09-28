package vm_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bosh-vmrun-cpi/vm"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("VMProps", func() {
	Describe("NewVMProps", func() {
		Describe("defaults", func() {
			It("sets the duration fields", func() {
				propsJson := []byte(`{}`)

				var cloudProps apiv1.CloudPropsImpl
				Expect(json.Unmarshal(propsJson, &cloudProps)).To(Succeed())

				vmProps, err := vm.NewVMProps(cloudProps)
				Expect(err).ToNot(HaveOccurred())

				Expect(vmProps.CPU).To(Equal(1))
				Expect(vmProps.RAM).To(Equal(1024))
				Expect(vmProps.Disk).To(Equal(0))
				Expect(vmProps.Bootstrap.Script_Content).To(Equal(""))
				Expect(vmProps.Bootstrap.Script_Path).To(Equal(""))
				Expect(vmProps.Bootstrap.Interpreter_Path).To(Equal(""))
				Expect(vmProps.Bootstrap.Ready_Process_Name).To(Equal(""))
				Expect(vmProps.Bootstrap.Username).To(Equal(""))
				Expect(vmProps.Bootstrap.Password).To(Equal(""))
				Expect(vmProps.Bootstrap.Min_Wait).To(Equal(time.Duration(0)))
				Expect(vmProps.Bootstrap.Max_Wait).To(Equal(600 * time.Second))
			})
		})

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
				Expect(json.Unmarshal(propsJson, &cloudProps)).To(Succeed())

				vmProps, err := vm.NewVMProps(cloudProps)
				Expect(err).ToNot(HaveOccurred())

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
