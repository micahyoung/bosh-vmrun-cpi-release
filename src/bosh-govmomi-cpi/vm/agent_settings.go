package vm

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	"github.com/rn/iso9660wrap"
)

type AgentSettingsImpl struct {
	fs            boshsys.FileSystem
	logger        boshlog.Logger
	parentTempDir string
}

func NewAgentSettings(fs boshsys.FileSystem, logger boshlog.Logger) AgentSettings {
	parentTempDir, _ := fs.TempDir("agent-settings-env-iso-")

	return &AgentSettingsImpl{
		fs:            fs,
		logger:        logger,
		parentTempDir: parentTempDir,
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

func (s AgentSettingsImpl) Cleanup() {
	err := s.fs.RemoveAll(s.parentTempDir)
	if err != nil {
		s.logger.Error("stemcell-client", "Cleaning up stemcell temp dir '%s'", s.parentTempDir)
	}
}
