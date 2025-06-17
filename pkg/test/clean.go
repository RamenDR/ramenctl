// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/e2e"
)

func Clean(configFile string, outputDir string) error {
	cfg, err := readConfig(configFile)
	if err != nil {
		return err
	}

	cmd, err := command.New("test-clean", cfg.Clusters, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Close()

	test := newCommand(cmd, cfg, e2e.Backend{}, Options{GatherData: true})
	return test.Clean()
}
