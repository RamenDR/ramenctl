// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"sync"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramen/e2e/util"
	"github.com/ramendr/ramen/e2e/validate"

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

	// Tests to run or clean.
	Tests []*Test

	// Command report, stored at the output directory on completion.
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

	testCmd := &Command{
		Command:         cmd,
		NamespacePrefix: "test-",
		PVCSpecs:        e2econfig.PVCSpecsMap(cmd.Config),
		Report:          newReport(name),
	}

	for _, tc := range cmd.Config.Tests {
		test := newTest(tc, testCmd)
		testCmd.Tests = append(testCmd.Tests, test)
	}

	return testCmd, nil
}

func (c *Command) Validate() bool {
	step := &Step{Name: ValidateStep}
	defer c.Report.AddStep(step)

	console.Step("Validate config")
	if err := validate.TestConfig(c.Env, c.Config, c.Logger); err != nil {
		c.fail("failed to validate config", err)
		step.Status = Failed
		return false
	}

	console.Pass("Config validated")
	step.Status = Passed
	return true
}

func (c *Command) Setup() bool {
	step := &Step{Name: SetupStep}
	defer c.Report.AddStep(step)

	console.Step("Setup environment")
	if err := util.EnsureChannel(c.Env.Hub, c.Config, c.Logger); err != nil {
		c.fail("failed to setup environment", err)
		step.Status = Failed
		return false
	}

	console.Pass("Environment setup")
	step.Status = Passed
	return true
}

func (c *Command) Cleanup() bool {
	step := &Step{Name: CleanupStep}
	defer c.Report.AddStep(step)

	console.Step("Clean environment")
	if err := util.EnsureChannelDeleted(c.Env.Hub, c.Config, c.Logger); err != nil {
		c.fail("failed to clean environment", err)
		step.Status = Failed
		return false
	}

	console.Pass("Environment cleaned")
	step.Status = Passed
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
		console.Error("failed to write report: %s", err)
	}
	return fmt.Errorf("failed (%d passed, %d failed, %d skipped)",
		c.Report.Summary.Passed, c.Report.Summary.Failed, c.Report.Summary.Skipped)
}

func (c *Command) Passed() {
	if err := c.WriteReport(c.Report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	console.Completed("passed (%d passed, %d failed, %d skipped)",
		c.Report.Summary.Passed, c.Report.Summary.Failed, c.Report.Summary.Skipped)
}

func (c *Command) fail(msg string, err error) {
	console.Error(msg)
	c.Logger.Error("%s: %s", msg, err)
}

func (c *Command) runFlowFunc(f flowFunc) bool {
	var wg sync.WaitGroup
	for _, test := range c.Tests {
		wg.Add(1)
		go func() {
			f(test)
			wg.Done()
		}()
	}
	wg.Wait()

	for _, test := range c.Tests {
		c.Report.AddTest(test)
	}

	return c.Report.Status == Passed
}

func (c *Command) runFlow(test *Test) {
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
	if !test.Unprotect() {
		return
	}
	test.Undeploy()
}
