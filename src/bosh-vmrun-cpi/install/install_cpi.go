package install

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"golang.org/x/crypto/ssh"
)

func (i *installerImpl) InstallCPI(version string) error {
	var err error

	cpiSrcPath := cpiSourcePath(i.cpiConfig)
	if _, err := os.Stat(cpiSrcPath); os.IsNotExist(err) {
		i.logger.Error("install-cpi", "opening CPI source file", err)
		return err
	}

	cpiDestPath := cpiDestPath(i.cpiConfig)

	matching, err := compareCPIVersionsSSH(version, cpiSrcPath, cpiDestPath, i.sshClient, i.logger)
	if err != nil {
		i.logger.Error("install-cpi", "comparing existing CPI version over ssh", err)
		return err
	}

	if matching {
		i.logger.Debug("install-cpi", "Using existing remote cpi")
	} else {
		i.logger.Debug("install-cpi", "Installing remote cpi")
		err = sshCopy(cpiSrcPath, cpiDestPath, i.sshClient, i.logger)
		if err != nil {
			//TODO handle failed install due to limited authorized key. Return manual instructions
			i.logger.Error("install-cpi", "creating CPI destination file over SSH", err)
			return err
		}
	}

	//output path
	escapedCpiDestPath := strings.Trim(strconv.Quote(cpiDestPath), `"`)
	fmt.Println(escapedCpiDestPath)
	return nil
}

func compareCPIVersionsSSH(version, cpiSrcPath, cpiDestPath string, client *ssh.Client, logger boshlog.Logger) (matching bool, err error) {
	session, err := client.NewSession()
	if err != nil {
		return false, err
	}

	cpiVersionOutputBytes, err := session.CombinedOutput(fmt.Sprintf("%s -version", cpiDestPath))
	cpiVersionOutput := strings.TrimSpace(string(cpiVersionOutputBytes))
	if err != nil {
		logger.Debug("install-cpi", "comparison error: %s output: %s", err.Error(), string(cpiVersionOutputBytes))
	}

	if strings.Contains(string(cpiVersionOutput), "The system cannot find the path specified") {
		return false, nil
	}

	if strings.Contains(string(cpiVersionOutput), "is not recognized as an internal or external command") {
		return false, nil
	}

	if strings.Contains(string(cpiVersionOutput), "No such file or directory") {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("Error: %s; Output: %s", err.Error(), cpiVersionOutput)
	}

	if cpiVersionOutput == version {
		return true, nil
	}
	return false, nil
}
