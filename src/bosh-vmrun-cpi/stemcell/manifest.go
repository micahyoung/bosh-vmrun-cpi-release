package stemcell

import (
	"github.com/go-yaml/yaml"
)

type StemcellManifest struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func NewStemcellManifest(data []byte) (*StemcellManifest, error) {
	newManifest := &StemcellManifest{}
	err := yaml.Unmarshal(data, newManifest)

	if err != nil {
		return nil, err
	}

	return newManifest, nil
}
