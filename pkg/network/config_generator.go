package network

import (
	"github.com/suse/elemental/v3/pkg/log"
	"github.com/suse/elemental/v3/pkg/sys"
)

type ConfigGenerator struct {
	sys.Runner
	log.Logger
}

func (g ConfigGenerator) GenerateNetworkConfig(configDir, outputDir string) error {
	b, err := g.Run("nmc", "generate",
		"--config-dir", configDir,
		"--output-dir", outputDir)
	if err != nil {
		return err
	}

	g.Debug("Generated network configuration: %s", string(b))

	return nil
}
