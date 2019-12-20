package stemcell_test

import (
	"archive/tar"
	"bosh-vmrun-cpi/stemcell"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

	Describe("GetByMetadata", func() {
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
					tarGzFile, err := os.Create(filepath.Join(storeDir, "valid-stecemll.tgz"))
					Expect(err).ToNot(HaveOccurred())

					gzipWriter := gzip.NewWriter(tarGzFile)
					tarWriter := tar.NewWriter(gzipWriter)

					manifestContent := strings.TrimSpace(`
name: bosh-vsphere-esxi-ubuntu-xenial-go_agent
version: "97.18"
					`)

					manifestHeader := &tar.Header{
						Name: "stemcell.MF",
						Size: int64(len(manifestContent)),
						Mode: 0666,
					}

					Expect(tarWriter.WriteHeader(manifestHeader)).To(Succeed())
					_, err = io.Copy(tarWriter, bytes.NewBuffer([]byte(manifestContent)))
					Expect(err).ToNot(HaveOccurred())

					imageHeader := &tar.Header{
						Name: "image",
						Size: 0,
						Mode: 0666,
					}
					Expect(tarWriter.WriteHeader(imageHeader)).To(Succeed())

					Expect(gzipWriter.Close()).To(Succeed())
					Expect(tarWriter.Close()).To(Succeed())
				})

				Context("with valid params", func() {
					It("returns path to the extracted stemcell image", func() {
						imagePath, err := stemcellStore.GetByMetadata("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "97.18")
						defer stemcellStore.Cleanup()

						Expect(err).ToNot(HaveOccurred())
						Expect(imagePath).To(HaveSuffix(filepath.Join("image")))
					})
				})

				Context("with non-matching params", func() {
					It("returns an error", func() {
						imagePath, err := stemcellStore.GetByMetadata("", "")
						defer stemcellStore.Cleanup()

						Expect(err).To(HaveOccurred())
						Expect(imagePath).To(Equal(""))
					})
				})
			})

			Context("when no stemcell exists", func() {
				It("returns empty string", func() {
					imagePath, err := stemcellStore.GetByMetadata("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "foobar")
					defer stemcellStore.Cleanup()

					Expect(err).ToNot(HaveOccurred())
					Expect(imagePath).To(Equal(""))
				})

				It("ignores empty files", func() {
					emptyFile, err := os.Create(filepath.Join(storeDir, "invalid.tgz"))
					Expect(err).ToNot(HaveOccurred())
					Expect(emptyFile.Close()).To(Succeed())

					imagePath, err := stemcellStore.GetByMetadata("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "foobar")
					defer stemcellStore.Cleanup()

					Expect(err).ToNot(HaveOccurred())
					Expect(imagePath).To(Equal(""))
				})

				It("ignores tar/gz without manifest", func() {
					tarGzFile, err := os.Create(filepath.Join(storeDir, "no-manifest.tgz"))
					Expect(err).ToNot(HaveOccurred())

					gzipWriter := gzip.NewWriter(tarGzFile)
					tarWriter := tar.NewWriter(gzipWriter)

					header := &tar.Header{
						Name: "empty",
						Size: 0,
						Mode: 0666,
					}

					Expect(tarWriter.WriteHeader(header)).To(Succeed())

					Expect(gzipWriter.Close()).To(Succeed())
					Expect(tarWriter.Close()).To(Succeed())

					imagePath, err := stemcellStore.GetByMetadata("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "foobar")
					defer stemcellStore.Cleanup()

					Expect(err).ToNot(HaveOccurred())
					Expect(imagePath).To(Equal(""))
				})

				It("ignores gzip files without tar", func() {
					var gzipBuffer bytes.Buffer
					gzipWriter := gzip.NewWriter(&gzipBuffer)

					_, err = gzipWriter.Write([]byte("hello"))
					Expect(err).ToNot(HaveOccurred())

					Expect(gzipWriter.Close()).To(Succeed())

					err = ioutil.WriteFile(filepath.Join(storeDir, "invalid.tgz"), gzipBuffer.Bytes(), 0666)
					Expect(err).ToNot(HaveOccurred())

					imagePath, err := stemcellStore.GetByMetadata("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "foobar")
					defer stemcellStore.Cleanup()

					Expect(err).ToNot(HaveOccurred())
					Expect(imagePath).To(Equal(""))
				})

				It("ignores directories", func() {
					err = os.Mkdir(filepath.Join(storeDir, "some-dir"), 0777)
					Expect(err).ToNot(HaveOccurred())

					imagePath, err := stemcellStore.GetByMetadata("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "foobar")
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
				imagePath, err := stemcellStore.GetByMetadata("bosh-vsphere-esxi-ubuntu-xenial-go_agent", "foobar")
				defer stemcellStore.Cleanup()

				Expect(err).ToNot(HaveOccurred())
				Expect(imagePath).To(Equal(""))
			})
		})
	})
})
