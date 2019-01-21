package vm

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

//go:generate counterfeiter -o fakes/fake_agent_settings.go agent_settings.go AgentSettings
type AgentSettings interface {
	Cleanup()
	GenerateAgentEnvIso([]byte) (string, error)
	GenerateMacAddress() (string, error)
	GetIsoAgentEnv(string) (apiv1.AgentEnv, error)
}
