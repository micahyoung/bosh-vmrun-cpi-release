package integration_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"bosh-govmomi-cpi/stemcell"
)

var _ = Describe("StemcellClient integration", func() {
	It("runs the stemcell conversion", func() {
		imageTarballPath := filepath.Join(extractedStemcellTempDir, "image")

		logger := boshlog.NewLogger(boshlog.LevelInfo)
		fs := boshsys.NewOsFileSystem(logger)
		cmdRunner := boshsys.NewExecCmdRunner(logger)
		compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)

		client := stemcell.NewClient(compressor, fs, logger)
		ovfPath, err := client.ExtractOvf(imageTarballPath)
		Expect(err).ToNot(HaveOccurred())

		err = client.Cleanup()
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(ovfPath)

		Expect(ovfPath).To(ContainSubstring("image.ovf"))
	})
})
