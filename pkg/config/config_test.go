// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"reflect"
	"testing"

	"github.com/ramendr/ramenctl/pkg/config"
)

func TestReadEnvFile(t *testing.T) {
	env, err := config.ReadEnvFile("testdata/regional-dr.yaml")
	if err != nil {
		t.Fatalf("Failed to read environment file: %v", err)
	}

	expected := &config.EnvFile{
		Name: "rdr",
		Ramen: config.Ramen{
			Hub:      "hub",
			Clusters: []string{"dr1", "dr2"},
		},
	}

	if !reflect.DeepEqual(expected, env) {
		t.Fatalf("expected %+v, got %+v", expected, env)
	}
}

func TestSample(t *testing.T) {
	sample := config.NewSample("ramenctl")
	expected := &config.Sample{
		CommandName:         "ramenctl",
		HubKubeconfig:       "hub/config",
		PrimaryKubeconfig:   "primary/config",
		SecondaryKubeconfig: "secondary/config",
	}
	if !reflect.DeepEqual(expected, sample) {
		t.Fatalf("expected %+v, got %+v", expected, sample)
	}
}

func TestSampleFromEnv(t *testing.T) {
	env := &config.EnvFile{
		Name: "rdr",
		Ramen: config.Ramen{
			Hub:      "hub",
			Clusters: []string{"dr1", "dr2"},
		},
	}
	sample := config.SampleFromEnv("ramenctl", env)
	expected := &config.Sample{
		CommandName:         "ramenctl",
		HubKubeconfig:       env.KubeconfigPath("hub"),
		PrimaryKubeconfig:   env.KubeconfigPath("dr1"),
		SecondaryKubeconfig: env.KubeconfigPath("dr2"),
	}
	if !reflect.DeepEqual(expected, sample) {
		t.Fatalf("expected %+v, got %+v", expected, sample)
	}
}
