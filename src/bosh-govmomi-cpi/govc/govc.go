package govc

//go:generate counterfeiter -o fakes/fake_govc_client.go $GOPATH/src/bosh-govmomi-cpi/govc/govc.go GovcClient
type GovcClient interface {
	ImportOvf(string) (string, error)
}

//go:generate counterfeiter -o fakes/fake_govc_runner.go $GOPATH/src/bosh-govmomi-cpi/govc/govc.go GovcRunner
type GovcRunner interface {
	CliCommand(string, map[string]string, []string) (string, error)
}
