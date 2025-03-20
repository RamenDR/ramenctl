// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/deployers"
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramen/e2e/workloads"
	"github.com/ramendr/ramenctl/pkg/console"
)

//go:embed sample.yaml
var sampleConfig string

func CreateSampleConfig(filename, commandName, envFile string) error {
	var sample *Sample
	if envFile != "" {
		console.Info("Using envfile %q", envFile)
		var err error
		sample, err = sampleFromEnvFile(envFile, commandName)
		if err != nil {
			return fmt.Errorf("failed to load environment file: %w", err)
		}
	} else {
		sample = defaultSample(commandName)
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

func ReadConfig(filename string) (*types.Config, error) {
	options := config.Options{
		Workloads: workloads.AvailableNames(),
		Deployers: deployers.AvailableNames(),
	}
	config, err := config.ReadConfig(filename, options)
	if err != nil {
		return nil, fmt.Errorf("unable to read config: %w", err)
	}
	return config, nil
}

type Sample struct {
	CommandName         string
	HubName             string
	HubKubeconfig       string
	PrimaryName         string
	PrimaryKubeconfig   string
	SecondaryName       string
	SecondaryKubeconfig string
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

func defaultSample(commandName string) *Sample {
	return &Sample{
		CommandName:         commandName,
		HubName:             "hub",
		HubKubeconfig:       "hub/config",
		PrimaryName:         "primary",
		PrimaryKubeconfig:   "primary/config",
		SecondaryName:       "secondary",
		SecondaryKubeconfig: "secondary/config",
	}
}
