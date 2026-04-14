// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	stdtime "time"

	"github.com/nirs/kubectl-gather/pkg/gather"
	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	basecmd "github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/logging"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/s3"
	"github.com/ramendr/ramenctl/pkg/time"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
	"github.com/ramendr/ramenctl/pkg/validation"
)

// HTMLWriter can write an HTML report.
type HTMLWriter interface {
	WriteHTML(io.Writer) error
}

type Command struct {
	// Backend implementing the validation interface.
	Backend validation.Validation

	// Report describes the command execution.
	Report *report.Report

	// Current validation step.
	Current *report.Step

	// S3Results stores S3 operation results for validation.
	S3Results []s3.Result

	cmd            *basecmd.Command
	config         *config.Config
	ctx            context.Context
	currentStarted time.Time
}

// Ensure that Command implements validation.Context.
var _ validation.Context = &Command{}

func New(
	cmd *basecmd.Command,
	cfg *config.Config,
	backend validation.Validation,
	r *report.Report,
) *Command {
	return &Command{
		cmd:     cmd,
		config:  cfg,
		Backend: backend,
		ctx:     cmd.Context(),
		Report:  r,
	}
}

// validation.Context interface.

func (c *Command) Env() *types.Env {
	return c.cmd.Env()
}

func (c *Command) Config() *config.Config {
	return c.config
}

func (c *Command) Logger() *zap.SugaredLogger {
	return c.cmd.Logger()
}

func (c *Command) Context() context.Context {
	return c.ctx
}

func (c *Command) LogFile() string {
	return c.cmd.LogFile()
}

// Validation.

// WithTimeout returns a derived command with a deadline. Call cancel to release resources
// associated with the context as soon as the operation running in the context complete.
func (c Command) WithTimeout(d stdtime.Duration) (*Command, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.ctx, d)
	c.ctx = ctx
	return &c, cancel
}

func (c *Command) ValidateConfig() bool {
	console.Step("Validate config")
	c.StartStep("validate config")
	timedCmd, cancel := c.WithTimeout(30 * stdtime.Second)
	defer cancel()
	if err := c.Backend.Validate(timedCmd); err != nil {
		return c.FailStep(err)
	}
	c.PassStep()
	console.Pass("Config validated")
	return true
}

func (c *Command) ValidatedDeleted(obj client.Object) report.ValidatedBool {
	validated := report.ValidatedBool{}
	if obj == nil {
		validated.Value = true
		validated.State = report.Problem
		validated.Description = "Resource does not exist"
	} else {
		if IsDeleted(obj) {
			validated.Value = true
			validated.State = report.Problem
			validated.Description = "Resource was deleted"
		} else {
			validated.State = report.OK
		}
	}
	summary.AddValidation(c.Report.Summary, &validated)
	return validated
}

func (c *Command) ValidatedConditions(
	obj client.Object,
	conditions []metav1.Condition,
) []report.ValidatedCondition {
	var validatedConditions []report.ValidatedCondition
	for i := range conditions {
		condition := &conditions[i]
		validated := ValidatedCondition(obj, condition, metav1.ConditionTrue)
		summary.AddValidation(c.Report.Summary, &validated)
		validatedConditions = append(validatedConditions, validated)
	}
	return validatedConditions
}

// Gathering data.

func (c *Command) GatherNamespaces(options gathering.Options) bool {
	start := time.Now()
	env := c.Env()
	clusters := []*types.Cluster{env.Hub, env.C1, env.C2}

	c.Logger().Infof("Gathering from clusters %q with options %+v",
		logging.ClusterNames(clusters), options)

	for r := range c.Backend.Gather(c, clusters, options) {
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
		c.Current.AddStep(step)
	}

	c.Logger().Infof("Gathered clusters in %.2f seconds", time.Since(start).Seconds())

	return c.Current.Status == report.Passed
}

func (c *Command) OutputReader(cluster string) gathering.OutputReader {
	clusterDir := filepath.Join(c.DataDir(), cluster)
	return gather.NewOutputReader(clusterDir)
}

func (c *Command) DataDir() string {
	return filepath.Join(c.cmd.OutputDir(), c.cmd.Name()+".data")
}

// WriteReport writes YAML, HTML, and CSS reports to the command output directory.
func (c *Command) WriteReport(r HTMLWriter) {
	c.cmd.WriteYAMLReport(r)

	file, err := c.cmd.OpenReport("html")
	if err != nil {
		console.Error("failed to open HTML report: %s", err)
		return
	}
	defer file.Close()
	if err := r.WriteHTML(file); err != nil {
		console.Error("failed to write HTML report: %s", err)
	}
	if err := file.Close(); err != nil {
		console.Error("failed to close HTML report: %s", err)
	}

	if err := report.WriteCSS(c.cmd.OutputDir()); err != nil {
		console.Error("failed to write report CSS: %s", err)
	}
}

// Managing steps.

func (c *Command) StartStep(name string) {
	c.Current = &report.Step{Name: name}
	c.currentStarted = time.Now()
	c.Logger().Infof("Step %q started", c.Current.Name)
}

func (c *Command) PassStep() bool {
	c.Current.Duration = time.Since(c.currentStarted).Seconds()
	c.Current.Status = report.Passed
	c.Logger().Infof("Step %q passed", c.Current.Name)
	c.Report.AddStep(c.Current)
	c.Current = nil
	return true
}

func (c *Command) FailStep(err error) bool {
	c.Current.Duration = time.Since(c.currentStarted).Seconds()
	if errors.Is(err, context.Canceled) {
		c.Current.Status = report.Canceled
		console.Error("Canceled %s", c.Current.Name)
	} else {
		c.Current.Status = report.Failed
		console.Error("Failed to %s", c.Current.Name)
	}
	c.Logger().Errorf("Step %q %s: %s", c.Current.Name, c.Current.Status, err)
	c.Report.AddStep(c.Current)
	c.Current = nil
	return false
}

func (c *Command) FinishStep() bool {
	c.Current.Duration = time.Since(c.currentStarted).Seconds()
	c.Logger().Infof("Step %q finished", c.Current.Name)
	c.Report.AddStep(c.Current)
	c.Current = nil
	return c.Report.Status == report.Passed
}
