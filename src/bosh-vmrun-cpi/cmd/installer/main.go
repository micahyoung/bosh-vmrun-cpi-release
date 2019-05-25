package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"golang.org/x/crypto/ssh"

	"bosh-vmrun-cpi/config"
)

var (
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
	//TODO, base64 encode internally
	configBase64JSONOpt = flag.String("configBase64JSON", "", "Base64-encoded JSON string of configuration")
	versionOpt          = flag.Bool("version", false, "Version")

	//set by X build flag
	version string
)

func main() {
	var err error
	var configJSON string
	var cpiConfig config.Config

	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)

	flag.Parse()

	if *versionOpt {
		fmt.Println(version)
		os.Exit(0)
	}

	if *configPathOpt != "" {
		configJSONBytes, err := ioutil.ReadFile(*configPathOpt)
		configJSON = string(configJSONBytes)
		if err != nil {
			logger.ErrorWithDetails("main", "loading cfg", err)
			os.Exit(1)
		}
	} else if *configBase64JSONOpt != "" {
		configJSONBytes, err := base64.StdEncoding.DecodeString(*configBase64JSONOpt)
		if err != nil {
			logger.ErrorWithDetails("main", "base64 decoding cfg", err)
			os.Exit(1)
		}
		configJSON = string(configJSONBytes)
	}

	cpiConfig, err = config.NewConfigFromJson(configJSON)
	if err != nil {
		logger.ErrorWithDetails("main", "config JSON is invalid", err, configJSON)
		os.Exit(1)
	}

	cpiDestPath := cpiDestPath(cpiConfig)
	sshHostname, sshPort, sshUsername, sshPrivateKey, sshPublicKey, err := sshCredentials(cpiConfig)
	if err != nil {
		logger.ErrorWithDetails("main", "loading ssh credentials", err)
		os.Exit(1)
	}

	sshClient, err := sshClient(sshHostname, sshPort, sshUsername, sshPrivateKey)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "ssh: handshake failed:"):
			errorMessage := sshAuthKeyMissingMessage(cpiDestPath, sshUsername, sshPublicKey)
			logger.ErrorWithDetails("main", errorMessage, err)
		default:
			logger.ErrorWithDetails("main", "creating SSH session", err)
		}

		os.Exit(1)
	}

	cpiSrcPath := cpiSourcePath(cpiConfig)
	if _, err := os.Stat(cpiSrcPath); os.IsNotExist(err) {
		logger.ErrorWithDetails("main", "opening CPI source file", err)
		os.Exit(1)
	}

	matching, err := compareCPIVersionsSSH(sshClient, cpiSrcPath, cpiDestPath)
	if err != nil {
		logger.ErrorWithDetails("main", "comparing existing CPI version over ssh", err)
		os.Exit(1)
	}

	if matching {
		logger.Debug("main", "Using existing remote cpi")
	} else {
		cpiDirFileInfo, err := os.Stat(filepath.Dir(cpiDestPath))
		if cpiDirFileInfo != nil && cpiDirFileInfo.IsDir() {
			logger.Debug("main", "Installing local cpi")
			err = writeCPIContentLocal(cpiSrcPath, cpiDestPath)
			if err != nil {
				logger.ErrorWithDetails("main", "creating CPI destination file locally", err)
				os.Exit(1)
			}
		} else {
			logger.Debug("main", "Installing remote cpi")
			err = writeCPIContentToSSH(sshClient, cpiSrcPath, cpiDestPath)
			if err != nil {
				//TODO handle failed install due to limited authorized key. Return manual instructions
				logger.ErrorWithDetails("main", "creating CPI destination file over SSH", err)
				os.Exit(1)
			}
		}
	}

	//output path
	escapedCpiDestPath := strings.Trim(strconv.Quote(cpiDestPath), `"`)
	fmt.Println(escapedCpiDestPath)
}

func sshClient(sshHostname, sshPort, sshUsername, sshPrivateKey string) (client *ssh.Client, err error) {
	//TODO: dry up
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
		return nil, err
	}

	return client, nil
}

func sshAuthKeyMissingMessage(cpiDestPath, sshUsername, sshPublicKey string) (output string) {
	tmpl := template.Must(template.New("tmpl").Parse(`
CPI: ssh connection failed. Command not set correctly in authorized_key:

add to your ~/.ssh/authorized_keys:
--------------------------
restrict,command="{{.CPIDestPath}} ${SSH_ORIGINAL_COMMAND#* }",port-forwarding,permitopen="*:6868",permitopen="*:25555" {{.SSHPublicKey}}
--------------------------
	`))

	var messageBuf bytes.Buffer
	_ = tmpl.Execute(&messageBuf, struct {
		CPIDestPath  string
		SSHPublicKey string
	}{cpiDestPath, sshPublicKey})
	return strings.TrimSpace(messageBuf.String())
}

func compareCPIVersionsSSH(client *ssh.Client, cpiSrcPath, cpiDestPath string) (matching bool, err error) {
	session, err := client.NewSession()
	if err != nil {
		return false, err
	}

	cpicpiVersionOutputBytes, err := session.CombinedOutput(fmt.Sprintf("%s -version", cpiDestPath))
	cpiVersionOutput := strings.TrimSpace(string(cpicpiVersionOutputBytes))
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

func writeCPIContentLocal(cpiSrcPath, cpiDestPath string) (err error) {
	srcFile, err := os.Open(cpiSrcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(cpiDestPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func writeCPIContentToSSH(client *ssh.Client, cpiSrcPath, cpiDestPath string) (err error) {
	cpiExeContent, err := ioutil.ReadFile(cpiSrcPath)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}

	cpiBinName := filepath.Base(cpiDestPath)
	cpiDestParentDirPath := filepath.Dir(cpiDestPath)
	go func(cpiExeContent []byte) {
		w, _ := session.StdinPipe()
		defer w.Close()
		cpiExeContentStr := string(cpiExeContent)
		fmt.Fprintln(w, "D0755", 0, ".")                            // mkdir: rwx for owner; rx for others
		fmt.Fprintln(w, "C0744", len(cpiExeContentStr), cpiBinName) // cpi file: rwx for owner; r for others
		fmt.Fprint(w, cpiExeContentStr)
		fmt.Fprint(w, "\x00") // transfer end with \x00
	}(cpiExeContent)

	scpOutput, err := session.CombinedOutput(fmt.Sprintf("scp -tr %s", cpiDestParentDirPath))
	if err != nil {
		info, _ := os.Stat(cpiDestParentDirPath)
		fmt.Printf("cpiDestPath parent path info: %+#v\n", info)
		fmt.Printf("scpOutput: %+#v\n", scpOutput)
		return err
	}

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
	return filepath.Join(dir, fmt.Sprintf("cpi-%s", hypervisorPlatform))
}

func cpiDestPath(cpiConfig config.Config) string {
	vmStorePath := cpiConfig.Cloud.Properties.Vmrun.Vm_Store_Path
	hypervisorPlatform := cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Platform

	cpiBinName := "cpi"
	if hypervisorPlatform == "windows" {
		cpiBinName = "cpi.exe"
	}

	return filepath.Join(vmStorePath, cpiBinName)
}
