// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/validate"
)

var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Detect disaster recovery problems",
}

var ValidateClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Detect problems in disaster recovery clusters",
	Run: func(c *cobra.Command, args []string) {
		if err := validate.Clusters(command.Options{
			ConfigFile:  configFile,
			OutputDir:   outputDir,
			Interactive: interactive,
		}); err != nil {
			os.Exit(1)
		}
	},
}

var ValidateApplicationCmd = &cobra.Command{
	Use:   "application",
	Short: "Detect problems in disaster recovery protected application",
	Run: func(c *cobra.Command, args []string) {
		if err := validate.Application(command.ApplicationOptions{
			Options: command.Options{
				ConfigFile:  configFile,
				OutputDir:   outputDir,
				Interactive: interactive,
			},
			DRPCName:      drpcName,
			DRPCNamespace: drpcNamespace,
		}); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	addDRPCFlags(ValidateApplicationCmd)
	addOutputFlags(ValidateCmd)
	ValidateCmd.AddCommand(ValidateClustersCmd)
	ValidateCmd.AddCommand(ValidateApplicationCmd)
}
