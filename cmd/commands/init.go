// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/spf13/cobra"
)

var envFile string

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create configuration file for your clusters",
	Run: func(c *cobra.Command, args []string) {
		if err := config.CreateSampleConfig(configFile, RootCmd.DisplayName(), envFile); err != nil {
			console.Fatal(fmt.Errorf("failed to create sample config: %v", err))
		}
		console.Info("Created new config file %q", configFile)
		console.Completed("Finalized config file %q - please modify as needed", configFile)
	},
}

func init() {
	// Register the --envfile flag
	InitCmd.Flags().StringVar(&envFile, "envfile", "", "ramen testing environment file")
}
