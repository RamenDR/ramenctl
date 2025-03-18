// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/command"
)

// Command is a ramenctl test command.
type Command struct {
	*command.Command

	// NamespacePrefix is used for all namespaces created by tests.
	NamespacePrefix string

	// PCCSpecs maps pvscpec name to pvcspec.
	PVCSpecs map[string]types.PVCSpecConfig
}

// newCommand return a new test command.
func newCommand(name, configFile, outputDir string) (*Command, error) {
	cmd, err := command.New(name, configFile, outputDir)
	if err != nil {
		return nil, err
	}

	// This is not user configurable. We use the same prefix for all namespaces created by the test.
	cmd.Config.Channel.Namespace = "test-gitops"

	return &Command{
		Command:         cmd,
		NamespacePrefix: "test-",
		PVCSpecs:        e2econfig.PVCSpecsMap(cmd.Config),
	}, nil
}
