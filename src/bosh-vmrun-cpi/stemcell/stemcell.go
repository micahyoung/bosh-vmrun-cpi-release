package stemcell

//go:generate counterfeiter -o fakes/fake_client.go $GOPATH/src/bosh-vmrun-cpi/stemcell/stemcell.go StemcellClient
type StemcellClient interface {
	ExtractOvf(string) (string, error)
	Cleanup()
}

//go:generate counterfeiter -o fakes/fake_store.go $GOPATH/src/bosh-vmrun-cpi/stemcell/stemcell.go StemcellStore
type StemcellStore interface {
	GetImagePath(string, string) (string, error)
	Cleanup()
}

//go:generate counterfeiter -o fakes/fake_config.go $GOPATH/src/bosh-vmrun-cpi/stemcell/stemcell.go Config
type Config interface {
	StemcellStorePath() string
}
