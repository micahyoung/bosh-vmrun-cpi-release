package install

import (
	"crypto/sha1"
	"fmt"
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

		mappingFileContents := remoteStemcellPath
		tempMappingFile, err := ioutil.TempFile("", remoteMappingFileName)
		if err != nil {
			return err
		}
		tempMappingFile.Close()
		tempMappingFilePath := tempMappingFile.Name()
		err = ioutil.WriteFile(tempMappingFilePath, []byte(mappingFileContents), 0666)
		if err != nil {
			return err
		}

		err = sshCopy(tempMappingFilePath, remoteMappingPath, i.sshClient, i.logger)
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
		tempTarFilePath, err := i.compressor.CompressFilesInDir(parentDirPath)
		if err != nil {
			i.logger.ErrorWithDetails("sync-director-stemcells", "tar failure", tempTarFilePath)
			return err
		}
		defer i.compressor.CleanUp(tempTarFilePath)

		err = sshCopy(tempTarFilePath, remoteStemcellPath, i.sshClient, i.logger)
		if err != nil {
			i.logger.ErrorWithDetails("sync-director-stemcells", "ssh failure", err)
			return err
		}
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
