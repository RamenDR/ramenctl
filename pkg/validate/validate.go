// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/validation"
)

func Clusters(configFile string, outputDir string) error {
	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		return err
	}

	cmd, err := command.New("validate-clusters", cfg.Clusters, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Close()

	validate := newCommand(cmd, cfg, validation.Backend{})
	return validate.Clusters()
}

func Application(configFile, outputDir, drpcName, drpcNamespace string) error {
	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		return err
	}

	cmd, err := command.New("validate-application", cfg.Clusters, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Close()

	validate := newCommand(cmd, cfg, validation.Backend{})
	return validate.Application(drpcName, drpcNamespace)
}
