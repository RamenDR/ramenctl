// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/gather"
)

var GatherCmd = &cobra.Command{
	Use:   "gather",
	Short: "Collect diagnostic data from your clusters",
}

var GatherApplicationCmd = &cobra.Command{
	Use:   "application",
	Short: "Collect data for a protected application",
	Run: func(c *cobra.Command, args []string) {
		if err := gather.Gather(configFile, outputDir, drpcName, drpcNamespace); err != nil {
			console.Fatal(err)
		}
	},
}

func init() {
	addOutputFlags(GatherCmd)
	addDRPCFlags(GatherApplicationCmd)
	GatherCmd.AddCommand(GatherApplicationCmd)
}
