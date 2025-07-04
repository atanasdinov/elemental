package network

import (
	"fmt"

	"github.com/suse/elemental/v3/pkg/sys/vfs"
)

type ConfiguratorInstaller struct {
	FS vfs.FS
}

func (i *ConfiguratorInstaller) InstallConfigurator(sourcePath, installPath string) error {
	if err := vfs.CopyFile(i.FS, sourcePath, installPath); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}

	return nil
}
