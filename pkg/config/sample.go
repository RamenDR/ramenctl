// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed sample.yaml
var sampleConfig string

type Sample struct {
	CommandName         string
	HubKubeconfig       string
	PrimaryKubeconfig   string
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

func defaultSample(commandName string) *Sample {
	return &Sample{
		CommandName:         commandName,
		HubKubeconfig:       "hub/config",
		PrimaryKubeconfig:   "primary/config",
		SecondaryKubeconfig: "secondary/config",
	}
}

func sampleFromEnvFile(envFile, commandName string) (*Sample, error) {
	env, err := ReadEnvFile(envFile)
	if err != nil {
		return nil, err
	}
	return &Sample{
		CommandName:         commandName,
		HubKubeconfig:       env.KubeconfigPath(env.Ramen.Hub),
		PrimaryKubeconfig:   env.KubeconfigPath(env.Ramen.Clusters[0]),
		SecondaryKubeconfig: env.KubeconfigPath(env.Ramen.Clusters[1]),
	}, nil
}
