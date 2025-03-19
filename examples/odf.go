// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

// Example for integrating a subset of ramenctl commamds as "odf dr" sub command.
package main

import (
	"log"

	dr "github.com/ramendr/ramenctl/cmd/commands"
	drbuild "github.com/ramendr/ramenctl/pkg/build"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "odf",
	Short: "Testing odf cli integration",

	// Not clear why odf is using this flag.
	TraverseChildren: true,

	// odf uses this to create k8s clients and validate the cluster. For ramenctl this is not relevant since we work on
	// multiple clusers that may or may not have odf installed.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Print("odf sub command pre-run")
	},
}

var fooCmd = &cobra.Command{
	Use:   "foo",
	Short: "Top level odf cli command",
	Run: func(c *cobra.Command, args []string) {
		log.Print("odf foo sub command")
	},
}

func init() {
	// This persistent flag is not relevant to the "odf dr" sub command so it should not be inherited by it. There seems
	// to be no way to avoid this inheritance now, so it will be need to change in odf.
	rootCmd.PersistentFlags().String("context", "", "kubectl context")
}

func main() {
	// Adapt ramenctl root command to the odf.
	dr.RootCmd.Use = "dr"
	dr.RootCmd.Short = "Troubleshoot OpenShift DR"

	// Set build information for odf dr reports.
	drbuild.Version = "v1.2.3"
	drbuild.Commit = "eb92ed81e2715d286bfd8ce173c76d4ecda9e2b4"

	// Add a subset of ramenctl command as the "odf dr" subcommand.
	dr.RootCmd.AddCommand(dr.InitCmd, dr.TestCmd, dr.ValidateCmd)

	rootCmd.AddCommand(fooCmd, dr.RootCmd)

	err := rootCmd.Execute()
	if err != nil {
		// odf uses its own logging infrastructure.
		log.Fatal(err)
	}
}
