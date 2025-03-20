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
