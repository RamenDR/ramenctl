// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/testing"
)

func Clean(opts command.Options) error {
	cfg, err := readConfig(opts.ConfigFile)
	if err != nil {
		return err
	}

	cmd, err := command.New("test-clean", cfg.Clusters, opts)
	if err != nil {
		return err
	}
	defer cmd.Close()

	test := newCommand(cmd, cfg, testing.Backend{})
	return test.Clean()
}
