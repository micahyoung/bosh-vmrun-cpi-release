package stemcell

import (
	"github.com/go-yaml/yaml"
)

type manifest struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func NewStemcellManifest(data []byte) (*manifest, error) {
	newManifest := &manifest{}
	err := yaml.Unmarshal(data, newManifest)

	if err != nil {
		return nil, err
	}

	return newManifest, nil
}
