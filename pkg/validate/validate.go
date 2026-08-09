// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/validate/application"
	"github.com/ramendr/ramenctl/pkg/validate/clusters"
	"github.com/ramendr/ramenctl/pkg/validation"
)

func Clusters(opts command.Options) error {
	cfg, err := config.ReadConfig(opts.ConfigFile)
	if err != nil {
		return console.Failed(err)
	}

	cmd, err := command.New(clusters.CommandName, cfg.Clusters, opts)
	if err != nil {
		return console.Failed(err)
	}
	defer cmd.Close()

	validate := clusters.NewCommand(cmd, cfg, validation.Backend{})

	var failed error
	if err := validate.Run(); err != nil {
		failed = console.Failed(err)
	}

	if opts.Interactive {
		cmd.BrowseReport()
	}

	return failed
}

func Application(opts command.ApplicationOptions) error {
	cfg, err := config.ReadConfig(opts.ConfigFile)
	if err != nil {
		return console.Failed(err)
	}

	cmd, err := command.New(application.CommandName, cfg.Clusters, opts.Options)
	if err != nil {
		return console.Failed(err)
	}
	defer cmd.Close()

	validate := application.NewCommand(cmd, cfg, validation.Backend{}, opts)

	var failed error
	if err := validate.Run(); err != nil {
		failed = console.Failed(err)
	}

	if opts.Interactive {
		cmd.BrowseReport()
	}

	return failed
}
