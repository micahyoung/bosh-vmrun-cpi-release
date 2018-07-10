package integration_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"bosh-vmrun-cpi/stemcell"
)

var _ = Describe("StemcellClient integration", func() {
	It("runs the stemcell conversion", func() {
		imageTarballPath := filepath.Join(ExtractedStemcellTempDir, "image")

		logger := boshlog.NewLogger(boshlog.LevelInfo)
		fs := boshsys.NewOsFileSystem(logger)
		cmdRunner := boshsys.NewExecCmdRunner(logger)
		compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)

		client := stemcell.NewClient(compressor, fs, logger)
		ovfPath, err := client.ExtractOvf(imageTarballPath)
		Expect(err).ToNot(HaveOccurred())

		client.Cleanup()

		Expect(ovfPath).To(ContainSubstring("image.ovf"))
	})
})
