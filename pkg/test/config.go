// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/deployers"
	"github.com/ramendr/ramen/e2e/workloads"
	"github.com/ramendr/ramenctl/pkg/console"
)

func readConfig(filename string) (*config.Config, error) {
	options := config.Options{
		Workloads: workloads.AvailableNames(),
		Deployers: deployers.AvailableTypes(),
	}
	config, err := config.ReadConfig(filename, options)
	if err != nil {
		return nil, fmt.Errorf("unable to read config: %w", err)
	}
	console.Info("Using config %q", filename)
	return config, nil
}
