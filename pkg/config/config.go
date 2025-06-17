// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"maps"
	"os"

	"github.com/spf13/viper"

	"github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramenctl/pkg/console"
)

// Config is used for all ramenctl commands except the test commands. It is a subset of
// ramen/e2e/config.Config.
type Config struct {
	// Clusters are part of this environment. Requires "hub", "c1", and "c2".
	Clusters map[string]config.Cluster `json:"clusters"`

	// ClusterSet name with the managed clusters.
	ClusterSet string `json:"clusterSet"`

	// Distro is either config.DistroK8s or config.DistroOcp. If unset the distro is detected
	// automatically when validating the config with the clusters.
	Distro string `json:"distro"`

	// Namespaces are set automatically based on Distro.
	Namespaces config.Namespaces `json:"namespaces"`
}

// CreateSampleConfig create a sample config that can be used by all commands. The file can be
// parsed using ReadConfig() or test.readConfig().
func CreateSampleConfig(filename, commandName, envFile string) error {
	var sample *Sample
	if envFile != "" {
		console.Info("Using envfile %q", envFile)
		env, err := ReadEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("failed to read environment file: %w", err)
		}
		sample = SampleFromEnv(commandName, env)
	} else {
		sample = NewSample(commandName)
	}

	content, err := sample.Bytes()
	if err != nil {
		return fmt.Errorf("failed to create sample config: %w", err)
	}

	if err := createFile(filename, content); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %q already exists", filename)
		}
		return fmt.Errorf("failed to create %q: %w", filename, err)
	}
	return nil
}

// ReadConfig reads the configuration file created by CreateSampleConfig, ignoring the test only
// configuration.
func ReadConfig(filename string) (*Config, error) {
	viper.SetDefault("ClusterSet", config.DefaultClusterSetName)
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	if err := cfg.validateDistro(); err != nil {
		return nil, err
	}

	if err := cfg.validateClusters(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Equal return true if config is equal to other config.
func (c *Config) Equal(o *Config) bool {
	if c == o {
		return true
	}
	if c.Distro != o.Distro {
		return false
	}
	if c.ClusterSet != o.ClusterSet {
		return false
	}
	if c.Namespaces != o.Namespaces {
		return false
	}
	return maps.Equal(c.Clusters, o.Clusters)
}

func (c *Config) validateDistro() error {
	if c.Distro == "" {
		// Will be detected when accessing the clusters.
		return nil
	}
	switch c.Distro {
	case config.DistroK8s:
		c.Namespaces = config.K8sNamespaces
	case config.DistroOcp:
		c.Namespaces = config.OcpNamespaces
	default:
		return fmt.Errorf("invalid distro %q: (choose one of %q, %q)",
			c.Distro, config.DistroK8s, config.DistroOcp)
	}
	return nil
}

func (c *Config) validateClusters() error {
	if c.Clusters["hub"].Kubeconfig == "" {
		return fmt.Errorf("failed to find hub cluster in configuration")
	}
	if c.Clusters["c1"].Kubeconfig == "" {
		return fmt.Errorf("failed to find c1 cluster in configuration")
	}
	if c.Clusters["c2"].Kubeconfig == "" {
		return fmt.Errorf("failed to find c2 cluster in configuration")
	}
	return nil
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
