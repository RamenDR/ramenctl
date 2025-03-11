// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"

	"path/filepath"
	"sigs.k8s.io/yaml"
)

type Ramen struct {
	Hub      string   `json:"hub" yaml:"hub"`
	Clusters []string `json:"clusters" yaml:"clusters"`
}

type EnvFile struct {
	Name  string `json:"name" yaml:"name"`
	Ramen Ramen  `json:"ramen" yaml:"ramen"`
}

// ReadEnvFile reads and parses an environment file into an EnvFile struct.
func ReadEnvFile(filePath string) (*EnvFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file %q: %w", filePath, err)
	}

	var env EnvFile
	if err := yaml.Unmarshal(content, &env); err != nil {
		return nil, fmt.Errorf("failed to parse environment file %q: %w", filePath, err)
	}

	return &env, nil
}

func (e *EnvFile) KubeconfigPath(name string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, ".config", "drenv", e.Name, "kubeconfigs", name)
}

func sampleFromEnvFile(envFile, commandName string) (*Sample, error) {
	envConfig, err := ReadEnvFile(envFile)
	if err != nil {
		return nil, err
	}
	return &Sample{
		CommandName:         commandName,
		HubName:             envConfig.Ramen.Hub,
		HubKubeconfig:       envConfig.KubeconfigPath(envConfig.Ramen.Hub),
		PrimaryName:         envConfig.Ramen.Clusters[0],
		PrimaryKubeconfig:   envConfig.KubeconfigPath(envConfig.Ramen.Clusters[0]),
		SecondaryName:       envConfig.Ramen.Clusters[1],
		SecondaryKubeconfig: envConfig.KubeconfigPath(envConfig.Ramen.Clusters[1]),
	}, nil
}
