// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/deployers"
	"github.com/ramendr/ramen/e2e/workloads"

	ramenctlconfig "github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
)

func readConfig(filename string) (*config.Config, error) {
	options := config.Options{
		Workloads: workloads.AvailableNames(),
		Deployers: deployers.AvailableTypes(),
	}
	cfg, err := config.ReadConfig(filename, options)
	if err != nil {
		return nil, fmt.Errorf("unable to read config: %w", err)
	}

	if err := ramenctlconfig.ValidateUniqueKubeconfigs(cfg.Clusters); err != nil {
		return nil, err
	}

	console.Info("Using config %q", filename)
	return cfg, nil
}
