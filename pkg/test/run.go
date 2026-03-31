// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/testing"
)

func Run(opts command.Options) error {
	cfg, err := readConfig(opts.ConfigFile)
	if err != nil {
		return console.Failed(err)
	}

	cmd, err := command.New("test-run", cfg.Clusters, opts)
	if err != nil {
		return console.Failed(err)
	}
	defer cmd.Close()

	test := newCommand(cmd, cfg, testing.Backend{})
	if err := test.Run(); err != nil {
		return console.Failed(err)
	}

	return nil
}
