package stemcell

//go:generate counterfeiter -o fakes/fake_client.go stemcell.go StemcellClient
type StemcellClient interface {
	ExtractOvf(string) (string, error)
	Cleanup()
}

//go:generate counterfeiter -o fakes/fake_store.go stemcell.go StemcellStore
type StemcellStore interface {
	GetImagePath(string, string) (string, error)
	Cleanup()
}

//go:generate counterfeiter -o fakes/fake_config.go stemcell.go Config
type Config interface {
	StemcellStorePath() string
}
