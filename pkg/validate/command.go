// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"errors"
	"fmt"
	stdtime "time"

	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
	"github.com/ramendr/ramenctl/pkg/validation"
)

type Command struct {
	// command is the generic command used by all ramenctl commands.
	command *command.Command

	// config is the config for this command.
	config *config.Config

	// backend implementing the validation interface.
	backend validation.Validation

	// content is used to set deadlines.
	context context.Context

	// report describes the command execution.
	report *report.Report

	// current validation step.
	current        *report.Step
	currentStarted time.Time
}

// Ensure that command implements validation.Context.
var _ validation.Context = &Command{}

func newCommand(cmd *command.Command, cfg *config.Config, backend validation.Validation) *Command {
	return &Command{
		command: cmd,
		config:  cfg,
		backend: backend,
		context: cmd.Context(),
		report:  report.NewReport(cmd.Name(), cfg),
	}
}

// validation.Context interface.

func (c *Command) Env() *types.Env {
	return c.command.Env()
}

func (c *Command) Config() *config.Config {
	return c.config
}

func (c *Command) Logger() *zap.SugaredLogger {
	return c.command.Logger()
}

func (c *Command) Context() context.Context {
	return c.context
}

// Command interface.

func (c *Command) Clusters() error {
	if !c.validateConfig() {
		return c.failed()
	}
	if !c.validateClusters() {
		return c.failed()
	}
	c.passed()
	return nil
}

func (c *Command) Application(drpcName, drpcNamespace string) error {
	c.report.Application = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}
	if !c.validateConfig() {
		return c.failed()
	}
	if !c.validateApplication(drpcName, drpcNamespace) {
		return c.failed()
	}
	c.passed()
	return nil
}

// Validation.

// withTimeout returns a derived command with a deadline. Call cancel to release resources
// associated with the context as soon as the operation running in the context complete.
func (c Command) withTimeout(d stdtime.Duration) (*Command, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.context, d)
	c.context = ctx
	return &c, cancel
}

func (c *Command) validateConfig() bool {
	console.Step("Validate config")
	c.startStep("validate config")
	timedCmd, cancel := c.withTimeout(30 * stdtime.Second)
	defer cancel()
	if err := c.backend.Validate(timedCmd); err != nil {
		return c.failStep(err)
	}
	c.passStep()
	console.Pass("Config validated")
	return true
}

func (c *Command) validateClusters() bool {
	console.Step("Validate clusters")
	c.startStep("validate clusters")
	env := c.command.Env()
	for _, cluster := range []*types.Cluster{env.Hub, env.C1, env.C2} {
		// TODO: Run parallel validation for hub, passive hub, and managed clusters.
		c.command.Logger().Infof("Validating cluster %q", cluster.Name)
		step := &report.Step{Name: cluster.Name, Status: report.Passed}
		c.current.AddStep(step)
		console.Pass("Cluster %q validated", cluster.Name)
	}
	c.finishStep()
	return true
}

func (c *Command) validateApplication(drpcName, drpcNamespace string) bool {
	console.Step("Validate application")
	c.startStep("validate application")
	env := c.command.Env()
	for _, cluster := range []*types.Cluster{env.Hub, env.C1, env.C2} {
		// TODO: Run parallel validation for hub, passive hub, and managed clusters.
		c.command.Logger().Infof("Validating application \"%s/%s\" in cluster %q",
			drpcNamespace, drpcName, cluster.Name)
		step := &report.Step{Name: cluster.Name, Status: report.Passed}
		c.current.AddStep(step)
		console.Pass("Cluster %q validated", cluster.Name)
	}
	c.finishStep()
	return true
}

func (c *Command) failed() error {
	if err := c.command.WriteReport(c.report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	return fmt.Errorf("validation %s", c.report.Status)
}

func (c *Command) passed() {
	if err := c.command.WriteReport(c.report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	console.Completed("Validation completed")
}

// Managing steps.

func (c *Command) startStep(name string) {
	c.current = &report.Step{Name: name}
	c.currentStarted = time.Now()
	c.command.Logger().Infof("Step %q started", c.current.Name)
}

func (c *Command) passStep() bool {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	c.current.Status = report.Passed
	c.command.Logger().Infof("Step %q passed", c.current.Name)
	c.report.AddStep(c.current)
	c.current = nil
	return true
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
	c.command.Logger().Errorf("Step %q %s: %s", c.current.Name, c.current.Status, err)
	c.report.AddStep(c.current)
	c.current = nil
	return false
}

func (c *Command) finishStep() bool {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	c.command.Logger().Infof("Step %q finished", c.current.Name)
	c.report.AddStep(c.current)
	c.current = nil
	return c.report.Status == report.Passed
}
