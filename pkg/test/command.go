// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/e2e"
	"github.com/ramendr/ramenctl/pkg/gather"
	"github.com/ramendr/ramenctl/pkg/time"
)

// namespacePrefix is used for all namespaces created by tests.
const namespacePrefix = "test-"

type Options struct {
	// Gather data for failed tests.
	GatherData bool
}

// Command is a ramenctl test command.
type Command struct {
	*command.Command
	Backend e2e.Testing
	Options Options

	// PCCSpecs maps pvscpec name to pvcspec.
	PVCSpecs map[string]types.PVCSpecConfig

	// Tests to run or clean.
	Tests []*Test

	// Current test step
	Current *Step

	// Command report, stored at the output directory on completion.
	Report *Report

	// stepStarted records the time when the step execution began, used for duration tracking.
	stepStarted time.Time
}

// flowFunc runs a test flow on with a test. The test logs progress messages and marked as failed if
// the flow failed.
type flowFunc func(t *Test)

// newCommand return a new test command.
func newCommand(cmd *command.Command, backend e2e.Testing, options Options) *Command {
	// This is not user configurable. We use the same prefix for all namespaces created by the test.
	cmd.Config().Channel.Namespace = namespacePrefix + "gitops"

	testCmd := &Command{
		Command:  cmd,
		Backend:  backend,
		Options:  options,
		PVCSpecs: e2econfig.PVCSpecsMap(cmd.Config()),
		Report:   newReport(cmd.Name(), cmd.Config()),
	}

	for _, tc := range cmd.Config().Tests {
		test := newTest(tc, testCmd)
		testCmd.Tests = append(testCmd.Tests, test)
	}

	return testCmd
}

// Run a test flow and return an error if one or more tests failed. When completed you need to call
// Clean() to remove resources created during the run.
func (c *Command) Run() error {
	if !c.validate() {
		return c.failed()
	}
	if !c.setup() {
		return c.failed()
	}
	if !c.runTests() {
		return c.failed()
	}
	c.passed()
	return nil
}

// Clean up after running a test flow and return an if cleaning one or more tests failed.
func (c *Command) Clean() error {
	if !c.validate() {
		return c.failed()
	}
	if !c.cleanTests() {
		return c.failed()
	}
	if !c.cleanup() {
		return c.failed()
	}
	c.passed()
	return nil
}

func (c *Command) validate() bool {
	c.startStep(ValidateStep)
	console.Step("Validate config")
	if err := c.Backend.Validate(c); err != nil {
		return c.failStep(err)
	}
	console.Pass("Config validated")
	return c.passStep()
}

func (c *Command) setup() bool {
	c.startStep(SetupStep)
	console.Step("Setup environment")
	if err := c.Backend.Setup(c); err != nil {
		return c.failStep(err)
	}
	console.Pass("Environment setup")
	return c.passStep()
}

func (c *Command) cleanup() bool {
	c.startStep(CleanupStep)
	console.Step("Clean environment")
	if err := c.Backend.Cleanup(c); err != nil {
		return c.failStep(err)
	}
	console.Pass("Environment cleaned")
	return c.passStep()
}

func (c *Command) runTests() bool {
	console.Step("Run tests")
	return c.runFlowFunc(c.runFlow)
}

func (c *Command) cleanTests() bool {
	console.Step("Clean tests")
	return c.runFlowFunc(c.cleanFlow)
}

func (c *Command) gatherData() {
	console.Step("Gather data")
	namespaces := c.namespacesToGather()
	outputDir := filepath.Join(c.OutputDir(), c.Name()+".gather")
	gather.Namespaces(c.Env(), namespaces, outputDir, c.Logger())
}

func (c *Command) failed() error {
	if err := c.WriteReport(c.Report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	return fmt.Errorf("%s (%s)", c.Report.Status, c.Report.Summary)
}

func (c *Command) passed() {
	if err := c.WriteReport(c.Report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	console.Completed("%s (%s)", c.Report.Status, c.Report.Summary)
}

func (c *Command) startStep(name string) {
	c.Current = &Step{Name: name}
	c.stepStarted = time.Now()
	c.Logger().Infof("Step %q started", c.Current.Name)
}

func (c *Command) failStep(err error) bool {
	c.Current.Duration = time.Since(c.stepStarted).Seconds()
	if errors.Is(err, context.Canceled) {
		c.Current.Status = Canceled
		console.Error("Canceled %s", c.Current.Name)
	} else {
		c.Current.Status = Failed
		console.Error("Failed to %s", c.Current.Name)
	}
	c.Logger().Errorf("Step %q %s: %s", c.Current.Name, c.Current.Status, err)
	c.Report.AddStep(c.Current)
	c.Current = nil
	return false
}

func (c *Command) passStep() bool {
	c.Current.Duration = time.Since(c.stepStarted).Seconds()
	c.Current.Status = Passed
	c.Logger().Infof("Step %q passed", c.Current.Name)
	c.Report.AddStep(c.Current)
	c.Current = nil
	return true
}

func (c *Command) finishStep() bool {
	c.Current.Duration = time.Since(c.stepStarted).Seconds()
	c.Logger().Infof("Step %q finished", c.Current.Name)
	c.Report.AddStep(c.Current)
	c.Current = nil
	return c.Report.Status == Passed
}

func (c *Command) runFlowFunc(f flowFunc) bool {
	c.startStep(TestsStep)

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
		c.Current.AddTest(test)
	}

	res := c.finishStep()
	if c.Report.Status == Failed && c.Options.GatherData {
		c.gatherData()
	}
	return res
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
	test.Cleanup()
}

func (c *Command) namespacesToGather() []string {
	cfg := c.Config()
	seen := map[string]struct{}{
		// Gather ramen namespaces to get ramen hub and dr-cluster logs and related resources.
		cfg.Namespaces.RamenHubNamespace:       {},
		cfg.Namespaces.RamenDRClusterNamespace: {},
	}

	// Add application resources for failed tests.
	for _, test := range c.Tests {
		if test.Status == Failed {
			seen[test.AppNamespace()] = struct{}{}
			seen[test.ManagementNamespace()] = struct{}{}
		}
	}

	var namespaces []string
	for ns := range seen {
		namespaces = append(namespaces, ns)
	}

	sort.Strings(namespaces)

	return namespaces
}
