package install

import (
	"bosh-vmrun-cpi/config"
	"bytes"
	"fmt"
	"strings"
	"text/template"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"golang.org/x/crypto/ssh"
)

func NewSshClient(cpiConfig config.Config, logger boshlog.Logger) (client *ssh.Client, err error) {
	sshHostname, sshPort, sshUsername, sshRawPrivateKey := sshCredentials(cpiConfig)

	privateKeyBytes, err := ssh.ParseRawPrivateKey([]byte(sshRawPrivateKey))
	if err != nil {
		return nil, err
	}

	signer, err := ssh.NewSignerFromKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	sshPublicKey := string(ssh.MarshalAuthorizedKey(signer.PublicKey()))

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
