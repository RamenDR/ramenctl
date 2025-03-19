// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/ramendr/ramenctl/pkg/build"
)

var (
	// configFile is shared by all commands, enabling access to all clusters.
	configFile string

	// userOutputDir is used by troubleshooting commands for creating a report.
	userOutputDir string
)

var RootCmd = &cobra.Command{
	Use:     "ramenctl",
	Short:   "Manage and troubleshoot Ramen",
	Version: build.Version,

	// When used as a subcommand in another tool, don't inherit persistent pre run commands.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
}

func init() {
	// Use plain, machine friendly version string.
	RootCmd.SetVersionTemplate("{{.Version}}\n")

	// These flags are used by all sub commands.
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "configuration file")
}

func addOutputFlag(c *cobra.Command) {
	// The actual output directory is known only when running the command.
	c.PersistentFlags().StringVarP(&userOutputDir, "output", "o", "", "report directory (default report.{timestamp})")
}

func outputDir() string {
	if userOutputDir == "" {
		return time.Now().Format("report.20060102150405")
	}
	return userOutputDir
}
