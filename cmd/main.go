// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/ramendr/ramenctl/cmd/commands"
)

func main() {
	// We add the sub commands here so other project can use the same root command with a subset of ramenctl commands.
	commands.RootCmd.AddCommand(
		commands.InitCmd,
		commands.TestCmd,
		commands.ValidateCmd,
	)

	err := commands.RootCmd.Execute()
	if err != nil {
		// When root command fails cobra already logged this error.
		os.Exit(1)
	}
}
