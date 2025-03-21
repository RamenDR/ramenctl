// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"sync"

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

	// Command report, stored at the output directory on completion.
	mutex  sync.Mutex
	Report *Report
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
		Report:          NewReport(name),
	}, nil
}

func (c *Command) Setup() bool {
	console.Progress("Setup environment")
	if err := util.EnsureChannel(c.Env.Hub, c.Config, c.Logger); err != nil {
		err := fmt.Errorf("failed to setup environment: %w", err)
		console.Error(err)
		c.Logger.Error(err)
		c.Report.AddSetup(false)
		return false
	}
	console.Completed("Environment setup")
	c.Report.AddSetup(true)
	return true
}

func (c *Command) Cleanup() bool {
	console.Progress("Clean environment")
	if err := util.EnsureChannelDeleted(c.Env.Hub, c.Config, c.Logger); err != nil {
		err := fmt.Errorf("failed to clean environment: %w", err)
		console.Error(err)
		c.Logger.Error(err)
		c.Report.AddCleanup(false)
		return false
	}
	console.Completed("Environment cleaned")
	c.Report.AddCleanup(true)
	return true
}

func (c *Command) RunTest(test *Test) bool {
	defer c.addTest(test)
	if !test.Deploy() {
		return false
	}
	if !test.Protect() {
		return false
	}
	if !test.Failover() {
		return false
	}
	if !test.Relocate() {
		return false
	}
	if !test.Unprotect() {
		return false
	}
	if !test.Undeploy() {
		return false
	}
	return true
}

func (c *Command) CleanTest(test *Test) bool {
	defer c.addTest(test)
	if !test.Unprotect() {
		return false
	}
	if !test.Undeploy() {
		return false
	}
	return true
}

func (c *Command) Failed() error {
	if err := c.WriteReport(c.Report); err != nil {
		console.Error(err)
	}
	return fmt.Errorf("failed (%d passed, %d failed, %d skipped)",
		c.Report.Summary.Passed, c.Report.Summary.Failed, c.Report.Summary.Skipped)
}

func (c *Command) Passed() {
	if err := c.WriteReport(c.Report); err != nil {
		console.Error(err)
	}
	console.Completed("passed (%d passed, %d failed, %d skipped)",
		c.Report.Summary.Passed, c.Report.Summary.Failed, c.Report.Summary.Skipped)
}

func (c *Command) addTest(test *Test) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Report.AddTest(test)
}
