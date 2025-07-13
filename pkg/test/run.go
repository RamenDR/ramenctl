// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/testing"
)

func Run(configFile string, outputDir string) error {
	cfg, err := readConfig(configFile)
	if err != nil {
		return err
	}

	cmd, err := command.New("test-run", cfg.Clusters, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Close()

	test := newCommand(cmd, cfg, testing.Backend{})
	return test.Run()
}
