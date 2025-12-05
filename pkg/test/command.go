// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"sync"
	stdtime "time"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/logging"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/testing"
	"github.com/ramendr/ramenctl/pkg/time"
)

// Command is a ramenctl test command.
type Command struct {
	// Command is the generic command used by all ramenctl commands.
	command *command.Command

	// config is the test config for this command.
	config *e2econfig.Config

	// content is used to set deadlines.
	context context.Context

	// backend is used to perform testing operations.
	backend testing.Testing

	// PCCSpecs maps pvscpec name to pvcspec.
	pvcSpecs map[string]e2econfig.PVCSpec

	// Deployers maps deployer name to deployer
	deployers map[string]e2econfig.Deployer

	// tests to run or clean.
	tests []*Test

	// Command report, stored at the output directory on completion.
	report *Report

	// current test step
	current        *report.Step
	currentStarted time.Time
}

// Ensure that command implements types.Context.
var _ types.Context = &Command{}

// flowFunc runs a test flow on with a test. The test logs progress messages and marked as failed if
// the flow failed.
type flowFunc func(t *Test)

// newCommand return a new test command.
func newCommand(
	cmd *command.Command,
	cfg *e2econfig.Config,
	backend testing.Testing,
) *Command {
	testCmd := &Command{
		command:   cmd,
		config:    cfg,
		context:   cmd.Context(),
		backend:   backend,
		pvcSpecs:  e2econfig.PVCSpecsMap(cfg),
		deployers: e2econfig.DeployersMap(cfg),
		report:    newReport(cmd.Name(), cfg),
	}

	for _, tc := range cfg.Tests {
		test := newTest(tc, testCmd)
		testCmd.tests = append(testCmd.tests, test)
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

// ramen/e2e/types.Context interface

func (c *Command) Logger() *zap.SugaredLogger {
	return c.command.Logger()
}

func (c *Command) Env() *types.Env {
	return c.command.Env()
}

func (c *Command) Context() context.Context {
	return c.context
}

func (c *Command) Config() *e2econfig.Config {
	return c.config
}

// withTimeout returns a derived command with a deadline. Call cancel to release resources
// associated with the context as soon as the operation running in the context complete.
func (c Command) withTimeout(d stdtime.Duration) (*Command, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.context, d)
	c.context = ctx
	return &c, cancel
}

func (c *Command) dataDir() string {
	return filepath.Join(c.command.OutputDir(), c.command.Name()+".data")
}

func (c *Command) validate() bool {
	c.startStep(ValidateStep)
	console.Step("Validate config")
	timedCmd, cancel := c.withTimeout(30 * stdtime.Second)
	defer cancel()
	if err := c.backend.Validate(timedCmd); err != nil {
		return c.failStep(err)
	}
	console.Pass("Config validated")
	return c.passStep()
}

func (c *Command) setup() bool {
	c.startStep(SetupStep)
	console.Step("Setup environment")
	timedCmd, cancel := c.withTimeout(30 * stdtime.Second)
	defer cancel()
	if err := c.backend.Setup(timedCmd); err != nil {
		return c.failStep(err)
	}
	console.Pass("Environment setup")
	return c.passStep()
}

func (c *Command) cleanup() bool {
	c.startStep(CleanupStep)
	console.Step("Clean environment")
	timedCmd, cancel := c.withTimeout(1 * stdtime.Minute)
	defer cancel()
	if err := c.backend.Cleanup(timedCmd); err != nil {
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
	env := c.Env()
	clusters := []*types.Cluster{env.Hub, env.C1, env.C2}
	options := gathering.Options{
		Namespaces: c.namespacesToGather(),
		OutputDir:  c.dataDir(),
	}
	start := time.Now()

	c.Logger().Infof("Gathering from clusters %q with options %+v",
		logging.ClusterNames(clusters), options)

	for r := range c.backend.Gather(c, clusters, options) {
		if r.Err != nil {
			msg := fmt.Sprintf("Failed to gather data from cluster %q", r.Name)
			console.Error(msg)
			c.Logger().Errorf("%s: %s", msg, r.Err)
		} else {
			console.Pass("Gathered data from cluster %q", r.Name)
		}
	}

	c.Logger().Infof("Gathered clusters in %.2f seconds", time.Since(start).Seconds())
}

func (c *Command) failed() error {
	if err := c.command.WriteReport(c.report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	return fmt.Errorf("%s (%s)", c.report.Status, c.report.Summary)
}

func (c *Command) passed() {
	if err := c.command.WriteReport(c.report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	console.Completed("%s (%s)", c.report.Status, c.report.Summary)
}

func (c *Command) startStep(name string) {
	c.current = &report.Step{Name: name}
	c.currentStarted = time.Now()
	c.Logger().Infof("Step %q started", c.current.Name)
}

func (c *Command) failStep(err error) bool {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	if errors.Is(err, context.Canceled) {
		c.current.Status = report.Canceled
		console.Error("Canceled %s", c.current.Name)
	} else {
		c.current.Status = report.Failed
		console.Error("Failed to %s", c.current.Name)
	}
	c.Logger().Errorf("Step %q %s: %s", c.current.Name, c.current.Status, err)
	c.report.AddStep(c.current)
	c.current = nil
	return false
}

func (c *Command) passStep() bool {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	c.current.Status = report.Passed
	c.Logger().Infof("Step %q passed", c.current.Name)
	c.report.AddStep(c.current)
	c.current = nil
	return true
}

func (c *Command) finishStep() bool {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	c.Logger().Infof("Step %q finished", c.current.Name)
	c.report.AddStep(c.current)
	c.current = nil
	return c.report.Status == report.Passed
}

func (c *Command) runFlowFunc(f flowFunc) bool {
	c.startStep(TestsStep)

	var wg sync.WaitGroup
	for _, test := range c.tests {
		wg.Add(1)
		go func() {
			f(test)
			wg.Done()
		}()
	}
	wg.Wait()

	for _, test := range c.tests {
		step := &report.Step{
			Name:     test.Name(),
			Status:   test.Status,
			Duration: test.Duration,
			Items:    test.Steps,
		}
		c.current.AddStep(step)
	}

	res := c.finishStep()
	if c.report.Status == report.Failed {
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
	test.Purge()
}

func (c *Command) namespacesToGather() []string {
	set := map[string]struct{}{
		// Gather ramen namespaces to get ramen hub and dr-cluster logs and related resources.
		c.config.Namespaces.RamenHubNamespace:       {},
		c.config.Namespaces.RamenDRClusterNamespace: {},
	}

	// Add application resources for failed tests.
	for _, test := range c.tests {
		if test.Status == report.Failed {
			set[test.AppNamespace()] = struct{}{}
			set[test.ManagementNamespace()] = struct{}{}
		}
	}

	return slices.Sorted(maps.Keys(set))
}
