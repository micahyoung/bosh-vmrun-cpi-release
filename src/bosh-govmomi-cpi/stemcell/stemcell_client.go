package stemcell

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type StemcellClient struct {
	fs         boshsys.FileSystem
	logger     boshlog.Logger
	compressor boshcmd.Compressor
}

func NewClient(compressor boshcmd.Compressor, fs boshsys.FileSystem, logger boshlog.Logger) StemcellClient {
	return StemcellClient{compressor: compressor, fs: fs, logger: logger}
}

func (c StemcellClient) ExtractOvf(stemcellTarballPath string) (string, error) {
	var err error

	if !c.fs.FileExists(stemcellTarballPath) {
		return "", bosherr.Error("stemcell not found")
	}

	stemcellTempDir, err := c.fs.TempDir("stemcell-")
	if err != nil {
		return "", bosherr.WrapError(err, "creating tempdirs for outer stemcell tarball")
	}
	c.logger.Debug("creating tempdirs for outer stemcell tarball", stemcellTempDir)

	c.logger.Debug("Unpacking outer stemcell tarball '%s' to '%s'", stemcellTarballPath, stemcellTempDir)
	err = c.compressor.DecompressFileToDir(stemcellTarballPath, stemcellTempDir, boshcmd.CompressorOptions{})
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Unpacking outer stemcell tarball '%s' to '%s'", stemcellTarballPath, stemcellTempDir)
	}
	defer c.fs.RemoveAll(stemcellTempDir)

	imageTarballPath := filepath.Join(stemcellTempDir, "image")
	imageTempDir, err := c.fs.TempDir("image-")
	if err != nil {
		return "", bosherr.WrapError(err, "creating tempdirs for inner stemcell tarball")
	}
	//defer c.fs.RemoveAll(imageTempDir)

	err = c.compressor.DecompressFileToDir(imageTarballPath, imageTempDir, boshcmd.CompressorOptions{})
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Unpacking inner stemcell tarball '%s' to '%s'", imageTarballPath, imageTempDir)
	}

	imageOvfPath := filepath.Join(imageTempDir, "image.ovf")

	return imageOvfPath, nil
}
