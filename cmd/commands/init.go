// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/skills"
)

var (
	envFile   string
	agentName string
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create configuration file and install AI skills",
	// Validate flags early so cobra shows usage on invalid input.
	PreRunE: func(c *cobra.Command, args []string) error {
		return skills.ValidateAgent(agentName)
	},
	Run: func(c *cobra.Command, args []string) {
		if err := runInit(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	InitCmd.Flags().StringVar(&envFile, "envfile", "", "ramen testing environment file")
	InitCmd.Flags().StringVarP(&agentName, "agent", "a", skills.AgentGeneric,
		fmt.Sprintf("AI agent to install skills for (%s)", strings.Join(skills.Agents(), ", ")))
}

func runInit() error {
	commandName := RootCmd.DisplayName()

	console.Info("Using config %q", configFile)
	console.Step("Initializing")

	if !config.Install(configFile, commandName, envFile) {
		return console.Failed(fmt.Errorf("init failed"))
	}

	if !skills.Install(commandName, agentName) {
		return console.Failed(fmt.Errorf("init failed"))
	}

	console.Completed("Init completed")
	return nil
}
