// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/e2e"
)

func Run(configFile string, outputDir string) error {
	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		return err
	}
	console.Info("Using config %q", configFile)

	cmd, err := command.New("test-run", cfg.Clusters, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Close()

	test := newCommand(cmd, cfg, e2e.Backend{}, Options{GatherData: true})
	return test.Run()
}
