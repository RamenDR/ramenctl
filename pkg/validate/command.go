// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	stdtime "time"

	"github.com/nirs/kubectl-gather/pkg/gather"
	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/logging"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/s3"
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

	// s3Results stores S3 operation results for validation.
	s3Results []s3.Result
}

// Ensure that command implements validation.Context.
var _ validation.Context = &Command{}

func newCommand(cmd *command.Command, cfg *config.Config, backend validation.Validation) *Command {
	r := report.NewReport(cmd.Name(), cfg)
	r.Summary = &report.Summary{}
	return &Command{
		command: cmd,
		config:  cfg,
		backend: backend,
		context: cmd.Context(),
		report:  r,
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

func (c *Command) validatedDeleted(obj client.Object) report.ValidatedBool {
	validated := report.ValidatedBool{}
	if obj == nil {
		// Resource is missing.
		validated.Value = true
		validated.State = report.Problem
		validated.Description = "Resource does not exist"
	} else {
		if isDeleted(obj) {
			// Resource was deleted.
			validated.Value = true
			validated.State = report.Problem
			validated.Description = "Resource was deleted"
		} else {
			// Resource not deleted.
			validated.State = report.OK
		}
	}
	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedConditions(
	obj client.Object,
	conditions []metav1.Condition,
) []report.ValidatedCondition {
	var validatedConditions []report.ValidatedCondition
	for i := range conditions {
		condition := &conditions[i]
		validated := validatedCondition(obj, condition, metav1.ConditionTrue)
		addValidation(c.report.Summary, &validated)
		validatedConditions = append(validatedConditions, validated)
	}
	return validatedConditions
}

// Gathering data.

func (c *Command) gatherNamespaces(options gathering.Options) bool {
	start := time.Now()
	env := c.Env()
	clusters := []*types.Cluster{env.Hub, env.C1, env.C2}

	c.Logger().Infof("Gathering from clusters %q with options %+v",
		logging.ClusterNames(clusters), options)

	for r := range c.backend.Gather(c, clusters, options) {
		step := &report.Step{Name: fmt.Sprintf("gather %q", r.Name), Duration: r.Duration}
		if r.Err != nil {
			msg := fmt.Sprintf("Failed to gather data from cluster %q", r.Name)
			console.Error(msg)
			c.Logger().Errorf("%s: %s", msg, r.Err)
			step.Status = report.Failed
		} else {
			console.Pass("Gathered data from cluster %q", r.Name)
			step.Status = report.Passed
		}
		c.current.AddStep(step)
	}

	c.Logger().Infof("Gathered clusters in %.2f seconds", time.Since(start).Seconds())

	return c.current.Status == report.Passed
}

func (c *Command) outputReader(cluster string) gathering.OutputReader {
	clusterDir := filepath.Join(c.dataDir(), cluster)
	return gather.NewOutputReader(clusterDir)
}

func (c *Command) dataDir() string {
	return filepath.Join(c.command.OutputDir(), c.command.Name()+".data")
}

// Completing commands.

func (c *Command) failed() error {
	c.command.WriteYAMLReport(c.report)
	return fmt.Errorf("validation %s (%s)", c.report.Status, summaryString(c.report.Summary))
}

func (c *Command) passed() {
	c.command.WriteYAMLReport(c.report)
	console.Completed("Validation completed (%s)", summaryString(c.report.Summary))
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
