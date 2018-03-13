package govc

//go:generate counterfeiter -o fakes/fake_govc_client.go $GOPATH/src/bosh-govmomi-cpi/govc/govc.go GovcClient
type GovcClient interface {
	ImportOvf(string) (bool, error)
}
