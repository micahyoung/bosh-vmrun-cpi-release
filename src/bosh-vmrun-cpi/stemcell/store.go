package stemcell

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
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
	var extractedImagePath string

	if name == "" || version == "" {
		return "", errors.New("stemcell store requires name and version from cloud properties")
	}

	filePaths, err := s.fs.Glob(filepath.Join(s.storePath, "*.tgz"))
	if err != nil {
		return "", err
	}

	s.logger.DebugWithDetails("stemcell-store", "Stemcell Store files:", filePaths)

	for _, filePath := range filePaths {
		var manifest *StemcellManifest

		err = s.WithTarballFile(filePath, "stemcell.MF", func(manifestReader io.Reader) error {
			var manifestContent []byte
			if manifestContent, err = ioutil.ReadAll(manifestReader); err != nil {
				return err
			}

			if manifest, err = NewStemcellManifest(manifestContent); err != nil {
				return err
			}

			return nil
		})

		if err == io.ErrUnexpectedEOF || err == io.EOF {
			s.logger.Warn("stemcell-store", "skipping invalid stemcell file %s:", filePath)
			continue
		}

		if err != nil {
			return "", err
		}

		if manifest == nil {
			s.logger.Warn("stemcell-store", "skipping invalid stemcell with no manifest:", filePath)
			continue
		}

		if manifest.Name == name && manifest.Version == version {
			err = s.WithTarballFile(filePath, "image", func(imageReader io.Reader) error {
				extractedImagePath = filepath.Join(s.parentTempDir, "image")
				var extractedImageFile *os.File
				if extractedImageFile, err = os.Create(extractedImagePath); err != nil {
					return err
				}
				defer extractedImageFile.Close()

				var bytesWritten int64
				if bytesWritten, err = io.Copy(extractedImageFile, imageReader); err != nil {
					return err
				}

				s.logger.Debug("stemcell-store", "wrote stemcell image to %s: %d bytes", extractedImagePath, bytesWritten)

				return nil
			})

			//exit if image was extracted
			break
		}

	}

	return extractedImagePath, nil
}

func (s stemcellStoreImpl) WithTarballFile(tarGzPath string, desiredTarballFilePath string, callback func(io.Reader) error) error {
	var err error
	var gzipFile *os.File

	if gzipFile, err = os.Open(tarGzPath); err != nil {
		return err
	}
	defer gzipFile.Close()

	var gzipReader *gzip.Reader
	gzipReader, err = gzip.NewReader(gzipFile)

	if err != nil {
		return err
	}

	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader) // no Close()

	for {
		var tarHeader *tar.Header
		tarHeader, err = tarReader.Next()

		if err == io.EOF {
			break // End of archive
		}

		if err != nil {
			return err
		}

		tarHeaderFilePath := tarHeader.Name

		switch tarHeader.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			s.logger.Debug("stemcell-store", "stemcell content file %s", tarHeaderFilePath)

			if tarHeaderFilePath == desiredTarballFilePath {

				//call callback and exit early if file was found
				return callback(tarReader)
			}
		default:
			s.logger.Error("stemcell-store", "unable to determine file type %c in file %s", tarHeader.Typeflag, tarHeaderFilePath)
		}
	}

	return nil
}

func (s stemcellStoreImpl) Cleanup() {
	os.RemoveAll(s.parentTempDir)
}
