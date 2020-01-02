package install

import (
	"fmt"
	"io"
	"regexp"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"golang.org/x/crypto/ssh"

	"bosh-vmrun-cpi/config"
)

type installerImpl struct {
	cpiConfig config.Config
	sshClient *ssh.Client
	logger    boshlog.Logger
}

func NewInstaller(cpiConfig config.Config, sshClient *ssh.Client, logger boshlog.Logger) (*installerImpl, error) {
	return &installerImpl{cpiConfig, sshClient, logger}, nil
}

func sshCopy(srcReader io.Reader, remoteDestPath string, contentSize int64, client *ssh.Client, logger boshlog.Logger) (err error) {
	session, err := client.NewSession()
	if err != nil {
		return err
	}

	pathParts := regexp.MustCompile(`[\\/]+`).Split(remoteDestPath, -1)
	rootDirPath := pathParts[0]
	if rootDirPath == "" {
		rootDirPath = `/`
	}
	destParentDirPathParts := pathParts[1 : len(pathParts)-1]
	remoteFileName := pathParts[len(pathParts)-1]
	logger.Debug("main", "remoteDestPath: root: %s parent paths: %+#v file: %s\n", rootDirPath, destParentDirPathParts, remoteFileName)

	go func(srcReader io.Reader, contentSize int64) {
		w, _ := session.StdinPipe()
		defer w.Close()
		for _, dirPart := range destParentDirPathParts {
			fmt.Fprintln(w, "D0755", 0, dirPart) // loop through all parent paths, creating ones that don't exist
		}
		fmt.Fprintln(w, "C0744", contentSize, remoteFileName) // create the file

		io.Copy(w, srcReader)

		fmt.Fprint(w, "\x00") // transfer end with \x00
	}(srcReader, contentSize)

	scpOutput, err := session.CombinedOutput(fmt.Sprintf("scp -tr %s", rootDirPath))
	if err != nil {
		logger.ErrorWithDetails("main", "remoteDestPath: root: %s parent paths: %+#v file: %s\n", rootDirPath, destParentDirPathParts, remoteFileName)
		logger.ErrorWithDetails("main", "scpOutput:", scpOutput)
		return err
	}

	logger.Debug("main", "wrote %s\n", remoteDestPath)

	return nil
}
