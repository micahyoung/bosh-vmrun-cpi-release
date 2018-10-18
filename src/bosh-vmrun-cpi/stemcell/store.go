package stemcell

import (
	"errors"
	"os"
	"path/filepath"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type stemcellStoreImpl struct {
	storePath     string
	parentTempDir string
	fs            boshsys.FileSystem
	logger        boshlog.Logger
	compressor    boshcmd.Compressor
}

func NewStemcellStore(config Config, compressor boshcmd.Compressor, fs boshsys.FileSystem, logger boshlog.Logger) StemcellStore {
	parentTempDir, _ := fs.TempDir("stemcell-store")
	return stemcellStoreImpl{storePath: config.StemcellStorePath(), parentTempDir: parentTempDir, compressor: compressor, fs: fs, logger: logger}
}

func (s stemcellStoreImpl) GetImagePath(name, version string) (string, error) {
	var err error

	if name == "" || version == "" {
		return "", errors.New("stemcell store requires name and version from cloud properties")
	}

	filePaths, err := s.fs.Glob(filepath.Join(s.storePath, "*.tgz"))
	if err != nil {
		return "", err
	}
	s.logger.DebugWithDetails("stemcell-store", "Stemcell Store files:", filePaths)

	extractedStemcellDir := filepath.Join(s.parentTempDir, "stemcell")
	err = os.Mkdir(extractedStemcellDir, 0700)
	if err != nil {
		return "", err
	}

	for _, filePath := range filePaths {
		err := s.compressor.DecompressFileToDir(filePath, extractedStemcellDir, boshcmd.CompressorOptions{})
		if err != nil {
			return "", err
		}

		imagePath := filepath.Join(extractedStemcellDir, "image")
		manifestPath := filepath.Join(extractedStemcellDir, "stemcell.MF")
		if !s.fs.FileExists(imagePath) || !s.fs.FileExists(manifestPath) {
			//contnue early if manifest or image don't exist
			continue
		}

		manifestContent, err := s.fs.ReadFile(manifestPath)
		if err != nil {
			return "", err
		}

		manifest, err := NewStemcellManifest(manifestContent)
		if err != nil {
			return "", err
		}

		if manifest.Name == name && manifest.Version == version {
			imagePath := filepath.Join(extractedStemcellDir, "image")
			return imagePath, nil
		}
	}

	return "", nil
}

func (s stemcellStoreImpl) Cleanup() {
	os.RemoveAll(s.parentTempDir)
}
