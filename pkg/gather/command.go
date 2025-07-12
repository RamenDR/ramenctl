// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package gather

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
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

func newCommand(cmd *command.Command, cfg *config.Config, backend validation.Validation) *Command {
	return &Command{
		command: cmd,
		config:  cfg,
		backend: backend,
		context: cmd.Context(),
		report:  report.NewReport(cmd.Name(), cfg),
	}
}

func (c *Command) Application(drpcName string, drpcNamespace string) error {
	if !c.validateConfig() {
		return c.failed()
	}
	if !c.gatherApplication(drpcName, drpcNamespace) {
		return c.failed()
	}
	c.passed()
	return nil
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

func (c *Command) gatherApplication(drpcName string, drpcNamespace string) bool {
	console.Step("Gather Application data")
	c.report.Application = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}
	c.startStep("gather application")

	env := c.Env()
	clusters := []*types.Cluster{env.Hub, env.C1, env.C2}
	namespaces, err := c.namespacesToGather(drpcName, drpcNamespace)
	if err != nil {
		return false
	}
	outputDir := filepath.Join(c.command.OutputDir(), c.command.Name()+".data")
	start := time.Now()

	c.Logger().Infof("Gathering namespaces %q from clusters %q", namespaces, clusterNames(clusters))
	for r := range Namespaces(clusters, namespaces, outputDir, c.Logger()) {
		if r.Err != nil {
			msg := fmt.Sprintf("Failed to gather data from cluster %q", r.Name)
			console.Error(msg)
			c.Logger().Errorf("%s: %s", msg, r.Err)
			step := &report.Step{Name: r.Name, Status: report.Failed, Duration: r.Duration}
			c.current.AddStep(step)
		} else {
			step := &report.Step{Name: r.Name, Status: report.Passed, Duration: r.Duration}
			c.current.AddStep(step)
			console.Pass("Gathered data from cluster %q", r.Name)
		}
	}

	c.Logger().Infof("Gathered clusters in %.2f seconds", time.Since(start).Seconds())

	return c.finishStep()
}

// withTimeout returns a derived command with a deadline. Call cancel to release resources
// associated with the context as soon as the operation running in the context complete.
func (c Command) withTimeout(d stdtime.Duration) (*Command, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.context, d)
	c.context = ctx
	return &c, cancel
}

func (c *Command) failed() error {
	if err := c.command.WriteReport(c.report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	return fmt.Errorf("Gather %s", c.report.Status)
}

func (c *Command) passed() {
	if err := c.command.WriteReport(c.report); err != nil {
		console.Error("failed to write report: %s", err)
	}
	console.Completed("Gather completed")
}

func (c *Command) namespacesToGather(drpcName string, drpcNamespace string) ([]string, error) {
	seen := map[string]struct{}{
		// Gather ramen namespaces to get ramen hub and dr-cluster logs and related resources.
		c.config.Namespaces.RamenHubNamespace:       {},
		c.config.Namespaces.RamenDRClusterNamespace: {},
		drpcNamespace: {},
	}

	appNamespaces, err := c.backend.ApplicationNamespaces(c, drpcName, drpcNamespace)
	if err != nil {
		return nil, err
	}
	var namespaces []string

	namespaces = append(namespaces, appNamespaces...)

	for ns := range seen {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	return namespaces, nil
}

func clusterNames(clusters []*types.Cluster) []string {
	names := []string{}
	for _, cluster := range clusters {
		names = append(names, cluster.Name)
	}
	return names
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
