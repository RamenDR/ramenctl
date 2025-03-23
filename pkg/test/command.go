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

// flowFunc runs a test flow on with a test. The test logs progress messages and marked as failed if the flow failed.
type flowFunc func(t *Test)

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
	console.Step("Setup environment")
	if err := util.EnsureChannel(c.Env.Hub, c.Config, c.Logger); err != nil {
		err := fmt.Errorf("failed to setup environment: %w", err)
		console.Error(err)
		c.Logger.Error(err)
		c.Report.AddSetup(false)
		return false
	}
	console.Pass("Environment setup")
	c.Report.AddSetup(true)
	return true
}

func (c *Command) Cleanup() bool {
	console.Step("Clean environment")
	if err := util.EnsureChannelDeleted(c.Env.Hub, c.Config, c.Logger); err != nil {
		err := fmt.Errorf("failed to clean environment: %w", err)
		console.Error(err)
		c.Logger.Error(err)
		c.Report.AddCleanup(false)
		return false
	}
	console.Pass("Environment cleaned")
	c.Report.AddCleanup(true)
	return true
}

func (c *Command) RunTests() bool {
	console.Step("Run tests")
	return c.runFlowFunc(c.runFlow)
}

func (c *Command) CleanTests() bool {
	console.Step("Clean tests")
	return c.runFlowFunc(c.cleanFlow)
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

func (c *Command) runFlowFunc(f flowFunc) bool {
	var wg sync.WaitGroup
	for _, tc := range c.Config.Tests {
		test := newTest(tc, c)
		wg.Add(1)
		go func() {
			f(test)
			wg.Done()
		}()
	}
	wg.Wait()

	return c.checkStatus()
}

func (c *Command) runFlow(test *Test) {
	defer c.addTest(test)

	if !test.Deploy() {
		return
	}
	if !test.Protect() {
		return
	}
	if !test.Failover() {
		return
	}
	if !test.Relocate() {
		return
	}
	if !test.Unprotect() {
		return
	}
	test.Undeploy()
}

func (c *Command) cleanFlow(test *Test) {
	defer c.addTest(test)

	if !test.Unprotect() {
		return
	}
	test.Undeploy()
}

func (c *Command) checkStatus() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Report.Status == Passed
}

func (c *Command) addTest(test *Test) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Report.AddTest(test)
}
