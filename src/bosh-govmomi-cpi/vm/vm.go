package vm

import "github.com/cppforlife/bosh-cpi-go/apiv1"

//go:generate counterfeiter -o fakes/fake_agent_settings.go $GOPATH/src/bosh-govmomi-cpi/vm/agent_settings.go AgentSettings
type AgentSettings interface {
	Cleanup()
	GenerateAgentEnvIso(apiv1.AgentEnv) (string, error)
}
