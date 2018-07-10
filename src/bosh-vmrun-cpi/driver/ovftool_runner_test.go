package driver_test

import (
	. "github.com/onsi/ginkgo"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	"bosh-vmrun-cpi/driver"
)

var _ = Describe("OvftoolRunner", func() {
	var ovftoolRunner driver.OvftoolRunner

	BeforeEach(func() {
		logger := &fakelogger.FakeLogger{}
		boshRunner := fakesys.NewFakeCmdRunner()
		ovftoolRunner = driver.NewOvftoolRunner("ovftool-bin", boshRunner, logger)
	})

	Describe("OvfToVmx", func() {
		It("execs ovftool", func() {

		})
	})

})
