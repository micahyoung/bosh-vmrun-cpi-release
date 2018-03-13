package stemcell

//go:generate counterfeiter -o fakes/fake_stemcell_client.go $GOPATH/src/bosh-govmomi-cpi/stemcell/stemcell.go StemcellClient
type StemcellClient interface {
	ExtractOvf(string) (string, error)
	Cleanup() error
}
