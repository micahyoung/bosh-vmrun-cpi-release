package action_test

import (
	"encoding/json"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakedriver "bosh-vmrun-cpi/driver/fakes"
	fakestemcell "bosh-vmrun-cpi/stemcell/fakes"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"bosh-vmrun-cpi/action"
)

var _ = Describe("CreateStemcell", func() {
	It("runs the cpi", func() {
		driverClient := &fakedriver.FakeClient{}
		stemcellClient := &fakestemcell.FakeStemcellClient{}
		stemcellStore := &fakestemcell.FakeStemcellStore{}
		logger := &fakelogger.FakeLogger{}
		fs := fakesys.NewFakeFileSystem()
		uuidGen := &fakeuuid.FakeGenerator{}

		stemcellStore.GetImagePathReturns("", nil)
		stemcellClient.ExtractOvfReturns("extracted-path", nil)

		var resourceCloudProps apiv1.CloudPropsImpl
		json.Unmarshal([]byte(`{}`), &resourceCloudProps)

		m := action.NewCreateStemcellMethod(driverClient, stemcellClient, stemcellStore, uuidGen, fs, logger)
		imageFile, _ := fs.TempFile("image")
		imagePath := imageFile.Name()
		var cid, err = m.CreateStemcell(imagePath, resourceCloudProps)
		Expect(err).ToNot(HaveOccurred())

		Expect(cid.AsString()).To(Equal("fake-uuid-0"))

		clientImportOvfPath, clientImportOvfVmId := driverClient.ImportOvfArgsForCall(0)
		Expect(clientImportOvfPath).To(Equal("extracted-path"))
		Expect(clientImportOvfVmId).To(Equal("cs-fake-uuid-0"))

		Expect(stemcellClient.ExtractOvfArgsForCall(0)).To(Equal(imagePath))
		Expect(stemcellClient.CleanupCallCount()).To(Equal(1))
	})
})
