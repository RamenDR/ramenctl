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
	Hub      string   `json:"hub"`
	Clusters []string `json:"clusters"`
}

type EnvFile struct {
	Name  string `json:"name"`
	Ramen Ramen  `json:"ramen"`
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
