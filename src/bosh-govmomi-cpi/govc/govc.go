package govc

//go:generate counterfeiter -o fakes/fake_govc_client.go $GOPATH/src/bosh-govmomi-cpi/govc/govc.go GovcClient
type GovcClient interface {
	ImportOvf(string, string) (string, error)
	CloneVM(string, string) (string, error)
	UpdateVMIso(string, string) (string, error)
	StartVM(string) (string, error)
	DestroyVM(string) (string, error)
}

//go:generate counterfeiter -o fakes/fake_govc_runner.go $GOPATH/src/bosh-govmomi-cpi/govc/govc.go GovcRunner
type GovcRunner interface {
	CliCommand(string, map[string]string, []string) (string, error)
}

//go:generate counterfeiter -o fakes/fake_govc_config.go $GOPATH/src/bosh-govmomi-cpi/govc/govc.go GovcConfig
type GovcConfig interface {
	EsxUrl() string
}
