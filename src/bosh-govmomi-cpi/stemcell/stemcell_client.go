package stemcell

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type StemcellClientImpl struct {
	fs            boshsys.FileSystem
	logger        boshlog.Logger
	compressor    boshcmd.Compressor
	parentTempDir string
}

func NewClient(compressor boshcmd.Compressor, fs boshsys.FileSystem, logger boshlog.Logger) StemcellClient {
	return StemcellClientImpl{compressor: compressor, fs: fs, logger: logger}
}

func (c StemcellClientImpl) ExtractOvf(stemcellTarballPath string) (string, error) {
	var err error

	if !c.fs.FileExists(stemcellTarballPath) {
		return "", bosherr.Error("stemcell not found")
	}

	c.parentTempDir, err = c.fs.TempDir("stemcell-")
	if err != nil {
		return "", bosherr.WrapError(err, "creating tempdir failed")
	}

	err = c.compressor.DecompressFileToDir(stemcellTarballPath, c.parentTempDir, boshcmd.CompressorOptions{})
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Unpacking outer stemcell tarball '%s' to '%s'", stemcellTarballPath, c.parentTempDir)
	}

	imageOvfPath := filepath.Join(c.parentTempDir, "image.ovf")
	if !c.fs.FileExists(imageOvfPath) {
		return "", bosherr.Error("stemcell does not contain 'image.ovf'")
	}

	return imageOvfPath, nil
}

func (c StemcellClientImpl) Cleanup() {
	err := c.fs.RemoveAll(c.parentTempDir)
	if err != nil {
		c.logger.Error("stemcell-client", "Cleaning up stemcell temp dir '%s'", c.parentTempDir)
	}
}
