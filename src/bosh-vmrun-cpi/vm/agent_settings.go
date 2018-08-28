package vm

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	"github.com/hooklift/iso9660"
	"github.com/rn/iso9660wrap"
)

type AgentSettingsImpl struct {
	fs              boshsys.FileSystem
	logger          boshlog.Logger
	agentEnvFactory apiv1.AgentEnvFactory
	parentTempDir   string
}

func NewAgentSettings(fs boshsys.FileSystem, logger boshlog.Logger, agentEnvFactory apiv1.AgentEnvFactory) AgentSettings {
	parentTempDir, _ := fs.TempDir("agent-settings-env-iso-")

	return &AgentSettingsImpl{
		fs:              fs,
		logger:          logger,
		agentEnvFactory: agentEnvFactory,
		parentTempDir:   parentTempDir,
	}
}

func (s AgentSettingsImpl) GenerateAgentEnvIso(agentEnv apiv1.AgentEnv) (string, error) {
	envBytes, _ := agentEnv.AsBytes()
	envIsoPath := filepath.Join(s.parentTempDir, "env.iso")

	isoFile, err := s.fs.OpenFile(envIsoPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", bosherr.WrapError(err, "opening tempfile failed")
	}
	defer isoFile.Close()

	iso9660wrap.WriteBuffer(isoFile, envBytes, "ENV")

	if err != nil {
		return "", bosherr.WrapError(err, "creating tempdir failed")
	}

	return envIsoPath, nil
}

func (s AgentSettingsImpl) GetIsoAgentEnv(isoPath string) (apiv1.AgentEnv, error) {
	var err error
	var image *os.File

	image, err = os.Open(isoPath)
	if err != nil {
		return nil, err
	}
	defer image.Close()

	reader, err := iso9660.NewReader(image)
	if err != nil {
		return nil, err
	}

	imageFileRef, err := reader.Next()
	if err != nil {
		return nil, err
	}

	imageFile := imageFileRef.(*iso9660.File)
	fileReader := imageFile.Sys().(io.Reader)
	content, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return nil, err
	}

	agentEnv, err := s.agentEnvFactory.FromBytes(content)
	if err != nil {
		return agentEnv, err
	}

	return agentEnv, nil
}

func (s AgentSettingsImpl) GenerateMacAddress() (string, error) {
	buf := make([]byte, 2)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	buf[0] |= 2
	//MAC OUI when manually set: https://pubs.vmware.com/vsphere-4-esx-vcenter/index.jsp?topic=/com.vmware.vsphere.server_configclassic.doc_41/esx_server_config/advanced_networking/c_setting_up_mac_addresses.html
	//note: naive implementation, limited to 64 hosts. The actual range is 00:50:56:00:00:00 to 00:50:56:3f:ff:ff
	return fmt.Sprintf("00:50:56:3f:%02x:%02x", buf[0], buf[1]), nil
}

func (s AgentSettingsImpl) Cleanup() {
	err := s.fs.RemoveAll(s.parentTempDir)
	if err != nil {
		s.logger.Error("stemcell-client", "Cleaning up stemcell temp dir '%s'", s.parentTempDir)
	}
}
