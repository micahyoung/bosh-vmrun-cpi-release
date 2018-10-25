package stemcell_test

import (
	"bosh-vmrun-cpi/stemcell"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakes "bosh-vmrun-cpi/stemcell/fakes"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

var _ = Describe("StemcellStore", func() {
	var (
		err           error
		stemcellStore stemcell.StemcellStore
		storeDir      string
		config        *fakes.FakeConfig
		logger        *fakelogger.FakeLogger
		cmdRunner     boshsys.CmdRunner
		fs            boshsys.FileSystem
		compressor    boshcmd.Compressor
	)

	BeforeEach(func() {
		config = &fakes.FakeConfig{}

		logger = &fakelogger.FakeLogger{}
		cmdRunner = boshsys.NewExecCmdRunner(logger)
		fs = boshsys.NewOsFileSystem(logger)
		compressor = boshcmd.NewTarballCompressor(cmdRunner, fs)
	})

	Describe("GetImagePath", func() {
		Context("when store path exists", func() {
			BeforeEach(func() {
				storeDir, err = ioutil.TempDir("", "stemcell-store-")
				Expect(err).ToNot(HaveOccurred())

				config.StemcellStorePathReturns(storeDir)

				stemcellStore = stemcell.NewStemcellStore(config, compressor, fs, logger)
			})

			AfterEach(func() {
				os.RemoveAll(storeDir)
			})

			Context("when stemcell exists", func() {
				BeforeEach(func() {
					stemcellSourcePath := filepath.Join("..", "test", "fixtures", "stemcell-store", "stemcell.tgz")
					stemcellDestPath := filepath.Join(storeDir, "stemcell.tgz")

					stemcellData, err := ioutil.ReadFile(stemcellSourcePath)
					Expect(err).ToNot(HaveOccurred())

					err = ioutil.WriteFile(stemcellDestPath, stemcellData, 0644)
					Expect(err).ToNot(HaveOccurred())

					//invalid tarball
					err = ioutil.WriteFile(filepath.Join(storeDir, "invalid.tgz"), []byte(""), 0777)
					Expect(err).ToNot(HaveOccurred())

					//sibling directory
					err = os.Mkdir(filepath.Join(storeDir, "some-dir"), 0777)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("with valid params", func() {
					It("returns path to the extracted stemcell image", func() {
						imagePath, err := stemcellStore.GetImagePath("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "97.18")
						defer stemcellStore.Cleanup()

						Expect(err).ToNot(HaveOccurred())
						Expect(imagePath).To(HaveSuffix(filepath.Join("stemcell", "image")))
					})
				})

				Context("with invalid params", func() {
					It("returns an error", func() {
						imagePath, err := stemcellStore.GetImagePath("", "")
						defer stemcellStore.Cleanup()

						Expect(err).To(HaveOccurred())
						Expect(imagePath).To(Equal(""))
					})
				})
			})

			Context("when no stemcell exists", func() {
				It("returns empty string", func() {
					imagePath, err := stemcellStore.GetImagePath("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "foobar")
					defer stemcellStore.Cleanup()

					Expect(err).ToNot(HaveOccurred())
					Expect(imagePath).To(Equal(""))
				})
			})
		})

		Context("when store path does not exist", func() {
			BeforeEach(func() {
				storeDir := filepath.Join("a", "fake", "dir")
				config.StemcellStorePathReturns(storeDir)
				stemcellStore = stemcell.NewStemcellStore(config, compressor, fs, logger)
			})

			It("returns empty string", func() {
				imagePath, err := stemcellStore.GetImagePath("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "foobar")
				defer stemcellStore.Cleanup()

				Expect(err).ToNot(HaveOccurred())
				Expect(imagePath).To(Equal(""))
			})
		})
	})
})
