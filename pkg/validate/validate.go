// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/validate/application"
	"github.com/ramendr/ramenctl/pkg/validate/clusters"
	"github.com/ramendr/ramenctl/pkg/validation"
)

func Clusters(opts command.Options) error {
	cfg, err := config.ReadConfig(opts.ConfigFile)
	if err != nil {
		return err
	}

	cmd, err := command.New(clusters.CommandName, cfg.Clusters, opts)
	if err != nil {
		return err
	}
	defer cmd.Close()

	validate := clusters.NewCommand(cmd, cfg, validation.Backend{})
	return validate.Run()
}

func Application(opts command.ApplicationOptions) error {
	cfg, err := config.ReadConfig(opts.ConfigFile)
	if err != nil {
		return err
	}

	cmd, err := command.New(application.CommandName, cfg.Clusters, opts.Options)
	if err != nil {
		return err
	}
	defer cmd.Close()

	validate := application.NewCommand(cmd, cfg, validation.Backend{})
	return validate.Run(opts.DRPCName, opts.DRPCNamespace)
}
