// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package failover

import (
	"fmt"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
)

func TestDryRun(configFile, outputDir, drpcName, drpcNamespace string) error {
	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		return fmt.Errorf("unable to read config: %w", err)
	}

	cmd, err := command.New("failover-dry-run", cfg.Clusters, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Close()

	failoverCmd := newCommand(cmd, cfg)
	return failoverCmd.TestDryRun(drpcName, drpcNamespace)
}

func AbortDryRun(configFile, drpcName, drpcNamespace string) error {
	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		return fmt.Errorf("unable to read config: %w", err)
	}

	// For abort, we don't need output directory
	cmd, err := command.New("failover-abort-dry-run", cfg.Clusters, "")
	if err != nil {
		return err
	}
	defer cmd.Close()

	failoverCmd := newCommand(cmd, cfg)
	return failoverCmd.AbortDryRun(drpcName, drpcNamespace)
}
