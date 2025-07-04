// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"

	"github.com/ramendr/ramenctl/pkg/console"
)

var gatherName string
var gatherNamespace string

var GatherCmd = &cobra.Command{
	Use:   "gather",
	Short: "Collect diagnostic data from your clusters",
}

var GatherApplicationCmd = &cobra.Command{
	Use:   "application",
	Short: "Collect data based on application",
	Run: func(c *cobra.Command, args []string) {
		console.Info("running gather command")
	},
}

func init() {
	addOutputFlag(GatherCmd)
	GatherCmd.AddCommand(GatherApplicationCmd)
	GatherApplicationCmd.Flags().StringVar(&gatherName, "drpc name", "", "name of drpc")
	GatherApplicationCmd.Flags().StringVar(&gatherNamespace, "drpc namespace", "", "namespace of drpc")
}
