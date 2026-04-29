// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/failover"
)

var (
	dryRun      bool
	abortDryRun bool
)

var FailoverCmd = &cobra.Command{
	Use:   "failover",
	Short: "Test failover without affecting the primary application",
	Long: `Test failover to the secondary cluster without affecting the primary application.

This performs a dry-run failover that:
- Starts the application on the secondary cluster
- Keeps the primary application running
- Allows you to verify DR readiness

The application reaches "TestingFailover" progression when the dry-run succeeds.

Use --abort to revert the dry-run test and return to the original state.`,
	Run: func(c *cobra.Command, args []string) {
		if abortDryRun {
			if err := failover.AbortDryRun(
				configFile, drpcName, drpcNamespace); err != nil {
				console.Fatal(err)
			}
		} else {
			if err := failover.TestDryRun(
				configFile, outputDir, drpcName, drpcNamespace); err != nil {
				console.Fatal(err)
			}
		}
	},
}

func init() {
	addDRPCFlags(FailoverCmd)
	FailoverCmd.Flags().BoolVar(&dryRun, "dry-run", false, "perform dry-run failover test")
	FailoverCmd.Flags().BoolVar(&abortDryRun, "abort", false,
		"abort the dry-run failover test and revert to original state")
	FailoverCmd.Flags().StringVarP(&outputDir, "output", "o", "",
		"output directory for test report (default: ./failover-dry-run-<name>-<timestamp>)")

	if err := FailoverCmd.MarkFlagRequired("dry-run"); err != nil {
		panic(err)
	}
}
