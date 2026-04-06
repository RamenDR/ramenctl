// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/spf13/cobra"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/failover"
)

var FailoverCmd = &cobra.Command{
	Use:   "failover",
	Short: "Manage application failover operations",
}

var FailoverDryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Test failover without affecting the primary application (DRY-RUN mode)",
	Long: `Test failover to the secondary cluster without affecting the primary application.

This is a DRY-RUN operation that:
- Starts the application on the secondary cluster
- Keeps the primary application running
- Allows you to verify DR readiness without risk

The application reaches "TestingFailover" progression when the dry-run succeeds.

Use --abort to revert the dry-run test and return to the original state.`,
	Run: func(c *cobra.Command, args []string) {
		if abortDryRun {
			if err := failover.AbortDryRun(configFile, drpcName, drpcNamespace); err != nil {
				console.Fatal(err)
			}
		} else {
			if err := failover.TestDryRun(configFile, outputDir, drpcName, drpcNamespace); err != nil {
				console.Fatal(err)
			}
		}
	},
}

var abortDryRun bool

func init() {
	addDRPCFlags(FailoverDryRunCmd)
	FailoverDryRunCmd.Flags().BoolVar(&abortDryRun, "abort", false, "abort the dry-run failover test and revert to original state")
	FailoverDryRunCmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory for test report (optional, only used without --abort)")

	FailoverCmd.AddCommand(FailoverDryRunCmd)
}
