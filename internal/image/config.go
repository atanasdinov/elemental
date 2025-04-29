package image

import (
	"bytes"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConfigDir string

func (dir ConfigDir) InstallFilepath() string {
	return filepath.Join(string(dir), "install.yaml")
}

func (dir ConfigDir) OSFilepath() string {
	return filepath.Join(string(dir), "os.yaml")
}

func (dir ConfigDir) ReleaseFilepath() string {
	return filepath.Join(string(dir), "release.yaml")
}

func ParseConfig(data []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	return decoder.Decode(target)
}
