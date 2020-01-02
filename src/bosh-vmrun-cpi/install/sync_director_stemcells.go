package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

func (i *installerImpl) SyncDirectorStemcells(directorTmpDirPath string) error {
	remoteStemcellStorePath := stemcellStorePath(i.cpiConfig)
	remotePathJoin := remotePathJoinFunc(i.cpiConfig)
	localPathJoin := localPathJoinFunc()

	if fileInfo, err := os.Stat(directorTmpDirPath); err != nil || !fileInfo.IsDir() {
		i.logger.Error("sync-director-stemcells", "director tmp path does not exist", err)
		return err
	}

	stemcellManifestPaths, err := filepath.Glob(localPathJoin(directorTmpDirPath, "*", "stemcell.MF"))
	if err != nil {
		return err
	}
	if len(stemcellManifestPaths) == 0 {
		i.logger.Info("sync-director-stemcells", "no stemcells found in %s", directorTmpDirPath)
		return nil
	}

	for _, manifestPath := range stemcellManifestPaths {
		i.logger.Debug("sync-director-stemcells", "found stemcell manifest: %s", manifestPath)
		stemcellFileName, err := parseStemcellFileName(manifestPath)
		if err != nil {
			i.logger.ErrorWithDetails("sync-director-stemcells", "parsing manifest for name", manifestPath)
			return err
		}
		remoteStemcellPath := remotePathJoin(remoteStemcellStorePath, stemcellFileName)
		matchingImagePath := localPathJoin(filepath.Dir(manifestPath), "image")
		remoteMappingFileName := fmt.Sprintf("%x.mapping", sha1.Sum([]byte(matchingImagePath)))
		remoteMappingPath := remotePathJoin(stemcellMappingsPath(i.cpiConfig), remoteMappingFileName)

		mappingFileContents := []byte(remoteStemcellPath)
		mappingFileReader := bytes.NewBuffer(mappingFileContents)
		err = sshCopy(mappingFileReader, remoteMappingPath, int64(len(mappingFileContents)), i.sshClient, i.logger)
		if err != nil {
			i.logger.ErrorWithDetails("sync-director-stemcells", "ssh failure", err)
			return err
		}
		i.logger.Debug("sync-director-stemcells", "generated mapping: %s for image path: %s", remoteMappingPath, matchingImagePath)

		exists, err := testRemoteStemcell(remoteStemcellPath, i.sshClient, i.logger)
		if err != nil {
			i.logger.Error("sync-director-stemcells", "testing remote stemcell", err)
			return err
		}
		if exists {
			i.logger.Debug("sync-director-stemcells", "remote stemcell already exists: %s\n", stemcellFileName)
			continue
		}

		parentDirPath := filepath.Dir(manifestPath)

		i.logger.Debug("sync-director-stemcells", "creating tar reader for path: %s\n", parentDirPath)

		srcSizeReader, srcSizeWriter := io.Pipe()
		go func() {
			defer srcSizeWriter.Close()
			err := tarGzipWrite(parentDirPath, srcSizeWriter, i.logger)
			if err != nil {
				i.logger.Error("sync-director-stemcells", "gofunc: copying tar file to calculate size", err)
			}
		}()

		srcSize, err := io.Copy(ioutil.Discard, srcSizeReader)
		if err != nil {
			return err
		}

		i.logger.Debug("sync-director-stemcells", "tar reader size: %d", srcSize)

		i.logger.Debug("sync-director-stemcells", "copying tar file to remote")
		srcContentReader, srcContentWriter := io.Pipe()
		go func() {
			defer srcContentWriter.Close()
			err := tarGzipWrite(parentDirPath, srcContentWriter, i.logger)
			if err != nil {
				i.logger.Error("sync-director-stemcells", "gofunc: copying tar file to remote", err)
			}
		}()

		err = sshCopy(srcContentReader, remoteStemcellPath, srcSize, i.sshClient, i.logger)
		if err != nil {
			i.logger.ErrorWithDetails("sync-director-stemcells", "ssh failure", err)
			return err
		}

		i.logger.Debug("sync-director-stemcells", "copied tar file")
	}

	return nil
}

func parseStemcellFileName(manifestPath string) (string, error) {
	var filename string
	var err error
	var manifestData struct {
		Name    string
		Version string
	}

	manifestContent, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(manifestContent, &manifestData)
	if err != nil {
		return "", err
	}

	stemcellType := strings.Replace(manifestData.Name, "bosh-", "", 1)
	filename = fmt.Sprintf("bosh-stemcell-%s-%s.tgz", manifestData.Version, stemcellType)
	return filename, nil
}

func testRemoteStemcell(remoteStemcellPath string, client *ssh.Client, logger boshlog.Logger) (existing bool, err error) {
	session, err := client.NewSession()
	if err != nil {
		return false, err
	}

	logger.Debug("sync-director-stemcells", "testing existing remote stemcell: %s", remoteStemcellPath)
	testStemcellOutputBytes, err := session.CombinedOutput(fmt.Sprintf("tar -t -z -f %s", remoteStemcellPath))
	testStemcellOutput := string(testStemcellOutputBytes)

	if strings.Contains(testStemcellOutput, "stemcell.MF") {
		return true, nil
	}
	return false, nil
}

func tarGzipWrite(src string, writer io.Writer, logger boshlog.Logger) error {
	gzipWriter, _ := gzip.NewWriterLevel(writer, gzip.NoCompression)
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		tarWriter.Close()
		gzipWriter.Close()
	}()

	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	// walk path
	err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		logger.Debug("sync-director-stemcells", "add tar file: %s", file)

		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tarWriter, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})

	if err != nil {
		return err
	}

	logger.Debug("sync-director-stemcells", "created in-memory tar")

	return nil
}
