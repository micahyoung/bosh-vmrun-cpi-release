package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/go-yaml/yaml"
	"golang.org/x/crypto/ssh"

	"bosh-vmrun-cpi/config"
)

var (
	configPathOpt     = flag.String("configPath", "", "Path to configuration file")
	directorTmpDirOpt = flag.String("directorTmpDirPath", "", "Path to director's temp file containing extracted stemcell images")
	versionOpt        = flag.Bool("version", false, "Version")

	//set by X build flag
	version string
)

func main() {
	var err error
	var configJSON string
	var directorTmpDirPath string

	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	cmdRunner := boshsys.NewExecCmdRunner(logger)
	compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)

	flag.Parse()

	if *configPathOpt != "" {
		configJSONBytes, err := ioutil.ReadFile(*configPathOpt)
		configJSON = string(configJSONBytes)
		if err != nil {
			logger.Error("main", "loading cfg", err)
			os.Exit(1)
		}
	}

	if *directorTmpDirOpt != "" {
		//optional, validated on use
		directorTmpDirPath = *directorTmpDirOpt
	}

	command := flag.Arg(0)

	if command == "version" {
		fmt.Println(version)
		os.Exit(0)
	}

	if command == "encoded-config" {
		err = encodedConfig(configJSON, logger)
		os.Exit(0)
	}

	cpiConfig, err := config.NewConfigFromJson(configJSON)
	if err != nil {
		logger.Error("main", "parsing config JSON", err)
		os.Exit(1)
	}
	sshClient, err := sshClient(cpiConfig, logger)
	if err != nil {
		logger.Error("main", "initializing ssh client", err)
		os.Exit(1)
	}

	switch command {
	case "install-cpi":
		err = installCPI(cpiConfig, sshClient, logger)
	case "sync-director-stemcells":
		err = syncDirectorStemcells(cpiConfig, directorTmpDirPath, sshClient, compressor, logger)
	default:
		err = errors.New("command required")
	}

	if err != nil {
		logger.Error("main", "command failed", err)
		os.Exit(1)
	}
}

func encodedConfig(configJSON string, logger boshlog.Logger) error {
	configBase64 := base64.StdEncoding.EncodeToString([]byte(configJSON))

	fmt.Println(configBase64)
	return nil
}

func syncDirectorStemcells(cpiConfig config.Config, directorTmpDirPath string, sshClient *ssh.Client, compressor boshcmd.Compressor, logger boshlog.Logger) error {
	remoteStemcellStorePath := stemcellStorePath(cpiConfig)
	remotePathJoin := remotePathJoinFunc(cpiConfig)
	localPathJoin := localPathJoinFunc()

	if fileInfo, err := os.Stat(directorTmpDirPath); err != nil || !fileInfo.IsDir() {
		logger.Error("sync-director-stemcells", "director tmp path does not exist", err)
		return err
	}

	stemcellManifestPaths, err := filepath.Glob(localPathJoin(directorTmpDirPath, "*", "stemcell.MF"))
	if err != nil {
		return err
	}
	if len(stemcellManifestPaths) == 0 {
		logger.Info("sync-director-stemcells", "no stemcells found in %s", directorTmpDirPath)
		return nil
	}

	for _, manifestPath := range stemcellManifestPaths {
		logger.Debug("sync-director-stemcells", "found stemcell manifest: %s", manifestPath)
		stemcellFileName, err := parseStemcellFileName(manifestPath)
		if err != nil {
			logger.ErrorWithDetails("sync-director-stemcells", "parsing manifest for name", manifestPath)
			return err
		}
		remoteStemcellPath := remotePathJoin(remoteStemcellStorePath, stemcellFileName)
		matchingImagePath := localPathJoin(filepath.Dir(manifestPath), "image")
		remoteMappingFileName := fmt.Sprintf("%x.mapping", sha1.Sum([]byte(matchingImagePath)))
		remoteMappingPath := remotePathJoin(stemcellMappingsPath(cpiConfig), remoteMappingFileName)

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

		err = sshCopy(tempMappingFilePath, remoteMappingPath, sshClient, logger)
		if err != nil {
			logger.ErrorWithDetails("sync-director-stemcells", "ssh failure", err)
			return err
		}
		logger.Debug("sync-director-stemcells", "generated mapping: %s for image path: %s", remoteMappingPath, matchingImagePath)

		exists, err := testRemoteStemcell(remoteStemcellPath, sshClient, logger)
		if err != nil {
			logger.Error("sync-director-stemcells", "testing remote stemcell", err)
			return err
		}
		if exists {
			logger.Debug("sync-director-stemcells", "remote stemcell already exists: %s\n", stemcellFileName)
			continue
		}

		parentDirPath := filepath.Dir(manifestPath)
		tempTarFilePath, err := compressor.CompressFilesInDir(parentDirPath)
		if err != nil {
			logger.ErrorWithDetails("sync-director-stemcells", "tar failure", tempTarFilePath)
			return err
		}
		defer compressor.CleanUp(tempTarFilePath)

		err = sshCopy(tempTarFilePath, remoteStemcellPath, sshClient, logger)
		if err != nil {
			logger.ErrorWithDetails("sync-director-stemcells", "ssh failure", err)
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

func installCPI(cpiConfig config.Config, sshClient *ssh.Client, logger boshlog.Logger) error {
	var err error

	cpiSrcPath := cpiSourcePath(cpiConfig)
	if _, err := os.Stat(cpiSrcPath); os.IsNotExist(err) {
		logger.Error("install-cpi", "opening CPI source file", err)
		return err
	}

	cpiDestPath := cpiDestPath(cpiConfig)

	matching, err := compareCPIVersionsSSH(sshClient, cpiSrcPath, cpiDestPath, logger)
	if err != nil {
		logger.Error("install-cpi", "comparing existing CPI version over ssh", err)
		return err
	}

	if matching {
		logger.Debug("install-cpi", "Using existing remote cpi")
	} else {
		logger.Debug("install-cpi", "Installing remote cpi")
		err = sshCopy(cpiSrcPath, cpiDestPath, sshClient, logger)
		if err != nil {
			//TODO handle failed install due to limited authorized key. Return manual instructions
			logger.Error("install-cpi", "creating CPI destination file over SSH", err)
			return err
		}
	}

	//output path
	escapedCpiDestPath := strings.Trim(strconv.Quote(cpiDestPath), `"`)
	fmt.Println(escapedCpiDestPath)
	return nil
}

func sshClient(cpiConfig config.Config, logger boshlog.Logger) (client *ssh.Client, err error) {
	sshHostname, sshPort, sshUsername, sshPrivateKey, sshPublicKey, err := sshCredentials(cpiConfig)
	if err != nil {
		logger.Error("main", "loading ssh credentials", err)
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey([]byte(sshPrivateKey))
	if err != nil {
		return nil, err
	}

	clientConfig := &ssh.ClientConfig{
		User: sshUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%s", sshHostname, sshPort), clientConfig)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "ssh: handshake failed:"):
			errorMessage := sshAuthKeyMissingMessage(sshUsername, sshPublicKey)
			logger.Error("main", errorMessage, err)
		default:
			logger.Error("main", "creating SSH session", err)
		}
		return nil, err
	}

	return client, nil
}

func sshAuthKeyMissingMessage(sshUsername, sshPublicKey string) (output string) {
	tmpl := template.Must(template.New("tmpl").Parse(`
CPI: ssh connection failed. Command not set correctly in authorized_key:

add to your ~{{.SSHUsername}}/.ssh/authorized_keys:
--------------------------
{{.SSHPublicKey}}
--------------------------
	`))

	var messageBuf bytes.Buffer
	_ = tmpl.Execute(&messageBuf, struct {
		SSHUsername  string
		SSHPublicKey string
	}{sshUsername, sshPublicKey})
	return strings.TrimSpace(messageBuf.String())
}

func compareCPIVersionsSSH(client *ssh.Client, cpiSrcPath, cpiDestPath string, logger boshlog.Logger) (matching bool, err error) {
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

func sshCopy(localSrcPath, remoteDestPath string, client *ssh.Client, logger boshlog.Logger) (err error) {
	session, err := client.NewSession()
	if err != nil {
		return err
	}

	srcReader, err := os.Open(localSrcPath)
	if err != nil {
		return err
	}
	defer srcReader.Close()

	fileInfo, err := os.Stat(localSrcPath)
	if err != nil {
		return err
	}
	srcSize := fileInfo.Size()

	pathParts := regexp.MustCompile(`[\\/]+`).Split(remoteDestPath, -1)
	rootDirPath := pathParts[0]
	if rootDirPath == "" {
		rootDirPath = `/`
	}
	destParentDirPathParts := pathParts[1 : len(pathParts)-1]
	remoteFileName := pathParts[len(pathParts)-1]
	logger.Debug("main", "remoteDestPath: root: %s parent paths: %+#v file: %s\n", rootDirPath, destParentDirPathParts, remoteFileName)

	go func(srcReader io.Reader, srcSize int64) {
		w, _ := session.StdinPipe()
		defer w.Close()
		for _, dirPart := range destParentDirPathParts {
			fmt.Fprintln(w, "D0755", 0, dirPart) // loop through all parent paths, creating ones that don't exist
		}
		fmt.Fprintln(w, "C0744", srcSize, remoteFileName) // create the file

		io.Copy(w, srcReader)

		fmt.Fprint(w, "\x00") // transfer end with \x00
	}(srcReader, srcSize)

	scpOutput, err := session.CombinedOutput(fmt.Sprintf("scp -tr %s", rootDirPath))
	if err != nil {
		logger.ErrorWithDetails("main", "remoteDestPath: root: %s parent paths: %+#v file: %s\n", rootDirPath, destParentDirPathParts, remoteFileName)
		logger.ErrorWithDetails("main", "scpOutput:", scpOutput)
		return err
	}

	logger.Debug("main", "wrote %s to %s\n", localSrcPath, remoteDestPath)

	return nil
}

func sshCredentials(cpiConfig config.Config) (sshHostname, sshPort, sshUsername, sshPrivateKey, sshPublicKey string, err error) {
	sshHostname = cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Host
	sshPort = cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Port
	sshUsername = cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Username
	sshPrivateKey = cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Private_Key

	privateKeyBytes, err := ssh.ParseRawPrivateKey([]byte(sshPrivateKey))
	if err != nil {
		return "", "", "", "", "", err
	}

	signer, err := ssh.NewSignerFromKey(privateKeyBytes)
	if err != nil {
		return "", "", "", "", "", err
	}

	sshPublicKey = string(ssh.MarshalAuthorizedKey(signer.PublicKey()))

	return sshHostname, sshPort, sshUsername, sshPrivateKey, sshPublicKey, nil
}

func cpiSourcePath(cpiConfig config.Config) string {
	hypervisorPlatform := cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Platform
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return localPathJoinFunc()(dir, fmt.Sprintf("cpi-%s", hypervisorPlatform))
}

func stemcellStorePath(cpiConfig config.Config) string {
	return cpiConfig.Cloud.Properties.Vmrun.Stemcell_Store_Path
}

func stemcellMappingsPath(cpiConfig config.Config) string {
	remotePathJoin := remotePathJoinFunc(cpiConfig)
	return remotePathJoin(cpiConfig.Cloud.Properties.Vmrun.Stemcell_Store_Path, "mappings")
}

func localPathJoinFunc() func(...string) string {
	return filepath.Join
}

func remotePathJoinFunc(cpiConfig config.Config) func(...string) string {
	sep := cpiConfig.Cloud.Properties.Vmrun.PlatformPathSeparator()

	return func(pathParts ...string) string {
		return strings.Join(pathParts, sep)
	}
}

func cpiDestPath(cpiConfig config.Config) string {
	vmStorePath := cpiConfig.Cloud.Properties.Vmrun.Vm_Store_Path
	hypervisorPlatform := cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Platform
	remotePathJoin := remotePathJoinFunc(cpiConfig)

	if hypervisorPlatform == "windows" {
		return remotePathJoin(vmStorePath, "cpi.exe")
	}

	return remotePathJoin(vmStorePath, "cpi")
}
