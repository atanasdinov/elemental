/*
Copyright © 2025 SUSE LLC
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package build

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/suse/elemental/v3/internal/image"
	"github.com/suse/elemental/v3/internal/template"
	"github.com/suse/elemental/v3/pkg/log"
	"github.com/suse/elemental/v3/pkg/sys/vfs"
)

//go:embed templates/network.sh.tpl
var configureNetworkScript string

const networkCustomScriptName = "configure-network.sh"

type networkConfigGenerator interface {
	GenerateNetworkConfig(configDir, outputDir string) error
}

type networkConfiguratorInstaller interface {
	InstallConfigurator(sourcePath, installPath string) error
}

type Network struct {
	FS                    vfs.FS
	Logger                log.Logger
	ConfigDir             string
	ConfigGenerator       networkConfigGenerator
	ConfiguratorInstaller networkConfiguratorInstaller
}

func NewNetwork(configDir string, fs vfs.FS, logger log.Logger, generator networkConfigGenerator, installer networkConfiguratorInstaller) *Network {
	return &Network{
		ConfigDir:             configDir,
		FS:                    fs,
		Logger:                logger,
		ConfigGenerator:       generator,
		ConfiguratorInstaller: installer,
	}
}

// Configure configures the network component if enabled.
//
//  1. Copies the nmc executable
//  2. Copies a custom network configuration script if provided
//  3. Generates network configurations and writes the configuration script template otherwise
//
// Example result file layout:
//
//	overlays
//	├── network
//	│   ├── node1.example.com
//	│   │   ├── eth0.nmconnection
//	│   │   └── eth1.nmconnection
//	│   ├── node2.example.com
//	│   │   └── eth0.nmconnection
//	│   ├── node3.example.com
//	│   │   ├── bond0.nmconnection
//	│   │   └── eth1.nmconnection
//	│   └── host_config.yaml
//	├── nmc
//	└── configure-network.sh
func (n *Network) Configure(buildDir image.BuildDir) error {
	entries, err := n.FS.ReadDir(n.ConfigDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			n.Logger.Info("Network configuration not provided, skipping.")
			return nil
		}

		return fmt.Errorf("reading network directory: %w", err)
	} else if len(entries) == 0 {
		return fmt.Errorf("network directory is present but empty")
	}

	//if err = n.installNetworkConfigurator(buildDir.OverlaysDir()); err != nil {
	//	return fmt.Errorf("installing configurator: %w", err)
	//}

	customScript := filepath.Join(n.ConfigDir, networkCustomScriptName)
	configScript := filepath.Join(buildDir.OverlaysDir(), networkCustomScriptName)

	// Copy custom network script if provided.
	// Proceed with generating configuration otherwise.
	err = vfs.CopyFile(n.FS, customScript, configScript)
	if err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("copying custom network script: %w", err)
	}

	outputDir := filepath.Join(buildDir.OverlaysDir(), buildDir.NetworkDir())
	if err = n.ConfigGenerator.GenerateNetworkConfig(n.ConfigDir, outputDir); err != nil {
		return fmt.Errorf("generating network config: %w", err)
	}

	if err = n.writeNetworkConfigurationScript(configScript, buildDir.OverlaysDir()); err != nil {
		return fmt.Errorf("writing network configuration script: %w", err)
	}

	return nil
}

func (n *Network) installNetworkConfigurator(outputDir string) error {
	sourcePath := "/usr/bin/nmc"

	exists, _ := vfs.Exists(n.FS, sourcePath)
	if !exists {
		return fmt.Errorf("%s not found, unable to install network configuration", sourcePath)
	}

	installPath := filepath.Join(outputDir, "nmc")

	return n.ConfiguratorInstaller.InstallConfigurator(sourcePath, installPath)
}

func (n *Network) writeNetworkConfigurationScript(scriptPath, configDir string) error {
	values := struct {
		ConfigDir string
	}{
		ConfigDir: configDir,
	}

	data, err := template.Parse("network", configureNetworkScript, &values)
	if err != nil {
		return fmt.Errorf("parsing network template: %w", err)
	}

	if err = n.FS.WriteFile(scriptPath, []byte(data), 0o744); err != nil {
		return fmt.Errorf("writing network script: %w", err)
	}

	return nil
}
