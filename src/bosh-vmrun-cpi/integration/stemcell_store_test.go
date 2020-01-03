package integration_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	cpiconfig "bosh-vmrun-cpi/config"
	"bosh-vmrun-cpi/stemcell"
)

var _ = Describe("StemcellStore integration", func() {
	var stemcellStore stemcell.StemcellStore
	BeforeEach(func() {
		generateCPIConfig(CpiConfigPath, DirectCPIConfig)

		logger := boshlog.NewWriterLogger(boshlog.LevelWarn, os.Stderr)
		fs := boshsys.NewOsFileSystem(logger)

		cmdRunner := boshsys.NewExecCmdRunner(logger)

		compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)
		cpiConfigJson, err := fs.ReadFileString(CpiConfigPath)
		Expect(err).ToNot(HaveOccurred())
		cpiConfig, err := cpiconfig.NewConfigFromJson(cpiConfigJson)
		Expect(err).ToNot(HaveOccurred())

		stemcellConfig := stemcell.NewConfig(cpiConfig)

		stemcellStore = stemcell.NewStemcellStore(stemcellConfig, compressor, fs, logger)
	})

	AfterEach(func() {
		stemcellStore.Cleanup()
	})

	Context("GetByMetadata", func() {
		It("finds stemcells by metadata", func() {
			stemcellImagePath, err := stemcellStore.GetByMetadata("bosh-vsphere-esxi-ubuntu-trusty-go_agent", "3586.42")
			Expect(err).ToNot(HaveOccurred())

			Expect(stemcellImagePath).To(BeAnExistingFile())
		})

		It("returns empty when metadata doesnt match", func() {
			stemcellImagePath, err := stemcellStore.GetByMetadata("non-existant-stemcell", "1234.5")
			Expect(err).ToNot(HaveOccurred())

			Expect(stemcellImagePath).To(Equal(""))
		})
	})

	Context("GetByImagePathMapping", func() {
		It("finds stemcells when path exists", func() {
			tmpImageFile, err := ioutil.TempFile("", "image")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpImageFile.Name())

			stemcellImagePath, err := stemcellStore.GetByImagePathMapping(tmpImageFile.Name())
			Expect(err).ToNot(HaveOccurred())

			Expect(stemcellImagePath).To(Equal(tmpImageFile.Name()))
		})

		It("finds stemcells when path does not exist but image path mapping exists", func() {
			stemcellImagePath, err := stemcellStore.GetByImagePathMapping("/test-stemcell-tmp-image-path")
			Expect(err).ToNot(HaveOccurred())

			Expect(stemcellImagePath).To(BeAnExistingFile())
		})

		It("returns empty when path and mapping don't exist", func() {
			stemcellImagePath, err := stemcellStore.GetByImagePathMapping("/non-existing/image")
			Expect(err).ToNot(HaveOccurred())

			Expect(stemcellImagePath).To(Equal(""))
		})
	})
})
