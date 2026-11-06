// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package recipe

import (
	"errors"
	"fmt"
	"os"

	recipev1alpha1 "github.com/ramendr/recipe/api/v1alpha1"
	"sigs.k8s.io/yaml"

	basecmd "github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
)

// CommandName is the name of the validate-recipe command.
const CommandName = "validate-recipe"

type Command struct {
	cmd            *basecmd.Command
	recipeFile     string
	Report         *Report
	current        *report.Step
	currentStarted time.Time
}

func NewCommand(cmd *basecmd.Command, recipeFile string) *Command {
	r := NewReport()
	return &Command{
		cmd:        cmd,
		recipeFile: recipeFile,
		Report:     r,
	}
}

func (c *Command) WriteReport(r any) {
	c.cmd.WriteYAMLReport(r)
}

func (c *Command) passed() {
	c.WriteReport(c.Report)
	console.Completed("Validation completed (%s)", summary.String(c.Report.Summary))
}

func (c *Command) failed() error {
	c.WriteReport(c.Report)
	return errors.New(c.Report.Error())
}

func (c *Command) startStep(name string) {
	c.current = &report.Step{Name: name}
	c.currentStarted = time.Now()
	c.cmd.Logger().Infof("Step %q started", c.current.Name)
}

func (c *Command) passStep() bool {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	c.current.Status = report.Passed
	c.cmd.Logger().Infof("Step %q passed", c.current.Name)
	c.Report.AddStep(c.current)
	c.current = nil
	return true
}

func (c *Command) failStep(err error) bool {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	c.current.Status = report.Failed
	c.current.Err = err.Error()
	console.Error(err.Error())
	c.cmd.Logger().Errorf("Step %q %s: %s", c.current.Name, c.current.Status, err)
	c.Report.AddStep(c.current)
	c.current = nil
	return false
}

func (c *Command) Run() error {
	console.Step("Validate recipe")
	c.startStep("validate recipe")

	data, err := os.ReadFile(c.recipeFile)
	if err != nil {
		c.failStep(fmt.Errorf("failed to read %q: %w", c.recipeFile, err))
		return c.failed()
	}

	r := &recipev1alpha1.Recipe{}
	if err := yaml.Unmarshal(data, r); err != nil {
		c.failStep(fmt.Errorf("failed to parse %q: %w", c.recipeFile, err))
		return c.failed()
	}

	console.Pass("Recipe parsed")

	status := Validate(r)
	c.Report.RecipeReport = *status

	for i := range status.Groups {
		summary.AddValidation(c.Report.Summary, &status.Groups[i])
	}
	for i := range status.Hooks {
		summary.AddValidation(c.Report.Summary, &status.Hooks[i])
	}
	for i := range status.Workflows {
		summary.AddValidation(c.Report.Summary, &status.Workflows[i])
	}

	for _, w := range status.Workflows {
		if w.State == report.Problem {
			c.failStep(fmt.Errorf("recipe %q has problems", c.recipeFile))
			return c.failed()
		}
	}

	console.Pass("Recipe validated")
	c.passStep()
	c.passed()
	return nil
}
