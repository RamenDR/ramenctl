// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramen/e2e/util"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/console"
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

func (c *Command) Setup() error {
	console.Progress("Setup environment")
	if err := util.EnsureChannel(c.Env.Hub, c.Config, c.Logger); err != nil {
		err := fmt.Errorf("failed to setup environment: %w", err)
		c.Logger.Error(err)
		return err
	}
	console.Completed("Environment setup")
	return nil
}

func (c *Command) Cleanup() error {
	console.Progress("Clean environment")
	if err := util.EnsureChannelDeleted(c.Env.Hub, c.Config, c.Logger); err != nil {
		err := fmt.Errorf("failed to clean environment: %w", err)
		c.Logger.Error(err)
		return err
	}
	console.Completed("Environment cleaned")
	return nil
}

func (c *Command) RunTest(test *Test) error {
	if err := test.Deploy(); err != nil {
		return err
	}
	if err := test.Protect(); err != nil {
		return err
	}
	if err := test.Failover(); err != nil {
		return err
	}
	if err := test.Relocate(); err != nil {
		return err
	}
	if err := test.Unprotect(); err != nil {
		return err
	}
	if err := test.Undeploy(); err != nil {
		return err
	}
	return nil
}

func (c *Command) CleanTest(test *Test) error {
	if err := test.Unprotect(); err != nil {
		return err
	}
	if err := test.Undeploy(); err != nil {
		return err
	}
	return nil
}
