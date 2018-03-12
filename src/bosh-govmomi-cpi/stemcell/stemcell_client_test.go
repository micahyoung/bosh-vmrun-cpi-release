package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"bosh-govmomi-cpi/stemcell"
)

var _ = Describe("CreateStemcell", func() {
	It("runs the cpi", func() {
		imagePath := "../../../ci/deploy-test/state/stemcell.tgz"

		logger := boshlog.NewLogger(boshlog.LevelInfo)
		fs := boshsys.NewOsFileSystem(logger)
		cmdRunner := boshsys.NewExecCmdRunner(logger)
		compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)

		client := stemcell.NewClient(compressor, fs, logger)
		ovfPath, err := client.ExtractOvf(imagePath)

		Expect(err).ToNot(HaveOccurred())
		Expect(ovfPath).To(ContainSubstring("image.ovf"))
	})
})
