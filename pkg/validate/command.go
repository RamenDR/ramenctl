// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"fmt"

	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
)

type Command struct {
	// command is the generic command used by all ramenctl commands.
	command *command.Command

	// config is the config for this command.
	config *config.Config

	// content is used to set deadlines.
	context context.Context

	// report describes the command execution.
	report *Report

	// current validation step.
	current        *report.Step
	currentStarted time.Time
}

func newCommand(cmd *command.Command, cfg *config.Config) *Command {
	return &Command{
		command: cmd,
		config:  cfg,
		context: cmd.Context(),
		report:  newReport(cmd.Name(), cfg),
	}
}

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

func (c *Command) validateConfig() bool {
	console.Step("Validate config")
	c.startStep("validate config")
	// TODO: Detect distro
	// TODO: Validate managed clusters in clusterset
	c.passStep()
	console.Pass("config validated")
	return true
}

func (c *Command) validateClusters() bool {
	console.Step("Validate clusters")
	c.startStep("validate clusters")
	env := c.command.Env()
	for _, cluster := range []*types.Cluster{env.Hub, env.C1, env.C2} {
		// TODO: Run parallel validation for hub, passive hub, and managed clusters.
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

func (c *Command) finishStep() bool {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	c.command.Logger().Infof("Step %q finished", c.current.Name)
	c.report.AddStep(c.current)
	c.current = nil
	return c.report.Status == report.Passed
}
