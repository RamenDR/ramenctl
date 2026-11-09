// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/validate/application"
	"github.com/ramendr/ramenctl/pkg/validate/clusters"
	"github.com/ramendr/ramenctl/pkg/validate/recipe"
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

// Recipe reads the Recipe YAML at path, runs offline semantic checks, and writes
// the recipe validation status report to the output directory.
// It returns a non-nil error when any problem-level issue is found or when the file
// cannot be read or parsed.
func Recipe(path string, opts command.Options) error {
	cmd, err := command.ForTest(recipe.CommandName, nil, opts.OutputDir)
	if err != nil {
		return console.Failed(err)
	}
	defer cmd.Close()

	console.Info("Using report %q", opts.OutputDir)

	validate := recipe.NewCommand(cmd, path)

	var failed error
	if err := validate.Run(); err != nil {
		failed = console.Failed(err)
	}

	return failed
}
