// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package gather

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	stdtime "time"

	"github.com/nirs/kubectl-gather/pkg/gather"
	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/logging"
	"github.com/ramendr/ramenctl/pkg/ramen"
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

func (c *Command) outputReader(cluster string) gathering.OutputReader {
	clusterDir := filepath.Join(c.dataDir(), cluster)
	return gather.NewOutputReader(clusterDir)
}

func (c *Command) dataDir() string {
	return filepath.Join(c.command.OutputDir(), c.command.Name()+".data")
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
	c.report.Application = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}
	if !c.validateConfig() {
		return c.failed()
	}
	if !c.gatherData(drpcName, drpcNamespace) {
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

func (c *Command) gatherData(drpcName string, drpcNamespace string) bool {
	console.Step("Gather application data")
	c.startStep("gather data")

	namespaces, ok := c.inspectApplication(drpcName, drpcNamespace)
	if !ok {
		return c.finishStep()
	}
	c.report.Namespaces = namespaces

	options := gathering.Options{
		Namespaces: namespaces,
		OutputDir:  c.dataDir(),
	}
	if !c.gatherApplication(options) {
		return c.finishStep()
	}

	profiles, prefix, ok := c.inspectS3Profiles(drpcName, drpcNamespace)
	if !ok {
		return c.finishStep()
	}

	if !c.gatherApplicationS3Data(profiles, prefix) {
		return c.finishStep()
	}

	c.finishStep()
	return true
}

func (c *Command) inspectApplication(drpcName, drpcNamespace string) ([]string, bool) {
	start := time.Now()
	step := &report.Step{Name: "inspect application"}
	c.Logger().Infof("Step %q started", step.Name)

	namespaces, err := c.namespacesToGather(drpcName, drpcNamespace)
	if err != nil {
		step.Duration = time.Since(start).Seconds()
		if errors.Is(err, context.Canceled) {
			console.Error("Canceled %s", step.Name)
			step.Status = report.Canceled
		} else {
			console.Error("Failed to %s", step.Name)
			step.Status = report.Failed
		}
		c.Logger().Errorf("Step %q %s: %s", c.current.Name, step.Status, err)
		c.current.AddStep(step)

		return nil, false
	}

	step.Duration = time.Since(start).Seconds()
	step.Status = report.Passed
	c.current.AddStep(step)

	console.Pass("Inspected application")
	c.Logger().Infof("Step %q passed", step.Name)

	return namespaces, true
}

func (c *Command) gatherApplication(options gathering.Options) bool {
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
			c.current.AddStep(step)
		} else {
			step.Status = report.Passed
			c.current.AddStep(step)
			console.Pass("Gathered data from cluster %q", r.Name)
		}
	}

	c.Logger().Infof("Gathered clusters in %.2f seconds", time.Since(start).Seconds())

	return c.current.Status == report.Passed
}

func (c *Command) inspectS3Profiles(
	drpcName, drpcNamespace string,
) ([]*s3.Profile, string, bool) {
	start := time.Now()
	step := &report.Step{Name: "inspect S3 profiles"}

	c.Logger().Infof("Step %q started", step.Name)

	profiles, prefix, err := c.applicationS3Info(drpcName, drpcNamespace)
	if err != nil {
		step.Duration = time.Since(start).Seconds()
		step.Status = report.Failed
		console.Error("Failed to %s", step.Name)
		c.Logger().Errorf("Step %q %s: %s", c.current.Name, step.Status, err)
		c.current.AddStep(step)
		return nil, "", false
	}

	step.Duration = time.Since(start).Seconds()
	step.Status = report.Passed
	c.current.AddStep(step)

	console.Pass("Inspected S3 profiles")
	c.Logger().Infof("Step %q passed", step.Name)

	return profiles, prefix, true
}

func (c *Command) gatherApplicationS3Data(profiles []*s3.Profile, prefix string) bool {
	start := time.Now()
	outputDir := c.dataDir()

	c.Logger().Infof("Gathering application S3 data from profiles %q with prefix %q",
		logging.ProfileNames(profiles), prefix)

	for r := range c.backend.GatherS3(c, profiles, prefix, outputDir) {
		step := &report.Step{
			Name:     fmt.Sprintf("gather S3 profile %q", r.ProfileName),
			Duration: r.Duration,
		}
		if r.Err != nil {
			if errors.Is(r.Err, context.Canceled) {
				msg := fmt.Sprintf("Canceled gather S3 profile %q", r.ProfileName)
				console.Error(msg)
				c.Logger().Errorf("%s: %s", msg, r.Err)
				step.Status = report.Canceled
			} else {
				msg := fmt.Sprintf("Failed to gather S3 profile %q", r.ProfileName)
				console.Error(msg)
				c.Logger().Errorf("%s: %s", msg, r.Err)
				step.Status = report.Failed
			}
		} else {
			step.Status = report.Passed
			console.Pass("Gathered S3 profile %q", r.ProfileName)
		}
		c.current.AddStep(step)
	}

	c.Logger().Infof("Gathered application S3 data in %.2f seconds", time.Since(start).Seconds())

	return c.current.Status == report.Passed
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
	set := map[string]struct{}{
		// Gather ramen namespaces to get ramen hub and dr-cluster logs and related resources.
		c.config.Namespaces.RamenHubNamespace:       {},
		c.config.Namespaces.RamenDRClusterNamespace: {},
	}

	appNamespaces, err := c.backend.ApplicationNamespaces(c, drpcName, drpcNamespace)
	if err != nil {
		return nil, err
	}

	for _, ns := range appNamespaces {
		set[ns] = struct{}{}
	}

	return slices.Sorted(maps.Keys(set)), nil
}

// applicationS3Info reads S3 profiles and application prefix from gathered hub data.
func (c *Command) applicationS3Info(
	drpcName, drpcNamespace string,
) ([]*s3.Profile, string, error) {
	// Read S3 profiles from the ramen hub configmap, the source of truth
	// synced to managed clusters.
	reader := c.outputReader(c.Env().Hub.Name)

	configMapName := ramen.HubOperatorConfigMapName
	configMapNamespace := c.config.Namespaces.RamenHubNamespace

	profiles, err := ramen.ClusterProfiles(reader, configMapName, configMapNamespace)
	if err != nil {
		return nil, "", err
	}

	prefix, err := ramen.ApplicationS3Prefix(reader, drpcName, drpcNamespace)
	if err != nil {
		return nil, "", err
	}

	return profiles, prefix, nil
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
