// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

// ClusterInfo represents each cluster in the YAML file.
type ClusterInfo struct {
	Name           string `json:"name" yaml:"name"`
	KubeconfigPath string `json:"kubeconfigpath" yaml:"kubeconfigpath"`
}

// EnvFile represents the full environment configuration.
type EnvFile struct {
	Clusters map[string]ClusterInfo `json:"clusters" yaml:"clusters"`
}

// LoadEnvFile reads and parses an environment file into an EnvFile struct.
func LoadEnvFile(filePath string) (*EnvFile, error) {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("environment file %q does not exist", filePath)
	}

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file %q: %w", filePath, err)
	}

	// Parse YAML content into the EnvFile struct
	var env EnvFile
	if err := yaml.Unmarshal(content, &env); err != nil {
		return nil, fmt.Errorf("failed to parse environment file %q: %w", filePath, err)
	}

	return &env, nil
}
