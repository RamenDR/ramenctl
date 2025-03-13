// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/template"

	"sigs.k8s.io/yaml"
)

//go:embed sample.yaml
var sampleConfig string

func CreateSampleConfig(filename, commandName, envFile string) error {
	// Generate default content from Sample
	sample := Sample{CommandName: commandName}
	content, err := sample.Bytes()
	if err != nil {
		return fmt.Errorf("failed to create sample config: %w", err)
	}

	// Parse the default content into a struct
	var defaultConfig map[string]interface{}
	if err := yaml.Unmarshal(content, &defaultConfig); err != nil {
		return fmt.Errorf("failed to parse default sample config: %w", err)
	}

	// Load environment file if provided
	if envFile != "" {
		envConfig, err := LoadEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("failed to load environment file: %w", err)
		}

		// Convert envConfig to map for merging
		envConfigMap, err := structToMap(envConfig)
		if err != nil {
			return fmt.Errorf("failed to convert envConfig to map: %w", err)
		}

		// Merge the two configs, giving priority to envConfig
		mergedConfig := mergeMaps(defaultConfig, envConfigMap)

		// Convert back to YAML
		content, err = yaml.Marshal(mergedConfig)
		if err != nil {
			return fmt.Errorf("failed to serialize merged config: %w", err)
		}
	}

	// Write the final config to a file
	if err := createFile(filename, content); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %q already exists", filename)
		}
		return fmt.Errorf("failed to create %q: %w", filename, err)
	}
	return nil
}

// Converts a struct to a map[string]interface{}
func structToMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v) // Convert to JSON first
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Merges two maps, giving priority to values from overrideMap
func mergeMaps(baseMap, overrideMap map[string]interface{}) map[string]interface{} {
	for key, value := range overrideMap {
		baseMap[key] = value
	}
	return baseMap
}

type Sample struct {
	CommandName string
}

func (s *Sample) Bytes() ([]byte, error) {
	t, err := template.New("sample").Parse(sampleConfig)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createFile(name string, content []byte) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		return err
	}
	return f.Close()
}
